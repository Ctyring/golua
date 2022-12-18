package state

import (
	. "lua/src/api"
)

// 返回栈顶索引
func (self *luaState) GetTop() int {
	return self.stack.top
}

// 将索引转换成绝对索引
func (self *luaState) AbsIndex(idx int) int {
	return self.stack.absIndex(idx)
}

// 扩容
func (self *luaState) CheckStack(n int) bool {
	self.stack.check(n)
	return true // 目前假设不会失败
}

// 弹出n个值
func (self *luaState) Pop(n int) {
	for i := 0; i < n; i++ {
		self.stack.pop()
	}
}

// 复制栈中的值到另一个位置
func (self *luaState) Copy(fromIdx, toIdx int) {
	val := self.stack.get(fromIdx)
	self.stack.set(toIdx, val)
}

// 把指定索引处的值移动到栈顶
func (self *luaState) PushValue(idx int) {
	val := self.stack.get(idx)
	self.stack.push(val)
}

// 把栈顶的值移动到指定索引处
func (self *luaState) Replace(idx int) {
	val := self.stack.pop()
	self.stack.set(idx, val)
}

// 把栈顶的值插入到指定索引处
func (self *luaState) Insert(idx int) {
	// 插入是旋转的一种特殊情况
	self.Rotate(idx, 1)
}

// 删除指定索引处的值
func (self *luaState) Remove(idx int) {
	// 删除是旋转的一种特殊情况
	self.Rotate(idx, -1)
	self.Pop(1)
}

// 旋转栈中的值
// 将区间[idx, top]内的值循环向上移动n个位置
func (self *luaState) Rotate(idx, n int) {
	t := self.stack.top - 1           // 栈顶索引
	p := self.stack.absIndex(idx) - 1 // 要旋转的值的索引
	var m int
	if n >= 0 {
		m = t - n
	} else {
		m = p - n - 1
	}
	self.stack.reverse(p, m)   // [p, m]
	self.stack.reverse(m+1, t) // [m+1, t]
	self.stack.reverse(p, t)   // [p, t]
}

// 设置栈顶索引
// 如果指定值小于当前栈顶元素，效果相当于弹出，如果大于当前栈顶元素，效果相当于压入nil
func (self *luaState) SetTop(idx int) {
	newTop := self.stack.absIndex(idx)
	if newTop < 0 {
		panic("stack underflow!")
	}

	n := self.stack.top - newTop
	if n > 0 {
		for i := 0; i < n; i++ {
			self.stack.pop()
		}
	} else if n < 0 {
		for i := 0; i > n; i-- {
			self.stack.push(nil)
		}
	}
}

func (self *luaState) XMove(to LuaState, n int) {
	vals := self.stack.popN(n)
	to.(*luaState).stack.pushN(vals, n)
}
