package state

import (
	. "lua/src/api"
)

type luaState struct {
	registry *luaTable // 注册表
	stack    *luaStack
	coCaller *luaState // 调用协程的协程
	coStatus int       // 协程状态
	coChan   chan int  // 协程通道
}

// 创建LuaState实例
func New() LuaState {
	ls := &luaState{}

	registry := newLuaTable(8, 0)
	registry.put(LUA_RIDX_MAINTHREAD, ls)              // 创建主线程
	registry.put(LUA_RIDX_GLOBALS, newLuaTable(0, 20)) // 全局环境

	ls.registry = registry
	ls.pushLuaStack(newLuaStack(LUA_MINSTACK, ls)) // 创建Lua栈
	return ls
}

// 向头部添加一个调用帧
func (self *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = self.stack
	self.stack = stack
}

// 从头部移除一个调用帧
func (self *luaState) popLuaStack() {
	stack := self.stack
	self.stack = stack.prev
	stack.prev = nil
}

// 判断是否是主线程
func (self *luaState) isMainThread() bool {
	return self.registry.get(LUA_RIDX_MAINTHREAD) == self
}
