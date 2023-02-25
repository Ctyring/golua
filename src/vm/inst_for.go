package vm

import (
	. "lua/src/api"
)

func forPrep(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1
	// R(A) -= R(A+2) 预先减去步长
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(LUA_OPSUB)
	vm.Replace(a)
	// pc += sBx
	vm.AddPC(sBx)
}

func forLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1
	vm.PushValue(a + 2) // R(A+2)
	vm.PushValue(a)     // R(A)
	vm.Arith(LUA_OPADD) // R(A) += R(A+2)
	vm.Replace(a)       // R(A) = R(A) + R(A+2)
	// 如果R(A) <= R(A+1) 则跳转
	isPositiveStep := vm.ToNumber(a+2) >= 0
	if isPositiveStep && vm.Compare(a, a+1, LUA_OPLE) ||
		!isPositiveStep && vm.Compare(a+1, a, LUA_OPLE) {
		vm.AddPC(sBx)
		vm.Copy(a, a+3)
	}
}

func tForLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	if !vm.IsNil(a + 1) {
		vm.Copy(a+1, a)
		vm.AddPC(sBx)
	}
}
