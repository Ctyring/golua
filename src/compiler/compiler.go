package compiler

import (
	"lua/src/binchunk"
	"lua/src/compiler/codegen"
	"lua/src/compiler/parser"
)

func Compile(chunk, chunkname string) *binchunk.Prototype {
	ast := parser.Parse(chunk, chunkname)
	return codegen.GenProto(ast)
}
