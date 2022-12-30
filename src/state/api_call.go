package state

import (
	"lua/src/api"
	"lua/src/binchunk"
	"lua/src/compiler"
	. "lua/src/vm"
)

// 加载二进制chunk，第一个参数是二进制chunk，第二个参数是chunk名字，第三个参数指定加载模式("b" 二进制 "t" 文本 "bt" 二进制或文本)
func (self *luaState) Load(chunk []byte, chunkName, mode string) int {
	var proto *binchunk.Prototype
	if binchunk.IsBinaryChunk(chunk) { // 如果是二进制chunk
		proto = binchunk.Undump(chunk) // 解析二进制chunk
	} else {
		proto = compiler.Compile(string(chunk), chunkName) // 编译文本chunk
	}
	//util.List(proto)
	c := newLuaClosure(proto)
	self.stack.push(c)
	// 判断是否需要Upvalue
	if len(proto.Upvalues) > 0 {
		env := self.registry.get(api.LUA_RIDX_GLOBALS) // 获取全局环境表
		c.upvals[0] = &upvalue{&env}                   // 把全局环境表作为第一个Upvalue
	}
	return api.LUA_OK
}

// 调用Lua函数
// 第一个参数是参数个数，第二个参数是返回值个数
func (self *luaState) Call(nArgs, nResults int) {
	// 根据索引取出函数，判断是否真的是Lua函数
	val := self.stack.get(-(nArgs + 1))
	c, ok := val.(*closure)

	if !ok { // 如果被调用值不是函数，就查找并调用元方法
		if mf := getMetafield(val, "__call", self); mf != nil {
			if c, ok = mf.(*closure); ok {
				self.stack.push(val)
				self.Insert(-(nArgs + 2))
				nArgs += 1
			}
		}
	}
	if ok {
		if c.proto != nil {
			//fmt.Printf("call %s<%d,%d>\n", c.proto.Source, c.proto.LineDefined, c.proto.LastLineDefined)
			self.callLuaClosure(nArgs, nResults, c) // 调用Lua函数
		} else {
			self.callGoClosure(nArgs, nResults, c) // 调用Go函数
		}
	} else {
		panic("not function!")
	}
}

// 调用Lua函数
func (self *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	// 拿到编译器为我们事先准备好的信息
	nRegs := int(c.proto.MaxStackSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	// 创建Lua栈帧
	newStack := newLuaStack(nRegs+api.LUA_MINSTACK, self)
	// 把闭包和调用帧联系起来
	newStack.closure = c

	// 把参数传递给新的Lua栈帧
	funcAndArgs := self.stack.popN(nArgs + 1) // 把函数和参数弹出
	newStack.pushN(funcAndArgs[1:], nParams)  // 把参数传递给新的Lua栈帧
	newStack.top = nRegs                      // 设置栈顶
	if nArgs > nParams && isVararg {          // 如果参数个数大于参数个数，且是可变参数
		newStack.varargs = funcAndArgs[nParams+1:] // 把多余的参数传递给可变参数
	}

	// 把新的Lua栈帧压入Lua虚拟机栈
	self.pushLuaStack(newStack)
	// 执行Lua函数
	self.runLuaClosure()
	// 弹出被调用帧
	self.popLuaStack()

	// 根据期望的返回值个数，从新的Lua栈帧中弹出返回值
	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs) // 弹出返回值
		self.stack.check(len(results))                 // 检查栈空间
		self.stack.pushN(results, nResults)            // 把返回值传递给调用者
	}
}

// 调用Go函数
func (self *luaState) callGoClosure(nArgs, nResults int, c *closure) {
	// 创建Lua栈帧
	newStack := newLuaStack(nArgs+api.LUA_MINSTACK, self)
	// 把闭包和调用帧联系起来
	newStack.closure = c

	// 拿到栈里的参数
	args := self.stack.popN(nArgs)
	// 压入调用帧
	newStack.pushN(args, nArgs)
	// 弹出函数名
	self.stack.pop()

	// 把新的Lua栈帧压入Lua虚拟机栈
	self.pushLuaStack(newStack)
	// 执行Go函数
	r := c.goFunc(self)
	// 弹出被调用帧
	self.popLuaStack()

	// 把返回值压入主调用帧，多退少补
	if nResults != 0 {
		results := newStack.popN(r)
		self.stack.check(len(results))
		self.stack.pushN(results, nResults)
	}
}

// 执行被调函数
func (self *luaState) runLuaClosure() {
	for {
		inst := Instruction(self.Fetch())
		//util.PrintStack(self)
		inst.Execute(self)
		if inst.Opcode() == OP_RETURN {
			break
		}
	}
}

// 从栈顶弹出n个值
func (self *luaStack) popN(n int) []luaValue {
	vals := make([]luaValue, n)
	for i := n - 1; i >= 0; i-- {
		vals[i] = self.pop()
	}
	return vals
}

// 把n个值压入栈顶
func (self *luaStack) pushN(vals []luaValue, n int) {
	nVals := len(vals)
	if n < 0 {
		n = nVals // n < 0时, 压入全部值
	}
	for i := 0; i < n; i++ {
		if i < nVals {
			self.push(vals[i])
		} else {
			self.push(nil) // 压入nil补齐
		}
	}
}

// 执行代码
func (self *luaState) PCall(nArgs, nResults, msgh int) (status int) {
	caller := self.stack
	status = api.LUA_ERRRUN

	// 定义一个匿名函数延时执行，用来做错误处理
	defer func() {
		if err := recover(); err != nil {
			if msgh != 0 {
				panic(err)
			}
			for self.stack != caller {
				self.popLuaStack()
			}
			self.stack.push(err)
		}
	}()

	self.Call(nArgs, nResults)
	status = api.LUA_OK
	return
}
