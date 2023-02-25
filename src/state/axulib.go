package state

import (
	"fmt"
	. "lua/src/api"
	. "lua/src/stdlib"
	"os"
)

// 增强报错函数
func (self *luaState) Error2(fmt string, a ...interface{}) int {
	self.PushFString(fmt, a...) // 添加个格式化
	return self.Error()
}

// 报参数错误
func (self *luaState) ArgError(arg int, extraMsg string) int {
	// bad argument #arg to 'funcname' (extramsg)
	return self.Error2("bad argument #%d (%s)", arg, extraMsg) // todo
}

// 增强检查并扩容占空间
func (self *luaState) CheckStack2(sz int, msg string) {
	if !self.CheckStack(sz) { // 增加了报错处理
		if msg != "" {
			self.Error2("stack overflow (%s)", msg)
		} else {
			self.Error2("stack overflow")
		}
	}
}

// 参数检查 三个参数 1.检查是否通过 2.参数索引 3.附加信息
func (self *luaState) ArgCheck(cond bool, arg int, extraMsg string) {
	if !cond {
		self.ArgError(arg, extraMsg)
	}
}

// 确保某个参数一定存在
func (self *luaState) CheckAny(arg int) {
	if self.Type(arg) == LUA_TNONE {
		self.ArgError(arg, "value expected")
	}
}

// 确保某个参数属于指定类型
func (self *luaState) CheckType(arg int, t LuaType) {
	if self.Type(arg) != t {
		self.tagError(arg, t)
	}
}

// 确保某个参数属于整数类型
func (self *luaState) CheckInteger(arg int) int64 {
	i, ok := self.ToIntegerX(arg)
	if !ok {
		self.intError(arg)
	}
	return i
}

// 确保某个参数是数字
func (self *luaState) CheckNumber(arg int) float64 {
	f, ok := self.ToNumberX(arg)
	if !ok {
		self.tagError(arg, LUA_TNUMBER)
	}
	return f
}

// 字符串转换检查，确保arg是字符串
func (self *luaState) CheckString(arg int) string {
	s, ok := self.ToStringX(arg) // 转换为字符串
	if !ok {
		self.tagError(arg, LUA_TSTRING)
	}
	return s
}

// 对可选参数进行检查，如果可选参数有值，确保该值属于指定类型，否则返回默认值
func (self *luaState) OptInteger(arg int, def int64) int64 {
	if self.IsNoneOrNil(arg) {
		return def
	}
	return self.CheckInteger(arg)
}

// 对可选参数进行检查，如果可选参数有值，确保该值属于指定类型，否则返回默认值
func (self *luaState) OptNumber(arg int, def float64) float64 {
	if self.IsNoneOrNil(arg) {
		return def
	}
	return self.CheckNumber(arg)
}

// 对可选参数进行检查，如果可选参数有值，确保该值属于指定类型，否则返回默认值
func (self *luaState) OptString(arg int, def string) string {
	if self.IsNoneOrNil(arg) {
		return def
	}
	return self.CheckString(arg)
}

// 加载并用保护模式执行文件
func (self *luaState) DoFile(filename string) bool {
	return self.LoadFile(filename) != LUA_OK ||
		self.PCall(0, LUA_MULTRET, 0) != LUA_OK
}

// 记载并用保护模式执行字符串
func (self *luaState) DoString(str string) bool {
	return self.LoadString(str) != LUA_OK ||
		self.PCall(0, LUA_MULTRET, 0) != LUA_OK
}

// 以默认模式加载文件
func (self *luaState) LoadFile(filename string) int {
	return self.LoadFileX(filename, "bt")
}

// 加载文件
func (self *luaState) LoadFileX(filename, mode string) int {
	if data, err := os.ReadFile(filename); err == nil {
		return self.Load(data, "@"+filename, mode)
	} else {
		self.PushString(err.Error())
		return LUA_ERRFILE
	}
}

// 加载字符串
func (self *luaState) LoadString(s string) int {
	return self.Load([]byte(s), s, "bt")
}

// 获取类型名
func (self *luaState) TypeName2(idx int) string {
	return self.TypeName(self.Type(idx))
}

// 增强获取长度方法
func (self *luaState) Len2(idx int) int64 {
	self.Len(idx)                   // 获取长度
	i, isNum := self.ToIntegerX(-1) // 转换为整数(考虑调用了元方法的情况下不一定是整数)
	if !isNum {                     // 不是整数
		self.Error2("object length is not an integer")
	}
	self.Pop(1)
	return i
}

// 增强字符串转化(适用于所有类型)
func (self *luaState) ToString2(idx int) string {
	if self.CallMeta(idx, "__tostring") { // 先判断是否使用元方法
		if !self.IsString(-1) {
			self.Error2("'__tostring' must return a string")
		}
	} else {
		switch self.Type(idx) {
		case LUA_TNUMBER: // 数字
			if self.IsInteger(idx) {
				self.PushString(fmt.Sprintf("%d", self.ToInteger(idx))) // todo
			} else {
				self.PushString(fmt.Sprintf("%g", self.ToNumber(idx))) // todo
			}
		case LUA_TSTRING: // 字符串
			self.PushValue(idx)
		case LUA_TBOOLEAN: // 布尔值
			if self.ToBoolean(idx) {
				self.PushString("true")
			} else {
				self.PushString("false")
			}
		case LUA_TNIL: // nil
			self.PushString("nil")
		default: // 其他类型
			tt := self.GetMetafield(idx, "__name") // 获取元表的__name域
			var kind string
			if tt == LUA_TSTRING { // 如果name是字符串
				kind = self.CheckString(-1) // 获取字符串
			} else { // 否则
				kind = self.TypeName2(idx) // 获取类型名
			}

			self.PushString(fmt.Sprintf("%s: %p", kind, self.ToPointer(idx)))
			if tt != LUA_TNIL {
				self.Remove(-2) // 如果tt不是nil, 则移除name
			}
		}
	}
	return self.CheckString(-1) // 检查栈顶并转换为string
}

