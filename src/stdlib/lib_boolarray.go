package stdlib

import (
	. "lua/src/api"
	"strconv"
)

var boolArrayLib = map[string]GoFunction{
	"new": newBoolArray,
}

var boolArrayOp = map[string]GoFunction{
	"get":        getBoolArray,
	"set":        setBoolArray,
	"size":       getSize,
	"__newindex": setBoolArray,
	"__index":    getBoolArray,
	"__len":      getSize,
	"__tostring": arraytostring,
}

var boolArrayMeta = map[string]GoFunction{}

func OpenBoolArrayLib(ls LuaState) int {
	ls.NewMetatable("boolarray")
	ls.SetFuncs(boolArrayOp, 0)
	ls.NewLib(boolArrayLib)
	return 1
}

type boolArray struct {
	size int64
	bits []uint
}

func arraytostring(ls LuaState) int {
	ls.CheckType(1, LUA_TUSERDATA)
	ba := *ls.ToUserdata(1)
	ls.ArgCheck(ba != nil, 1, "bool array expected")
	if array, ok := ba.(boolArray); ok {
		ls.PushString("boolArray" + strconv.FormatInt(array.size, 10))
	} else {
		ls.Error2("bool array expected")
	}
	return 1
}

func newBoolArray(ls LuaState) int {
	n := ls.CheckInteger(1)
	ls.ArgCheck(n >= 0, 1, "invalid size")
	var size uint
	if n%32 == 0 {
		size = uint(n / 32)
	} else {
		size = uint(n/32 + 1)
	}
	ud := boolArray{n, make([]uint, size)}
	ls.NewUserdata(ud)
	ls.GetMetatableFromRegistry("boolarray")
	ls.SetMetatable(-2)
	return 1
}

func setBoolArray(ls LuaState) int {
	ls.CheckType(1, LUA_TUSERDATA)
	ba := *ls.ToUserdata(1)
	ls.ArgCheck(ba != nil, 1, "bool array expected")
	if array, ok := ba.(boolArray); ok {
		index := ls.CheckInteger(2) - 1
		ls.ArgCheck(0 <= index && index < int64(len(array.bits)*32), 2, "index out of range")
		value := ls.ToBoolean(3)
		if value {
			array.bits[index/32] |= 1 << (index % 32)
		} else {
			array.bits[index/32] &= ^(1 << (index % 32))
		}
	} else {
		ls.Error2("bool array expected")
	}
	return 1
}

func getBoolArray(ls LuaState) int {
	ls.CheckType(1, LUA_TUSERDATA)
	ba := *ls.ToUserdata(1)
	ls.ArgCheck(ba != nil, 1, "bool array expected")
	if array, ok := ba.(boolArray); ok {
		index := ls.CheckInteger(2) - 1
		ls.ArgCheck(0 <= index && index < int64(len(array.bits)*32), 2, "index out of range")
		ls.PushBoolean(array.bits[index/32]&(1<<(index%32)) != 0)
	} else {
		ls.Error2("bool array expected")
	}
	return 1
}

func getSize(ls LuaState) int {
	ls.CheckType(1, LUA_TUSERDATA)
	ba := *ls.ToUserdata(1)
	ls.ArgCheck(ba != nil, 1, "bool array expected")
	if array, ok := ba.(boolArray); ok {
		ls.PushInteger(array.size)
	} else {
		ls.Error2("bool array expected")
	}
	return 1
}
