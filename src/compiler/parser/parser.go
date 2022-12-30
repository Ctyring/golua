package parser

import (
	"lua/src/compiler/ast"
	. "lua/src/compiler/lexer"
)

func Parse(chunk, chunkName string) *ast.Block {
	l := NewLexer(chunk, chunkName)
	block := parseBlock(l)
	l.NextTokenOfKind(TOKEN_EOF)
	return block
}
