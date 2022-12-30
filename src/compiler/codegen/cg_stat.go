package codegen

import (
	. "lua/src/compiler/ast"
)

func cgStat(fi *funcInfo, node Stat) {
	switch stat := node.(type) {
	case *FuncCallStat:
		cgFuncCallStat(fi, stat)
	case *BreakStat:
		cgBreakStat(fi, stat)
	case *DoStat:
		cgDoStat(fi, stat)
	case *WhileStat:
		cgWhileStat(fi, stat)
	case *RepeatStat:
		cgRepeatStat(fi, stat)
	case *IfStat:
		cgIfStat(fi, stat)
	case *ForNumStat:
		cgForNumStat(fi, stat)
	case *ForInStat:
		cgForInStat(fi, stat)
	case *AssignStat:
		cgAssignStat(fi, stat)
	case *LocalVarDeclStat:
		cgLocalVarDeclStat(fi, stat)
	case *LocalFuncDefStat:
		cgLocalFuncDefStat(fi, stat)
	case *LabelStat, *GotoStat:
		panic("label and goto statements are not supported!")
	}
}

// 生成局部函数定义语句
func cgLocalFuncDefStat(fi *funcInfo, node *LocalFuncDefStat) {
	r := fi.addLocVar(node.Name)  // 为函数名分配一个寄存器
	cgFuncDefExp(fi, node.Exp, r) // 生成函数定义指令
}

// 生成函数调用语句
func cgFuncCallStat(fi *funcInfo, node *FuncCallStat) {
	r := fi.allocReg()            // 为函数调用分配一个寄存器
	cgFuncCallExp(fi, node, r, 0) // 生成函数调用指令
	fi.freeReg()                  // 释放寄存器
}

// 生成break语句
func cgBreakStat(fi *funcInfo, node *BreakStat) {
	pc := fi.emitJmp(0, 0) // 生成跳转指令(等到确定跳转位置时再填充跳转偏移)
	fi.addBreakJmp(pc)     // 将跳转指令的pc加入break列表
}

// 生成do语句
func cgDoStat(fi *funcInfo, node *DoStat) {
	fi.enterScope(false) // 非循环块
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals() // 关闭未关闭的upvalue
	fi.exitScope()       // 退出块
}

// 生成while语句
func cgWhileStat(fi *funcInfo, node *WhileStat) {
	pcBeforeExp := fi.pc()                    // 记录下while语句的起始位置
	r := fi.allocReg()                        // 为while表达式分配一个寄存器
	cgExp(fi, node.Exp, r, 1)                 // 生成while表达式
	fi.freeReg()                              // 释放寄存器
	fi.emitTest(r, 0)                         // 生成测试指令
	pcJmpToEnd := fi.emitJmp(0, 0)            // 生成跳转指令(等到确定跳转位置时再填充跳转偏移)
	fi.enterScope(true)                       // 进入循环块
	cgBlock(fi, node.Block)                   // 生成块
	fi.closeOpenUpvals()                      // 关闭未关闭的upvalue
	fi.emitJmp(0, pcBeforeExp-fi.pc()-1)      // 生成跳转指令(跳转到while语句的起始位置)
	fi.exitScope()                            // 退出块
	fi.fixSbx(pcJmpToEnd, fi.pc()-pcJmpToEnd) // 填充跳转指令的跳转偏移
}

// 生成repeat语句
func cgRepeatStat(fi *funcInfo, node *RepeatStat) {
	fi.enterScope(true)                    // 进入循环块
	pcBeforeBlock := fi.pc()               // 记录下repeat语句的起始位置
	cgBlock(fi, node.Block)                // 生成块
	r := fi.allocReg()                     // 为repeat表达式分配一个寄存器
	cgExp(fi, node.Exp, r, 1)              // 生成repeat表达式
	fi.freeReg()                           // 释放寄存器
	fi.emitTest(r, 1)                      // 生成测试指令
	fi.emitJmp(0, pcBeforeBlock-fi.pc()-1) // 生成跳转指令(跳转到repeat语句的起始位置)
	fi.closeOpenUpvals()                   // 关闭未关闭的upvalue
	fi.exitScope()                         // 退出块
}

