package codegen

import (
	. "lua/src/compiler/ast"
)

func cgBlock(fi *funcInfo, node *Block) {
	for _, stat := range node.Stats { // 遍历语句序列
		cgStat(fi, stat) // 生成语句
	}

	if node.RetExps != nil { // 如果有返回值
		cgRetStat(fi, node.RetExps) // 生成返回指令
	}
}

// 处理并生成返回指令
func cgRetStat(fi *funcInfo, exps []Exp) {
	nExps := len(exps)
	if nExps == 0 { // 如果没有返回值
		fi.emitReturn(0, 0) // 生成返回指令
		return
	}

	if nExps == 1 { // 如果只有一个返回值
		if nameExp, ok := exps[0].(*NameExp); ok { // 如果是变量
			if r := fi.slotOfLocVar(nameExp.Name); r >= 0 {
				fi.emitReturn(r, 1)
				return
			}
		}
		if fcExp, ok := exps[0].(*FuncCallExp); ok { // 如果是函数调用
			r := fi.allocReg()
			cgTailCallExp(fi, fcExp, r) // 生成尾调用指令
			fi.freeReg()
			fi.emitReturn(r, -1)
			return
		}
	}

	multRet := isVarargOrFuncCall(exps[nExps-1]) // 判断最后一个表达式是否是可变参数或者函数调用
	for i, exp := range exps {
		r := fi.allocReg()           // 为表达式分配一个寄存器
		if i == nExps-1 && multRet { // 如果是最后一个表达式且是可变参数或者函数调用
			cgExp(fi, exp, r, -1) // 生成表达式指令
		} else {
			cgExp(fi, exp, r, 1) // 生成表达式指令
		}
	}
	fi.freeRegs(nExps) // 释放寄存器
	a := fi.usedRegs
	if multRet {
		fi.emitReturn(a, -1) // 生成返回指令
	} else {
		fi.emitReturn(a, nExps) // 生成返回指令
	}
}
