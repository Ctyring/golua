package state

import (
	. "lua/src/api"
	. "lua/src/binchunk"
)

// 闭包
// proto和goFunc必须有一个不为空
type closure struct {
	proto  *Prototype // Lua函数原型
	goFunc GoFunction // Go函数原型
	upvals []*upvalue // upvalue表
}

type upvalue struct {
	val *luaValue // 指向upvalue的值
}

// 创建lua闭包
func newLuaClosure(proto *Prototype) *closure {
	c := &closure{proto: proto}
	// 判断是否有upvalue，有的话创建upvalue表
	if nUpvals := len(proto.Upvalues); nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}

// 创建go闭包
func newGoClosure(f GoFunction, nUpvals int) *closure {
	c := &closure{goFunc: f}
	if nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}
