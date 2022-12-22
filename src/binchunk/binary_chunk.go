package binchunk

// 二进制 chunk 定义
type binaryChunk struct {
	header                  // 头部
	sizeUpvalues byte       // 主函数 upvalue数量
	mainFunc     *Prototype // 主函数原型
}

// 头部
type header struct {
	signature       [4]byte // 签名(魔数) 0x1B4C7561 是 ESC L u a 的 ASCII 码
	version         byte    // 版本号
	format          byte    // 格式号
	luacData        [6]byte // 编译器信息 0x19 93 13 0A 1A 0A 即 "\x19\x93\r\n\x1a\n"
	cintSize        byte    // int大小
	sizetSize       byte    // size_t大小
	instructionSize byte    // 指令大小
	luaIntegerSize  byte    // lua 整数大小
	luaNumberSize   byte    // lua 浮点数大小
	luacInt         int64   // 一个整数，用于测试字节序 0x5678
	luacNum         float64 // 一个浮点数，检测浮点数格式 370.5
}

// 默认值
const (
	LUA_SIGNATURE    = "\x1bLua"            // 二进制块签名
	LUAC_VERSION     = 0x53                 // 版本号
	LUAC_FORMAT      = 0                    // 格式号
	LUAC_DATA        = "\x19\x93\r\n\x1a\n" // 编译器信息
	CINT_SIZE        = 4                    // int大小
	CSIZET_SIZE      = 8                    // size_t大小
	INSTRUCTION_SIZE = 4                    // 指令大小
	LUA_INTEGER_SIZE = 8                    // lua 整数大小
	LUA_NUMBER_SIZE  = 8                    // lua 浮点数大小
	LUAC_INT         = 0x5678               // 一个整数，用于测试字节序
	LUAC_NUM         = 370.5                // 一个浮点数，检测浮点数格式
)

// 原型
type Prototype struct {
	Source          string        // 源文件名
	LineDefined     uint32        // 起始行号
	LastLineDefined uint32        // 最后行号
	NumParams       byte          // 固定参数个数
	IsVararg        byte          // 是否有变长参数
	MaxStackSize    byte          // 寄存器数量
	Code            []uint32      // 指令表，每条指令占四个字节
	Constants       []interface{} // 常量表，用于存放Lua代码里出现的字面量，每个常量都以1字节tag开头，表示常量类型
	Upvalues        []Upvalue     // upvalue表，每个元素占用两个字节
	Protos          []*Prototype  // 子函数原型表
	LineInfo        []uint32      // 行号表，行号表和指令表一一对应，记录了每条指令对应的源代码行号
	LocVars         []LocVar      // 局部变量表
	UpvalueNames    []string      // upvalue名列表，和前面的Upvalue表一一对应，记录每个Upvalue在源代码中的名字
}

// go语言中的空接口可以等效c语言中的union的效果
const (
	TAG_NIL       = 0x00 // nil
	TAG_BOOLEAN   = 0x01 // 布尔值
	TAG_NUMBER    = 0x03 // 浮点数
	TAG_INTEGER   = 0x13 // 整数
	TAG_SHORT_STR = 0x04 // 短字符串
	TAG_LONG_STR  = 0x14 // 长字符串
)

type Upvalue struct {
	Instack byte // 是否在寄存器(栈)中
	Idx     byte // 寄存器索引或upvalue索引
}

type LocVar struct {
	VarName string // 变量名
	StartPC uint32 // 起始指令索引
	EndPC   uint32 // 结束指令索引
}

// 解析二进制chunk
func Undump(data []byte) *Prototype {
	reader := &reader{data}
	reader.checkHeader()        // 检查头部
	reader.readByte()           // 跳过upvalue数量
	return reader.readProto("") // 读取主函数原型
}

func IsBinaryChunk(data []byte) bool {
	return len(data) > 4 && string(data[:4]) == LUA_SIGNATURE
}
