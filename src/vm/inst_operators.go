package vm

import (
	. "lua/src/api"
)

func _binaryArith(i Instruction, vm LuaVM, op ArithOp) {
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

func _unaryArith(i Instruction, vm LuaVM, op ArithOp) {
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

func add(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPADD) }
func sub(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPSUB) }
func mul(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPMUL) }
func mod(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPMOD) }
func pow(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPPOW) }
func div(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPDIV) }
func idiv(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPIDIV) }
func band(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPBAND) }
func bor(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPBOR) }
func bxor(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPBXOR) }
func shl(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPSHL) }
func shr(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPSHR) }
func unm(i Instruction, vm LuaVM)  { _unaryArith(i, vm, LUA_OPUNM) }
func bnot(i Instruction, vm LuaVM) { _unaryArith(i, vm, LUA_OPBNOT) }

func length(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.Len(b)
	vm.Replace(a)
}

func concat(i Instruction, vm LuaVM) {
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

func _compare(i Instruction, vm LuaVM, op CompareOp) {
	a, b, c := i.ABC()
	vm.GetRK(b)
	vm.GetRK(c)
	if vm.Compare(-2, -1, op) != (a != 0) { // 如果比较结果和a匹配则跳过下一条指令
		vm.AddPC(1)
	}
	vm.Pop(2)
}

func eq(i Instruction, vm LuaVM) { _compare(i, vm, LUA_OPEQ) }
func lt(i Instruction, vm LuaVM) { _compare(i, vm, LUA_OPLT) }
func le(i Instruction, vm LuaVM) { _compare(i, vm, LUA_OPLE) }

func not(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.PushBoolean(!vm.ToBoolean(b))
	vm.Replace(a)
}

func testSet(i Instruction, vm LuaVM) {
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
func test(i Instruction, vm LuaVM) {
	a, _, c := i.ABC()
	a += 1

	// 如果a与c一致则跳过下一条指令
	if vm.ToBoolean(a) != (c != 0) {
		vm.AddPC(1)
	}
}