// 生成if语句
func cgIfStat(fi *funcInfo, node *IfStat) {
	pcJmpToEnds := make([]int, len(node.Exps)) // 用于记录每个分支的跳转指令的pc
	pcJmpToNextExp := -1                       // 用于记录跳转到下一个分支的跳转指令的pc

	for i, exp := range node.Exps { // 生成每个分支的测试指令和跳转指令
		if pcJmpToNextExp >= 0 { // 如果有跳转到下一个分支的跳转指令
			fi.fixSbx(pcJmpToNextExp, fi.pc()-pcJmpToNextExp) // 填充跳转指令的跳转偏移
		}

		r := fi.allocReg() // 为if表达式分配一个寄存器
		cgExp(fi, exp, r, 1)
		fi.freeReg() // 释放寄存器

		fi.emitTest(r, 0)                 // 生成测试指令
		pcJmpToNextExp = fi.emitJmp(0, 0) // 生成跳转指令(等到确定跳转位置时再填充跳转偏移)
		fi.enterScope(false)              // 进入非循环块
		cgBlock(fi, node.Blocks[i])
		fi.closeOpenUpvals()      // 关闭未关闭的upvalue
		fi.exitScope()            // 退出块
		if i < len(node.Exps)-1 { // 如果不是最后一个分支
			pcJmpToEnds[i] = fi.emitJmp(0, 0) // 生成跳转指令(等到确定跳转位置时再填充跳转偏移)
		} else {
			pcJmpToEnds[i] = pcJmpToNextExp // 最后一个分支的跳转指令的pc就是跳转到下一个分支的跳转指令的pc
		}
	}

	for _, pc := range pcJmpToEnds { // 填充每个分支的跳转指令的跳转偏移
		fi.fixSbx(pc, fi.pc()-pc)
	}
}

// 生成数值for语句
func cgForNumStat(fi *funcInfo, node *ForNumStat) {
	fi.enterScope(true)                       // 进入循环块
	cgLocalVarDeclStat(fi, &LocalVarDeclStat{ // 生成局部变量声明语句。三个特殊的局部变量分别是循环变量、循环变量的初始值、循环变量的终止值
		NameList: []string{"(for index)", "for limit", "for step"},
		ExpList:  []Exp{node.InitExp, node.LimitExp, node.StepExp},
	})
	fi.addLocVar(node.VarName) // 添加循环变量
	a := fi.usedRegs - 4
	pcForPrep := fi.emitForPrep(a, 0)           // 生成for prep指令(等到确定跳转位置时再填充跳转偏移)
	cgBlock(fi, node.Block)                     // 生成块
	fi.closeOpenUpvals()                        // 关闭未关闭的upvalue
	pcForLoop := fi.emitForLoop(a, 0)           // 生成for loop指令(等到确定跳转位置时再填充跳转偏移)
	fi.fixSbx(pcForPrep, pcForLoop-pcForPrep-1) // 填充for prep指令的跳转偏移
	fi.fixSbx(pcForLoop, pcForPrep-pcForLoop)   // 填充for loop指令的跳转偏移
	fi.exitScope()                              // 退出块
}

// 生成泛型for语句
func cgForInStat(fi *funcInfo, node *ForInStat) {
	fi.enterScope(true) // 进入循环块
	cgLocalVarDeclStat(fi, &LocalVarDeclStat{
		NameList: []string{"(for generator)", "(for state)", "(for control)"},
		ExpList:  node.ExpList,
	})
	for _, name := range node.NameList {
		fi.addLocVar(name)
	}
	pcJmpToTFC := fi.emitJmp(0, 0)            // 生成跳转指令(等到确定跳转位置时再填充跳转偏移)
	cgBlock(fi, node.Block)                   // 生成块
	fi.fixSbx(pcJmpToTFC, fi.pc()-pcJmpToTFC) // 填充跳转指令的跳转偏移
	rGenerator := fi.slotOfLocVar("(for generator)")
	fi.emitTForCall(rGenerator, len(node.NameList))
	fi.emitTForLoop(rGenerator+2, pcJmpToTFC-fi.pc()-1)

	fi.exitScope()
}

