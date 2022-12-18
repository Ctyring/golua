package state

import (
	"lua/src/api"
	"lua/src/binchunk"
)

// 闭包
// proto和goFunc必须有一个不为空
type closure struct {
	proto  *binchunk.Prototype // Lua函数原型
	goFunc api.GoFunction      // Go函数原型
	upvals []*upvalue          // upvalue表
}

type upvalue struct {
	val *luaValue // 指向upvalue的值
}

// 创建lua闭包
func newLuaClosure(proto *binchunk.Prototype) *closure {
	c := &closure{proto: proto}
	// 判断是否有upvalue，有的话创建upvalue表
	if nUpvals := len(proto.Upvalues); nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}

// 创建go闭包
func newGoClosure(f api.GoFunction, nUpvals int) *closure {
	c := &closure{goFunc: f}
	if nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}
