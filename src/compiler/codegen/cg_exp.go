package codegen

import (
	. "lua/src/compiler/ast"
	. "lua/src/compiler/lexer"
	. "lua/src/vm"
)

func cgExp(fi *funcInfo, node Exp, a, n int) {
	switch exp := node.(type) {
	case *NilExp:
		fi.emitLoadNil(a, n)
	case *FalseExp:
		fi.emitLoadBool(a, 0, 0)
	case *TrueExp:
		fi.emitLoadBool(a, 1, 0)
	case *IntegerExp:
		fi.emitLoadK(a, exp.Val)
	case *FloatExp:
		fi.emitLoadK(a, exp.Val)
	case *StringExp:
		fi.emitLoadK(a, exp.Str)
	case *ParensExp:
		cgExp(fi, exp.Exp, a, 1)
	case *VarargExp:
		cgVarargExp(fi, exp, a, n)
	case *FuncDefExp:
		cgFuncDefExp(fi, exp, a)
	case *TableConstructorExp:
		cgTableConstructorExp(fi, exp, a)
	case *UnopExp:
		cgUnopExp(fi, exp, a)
	case *BinopExp:
		cgBinopExp(fi, exp, a)
	case *ConcatExp:
		cgConcatExp(fi, exp, a)
	case *NameExp:
		cgNameExp(fi, exp, a)
	case *TableAccessExp:
		cgTableAccessExp(fi, exp, a)
	case *FuncCallExp:
		cgFuncCallExp(fi, exp, a, n)
	}
}

// 生成vararg表达式
func cgVarargExp(fi *funcInfo, exp *VarargExp, a, n int) {
	if !fi.isVararg {
		panic("cannot use '...' outside a vararg function")
	}
	fi.emitVararg(a, n)
}

// 生成函数定义表达式
func cgFuncDefExp(fi *funcInfo, node *FuncDefExp, a int) {
	// 与外围函数形成父子关系
	subFI := newFuncInfo(fi, node) // 这里实现了进入作用域的功能
	fi.subFuncs = append(fi.subFuncs, subFI)

	// 处理函数表达式
	for _, param := range node.ParList { // 参数列表
		subFI.addLocVar(param)
	}
	cgBlock(subFI, node.Block) // 函数体
	subFI.exitScope()          // 退出作用域
	subFI.emitReturn(0, 0)     // 返回

	bx := len(fi.subFuncs) - 1
	fi.emitClosure(a, bx)
}

// 生成表构造表达式
func cgTableConstructorExp(fi *funcInfo, node *TableConstructorExp, a int) {
	// 计算数组的长度
	nArr := 0
	for _, keyExp := range node.KeyExps {
		if keyExp == nil { // 没有key，是数组元素
			nArr++
		}
	}
	// 计算哈希表的长度
	nExps := len(node.KeyExps)
	multRet := nExps > 0 && isVarargOrFuncCall(node.ValExps[nExps-1])

	fi.emitNewTable(a, nArr, nExps-nArr) // 创建表指令

	// 遍历处理每一个键值对
	arrIdx := 0
	for i, keyExp := range node.KeyExps {
		valExp := node.ValExps[i]

		// 处理数组
		if keyExp == nil {
			arrIdx++
			tmp := fi.allocReg()
			if i == nExps-1 && multRet {
				cgExp(fi, valExp, tmp, -1)
			} else {
				cgExp(fi, valExp, tmp, 1)
			}

			// 攒在一起处理
			if arrIdx%50 == 0 || arrIdx == nArr { // LFIELDS_PER_FLUSH
				n := arrIdx % 50
				if n == 0 {
					n = 50
				}
				fi.freeRegs(n)
				c := (arrIdx-1)/50 + 1 // todo: c > 0xFF
				if i == nExps-1 && multRet {
					fi.emitSetList(a, 0, c)
				} else {
					fi.emitSetList(a, n, c)
				}
			}

			continue
		}

		// 处理哈希表
		b := fi.allocReg()
		cgExp(fi, keyExp, b, 1)
		c := fi.allocReg()
		cgExp(fi, valExp, c, 1)
		fi.freeRegs(2)

		fi.emitSetTable(a, b, c)
	}
}

// 生成一元表达式
func cgUnopExp(fi *funcInfo, node *UnopExp, a int) {
	b := fi.allocReg()            // 申请一个寄存器
	cgExp(fi, node.Exp, b, 1)     // 处理右侧表达式
	fi.emitUnaryOp(node.Op, a, b) // 生成一元操作指令
	fi.freeReg()                  // 释放寄存器
}

