package vm

import (
	"lua/src/api"
)

// R(A) := closure(KPROTO[Bx])
// 把当前Lua函数的子函数原型实例化为闭包，放入寄存器A中，子函数原型来自于当前函数原型的子函数原型表，索引为Bx
func closure(i Instruction, vm api.LuaVM) {
	a, bx := i.ABx()
	a += 1
	// 加载子函数原型表
	vm.LoadProto(bx)
	vm.Replace(a)
}

// 调用Lua函数
// R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))
func call(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1

	nArgs := _pushFuncAndArgs(a, b, vm) // 把参数和函数压入栈顶
	vm.Call(nArgs, c-1)                 // 调用函数
	_popResults(a, c, vm)               // 弹出返回值
}

// 把参数和函数压入栈顶
func _pushFuncAndArgs(a, b int, vm api.LuaVM) int {
	if b >= 1 {
		// b-1个参数
		vm.CheckStack(b)
		for i := a; i < a+b; i++ {
			vm.PushValue(i)
		}
		return b - 1 // 除去函数外的参数个数
	} else {
		// 这种情况是要求把所有参数压入栈顶
		_fixStack(a, vm)                            // 把参数压入栈顶
		return vm.GetTop() - vm.RegisterCount() - 1 // 除去函数外的参数个数
	}
}

func _popResults(a, c int, vm api.LuaVM) {
	if c == 1 {
		// 无返回值
	} else if c > 1 {
		// 有返回值
		for i := a + c - 2; i >= a; i-- {
			vm.Replace(i)
		}
	} else {
		// 对应返回所有参数，就把参数留在栈中，因为后面还有需求
		// 推入一个整数，标记这些返回值原本是需要移动到哪些寄存器中的
		// 这种情况的一个情形: f(1, 2, g(3, 4)) 此时g的返回值需要留在栈中
		vm.CheckStack(1)
		vm.PushInteger(int64(a))
		//vm.Replace(a)
	}
}

func _fixStack(a int, vm api.LuaVM) {
	x := int(vm.ToInteger(-1)) // 取出最后的整数
	vm.Pop(1)                  // 弹出标记

	vm.CheckStack(x - a)
	// 把函数和前半部分参数压入栈顶
	for i := a; i < x; i++ {
		vm.PushValue(i)
	}
	vm.Rotate(vm.RegisterCount()+1, x-a) // 旋转栈顶的x-a个元素，使得参数在栈顶
}

func _return(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	if b == 1 {
		// 无返回值
	} else if b > 1 {
		vm.CheckStack(b - 1)
		for i := a; i <= a+b-2; i++ {
			vm.PushValue(i)
		}
	} else {
		// 如果有部分返回值已经在栈中，只需要返回一部分
		_fixStack(a, vm)
	}
}

// 把传递给当前函数的变长参数加载到连续多个寄存器中
func vararg(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	if b != 1 {
		// 把变长参数压入栈顶
		vm.LoadVararg(b - 1)
		// 把变长参数从栈顶移动到连续多个寄存器中
		_popResults(a, b, vm)
	}
}

// 尾调用优化
func tailCall(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	c := 0
	nArgs := _pushFuncAndArgs(a, b, vm) // 把参数和函数压入栈顶
	vm.Call(nArgs, c-1)                 // 调用函数
	_popResults(a, c, vm)               // 弹出返回值
}

// SELF指令 用来优化语法糖，把对象和方法拷贝到连续的两个寄存器中，这样在调用方法时就不需要再次拷贝了(节约一条指令)
// R(A+1) := R(B); R(A) := R(B)[RK(C)]
func self(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.Copy(b, a+1) // R(A+1) := R(B)
	vm.GetRK(c)     // R(A) := R(B)[RK(C)]
	vm.GetTable(b)
	vm.Replace(a)
}

func tForCall(i Instruction, vm api.LuaVM) {
	a, _, c := i.ABC()
	a += 1

	_pushFuncAndArgs(a, 3, vm) // 把函数和参数压入栈顶
	vm.Call(2, c)              // 调用函数
	_popResults(a+3, c+1, vm)  // 弹出返回值
}
