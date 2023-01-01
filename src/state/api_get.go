package state

import (
	api2 "lua/src/api"
)

// 创建一个空lua表，将其推入栈顶，两个参数指定数组部分和哈希表部分的初始大小
func (self *luaState) CreateTable(nArr, nRec int) {
	t := newLuaTable(nArr, nRec)
	self.stack.push(t)
}

// 属于CreateTable的特殊情况，无法预估大小，所以直接创建一个空表
func (self *luaState) NewTable() {
	self.CreateTable(0, 0)
}

// 根据键(从栈顶弹出)从表中取值，将值推入栈顶
func (self *luaState) GetTable(idx int) api2.LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k, false)
}

// 从表中取值，将值推入栈顶，raw表示是否忽略元方法
func (self *luaState) getTable(t, k luaValue, raw bool) api2.LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		// 如果t是表，表里有v或者需要忽略元方法，或者表里没有__index字段，直接返回
		if raw || v != nil || !tbl.hasMetafield("__index") {
			self.stack.push(v)
			return typeOf(v)
		}
	}
	if !raw {
		if mf := getMetafield(t, "__index", self); mf != nil {
			switch x := mf.(type) {
			case *luaTable: // 如果元方法是表，继续从表中取值
				return self.getTable(x, k, raw)
			case *closure: // 如果元方法是函数，调用函数
				self.stack.push(mf)
				self.stack.push(t)
				self.stack.push(k)
				self.Call(2, 1)
				v := self.stack.get(-1)
				return typeOf(v)
			}
		}
	}
	panic("not a table!")
}

// 根据参数传入的字符串键从表中取值，将值推入栈顶
func (self *luaState) GetField(idx int, k string) api2.LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, k, false)
}

// 传入数字键从表中取值，将值推入栈顶
func (self *luaState) GetI(idx int, i int64) api2.LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i, false)
}

// 将t表的k对应的值压入栈顶
func (self *luaState) RawGet(idx int) api2.LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k, true)
}

func (self *luaState) RawGetI(idx int, i int64) api2.LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i, true)
}

// 把全局环境的某个字段推入栈顶
func (self *luaState) GetGlobal(name string) api2.LuaType {
	t := self.registry.get(api2.LUA_RIDX_GLOBALS)
	return self.getTable(t, name, false)
}

// 查看指定索引处是否有元表，如果有，将元表推入栈顶
func (self *luaState) GetMetatable(idx int) bool {
	val := self.stack.get(idx)
	if mt := getMetatable(val, self); mt != nil {
		self.stack.push(mt)
		return true
	}
	return false
}

func (self *luaState) GetUpvalue(idx, n int) {
	c := self.stack.get(idx).(*closure)
	if c != nil {
		uvInfo := c.proto.Upvalues[n-1]
		if uvInfo.Instack == 1 {
			self.stack.push(c.upvals[uvInfo.Idx].val)
		} else {
			self.stack.push(c.upvals[uvInfo.Idx])
		}
	}
}
