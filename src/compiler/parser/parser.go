package parser

import (
	"lua/src/compiler/ast"
	lexer2 "lua/src/compiler/lexer"
)

func Parse(chunk, chunkName string) *ast.Block {
	l := lexer2.NewLexer(chunk, chunkName)
	block := parseBlock(l)
	l.NextTokenOfKind(lexer2.TOKEN_EOF)
	return block
}
