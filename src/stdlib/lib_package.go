package stdlib

import (
	. "lua/src/api"
	"os"
)
import "strings"

/* key, in the registry, for table of loaded modules */
const LUA_LOADED_TABLE = "_LOADED"

/* key, in the registry, for table of preloaded loaders */
const LUA_PRELOAD_TABLE = "_PRELOAD"

const (
	LUA_DIRSEP    = string(os.PathSeparator)
	LUA_PATH_SEP  = ";"
	LUA_PATH_MARK = "?"
	LUA_EXEC_DIR  = "!"
	LUA_IGMARK    = "-"
)

var pkgFuncs = map[string]GoFunction{
	"searchpath": pkgSearchPath,
	/* placeholders */
	"preload":   nil,
	"cpath":     nil,
	"path":      nil,
	"searchers": nil,
	"loaded":    nil,
}

var llFuncs = map[string]GoFunction{
	"require": pkgRequire,
}

// 包模块的开启函数
func OpenPackageLib(ls LuaState) int {
	ls.NewLib(pkgFuncs) /* create 'package' table */
	createSearchersTable(ls)
	/* set paths */
	ls.PushString("./?.lua;./?/init.lua")
	ls.SetField(-2, "path")
	/* store config information */
	ls.PushString(LUA_DIRSEP + "\n" + LUA_PATH_SEP + "\n" +
		LUA_PATH_MARK + "\n" + LUA_EXEC_DIR + "\n" + LUA_IGMARK + "\n")
	ls.SetField(-2, "config")
	/* set field 'loaded' */
	ls.GetSubTable(LUA_REGISTRYINDEX, LUA_LOADED_TABLE)
	ls.SetField(-2, "loaded")
	/* set field 'preload' */
	ls.GetSubTable(LUA_REGISTRYINDEX, LUA_PRELOAD_TABLE)
	ls.SetField(-2, "preload")
	ls.PushGlobalTable()
	ls.PushValue(-2)        /* set 'package' as upvalue for next lib */
	ls.SetFuncs(llFuncs, 1) /* open lib into global table */
	ls.Pop(1)               /* pop global table */
	return 1                /* return 'package' table */
}

// 初始化package.searchers表
func createSearchersTable(ls LuaState) {
	searchers := []GoFunction{ // 目前实现了两个搜索器
		preloadSearcher,
		luaSearcher,
	}
	/* create 'searchers' table */
	ls.CreateTable(len(searchers), 0)
	/* fill it with predefined searchers */
	for idx, searcher := range searchers {
		ls.PushValue(-2) /* set 'package' as upvalue for all searchers */
		ls.PushGoClosure(searcher, 1)
		ls.RawSetI(-2, int64(idx+1))
	}
	ls.SetField(-2, "searchers") /* put it in field 'searchers' */
}

// preload搜索器
func preloadSearcher(ls LuaState) int {
	name := ls.CheckString(1)
	ls.GetField(LUA_REGISTRYINDEX, "_PRELOAD") // 获取预加载表
	if ls.GetField(-1, name) == LUA_TNIL {     /* not found? */
		ls.PushString("\n\tno field package.preload['" + name + "']")
	}
	return 1
}

// lua搜索器
func luaSearcher(ls LuaState) int {
	// 首先检查第一个参数是否是一个字符串，并保存为name
	name := ls.CheckString(1)
	ls.GetField(LuaUpvalueIndex(1), "path")
	path, ok := ls.ToStringX(-1)
	if !ok {
		ls.Error2("'package.path' must be a string")
	}

	filename, errMsg := _searchPath(name, path, ".", LUA_DIRSEP)
	if errMsg != "" {
		ls.PushString(errMsg)
		return 1
	}

	// 判断是否加载模块成功
	if ls.LoadFile(filename) == LUA_OK {
		ls.PushString(filename) /* will be 2nd argument to module */
		return 2                /* return open function and file name */
	} else {
		return ls.Error2("error loading module '%s' from file '%s':\n\t%s",
			ls.CheckString(1), filename, ls.CheckString(-1))
	}
}

