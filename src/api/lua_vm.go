package api

// 扩展LuaState接口
type LuaVM interface {
	LuaState
	PC() int             // 获取当前指令指针
	AddPC(n int)         // 修改指令指针
	Fetch() uint32       // 取出当前指令，将PC指向下一条指令
	GetConst(idx int)    // 将指定常量推入栈顶
	GetRK(rk int)        // 将指定常量或栈值推入栈顶
	RegisterCount() int  // 获取寄存器数量
	LoadVararg(n int)    // 将可变参数推入栈顶
	LoadProto(idx int)   // 将指定子函数原型推入栈顶
	CloseUpvalues(a int) // 关闭指定索引处的Upvalue
}
