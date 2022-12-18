package state

import (
	. "lua/src/api"
)

// 创建一个新的线程，把它推入栈顶，同时作为返回值返回
func (self *luaState) NewThread() LuaState {
	t := &luaState{
		registry: self.registry,
	}
	t.pushLuaStack(newLuaStack(LUA_MINSTACK, t))
	self.stack.push(t)
	return t
}

// 让线程进入运行状态
func (self *luaState) Resume(from LuaState, nArgs int) int {
	lsFrom := from.(*luaState)
	if lsFrom.coChan == nil { // 首次使用前需要初始化
		lsFrom.coChan = make(chan int)
	}

	if self.coChan == nil { // 首次使用前需要初始化
		self.coChan = make(chan int)
		self.coCaller = lsFrom
		go func() { // 执行主函数
			self.coStatus = self.PCall(nArgs, LUA_MULTRET, 0)
			lsFrom.coChan <- 1 // 向通道中随意写一个数值
		}()
	} else {
		// resume coroutine
		if self.coStatus != LUA_YIELD { // todo
			self.stack.push("cannot resume non-suspended coroutine")
			return LUA_ERRRUN
		}
		self.coStatus = LUA_OK
		self.coChan <- 1
	}

	<-lsFrom.coChan // 等待协程结束
	return self.coStatus
}

// 让线程进入挂起状态
func (self *luaState) Yield(nResult int) int {
	self.coStatus = LUA_YIELD
	self.coCaller.coChan <- 1 // 通知协作方恢复运行
	<-self.coChan             // 等待再次恢复运行
	return self.coStatus
}

// 返回当前线程状态
func (self *luaState) Status() int {
	return self.coStatus
}

// 返回当前线程是否在调用中(简化版本，本来是用于调试的)
func (self *luaState) GetStack() bool {
	return self.stack.prev != nil
}

func (self *luaState) IsYieldable() bool {
	if self.isMainThread() {
		return false
	}
	return self.coStatus != LUA_YIELD // todo
}
