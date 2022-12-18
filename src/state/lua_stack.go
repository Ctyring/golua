package state

import (
	"lua/src/api"
)

// lua栈
type luaStack struct {
	slots   []luaValue // 栈的数据
	top     int        // 栈顶
	prev    *luaStack
	closure *closure   // 当前栈对应的闭包
	varargs []luaValue // 可变参数
	pc      int
	state   *luaState
	openuvs map[int]*upvalue // 存放所有打开的upvalue
}

// 创建指定容量的栈
func newLuaStack(size int, state *luaState) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size),
		top:   0,
		state: state,
	}
}

// 检查空闲空间是否还可以容纳至少n个值
func (self *luaStack) check(n int) {
	free := len(self.slots) - self.top
	// 如果空闲空间不足，则扩容
	for i := free; i < n; i++ {
		self.slots = append(self.slots, nil)
	}
}

// 将值压入栈顶
func (self *luaStack) push(val luaValue) {
	// 如果溢出。终止
	if self.top == len(self.slots) {
		panic("stack overflow!")
	}
	self.slots[self.top] = val
	self.top++
}

// 将栈顶的值弹出
func (self *luaStack) pop() luaValue {
	// 如果栈顶是空的，终止
	if self.top < 1 {
		panic("stack underflow!")
	}
	self.top--
	val := self.slots[self.top]
	self.slots[self.top] = nil // GC
	return val
}

// 把索引转换成绝对索引
func (self *luaStack) absIndex(idx int) int {
	if idx >= 0 || idx <= api.LUA_REGISTRYINDEX {
		return idx
	}
	return idx + self.top + 1
}

// 检查索引是否合法
func (self *luaStack) isValid(idx int) bool {
	// 判断是否是Upvalue
	if idx < api.LUA_REGISTRYINDEX {
		uvIdx := api.LUA_REGISTRYINDEX - idx - 1
		c := self.closure
		return c != nil && uvIdx < len(c.upvals)
	}
	// 伪索引属于有效索引
	if idx == api.LUA_REGISTRYINDEX {
		return true
	}
	absIds := self.absIndex(idx)
	return absIds > 0 && absIds <= self.top
}

// 根据索引从栈里取值
func (self *luaStack) get(idx int) luaValue {
	// 判断是否是Upvalue
	if idx < api.LUA_REGISTRYINDEX {
		uvidx := api.LUA_REGISTRYINDEX - idx - 1
		c := self.closure
		if c == nil || uvidx >= len(c.upvals) {
			return nil
		}
		return *(c.upvals[uvidx].val)
	}
	if idx == api.LUA_REGISTRYINDEX {
		return self.state.registry
	}
	absIds := self.absIndex(idx)
	if absIds > 0 && absIds <= self.top {
		return self.slots[absIds-1]
	}
	return nil
}

// 根据索引设置栈里的值
func (self *luaStack) set(idx int, val luaValue) {
	// 判断是否是Upvalue
	if idx < api.LUA_REGISTRYINDEX {
		uvIdx := api.LUA_REGISTRYINDEX - idx - 1
		c := self.closure
		if c != nil && uvIdx < len(c.upvals) {
			*(c.upvals[uvIdx].val) = val
		}
		return
	}
	// 判断是否是注册表
	if idx == api.LUA_REGISTRYINDEX {
		self.state.registry = val.(*luaTable)
		return
	}
	absIds := self.absIndex(idx)
	if absIds > 0 && absIds <= self.top {
		self.slots[absIds-1] = val
		return
	}
	// 如果索引无效，终止
	panic("invalid index!")
}

// 反转
func (self *luaStack) reverse(from, to int) {
	slots := self.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}
