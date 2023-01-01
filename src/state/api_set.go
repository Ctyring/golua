package state

import (
	. "lua/src/api"
)

// 把键值写入表，键和值都从栈顶弹出
func (self *luaState) SetTable(idx int) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	k := self.stack.pop()
	self.setTable(t, k, v, false)
}

func (self *luaState) setTable(t, k, v luaValue, raw bool) {
	if tbl, ok := t.(*luaTable); ok {
		// 如果t是表，表里有k，或者忽略元方法，或者没有元方法
		if raw || tbl.get(k) != nil || !tbl.hasMetafield("__newindex") {
			tbl.put(k, v)
			return
		}
	}
	if !raw {
		if mf := getMetafield(t, "__newindex", self); mf != nil {
			switch x := mf.(type) {
			case *luaTable: // 如果元方法是表，把k和v写入表
				self.setTable(x, k, v, false)
			case *closure: // 如果元方法是函数，调用函数
				self.stack.push(t)
				self.stack.push(k)
				self.stack.push(v)
				self.Call(3, 0)
				return
			}
		}
	}
	panic("not a table!")
}

// 把值写入表，键从参数传入(字符串)，值从栈顶弹出
func (self *luaState) SetField(idx int, k string) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, k, v, false)
}

// 把值写入表，键从参数传入(数字)，值从栈顶弹出
func (self *luaState) SetI(idx int, i int64) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, i, v, false)
}

func (self *luaState) RawSet(idx int) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	k := self.stack.pop()
	self.setTable(t, k, v, true)
}

func (self *luaState) RawSetI(idx int, i int64) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, i, v, true)
}

// 向全局变量写入一个值
func (self *luaState) SetGlobal(name string) {
	t := self.registry.get(LUA_RIDX_GLOBALS)
	v := self.stack.pop()
	self.setTable(t, name, v, false)
}

// 给全局环境注册Go函数值
func (self *luaState) Register(name string, f GoFunction) {
	// 把go函数压入栈
	self.PushGoFunction(f)
	// 放入全局环境
	self.SetGlobal(name)
}

// 把值写入表，键和值都从参数传入
func (self *luaState) SetMetatable(idx int) {
	// 从栈顶弹出值
	val := self.stack.get(idx)
	// 从栈顶弹出表
	mtVal := self.stack.pop()
	if mtVal == nil { // 如果mtVal是nil，把元表val设为nil
		setMetatable(val, nil, self)
	} else if tbl, ok := mtVal.(*luaTable); ok { // 如果mtVal是表，把元表val设为mtVal
		setMetatable(val, tbl, self)
	} else {
		panic("table expected!")
	}
}

func (self *luaState) SetUpvalue(idx int, n int) {
	fi := self.stack.get(idx)
	if c, ok := fi.(*closure); ok {
		val := self.stack.pop()
		if uv, ok := val.(*upvalue); ok {
			c.upvals[n-1] = uv
		}
	}
}
