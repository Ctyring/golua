package compiler

import (
	. "lua/src/binchunk"
	. "lua/src/compiler/codegen"
	. "lua/src/compiler/parser"
)

func Compile(chunk, chunkname string) *Prototype {
	ast := Parse(chunk, chunkname)
	return GenProto(ast)
}
