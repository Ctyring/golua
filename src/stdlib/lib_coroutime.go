package stdlib

import (
	. "lua/src/api"
)

var coFuncs = map[string]GoFunction{
	"create":      coCreate,    // 创建一个协程
	"resume":      coResume,    // 恢复一个协程
	"yield":       coYield,     // 挂起一个协程
	"status":      coStatus,    // 获取协程状态
	"isyieldable": coYieldable, // 判断协程是否可挂起
	"running":     coRunning,   // 获取当前协程
	"wrap":        coWrap,      // 创建一个协程包装器
}

func OpenCoroutineLib(ls LuaState) int {
	ls.NewLib(coFuncs)
	return 1
}

// coroutine.create (f)
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.create
// lua-5.3.4/src/lcorolib.c#luaB_cocreate()
// 创建协程
func coCreate(ls LuaState) int {
	ls.CheckType(1, LUA_TFUNCTION) // 检查类型
	ls2 := ls.NewThread()          // 创建新的协程
	ls.PushValue(1)                // 把函数压入栈(这个函数就是新协程的主函数)
	ls.XMove(ls2, 1)               // 移动给新协程
	return 1
}

// coroutine.resume (co [, val1, ···])
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.resume
// lua-5.3.4/src/lcorolib.c#luaB_coresume()
// 恢复协程
func coResume(ls LuaState) int {
	// 转换参数并获取协程
	co := ls.ToThread(1)
	ls.ArgCheck(co != nil, 1, "thread expected")

	// 调用辅助函数恢复协程
	if r := _auxResume(ls, co, ls.GetTop()-1); r < 0 {
		ls.PushBoolean(false)
		ls.Insert(-2)
		return 2 /* return false + error message */
	} else {
		ls.PushBoolean(true)
		ls.Insert(-(r + 1))
		return r + 1 /* return true + 'resume' returns */
	}
}

// 恢复协程的辅助函数
func _auxResume(ls, co LuaState, narg int) int {
	if !ls.CheckStack(narg) {
		ls.PushString("too many arguments to resume")
		return -1 /* error flag */
	}
	if co.Status() == LUA_OK && co.GetTop() == 0 { // 协程已经结束
		ls.PushString("cannot resume dead coroutine")
		return -1 /* error flag */
	}
	ls.XMove(co, narg)            // 移动参数给协程
	status := co.Resume(ls, narg) // 调用api恢复协程
	// 等待协程返回后，判断状态
	if status == LUA_OK || status == LUA_YIELD { // 协程正常结束或者挂起
		nres := co.GetTop()           // 获取返回值个数
		if !ls.CheckStack(nres + 1) { // 检查栈空间
			co.Pop(nres) /* remove results anyway */
			ls.PushString("too many results to resume")
			return -1 /* error flag */
		}
		co.XMove(ls, nres) // 移动返回值给主协程
		return nres
	} else { // 如果失败会提供报错信息
		co.XMove(ls, 1) /* move error message */
		return -1       /* error flag */
	}
}

// coroutine.yield (···)
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.yield
// lua-5.3.4/src/lcorolib.c#luaB_yield()
// 挂起协程
func coYield(ls LuaState) int {
	return ls.Yield(ls.GetTop()) // 调用api挂起协程，将栈中的所有参数都返回给主协程
}

// coroutine.status (co)
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.status
// lua-5.3.4/src/lcorolib.c#luaB_costatus()
// 获取协程状态
func coStatus(ls LuaState) int {
	co := ls.ToThread(1)
	ls.ArgCheck(co != nil, 1, "thread expected")
	if ls == co {
		ls.PushString("running")
	} else {
		switch co.Status() {
		case LUA_YIELD:
			ls.PushString("suspended")
		case LUA_OK:
			if co.GetStack() { /* does it have frames? */
				ls.PushString("normal") /* it is running */
			} else if co.GetTop() == 0 {
				ls.PushString("dead")
			} else {
				ls.PushString("suspended")
			}
		default: /* some error occurred */
			ls.PushString("dead")
		}
	}

	return 1
}

// coroutine.isyieldable ()
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.isyieldable
// 判断协程是否能挂起，只有mainthread不能挂起
func coYieldable(ls LuaState) int {
	ls.PushBoolean(ls.IsYieldable())
	return 1
}

// coroutine.running ()
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.running
// 返回正在运行的协程，并把是否是主线程压入栈
func coRunning(ls LuaState) int {
	isMain := ls.PushThread()
	ls.PushBoolean(isMain)
	return 2
}

// coroutine.wrap (f)
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.wrap
// 创建一个协程，每次调用f会恢复协程，f结束协程结束
func coWrap(ls LuaState) int {
	coCreate(ls)
	ls.PushGoClosure(_auxWrap, 1)
	return 1
}

// 创建协程包装的辅助函数
func _auxWrap(ls LuaState) int {
	co := ls.ToThread(LUA_REGISTRYINDEX - 1)
	r := _auxResume(ls, co, ls.GetTop())
	if r < 0 {
		if ls.IsString(-1) { /* error object is a string? */
			ls.PushString("error in wrapped function: " + ls.ToString(-1))
		}
		return ls.Error()
	}
	return r
}