// 检查索引处的表的某个字段表，如果该字段不是表，创建一个空表赋值给该字段并返回false
func (self *luaState) GetSubTable(idx int, fname string) bool {
	if self.GetField(idx, fname) == LUA_TTABLE {
		return true /* table already there */
	}
	self.Pop(1) /* remove previous result */
	idx = self.stack.absIndex(idx)
	self.NewTable()
	self.PushValue(-1)        /* copy to be left at top */
	self.SetField(idx, fname) /* assign new table to field */
	return false              /* false, because did not find table there */
}

// 获取元表obj的event字段
func (self *luaState) GetMetafield(obj int, event string) LuaType {
	if !self.GetMetatable(obj) { // 如果获取不到元表，直接返回nil
		return LUA_TNIL
	}

	self.PushString(event)
	tt := self.RawGet(-2) // 取值元表中的event字段
	if tt == LUA_TNIL {   // 如果event字段不存在
		self.Pop(2) // 弹出nil和元表
	} else {
		self.Remove(-2) // 只弹出表，保留值
	}
	return tt // 返回值
}

// 调用元方法
func (self *luaState) CallMeta(obj int, event string) bool {
	obj = self.AbsIndex(obj)
	if self.GetMetafield(obj, event) == LUA_TNIL { // 如果没有元方法
		return false
	}
	// 此时栈顶为元方法
	self.PushValue(obj) // 将obj压入栈顶
	self.Call(1, 1)
	return true
}

// 开启标准库
func (self *luaState) OpenLibs() {
	// 声明要开启的标准库
	libs := map[string]GoFunction{
		"_G":        OpenBaseLib,
		"math":      OpenMathLib,
		"table":     OpenTableLib,
		"string":    OpenStringLib,
		"utf8":      OpenUTF8Lib,
		"os":        OpenOSLib,
		"package":   OpenPackageLib,
		"coroutine": OpenCoroutineLib,
	}

	// 循环调用各个标准库的开启函数
	for name, fun := range libs {
		self.RequireF(name, fun, true)
		self.Pop(1)
	}
}

// 开启单个标准库
func (self *luaState) RequireF(modname string, openf GoFunction, glb bool) {
	self.GetSubTable(LUA_REGISTRYINDEX, "_LOADED")
	self.GetField(-1, modname) /* LOADED[modname] */
	if !self.ToBoolean(-1) {   /* package not already loaded? */
		self.Pop(1) /* remove field */
		self.PushGoFunction(openf)
		self.PushString(modname)   /* argument to open function */
		self.Call(1, 1)            /* call 'openf' to open module */
		self.PushValue(-1)         /* make copy of module (call result) */
		self.SetField(-3, modname) /* _LOADED[modname] = module */
	}
	self.Remove(-2) /* remove _LOADED table */
	if glb {
		self.PushValue(-1)      /* copy of module */
		self.SetGlobal(modname) /* _G[modname] = module */
	}
}

// 创建一个新库
func (self *luaState) NewLib(l FuncReg) {
	// 创建新表
	self.NewLibTable(l)
	// 设置函数到对应位置
	self.SetFuncs(l, 0)
}

// 创建一个新的表，用于存放库函数
func (self *luaState) NewLibTable(l FuncReg) {
	self.CreateTable(0, len(l))
}

// 将库函数注册到表中
func (self *luaState) SetFuncs(l FuncReg, nup int) {
	self.CheckStack2(nup, "too many upvalues")
	for name, fun := range l { /* fill the table with given functions */
		for i := 0; i < nup; i++ { // 把upvalue放到栈顶，方便后面的注册
			self.PushValue(-nup)
		}
		// r[-(nup+2)][name]=fun
		self.PushGoClosure(fun, nup) /* closure with those upvalues */
		self.SetField(-(nup + 2), name)
	}
	self.Pop(nup) /* remove upvalues */
}

// int类型错误
func (self *luaState) intError(arg int) {
	if self.IsNumber(arg) {
		self.ArgError(arg, "number has no integer representation")
	} else {
		self.tagError(arg, LUA_TNUMBER)
	}
}

func (self *luaState) tagError(arg int, tag LuaType) {
	self.typeError(arg, self.TypeName(LuaType(tag)))
}

// 类型不匹配错误
func (self *luaState) typeError(arg int, tname string) int {
	var typeArg string                                   /* name for the type of the actual argument */
	if self.GetMetafield(arg, "__name") == LUA_TSTRING { // 查找元表中有没有对类型名称的定义
		typeArg = self.ToString(-1) /* use the given type name */
	} else if self.Type(arg) == LUA_TLIGHTUSERDATA { // 判断是否是lightuserdata
		typeArg = "light userdata" /* special name for messages */
	} else {
		typeArg = self.TypeName2(arg) //其他标准类型都可以直接转换
	}
	msg := tname + " expected, got " + typeArg
	self.PushString(msg)
	return self.ArgError(arg, msg)
}
