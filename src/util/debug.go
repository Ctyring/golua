package util

import (
	"fmt"
	api2 "lua/src/api"
)

func printStack(ls api2.LuaState) {
	top := ls.GetTop()
	for i := 1; i <= top; i++ {
		t := ls.Type(i)
		switch t {
		case api2.LUA_TBOOLEAN:
			fmt.Printf("[%t]", ls.ToBoolean(i))
		case api2.LUA_TNUMBER:
			fmt.Printf("[%g]", ls.ToNumber(i))
		case api2.LUA_TSTRING:
			fmt.Printf("[%q]", ls.ToString(i))
		default:
			fmt.Printf("[%s]", ls.TypeName(t))
		}
	}
	fmt.Println()
}