// 生成拼接表达式
func cgConcatExp(fi *funcInfo, node *ConcatExp, a int) {
	for _, subExp := range node.Exps {
		a := fi.allocReg()
		cgExp(fi, subExp, a, 1)
	}

	c := fi.usedRegs - 1
	b := c - len(node.Exps) + 1
	fi.freeRegs(c - b + 1)
	fi.emitABC(OP_CONCAT, a, b, c)
}

// 生成二元表达式
func cgBinopExp(fi *funcInfo, node *BinopExp, a int) {
	switch node.Op {
	case TOKEN_OP_AND, TOKEN_OP_OR:
		b := fi.allocReg()
		cgExp(fi, node.Exp1, b, 1) // 处理左侧表达式
		fi.freeReg()
		if node.Op == TOKEN_OP_AND { // 判断是否需要短路
			fi.emitTestSet(a, b, 0)
		} else {
			fi.emitTestSet(a, b, 1)
		}
		pcOfJmp := fi.emitJmp(0, 0) // 短路

		b = fi.allocReg()
		cgExp(fi, node.Exp2, b, 1) // 处理右侧表达式
		fi.freeReg()
		fi.emitMove(a, b)                   // 将右侧表达式的结果赋值给a
		fi.fixSbx(pcOfJmp, fi.pc()-pcOfJmp) // 修复跳转指令的偏移量
	default:
		b := fi.allocReg()
		cgExp(fi, node.Exp1, b, 1) // 处理左侧表达式
		c := fi.allocReg()
		cgExp(fi, node.Exp2, c, 1) // 处理右侧表达式
		fi.emitBinaryOp(node.Op, a, b, c)
		fi.freeRegs(2)
	}
}

// 名字表达式
func cgNameExp(fi *funcInfo, node *NameExp, a int) {
	if r := fi.slotOfLocVar(node.Name); r >= 0 { // 局部变量
		fi.emitMove(a, r)
	} else if idx := fi.indexOfUpval(node.Name); idx >= 0 { // upvalue
		fi.emitGetUpval(a, idx)
	} else { // 全局变量
		taExp := &TableAccessExp{
			PrefixExp: &NameExp{Line: node.Line, Name: "_ENV"},
			KeyExp:    &StringExp{Line: node.Line, Str: node.Name},
		}
		cgTableAccessExp(fi, taExp, a)
	}
}

// 表访问表达式
func cgTableAccessExp(fi *funcInfo, node *TableAccessExp, a int) {
	b := fi.allocReg()
	cgExp(fi, node.PrefixExp, b, 1) // 处理前缀表达式
	c := fi.allocReg()
	cgExp(fi, node.KeyExp, c, 1) // 处理键表达式
	fi.emitGetTable(a, b, c)
	fi.freeRegs(2)
}

// 函数调用表达式
func cgFuncCallExp(fi *funcInfo, node *FuncCallExp, a, n int) {
	nArgs := prepFuncCall(fi, node, a) // 准备函数调用
	fi.emitCall(a, nArgs, n)
}

// 尾调用
func cgTailCallExp(fi *funcInfo, node *FuncCallExp, a int) {
	nArgs := prepFuncCall(fi, node, a)
	fi.emitTailCall(a, nArgs)
}

func prepFuncCall(fi *funcInfo, node *FuncCallExp, a int) int {
	nArgs := len(node.Args)
	lastArgIsVarargOrFuncCall := false

	cgExp(fi, node.PrefixExp, a, 1) // 处理前缀表达式
	if node.NameExp != nil {        // 处理语法糖
		c := 0x100 + fi.indexOfConstant(node.NameExp.Name)
		fi.emitSelf(a, a, c)
	}
	for i, arg := range node.Args { // 处理参数
		tmp := fi.allocReg()
		if i == nArgs-1 && isVarargOrFuncCall(arg) {
			lastArgIsVarargOrFuncCall = true
			cgExp(fi, arg, tmp, -1)
		} else {
			cgExp(fi, arg, tmp, 1)
		}
	}
	fi.freeRegs(nArgs)

	if node.NameExp != nil { // 如果是语法糖(self)参数，需要多传递一个self参数
		nArgs++
	}
	if lastArgIsVarargOrFuncCall {
		nArgs = -1
	}

	return nArgs
}
