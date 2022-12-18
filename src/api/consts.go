package api

const LUA_MINSTACK = 20
const LUAI_MAXSTACK = 1000000
const LUA_REGISTRYINDEX = -LUAI_MAXSTACK - 1000 // 注册表的"伪"索引
const LUA_RIDX_GLOBALS int64 = 2
const LUA_MULTRET = -1
const LUA_RIDX_MAINTHREAD int64 = 1

const (
	LUA_MAXINTEGER = 1<<63 - 1
	LUA_MININTEGER = -1 << 63
)

const (
	LUA_TNONE = iota - 1 // -1 无效索引
	LUA_TNIL
	LUA_TBOOLEAN
	LUA_TLIGHTUSERDATA
	LUA_TNUMBER
	LUA_TSTRING
	LUA_TTABLE
	LUA_TFUNCTION
	LUA_TUSERDATA
	LUA_TTHREAD
)

const (
	LUA_OPADD  = iota // +
	LUA_OPSUB         // -
	LUA_OPMUL         // *
	LUA_OPMOD         // %
	LUA_OPPOW         // ^
	LUA_OPDIV         // /
	LUA_OPIDIV        // //
	LUA_OPBAND        // &
	LUA_OPBOR         // |
	LUA_OPBXOR        // ~
	LUA_OPSHL         // <<
	LUA_OPSHR         // >>
	LUA_OPUNM         // - (无符号类型减号)
	LUA_OPBNOT        // ~
)

const (
	LUA_OPEQ = iota // ==
	LUA_OPLT        // <
	LUA_OPLE        // <=
)

const (
	LUA_OK = iota // 0
	LUA_YIELD
	LUA_ERRRUN
	LUA_ERRSYNTAX
	LUA_ERRMEM
	LUA_ERRGCMM
	LUA_ERRERR
	LUA_ERRFILE
)
