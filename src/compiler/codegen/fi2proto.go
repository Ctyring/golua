package codegen

import (
	. "lua/src/binchunk"
)

func toProto(fi *funcInfo) *Prototype {
	proto := &Prototype{
		NumParams:    byte(fi.numParams),    // 参数个数
		MaxStackSize: byte(fi.maxRegs),      // 最大栈空间
		Code:         fi.insts,              // 指令表
		Constants:    getConstants(fi),      // 常量表
		Upvalues:     getUpvalues(fi),       // upvalue表
		Protos:       toProtos(fi.subFuncs), // 子函数原型表
		LineInfo:     []uint32{},            // debug info
		LocVars:      []LocVar{},            // debug info
		UpvalueNames: fi.upvalNames,         // debug info
	}

	if proto.MaxStackSize < 2 {
		proto.MaxStackSize = 2 // todo
	}
	if fi.isVararg {
		proto.IsVararg = 1 // todo
	}

	return proto
}

func toProtos(fis []*funcInfo) []*Prototype {
	protos := make([]*Prototype, len(fis))
	for i, fi := range fis {
		protos[i] = toProto(fi)
	}
	return protos
}

func getConstants(fi *funcInfo) []interface{} {
	consts := make([]interface{}, len(fi.constants))
	for k, idx := range fi.constants {
		consts[idx] = k
	}
	return consts
}

func getUpvalues(fi *funcInfo) []Upvalue {
	upvals := make([]Upvalue, len(fi.upvalues))
	for _, uv := range fi.upvalues {
		if uv.locVarSlot >= 0 { // instack
			upvals[uv.index] = Upvalue{1, byte(uv.locVarSlot)}
		} else {
			upvals[uv.index] = Upvalue{0, byte(uv.upvalIndex)}
		}
	}
	return upvals
}
