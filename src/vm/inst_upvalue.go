package vm

import (
	. "lua/src/api"
)

// 把当前闭包的某个Upvalue值保存到目标寄存器
// R(A) := UpValue[B]
func getUpval(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Copy(LuaUpvalueIndex(b), a)
}

// 使用寄存器中的值给当前闭包赋值
// UpValue[B] := R(A)
func setUpval(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Copy(a, LuaUpvalueIndex(b))
}

// 如果当前闭包的某个Upvalue是表，则GETTABUP可以根据键从该表取值，相当于更高效的GETUPVAL+GETTABLE
// R(A) := UpValue[B][RK(C)]
func getTabUp(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	// 把键推入栈
	vm.GetRK(c)
	// 从Upvalue中取出表，并根据栈中的c取出对应的值压入栈
	vm.GetTable(LuaUpvalueIndex(b))
	// 弹出并保存到a
	vm.Replace(a)
}

// 如果当前闭包的某个Upvalue是表，则SETTABUP可以根据键给该表赋值，相当于更高效的GETUPVAL+SETTABLE
// UpValue[A][RK(B)] := RK(C)
func setTabUp(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1

	// 把键和值推入栈
	vm.GetRK(b)
	vm.GetRK(c)
	// 从Upvalue中取出表，并根据栈中的b和c给表赋值
	vm.SetTable(LuaUpvalueIndex(a))
}
