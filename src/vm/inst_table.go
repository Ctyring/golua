package vm

import (
	"lua/src/api"
)

const LFIELDS_PER_FLUSH = 50

// 创建空表并放入指定寄存器
func newTable(i Instruction, vm api.LuaVM) {
	// b和c使用浮点字节的编码模式
	a, b, c := i.ABC()
	a += 1
	vm.CreateTable(Fb2int(b), Fb2int(c))
	vm.Replace(a)
}

// int --> "floating point byte"
func Int2fb(x int) int {
	e := 0 // exponent
	if x < 8 {
		return x
	}
	for x >= (8 << 4) { // 8*16=128
		x = (x + 0xf) >> 4 // x = ceil(x/16)
		e += 4
	}
	for x >= (8 << 1) { // 8*2=16
		x = (x + 1) >> 1 // x = ceil(x/2)
		e += 1
	}
	return ((e + 1) << 3) | (x - 8)
}

// "floating point byte" --> int
func Fb2int(x int) int {
	if x < 8 {
		return x
	}
	return ((x & 7) + 8) << uint((x>>3)-1)
}

// 根据键从表里取值
func getTable(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

// R(A)[RK(B)] := RK(C)
// A指定表的寄存器，BC指定键和值的寄存器或常量表
func setTable(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(a)
}

// 按索引批量设置表的值
// R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B
// A指定表的寄存器，B指定值的数量，C指定数组起始索引的批次数，FPF指定每批次的数量。用批次数乘以每批次的数量，再加上1，就是数组的起始索引
// 如果还不够，会在SETLIST指令后面添加一个EXTRAARG指令，用其Ax操作数来保存批次数
func setList(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	bIsZero := b == 0
	if bIsZero { // 如果b为0，表示要收集所有的值
		b = int(vm.ToInteger(-1)) - a - 1
		vm.Pop(1)
	}
	if c > 0 {
		c = c - 1
	} else {
		c = Instruction(vm.Fetch()).Ax()
	}

	vm.CheckStack(1)
	idx := int64(c * LFIELDS_PER_FLUSH)
	for j := 1; j <= b; j++ {
		idx++
		vm.PushValue(a + j)
		vm.SetI(a, idx)
	}

	// 处理完寄存器的值后，来处理栈内的值
	if bIsZero {
		for j := vm.RegisterCount() + 1; j <= vm.GetTop(); j++ {
			idx++
			vm.PushValue(j)
			vm.SetI(a, idx)
		}
		vm.SetTop(vm.RegisterCount())
	}
}
