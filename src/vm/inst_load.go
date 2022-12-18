package vm

import (
	"lua/src/api"
)

// 给连续几个寄存器设置nil值
func loadNil(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	// 先压入一个nil
	vm.PushNil()
	for j := a; j <= a+b; j++ {
		// 全部复制为栈顶的nil
		vm.Copy(-1, j)
	}
	// 弹出栈顶的nil
	vm.Pop(1)
}

// 给单个寄存器设置布尔值
func loadBool(i Instruction, vm api.LuaVM) {
	// a指定寄存器索引，b指定布尔值，c指定是否跳过下一条指令
	a, b, c := i.ABC()
	a += 1
	vm.PushBoolean(b != 0)
	vm.Replace(a)
	if c != 0 {
		vm.AddPC(1)
	}
}

// 将常量表的某个常量加载到寄存器
func loadK(i Instruction, vm api.LuaVM) {
	a, bx := i.ABx()
	a += 1
	vm.GetConst(bx)
	vm.Replace(a)
}

func loadKx(i Instruction, vm api.LuaVM) {
	a, _ := i.ABx()
	a += 1
	ax := Instruction(vm.Fetch()).Ax()
	vm.GetConst(ax)
	vm.Replace(a)
}
