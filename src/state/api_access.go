package state

import (
	"fmt"
	. "lua/src/api"
	"lua/src/binchunk"
)

func (self *luaState) RawLen(idx int) uint {
	val := self.stack.get(idx)
	switch x := val.(type) {
	case string:
		return uint(len(x))
	case *luaTable:
		return uint(x.len())
	default:
		return 0
	}
}

// 把给定Lua类型转换成对应的字符串表示
func (self *luaState) TypeName(tp LuaType) string {
	switch tp {
	case LUA_TNONE:
		return "no value"
	case LUA_TNIL:
		return "nil"
	case LUA_TBOOLEAN:
		return "boolean"
	case LUA_TNUMBER:
		return "number"
	case LUA_TSTRING:
		return "string"
	case LUA_TTABLE:
		return "table"
	case LUA_TFUNCTION:
		return "function"
	case LUA_TTHREAD:
		return "thread"
	case LUA_TUSERDATA:
		return "userdata"
	default:
		return "userdata"
	}
}

// 根据索引返回值的类型
func (self *luaState) Type(idx int) LuaType {
	if self.stack.isValid(idx) {
		val := self.stack.get(idx)
		return typeOf(val)
	}
	return LUA_TNONE
}

// 判定索引处的值是否属于特定类型
func (self *luaState) IsNone(idx int) bool {
	return self.Type(idx) == LUA_TNONE
}

func (self *luaState) IsNil(idx int) bool {
	return self.Type(idx) == LUA_TNIL
}

func (self *luaState) IsNoneOrNil(idx int) bool {
	return self.Type(idx) <= LUA_TNIL
}

func (self *luaState) IsBoolean(idx int) bool {
	return self.Type(idx) == LUA_TBOOLEAN
}

// 判定索引处的值是否是数字或是字符串(考虑类型转换)
func (self *luaState) IsString(idx int) bool {
	t := self.Type(idx)
	return t == LUA_TSTRING || t == LUA_TNUMBER
}

func (self *luaState) IsNumber(idx int) bool {
	_, ok := self.ToNumberX(idx)
	return ok
}

func (self *luaState) IsInteger(idx int) bool {
	val := self.stack.get(idx)
	_, ok := val.(int64)
	return ok
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_istable
func (self *luaState) IsTable(idx int) bool {
	return self.Type(idx) == LUA_TTABLE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isfunction
func (self *luaState) IsFunction(idx int) bool {
	return self.Type(idx) == LUA_TFUNCTION
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isthread
func (self *luaState) IsThread(idx int) bool {
	return self.Type(idx) == LUA_TTHREAD
}

// 在索引处取出一个布尔值，如果索引处的值不是布尔值，那么进行类型转换
func (self *luaState) ToBoolean(idx int) bool {
	val := self.stack.get(idx)
	return convertToBoolean(val)
}

func convertToBoolean(val luaValue) bool {
	switch x := val.(type) {
	case nil:
		return false
	case bool:
		return x
	default:
		return true
	}
}

// 在索引处取出一个数字，如果索引处的值不是数字，那么进行类型转换
func (self *luaState) ToNumber(idx int) float64 {
	n, _ := self.ToNumberX(idx)
	return n
}

func (self *luaState) ToNumberX(idx int) (float64, bool) {
	val := self.stack.get(idx)
	return convertToFloat(val)
}

// 取整数
func (self *luaState) ToInteger(idx int) int64 {
	i, _ := self.ToIntegerX(idx)
	return i
}

func (self *luaState) ToIntegerX(idx int) (int64, bool) {
	val := self.stack.get(idx)
	return convertToInteger(val)
}

// 取字符串
func (self *luaState) ToString(idx int) string {
	s, _ := self.ToStringX(idx)
	return s
}

func (self *luaState) ToStringX(idx int) (string, bool) {
	val := self.stack.get(idx)
	switch x := val.(type) {
	case string:
		return x, true
	case int64, float64:
		s := fmt.Sprintf("%v", x)
		self.stack.set(idx, s)
		return s, true
	default:
		return "", false
	}
}

func (self *luaState) IsGoFunction(idx int) bool {
	// 先拿到索引处的值
	val := self.stack.get(idx)
	// 判断是否能转换为闭包
	if c, ok := val.(*closure); ok {
		// 判断是否是Go函数
		return c.goFunc != nil
	}
	return false
}

func (self *luaState) ToGoFunction(idx int) GoFunction {
	val := self.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc
	}
	return nil
}

// 将指定索引处的值转换为指针
func (self *luaState) ToPointer(idx int) interface{} {
	// todo
	return self.stack.get(idx)
}

// 将指定索引处的值转换为线程
func (self *luaState) ToThread(idx int) LuaState {
	val := self.stack.get(idx)
	if val != nil {
		if thread, ok := val.(*luaState); ok {
			return thread
		}
	}
	return nil
}

// 将制定索引处的值转换为函数原型
func (self *luaState) ToProto(idx int) *binchunk.Prototype {
	val := self.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.proto
	}
	return nil
}

func (self *luaState) ToUserdata(idx int) *interface{} {
	val := self.stack.get(idx)
	if udata, ok := val.(*userdata); ok {
		return &udata.val
	}
	return nil
}
