package binchunk

import (
	"encoding/binary"
	"math"
)

// 完成二进制chunk解析工作
type reader struct {
	data []byte
}

// 检查头部
func (self *reader) checkHeader() {
	if string(self.readBytes(4)) != LUA_SIGNATURE {
		panic("not a precompiled chunk!")
	} else if self.readByte() != LUAC_VERSION {
		panic("version mismatch!")
	} else if self.readByte() != LUAC_FORMAT {
		panic("format mismatch!")
	} else if string(self.readBytes(6)) != LUAC_DATA {
		panic("corrupted!")
	} else if self.readByte() != CINT_SIZE {
		panic("int size mismatch!")
	} else if self.readByte() != CSIZET_SIZE {
		panic("size_t size mismatch!")
	} else if self.readByte() != INSTRUCTION_SIZE {
		panic("instruction size mismatch!")
	} else if self.readByte() != LUA_INTEGER_SIZE {
		panic("lua_Integer size mismatch!")
	} else if self.readByte() != LUA_NUMBER_SIZE {
		panic("lua_Number size mismatch!")
	} else if self.readLuaInteger() != LUAC_INT {
		panic("endianness mismatch!")
	} else if self.readLuaNumber() != LUAC_NUM {
		panic("float format mismatch!")
	}
}

// 读取函数原型
func (self *reader) readProto(parentSource string) *Prototype {
	source := self.readString()
	if source == "" {
		source = parentSource
	}
	return &Prototype{
		Source:          source,
		LineDefined:     self.readUint32(),
		LastLineDefined: self.readUint32(),
		NumParams:       self.readByte(),
		IsVararg:        self.readByte(),
		MaxStackSize:    self.readByte(),
		Code:            self.readCode(),
		Constants:       self.readConstants(),
		Upvalues:        self.readUpvalues(),
		Protos:          self.readProtos(source),
		LineInfo:        self.readLineInfo(),
		LocVars:         self.readLocVars(),
		UpvalueNames:    self.readUpvalueNames(),
	}
}

// 读取基本数据类型
// 读取基本数据类型的方法一共有七种，其他方法通过调用这7种方法来从二进制chunk里提取数据

// 读取一个字节
func (self *reader) readByte() byte {
	b := self.data[0]
	self.data = self.data[1:]
	return b
}

// 读取n个字节
func (self *reader) readBytes(n uint) []byte {
	bytes := self.data[:n]
	self.data = self.data[n:]
	return bytes
}

// 用小端的方式从字节流读取一个cint存储类型(在go中对应uint32)
func (self *reader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(self.data)
	self.data = self.data[4:]
	return i
}

// 用小端的方式从字节流读取一个size_t存储类型(在go中对应uint64)
func (self *reader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(self.data)
	self.data = self.data[8:]
	return i
}

// 读取Lua整数
func (self *reader) readLuaInteger() int64 {
	return int64(self.readUint64())
}

// 读取Lua浮点数
func (self *reader) readLuaNumber() float64 {
	return math.Float64frombits(self.readUint64())
}

// 读取Lua字符串
func (self *reader) readString() string {
	size := uint(self.readByte())
	if size == 0 {
		return ""
	}
	if size == 0xFF {
		size = uint(self.readUint64())
	}
	bytes := self.readBytes(size - 1)
	return string(bytes)
}

// 读取指令表
func (self *reader) readCode() []uint32 {
	size := self.readUint32()
	code := make([]uint32, size)
	for i := range code {
		code[i] = self.readUint32()
	}
	return code
}

// 读取一个常量
func (self *reader) readConstant() interface{} {
	switch self.readByte() {
	case TAG_NIL:
		return nil
	case TAG_BOOLEAN:
		return self.readByte() != 0
	case TAG_NUMBER:
		return self.readLuaNumber()
	case TAG_INTEGER:
		return self.readLuaInteger()
	case TAG_SHORT_STR:
		return self.readString()
	case TAG_LONG_STR:
		return self.readString()
	default:
		panic("corrupted!")
	}
}

// 读取常量表
func (self *reader) readConstants() []interface{} {
	size := self.readUint32()
	k := make([]interface{}, size)
	for i := range k {
		k[i] = self.readConstant()
	}
	return k
}

// 读取Upvalue表
func (self *reader) readUpvalues() []Upvalue {
	size := self.readUint32()
	upvalues := make([]Upvalue, size)
	for i := range upvalues {
		upvalues[i] = Upvalue{
			Instack: self.readByte(),
			Idx:     self.readByte(),
		}
	}
	return upvalues
}

// 读取子函数原型表
func (self *reader) readProtos(source string) []*Prototype {
	size := self.readUint32()
	p := make([]*Prototype, size)
	for i := range p {
		p[i] = self.readProto(source)
	}
	return p
}

// 读取行号表
func (self *reader) readLineInfo() []uint32 {
	size := self.readUint32()
	lineInfo := make([]uint32, size)
	for i := range lineInfo {
		lineInfo[i] = self.readUint32()
	}
	return lineInfo
}

// 读取局部变量表
func (self *reader) readLocVars() []LocVar {
	size := self.readUint32()
	locVars := make([]LocVar, size)
	for i := range locVars {
		locVars[i] = LocVar{
			VarName: self.readString(),
			StartPC: self.readUint32(),
			EndPC:   self.readUint32(),
		}
	}
	return locVars
}

// 读取Upvalue名表
func (self *reader) readUpvalueNames() []string {
	size := self.readUint32()
	upvalueNames := make([]string, size)
	for i := range upvalueNames {
		upvalueNames[i] = self.readString()
	}
	return upvalueNames
}