// 封装了一个搜索路径的函数
func pkgSearchPath(ls LuaState) int {
	name := ls.CheckString(1)                                                // 模块名
	path := ls.CheckString(2)                                                // 搜索路径
	sep := ls.OptString(3, ".")                                              // 路径分隔符
	rep := ls.OptString(4, LUA_DIRSEP)                                       // 目录分隔符
	if filename, errMsg := _searchPath(name, path, sep, rep); errMsg == "" { // 搜索成功
		ls.PushString(filename)
		return 1
	} else { // 搜索失败
		ls.PushNil()
		ls.PushString(errMsg)
		return 2
	}
}

// 在搜索路径中搜索Lua文件
// 参数：文件名，路径字符串，路径分隔符，目录分隔符
func _searchPath(name, path, sep, dirSep string) (filename, errMsg string) {
	if sep != "" {
		name = strings.Replace(name, sep, dirSep, -1) // 将路径分隔符替换为目录分隔符
	}

	for _, filename := range strings.Split(path, LUA_PATH_SEP) { // 将路径拆分为多个路径
		filename = strings.Replace(filename, LUA_PATH_MARK, name, -1)
		if _, err := os.Stat(filename); !os.IsNotExist(err) { // 检查路径是否存在
			return filename, ""
		}
		errMsg += "\n\tno file '" + filename + "'"
	}

	return "", errMsg
}

// require (modname)
// http://www.lua.org/manual/5.3/manual.html#pdf-require
// 加载模块
func pkgRequire(ls LuaState) int {
	name := ls.CheckString(1)                        // 模块名
	ls.SetTop(1)                                     /* LOADED table will be at index 2 */
	ls.GetField(LUA_REGISTRYINDEX, LUA_LOADED_TABLE) // 获取已加载表
	ls.GetField(2, name)                             // 获取name模块
	if ls.ToBoolean(-1) {                            // 判断name模块是否在已加载表中
		return 1 /* package is already loaded */
	}
	/* else must load package */
	ls.Pop(1)             // 将上面的获取结果弹出
	_findLoader(ls, name) // 查找加载器
	ls.PushString(name)   // 传参
	ls.Insert(-2)         // 将name插入到加载器之前
	ls.Call(2, 1)         // 调用加载器
	if !ls.IsNil(-1) {    // 判断返回值是否为nil
		ls.SetField(2, name) /* LOADED[name] = returned value */
	}
	if ls.GetField(2, name) == LUA_TNIL { /* module set no value? */
		ls.PushBoolean(true) /* use true as result */
		ls.PushValue(-1)     /* extra copy to be returned */
		ls.SetField(2, name) /* LOADED[name] = true */
	}
	return 1
}

// 搜索加载器
func _findLoader(ls LuaState, name string) {
	/* push 'package.searchers' to index 3 in the stack */
	if ls.GetField(LuaUpvalueIndex(1), "searchers") != LUA_TTABLE { // 获取搜索器表
		ls.Error2("'package.searchers' must be a table")
	}

	/* to build error message */
	errMsg := "module '" + name + "' not found:"

	/*  iterate over available searchers to find a loader */
	for i := int64(1); ; i++ { // 遍历搜索器表寻找加载器
		if ls.RawGetI(3, i) == LUA_TNIL { // 判断是否已经没有下一个搜索器
			ls.Pop(1)         /* remove nil */
			ls.Error2(errMsg) /* create error message */
		}

		ls.PushString(name)
		ls.Call(1, 2)          // 调用搜索器
		if ls.IsFunction(-2) { // 判断是否找到加载器
			return /* module loader found */
		} else if ls.IsString(-2) { // 判断是否找到错误信息
			ls.Pop(1)                    /* remove extra return */
			errMsg += ls.CheckString(-1) /* concatenate error message */
		} else {
			ls.Pop(2) /* remove both returns */
		}
	}
}
