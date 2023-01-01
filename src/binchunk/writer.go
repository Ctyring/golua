package binchunk

import (
	"encoding/binary"
	"math"
)

type writer struct {
	prototype Prototype
	data      []byte
}

func (self *writer) writeHeader() {
	self.writeBytes([]byte(LUA_SIGNATURE))
	self.writeByte(LUAC_VERSION)
	self.writeByte(LUAC_FORMAT)
	self.writeBytes([]byte(LUAC_DATA))
	self.writeByte(CINT_SIZE)
	self.writeByte(CSIZET_SIZE)
	self.writeByte(INSTRUCTION_SIZE)
	self.writeByte(LUA_INTEGER_SIZE)
	self.writeByte(LUA_NUMBER_SIZE)
	self.writeLuaInteger(LUAC_INT)
	self.writeLuaNumber(LUAC_NUM)
}

func (self *writer) writeProto(proto *Prototype) {
	self.writeString(proto.Source)
	self.writeUint32(proto.LineDefined)
	self.writeUint32(proto.LastLineDefined)
	self.writeByte(proto.NumParams)
	self.writeByte(proto.IsVararg)
	self.writeByte(proto.MaxStackSize)
	self.writeCode(proto.Code)
	self.writeConstants(proto.Constants)
	self.writeUpvalues(proto.Upvalues)
	self.writeProtos(proto.Protos)
	self.writeLineInfo(proto.LineInfo)
	self.writeLocVars(proto.LocVars)
}

func (self *writer) writeByte(b byte) {
	self.data = append(self.data, b)
}

func (self *writer) writeBytes(b []byte) {
	self.data = append(self.data, b...)
}

func (self *writer) writeUint32(i uint32) {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, i)
	self.writeBytes(bytes)
}

func (self *writer) writeUint64(i uint64) {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, i)
	self.writeBytes(bytes)
}

func (self *writer) writeLuaInteger(i int64) {
	self.writeUint64(uint64(i))
}

func (self *writer) writeLuaNumber(n float64) {
	self.writeUint64(math.Float64bits(n))
}

func (self *writer) writeString(s string) {
	self.writeByte(byte(len(s) + 1))
	self.writeBytes([]byte(s))
}

func (self *writer) writeCode(code []uint32) {
	self.writeUint32(uint32(len(code)))
	for _, c := range code {
		self.writeUint32(c)
	}
}

func (self *writer) writeConstants(constants []interface{}) {
	self.writeUint32(uint32(len(constants)))
	for _, c := range constants {
		switch c.(type) {
		case nil:
			self.writeByte(TAG_NIL)
		case bool:
			self.writeByte(TAG_BOOLEAN)
			if c.(bool) {
				self.writeByte(1)
			} else {
				self.writeByte(0)
			}
		case int64:
			self.writeByte(TAG_INTEGER)
			self.writeLuaInteger(c.(int64))
		case float64:
			self.writeByte(TAG_NUMBER)
			self.writeLuaNumber(c.(float64))
		case string:
			self.writeByte(TAG_LONG_STR)
			self.writeString(c.(string))
		}
	}
}

func (self *writer) writeUpvalues(upvalues []Upvalue) {
	self.writeUint32(uint32(len(upvalues)))
	for _, u := range upvalues {
		self.writeString(u.Name)
		self.writeByte(u.Instack)
		self.writeByte(u.Idx)
	}
}

func (self *writer) writeProtos(protos []*Prototype) {
	self.writeUint32(uint32(len(protos)))
	for _, p := range protos {
		self.writeProto(p)
	}
}

func (self *writer) writeLineInfo(lineInfo []uint32) {
	self.writeUint32(uint32(len(lineInfo)))
	for _, line := range lineInfo {
		self.writeUint32(line)
	}
}

func (self *writer) writeLocVars(locVars []LocVar) {
	self.writeUint32(uint32(len(locVars)))
	for _, locVar := range locVars {
		self.writeString(locVar.VarName)
		self.writeUint32(locVar.StartPC)
		self.writeUint32(locVar.EndPC)
	}
}

// 暂时废弃
func (self *writer) writeUpvalueNames(upvalueNames []string) {
	self.writeUint32(uint32(len(upvalueNames)))
	for _, name := range upvalueNames {
		self.writeString(name)
	}
}