// 生成局部变量声明语句
func cgLocalVarDeclStat(fi *funcInfo, node *LocalVarDeclStat) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nName := len(node.NameList)

	oldRegs := fi.usedRegs
	if nExps == nName { // 如果变量个数和表达式个数相等
		for _, exp := range node.ExpList {
			a := fi.allocReg()
			cgExp(fi, exp, a, 1)
		}
	} else if nExps > nName { // 如果表达式个数大于变量个数
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) { // 如果是最后一个表达式且是可变参数或函数调用
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else { // 如果表达式个数小于变量个数
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) { // 如果是最后一个表达式且是可变参数或函数调用
				multRet = true
				n := nName - nExps + 1
				cgExp(fi, exp, a, n) // 进行多重赋值
				fi.allocRegs(n - 1)  // 为多余的变量分配寄存器
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
		if !multRet { // 如果不存在可变参数或函数调用
			n := nName - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(a, n) // 生成load nil指令
		}
	}
	fi.usedRegs = oldRegs
	for _, name := range node.NameList {
		fi.addLocVar(name)
	}
}

// 赋值语句
func cgAssignStat(fi *funcInfo, node *AssignStat) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nVars := len(node.VarList)

	tRegs := make([]int, nVars)
	kRegs := make([]int, nVars)
	vRegs := make([]int, nVars)
	oldRegs := fi.usedRegs

	for i, exp := range node.VarList {
		if taExp, ok := exp.(*TableAccessExp); ok { // 如果是表访问表达式
			tRegs[i] = fi.allocReg()                // 为表分配寄存器
			cgExp(fi, taExp.PrefixExp, tRegs[i], 1) // 生成表达式
			kRegs[i] = fi.allocReg()                // 为键分配寄存器
			cgExp(fi, taExp.KeyExp, kRegs[i], 1)    // 生成键表达式
		} else { // 如果是变量
			name := exp.(*NameExp).Name
			if fi.slotOfLocVar(name) < 0 && fi.indexOfUpval(name) < 0 { // 如果变量不是局部变量也不是upvalue，说明是全局变量
				// global var
				kRegs[i] = -1
				if fi.indexOfConstant(name) > 0xFF {
					kRegs[i] = fi.allocReg()
				}
			}
		}
	}
	for i := 0; i < nVars; i++ {
		vRegs[i] = fi.usedRegs + i
	}

	if nExps >= nVars {
		for i, exp := range exps {
			a := fi.allocReg()
			if i >= nVars && i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else { // nVars > nExps
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1
				cgExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
		if !multRet {
			n := nVars - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(a, n)
		}
	}

	for i, exp := range node.VarList { // 遍历变量列表
		if nameExp, ok := exp.(*NameExp); ok { // 如果是变量名
			varName := nameExp.Name
			if a := fi.slotOfLocVar(varName); a >= 0 {
				fi.emitMove(a, vRegs[i]) // 生成move指令
			} else if b := fi.indexOfUpval(varName); b >= 0 {
				fi.emitSetUpval(vRegs[i], b) // 生成setupval指令
			} else if a := fi.slotOfLocVar("_ENV"); a >= 0 { // 检查是否存在一个局部变量名为_ENV
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfConstant(varName)
					fi.emitSetTable(a, b, vRegs[i]) // 生成settable指令
				} else {
					fi.emitSetTable(a, kRegs[i], vRegs[i]) // 生成settable指令
				}
			} else { // global var
				a := fi.indexOfUpval("_ENV") // 绑定_ENV到upvalue
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfConstant(varName)
					fi.emitSetTabUp(a, b, vRegs[i]) // 生成settabup指令
				} else {
					fi.emitSetTabUp(a, kRegs[i], vRegs[i]) // 生成settabup指令
				}
			}
		} else { // 如果是表的访问表达式
			fi.emitSetTable(tRegs[i], kRegs[i], vRegs[i]) // 生成settable指令
		}
	}

	// todo
	fi.usedRegs = oldRegs
}
