package vm

import (
	"lua/src/api"
)

// 指令解码
type Instruction uint32

const MAXARG_Bx = 1<<18 - 1       // 262143
const MAXARG_sBx = MAXARG_Bx >> 1 // 262143 / 2 = 131071
// 定义从指令中提取操作码的方法
func (self Instruction) Opcode() int {
	return int(self & 0x3F)
}

// 从iABC模式指令中提取参数
func (self Instruction) ABC() (a, b, c int) {
	return int(self >> 6 & 0xFF), int(self >> 23 & 0x1FF), int(self >> 14 & 0x1FF)
}

// 从iABx模式指令中提取参数
func (self Instruction) ABx() (a, bx int) {
	return int(self >> 6 & 0xFF), int(self >> 14)
}

// 从iAsBx模式指令中提取参数
// Lua虚拟机采取将有符号整数编码成比特序列的方式为Excess-K，也叫偏移二进制码的编码模式
// 如果将sBx解释成无符号整数时它的值是x，那么解释成有符号整数时的值就是x-MAXARG_sBx
// MAXARG_sBx取sBx所能表示的最大无符号整数值的一半
func (self Instruction) AsBx() (a, sbx int) {
	a, bx := self.ABx()
	return a, bx - MAXARG_sBx
}

// 从iAx模式指令中提取参数
func (self Instruction) Ax() (ax int) {
	return int(self >> 6)
}

// 返回指令的操作码名字
func (self Instruction) OpName() string {
	return opcodes[self.Opcode()].name
}

// 返回指令的编码模式
func (self Instruction) OpMode() byte {
	return opcodes[self.Opcode()].opMode
}

// 返回操作数B的使用方式
func (self Instruction) BMode() byte {
	return opcodes[self.Opcode()].argBMode
}

// 返回操作数C的使用方式
func (self Instruction) CMode() byte {
	return opcodes[self.Opcode()].argCMode
}

func (self Instruction) Execute(vm api.LuaVM) {
	action := opcodes[self.Opcode()].action
	println("action:", opcodes[self.Opcode()].name)
	if action != nil {
		action(self, vm)
	} else {
		panic(self.OpName())
	}
}
