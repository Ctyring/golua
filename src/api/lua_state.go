package api

import . "lua/src/binchunk"

type LuaType = int
type ArithOp = int
type CompareOp = int

// Upvalue索引
func LuaUpvalueIndex(i int) int {
	// 注册表索引 - i
	return LUA_REGISTRYINDEX - i
}

type BasicAPI interface {
	GetTop() int                                   // 获取栈顶索引
	AbsIndex(idx int) int                          // 将索引转换成绝对索引
	CheckStack(n int) bool                         // 扩容
	Pop(n int)                                     // 弹出n个值
	Copy(fromIdx, toIdx int)                       // 复制栈中的值到另一个位置
	PushValue(idx int)                             // 把指定索引处的值移动到栈顶
	Replace(idx int)                               // 把栈顶的值移动到指定索引处
	Insert(idx int)                                // 把栈顶的值插入到指定索引处
	Remove(idx int)                                // 删除指定索引处的值
	Rotate(idx, n int)                             // 旋转栈中的值
	SetTop(idx int)                                // 设置栈顶索引
	TypeName(tp LuaType) string                    // 获取类型名称
	Type(idx int) LuaType                          // 获取指定索引处的值的类型
	IsNone(idx int) bool                           // 判断指定索引处的值是否是无效值
	IsNil(idx int) bool                            // 判断指定索引处的值是否是nil
	IsNoneOrNil(idx int) bool                      // 判断指定索引处的值是否是无效值或nil
	IsBoolean(idx int) bool                        // 判断指定索引处的值是否是布尔值
	IsInteger(idx int) bool                        // 判断指定索引处的值是否是整数
	IsNumber(idx int) bool                         // 判断指定索引处的值是否是数字
	IsString(idx int) bool                         // 判断指定索引处的值是否是字符串
	IsTable(idx int) bool                          // 判断指定索引处的值是否是表
	IsThread(idx int) bool                         // 判断指定索引处的值是否是协程
	IsFunction(idx int) bool                       // 判断指定索引处的值是否是函数
	ToBoolean(idx int) bool                        // 将指定索引处的值转换成布尔值
	ToInteger(idx int) int64                       // 将指定索引处的值转换成整数
	ToIntegerX(idx int) (int64, bool)              // 将指定索引处的值转换成整数
	ToNumber(idx int) float64                      // 将指定索引处的值转换成浮点数
	ToNumberX(idx int) (float64, bool)             // 将指定索引处的值转换成浮点数
	ToString(idx int) string                       // 将指定索引处的值转换成字符串
	ToStringX(idx int) (string, bool)              // 将指定索引处的值转换成字符串
	PushNil()                                      // 将nil压入栈顶
	PushBoolean(b bool)                            // 将布尔值压入栈顶
	PushInteger(n int64)                           // 将整数压入栈顶
	PushNumber(n float64)                          // 将浮点数压入栈顶
	PushString(s string)                           // 将字符串压入栈顶
	Arith(op ArithOp)                              // 对栈顶的两个值进行算术运算
	Compare(idx1, idx2 int, op CompareOp) bool     // 比较栈中的两个值
	Len(idx int)                                   // 获取指定索引处的值的长度
	Concat(n int)                                  // 将栈顶的n个值弹出并拼接成一个字符串，再将结果压入栈顶
	NewTable()                                     // 创建一个空表并将其压入栈顶
	CreateTable(nArr, nRec int)                    // 创建一个空表并将其压入栈顶
	GetTable(idx int) LuaType                      // 获取指定索引处的表中指定键的值
	GetField(idx int, k string) LuaType            // 获取指定索引处的表中指定键的值
	GetI(idx int, i int64) LuaType                 // 获取指定索引处的表中指定键的值
	SetTable(idx int)                              // 设置指定索引处的表中指定键的值
	SetField(idx int, k string)                    // 设置指定索引处的表中指定键的值
	SetI(idx int, n int64)                         // 设置指定索引处的表中指定键的值
	Load(chunk []byte, chunkName, mode string) int // 加载一个块
	Call(nArgs, nResults int)                      // 调用一个函数
	PushGoFunction(f GoFunction)                   // 将Go函数压入栈顶
	IsGoFunction(idx int) bool                     // 判断指定索引处的值是否是Go函数
	ToGoFunction(idx int) GoFunction               // 将指定索引处的值转换成Go函数
	PushGlobalTable()                              // 将全局表压入栈顶
	GetGlobal(name string) LuaType                 // 获取全局变量的值
	SetGlobal(name string)                         // 设置全局变量的值
	Register(name string, f GoFunction)            // 注册一个Go函数
	PushGoClosure(f GoFunction, n int)             // 将Go闭包压入栈顶
	GetMetatable(idx int) bool                     // 获取指定索引处的值的元表
	SetMetatable(idx int)                          // 设置指定索引处的值的元表
	NewMetatable(tname string)                     // 创建一个新的元表
	GetMetatableFromRegistry(tname string)         // 从注册表中获取指定名称的元表
	RawLen(idx int) uint                           // 获取指定索引处的值的长度
	RawEqual(idx1, idx2 int) bool                  // 比较栈中的两个值
	RawGet(idx int) LuaType                        // 获取指定索引处的表中指定键的值
	RawGetI(idx int, i int64) LuaType              // 获取指定索引处的表中指定键的值
	RawSet(idx int)                                // 设置指定索引处的表中指定键的值
	RawSetI(idx int, i int64)                      // 设置指定索引处的表中指定键的值
	Next(idx int) bool                             // 将指定索引处的表中的下一个键值对压入栈顶
	Error() int                                    // 将栈顶的值作为错误对象抛出
	PCall(nArgs, nResults, msgh int) int           // 调用一个函数
	StringToNumber(s string) bool                  // 将字符串转换成数字并压入栈顶
	ToPointer(idx int) interface{}                 // 将指定索引处的值转换成指针
	NewThread() LuaState                           // 创建一个协程并将其压入栈顶
	Resume(from LuaState, nArgs int) int           // 恢复一个协程
	Yield(nResults int) int                        // 挂起一个协程
	Status() int                                   // 获取协程的状态
	IsYieldable() bool                             // 判断当前协程是否可以挂起
	ToThread(idx int) LuaState                     // 将指定索引处的值转换成协程
	PushThread() bool                              // 将当前协程压入栈顶
	XMove(to LuaState, n int)                      // 用于在两个协程栈之间移动元素
	GetStack() bool                                // 获取栈帧
	ToProto(idx int) *Prototype                    // 将指定索引处的值转换成原型
	SetUpvalue(idx, n int)                         // 设置指定索引处的闭包的指定upvalue的值
	GetUpvalue(idx, n int)                         // 获取指定索引处的闭包的指定upvalue的值
	NewUserdata(data interface{})                  // 创建一个新的userdata并将其压入栈顶
	ToUserdata(idx int) *interface{}               // 将指定索引处的值转换成userdata
}

type LuaState interface {
	BasicAPI
	AuxLib
}

// Go函数类型
type GoFunction func(LuaState) int
