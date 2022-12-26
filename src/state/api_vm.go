package state

func (self *luaState) PC() int {
	return self.stack.pc
}

// 修改指令指针
func (self *luaState) AddPC(n int) {
	self.stack.pc += n
}

// 取出当前指令，将PC指向下一条指令
func (self *luaState) Fetch() uint32 {
	i := self.stack.closure.proto.Code[self.stack.pc]
	self.stack.pc++
	return i
}

// 从常量表取出一个常量值，然后推入栈顶
func (self *luaState) GetConst(idx int) {
	c := self.stack.closure.proto.Constants[idx]
	self.stack.push(c)
}

func (self *luaState) GetRK(rk int) {
	if rk > 0xFF { // constant
		// 把常量推入栈顶
		self.GetConst(rk & 0xFF)
	} else { // register
		// 把某个索引处的栈值推入栈顶
		self.PushValue(rk + 1)
	}
}

// 获取寄存器数量
func (self *luaState) RegisterCount() int {
	return int(self.stack.closure.proto.MaxStackSize)
}

// 将可变参数推入栈顶
func (self *luaState) LoadVararg(n int) {
	if n < 0 {
		n = len(self.stack.varargs)
	}
	self.stack.check(n)
	self.stack.pushN(self.stack.varargs, n)
}

// 将指定子函数原型推入栈顶
func (self *luaState) LoadProto(idx int) {
	stack := self.stack
	subProto := stack.closure.proto.Protos[idx]
	closure := newLuaClosure(subProto)
	stack.push(closure)
	// 遍历子函数的upvalue表
	// 将子函数原型转换为闭包，并确保闭包中的Upvalue能够正确地引用外部变量。
	for i, uvInfo := range subProto.Upvalues {
		uvIdx := int(uvInfo.Idx)
		// 判断Upvalue是开放状态(捕获自当前函数)
		if uvInfo.Instack == 1 {
			// 初始化
			if stack.openuvs == nil {
				stack.openuvs = map[int]*upvalue{}
			}
			// 如果Upvalue已经被打开，将该Upvalue赋值给闭包的upvals数组
			if openuv, found := stack.openuvs[uvIdx]; found {
				closure.upvals[i] = openuv
			} else {
				// 将一个新的Upvalue赋值给闭包的upvals数组
				closure.upvals[i] = &upvalue{&stack.slots[uvIdx]}
				// 添加到栈的openuvs表中
				stack.openuvs[uvIdx] = closure.upvals[i]
			}
		} else { // 来自更外围函数(闭合状态)
			closure.upvals[i] = stack.closure.upvals[uvIdx]
		}
	}
}

// 关闭指定索引处的Upvalue
func (self *luaState) CloseUpvalues(a int) {
	for i, openuv := range self.stack.openuvs {
		if i >= a-1 {
			// 关闭Upvalue
			val := *openuv.val
			openuv.val = &val
			// 从openuvs表中删除该Upvalue
			delete(self.stack.openuvs, i)
		}
	}
}
