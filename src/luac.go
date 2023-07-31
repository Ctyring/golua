package main

import (
	"lua/src/Tools"
	"lua/src/api"
	. "lua/src/binchunk"
	"os"
)

var progname = "luac"
var output = "luac.out"
var listing = false
var dumping = true
var stripping = false

// 实现多个文件的编译聚合
func combine(L api.LuaState, n int) *Prototype {
	if n == 1 {
		return L.ToProto(-1)
	}
	proto := new(Prototype)
	proto.Source = "=(" + progname + ")"
	for i := 0; i < n; i++ {
		subProto := L.ToProto(-1)
		L.Remove(-1)
		proto.Protos = append(proto.Protos, subProto)
		if len(subProto.Upvalues) > 0 {
			proto.Protos[i].Upvalues[0].Instack = 0
		}
	}
	return proto
}

// 参数处理
func doArgs(argc int, argv []string) int {
	version := 0
	i := 1
	if argc > 0 {
		progname = argv[0]
	}
	for ; i < argc; i++ {
		if argv[i][0] != '-' { // 非选项
			break
		} else if argv[i] == "--" { // 选项结束
			i++
			if version != 0 {
				version++
			}
			break
		} else if argv[i] == "-" {
			break
		} else if argv[i] == "-l" {
			listing = true
		} else if argv[i] == "-o" {
			i++
			output = argv[i]
			if output == "" || (output[0] == '-' && output != "-") {
				panic("invalid -o option")
			}
		} else if argv[i] == "-p" {
			dumping = false
		} else if argv[i] == "-s" {
			stripping = true
		} else if argv[i] == "-v" || argv[i] == "--version" {
			version++
		} else {
			panic("invalid option " + argv[i])
		}
	}
	if i == argc && (listing || !dumping) {
		dumping = false
		argv[i] = output
		i--
	}
	if version != 0 {
		println("luac 5.3.4")
		if version == argc-1 {
			os.Exit(0)
		}
	}
	return i
}

func pmain(L api.LuaState) int {
	if !L.CheckStack(len(os.Args)) {
		panic("too many input files")
	}
	for i := 0; i < len(os.Args); i++ {
		filename := os.Args[i]
		if filename == "-" {
			filename = ""
		}
		if L.LoadFile(filename) != api.LUA_OK {
			panic(L.ToString(-1))
		}
	}
	f := combine(L, len(os.Args))
	if listing {
		Tools.List(f)
	}
	if dumping {
		data := Dump(*f)
		os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0666)
		os.WriteFile(output, data, 0666)
	}
	return 0
}

//func main() {
//	i := doArgs(len(os.Args), os.Args)
//	os.Args = append(os.Args[:0], os.Args[i:]...)
//	if len(os.Args) <= 0 {
//		panic("no input files")
//	}
//	L := state.New()
//	if L == nil {
//		panic("cannot create state: not enough memory")
//	}
//	L.PushGoFunction(pmain)
//	L.Call(0, 0)
//}
