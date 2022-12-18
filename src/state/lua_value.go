package state

import (
	"fmt"
	api2 "lua/src/api"
	"lua/src/number"
)

type luaValue interface{}

func typeOf(val luaValue) api2.LuaType {
	switch val.(type) {
	case nil:
		return api2.LUA_TNIL
	case bool:
		return api2.LUA_TBOOLEAN
	case int64:
		return api2.LUA_TNUMBER
	case float64:
		return api2.LUA_TNUMBER
	case string:
		return api2.LUA_TSTRING
	case *luaTable:
		return api2.LUA_TTABLE
	case *closure:
		return api2.LUA_TFUNCTION
	case *luaState:
		return api2.LUA_TTHREAD
	default:
		panic("todo!")
	}
}

func convertToFloat(val luaValue) (float64, bool) {
	switch x := val.(type) {
	case int64:
		return float64(x), true
	case float64:
		return x, true
	case string:
		return number.ParseFloat(x)
	default:
		return 0, false
	}
}

func convertToInteger(val luaValue) (int64, bool) {
	switch x := val.(type) {
	case int64:
		return x, true
	case float64:
		return int64(x), true
	case string:
		return number.ParseInteger(x)
	default:
		return 0, false
	}
}

// 给值关联元表
func setMetatable(val luaValue, mt *luaTable, ls *luaState) {
	// 先判断是否是表，如果是表，直接修改其元表字段
	if t, ok := val.(*luaTable); ok {
		t.metatable = mt
		return
	}
	// 否则把元表存储到注册表
	key := fmt.Sprintf("_MT%d", typeOf(val))
	ls.registry.put(key, mt)
}

// 返回与给定值关联的元表
func getMetatable(val luaValue, ls *luaState) *luaTable {
	// 如果是表，直接返回其元表字段
	if t, ok := val.(*luaTable); ok {
		return t.metatable
	}
	// 否则从注册表中取出元表，还要判断是否存在
	key := fmt.Sprintf("_MT%d", typeOf(val))
	if mt := ls.registry.get(key); mt != nil {
		return mt.(*luaTable)
	}
	return nil
}

// 调用元方法
// 四个参数分别是：操作数1，操作数2，元方法名，Lua状态机(如果操作数不是表，则需要从注册表中取出元表)
func callMetamethod(a, b luaValue, mmName string, ls *luaState) (luaValue, bool) {
	var mm luaValue
	// 依次查看操作数a和b是否有对应的元方法
	if mm = getMetafield(a, mmName, ls); mm == nil {
		if mm = getMetafield(b, mmName, ls); mm == nil {
			return nil, false
		}
	}

	// 入栈
	ls.stack.check(4)
	ls.stack.push(mm)
	ls.stack.push(a)
	ls.stack.push(b)
	// 调用
	ls.Call(2, 1)
	return ls.stack.pop(), true
}

// 获取元方法
func getMetafield(val luaValue, fieldName string, ls *luaState) luaValue {
	if mt := getMetatable(val, ls); mt != nil {
		return mt.get(fieldName)
	}
	return nil
}
