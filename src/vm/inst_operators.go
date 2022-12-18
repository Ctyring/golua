package vm

import (
	api2 "lua/src/api"
)

func _binaryArith(i Instruction, vm api2.LuaVM, op api2.ArithOp) {
	a, b, c := i.ABC()
	a += 1
	// 压入两个操作数
	vm.GetRK(b)
	vm.GetRK(c)
	// 运算
	vm.Arith(op)
	// 把结果放到指定寄存器
	vm.Replace(a)
}

func _unaryArith(i Instruction, vm api2.LuaVM, op api2.ArithOp) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	// 压入操作数
	vm.PushValue(b)
	// 运算
	vm.Arith(op)
	// 把结果放到指定寄存器
	vm.Replace(a)
}

func add(i Instruction, vm api2.LuaVM)  { _binaryArith(i, vm, api2.LUA_OPADD) }
func sub(i Instruction, vm api2.LuaVM)  { _binaryArith(i, vm, api2.LUA_OPSUB) }
func mul(i Instruction, vm api2.LuaVM)  { _binaryArith(i, vm, api2.LUA_OPMUL) }
func mod(i Instruction, vm api2.LuaVM)  { _binaryArith(i, vm, api2.LUA_OPMOD) }
func pow(i Instruction, vm api2.LuaVM)  { _binaryArith(i, vm, api2.LUA_OPPOW) }
func div(i Instruction, vm api2.LuaVM)  { _binaryArith(i, vm, api2.LUA_OPDIV) }
func idiv(i Instruction, vm api2.LuaVM) { _binaryArith(i, vm, api2.LUA_OPIDIV) }
func band(i Instruction, vm api2.LuaVM) { _binaryArith(i, vm, api2.LUA_OPBAND) }
func bor(i Instruction, vm api2.LuaVM)  { _binaryArith(i, vm, api2.LUA_OPBOR) }
func bxor(i Instruction, vm api2.LuaVM) { _binaryArith(i, vm, api2.LUA_OPBXOR) }
func shl(i Instruction, vm api2.LuaVM)  { _binaryArith(i, vm, api2.LUA_OPSHL) }
func shr(i Instruction, vm api2.LuaVM)  { _binaryArith(i, vm, api2.LUA_OPSHR) }
func unm(i Instruction, vm api2.LuaVM)  { _unaryArith(i, vm, api2.LUA_OPUNM) }
func bnot(i Instruction, vm api2.LuaVM) { _unaryArith(i, vm, api2.LUA_OPBNOT) }

func length(i Instruction, vm api2.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.Len(b)
	vm.Replace(a)
}

func concat(i Instruction, vm api2.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	c += 1
	n := c - b + 1
	vm.CheckStack(n)
	for j := b; j <= c; j++ {
		vm.PushValue(j)
	}
	vm.Concat(n)
	vm.Replace(a)
}

func _compare(i Instruction, vm api2.LuaVM, op api2.CompareOp) {
	a, b, c := i.ABC()
	vm.GetRK(b)
	vm.GetRK(c)
	if vm.Compare(-2, -1, op) != (a != 0) { // 如果比较结果和a匹配则跳过下一条指令
		vm.AddPC(1)
	}
	vm.Pop(2)
}

func eq(i Instruction, vm api2.LuaVM) { _compare(i, vm, api2.LUA_OPEQ) }
func lt(i Instruction, vm api2.LuaVM) { _compare(i, vm, api2.LUA_OPLT) }
func le(i Instruction, vm api2.LuaVM) { _compare(i, vm, api2.LUA_OPLE) }

func not(i Instruction, vm api2.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.PushBoolean(!vm.ToBoolean(b))
	vm.Replace(a)
}

func testSet(i Instruction, vm api2.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	// 如果b与c一致则跳过下一条指令
	if vm.ToBoolean(b) == (c != 0) {
		vm.Copy(b, a)
	} else {
		vm.AddPC(1)
	}
}

// test 是 testset 的一种特殊情况， 即 a = b
func test(i Instruction, vm api2.LuaVM) {
	a, _, c := i.ABC()
	a += 1

	// 如果a与c一致则跳过下一条指令
	if vm.ToBoolean(a) != (c != 0) {
		vm.AddPC(1)
	}
}
