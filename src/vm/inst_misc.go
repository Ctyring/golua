package vm

import (
	"lua/src/api"
)

// iABC模式，把源寄存器(B指定)的值拷贝到目标寄存器(A指定)。
// 通过move函数得知，Lua代码的局部变量就保存在寄存器里。
func move(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.Copy(b, a)
}

// jmp指令负责无条件跳转和闭合处于开启状态的Upvalue
// pc += sBx; if (A) close all upvalues >= R(A - 1)
func jmp(i Instruction, vm api.LuaVM) {
	a, sBx := i.AsBx()
	vm.AddPC(sBx)
	if a != 0 {
		vm.CloseUpvalues(a)
	}
}
