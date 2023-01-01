package util

import (
	"fmt"
	. "lua/src/api"
	"lua/src/binchunk"
)

func PrintStack(ls LuaState) {
	top := ls.GetTop()
	for i := 1; i <= top; i++ {
		t := ls.Type(i)
		switch t {
		case LUA_TBOOLEAN:
			fmt.Printf("[%t]", ls.ToBoolean(i))
		case LUA_TNUMBER:
			fmt.Printf("[%g]", ls.ToNumber(i))
		case LUA_TSTRING:
			fmt.Printf("[%q]", ls.ToString(i))
		default:
			fmt.Printf("[%s]", ls.TypeName(t))
		}
	}
	fmt.Println()
}

// 把函数原型的信息打印到控制台
func List(f *binchunk.Prototype) {
	printHeader(f)
	printCode(f)
	printDeatils(f)
	for _, p := range f.Protos {
		List(p)
	}
}

// 打印函数原型的头部信息
func printHeader(f *binchunk.Prototype) {
	funcType := "main"
	if f.LineDefined > 0 {
		funcType = "function"
	}
	varargFlag := ""
	if f.IsVararg > 0 {
		varargFlag = "+"
	}
	fmt.Printf("\n%s <%s:%d,%d> (%d instructions)\n", funcType, f.Source, f.LineDefined, f.LastLineDefined, len(f.Code))
	fmt.Printf("%d%s params, %d slots, %d upvalues, ", f.NumParams, varargFlag, f.MaxStackSize, len(f.Upvalues))
	fmt.Printf("%d locals, %d constants, %d functions\n", len(f.LocVars), len(f.Constants), len(f.Protos))
}

// 打印指令的序号、行号和十六进制表示
func printCode(f *binchunk.Prototype) {
	for pc, c := range f.Code {
		line := "-"
		if len(f.LineInfo) > 0 {
			line = fmt.Sprintf("%d", f.LineInfo[pc])
		}
		fmt.Printf("\t%d\t[%s]\t0x%08X\n", pc+1, line, c)
	}
}

// 打印常量表、局部变量表和Upvalue表
func printDeatils(f *binchunk.Prototype) {
	fmt.Printf("constants (%d):\n", len(f.Constants))
	for i, c := range f.Constants {
		fmt.Printf("\t%d\t%s\n", i+1, constantToString(c))
	}

	fmt.Printf("locals (%d):\n", len(f.LocVars))
	for i, v := range f.LocVars {
		fmt.Printf("\t%d\t%s\t%d\t%d\n", i, v.VarName, v.StartPC+1, v.EndPC+1)
	}

	fmt.Printf("upvalues (%d):\n", len(f.Upvalues))
	for i, u := range f.Upvalues {
		fmt.Printf("\t%d\t%s\t%d\t%d\n", i, u.Name, u.Instack, u.Idx)
	}
}

// 把常量转换成字符串表示
func constantToString(c interface{}) string {
	switch c.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprint(c)
	case int64:
		return fmt.Sprint(c)
	case float64:
		return fmt.Sprint(c)
	case string:
		return fmt.Sprintf("%q", c)
	default:
		return "?"
	}
}

// 获取Upvalue的名字
func upvalName(f *binchunk.Prototype, idx int) string {
	if len(f.UpvalueNames) > 0 {
		return f.UpvalueNames[idx]
	}
	return "-"
}
