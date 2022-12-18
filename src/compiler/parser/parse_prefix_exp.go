package parser

import (
	"lua/src/compiler/ast"
	lexer2 "lua/src/compiler/lexer"
)

// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
// functioncall ::=  prefixexp args | prefixexp ‘:’ Name args

/*
prefixexp ::= Name

	| ‘(’ exp ‘)’
	| prefixexp ‘[’ exp ‘]’
	| prefixexp ‘.’ Name
	| prefixexp [‘:’ Name] args
*/
func parsePrefixExp(l *lexer2.Lexer) ast.Exp {
	var exp ast.Exp
	if l.LookAhead() == lexer2.TOKEN_IDENTIFIER { // 先前瞻一个token看是不是标识符
		line, name := l.NextIdentifier() // Name
		exp = &ast.NameExp{line, name}
	} else { // ‘(’ exp ‘)’
		exp = parseParensExp(l) // 圆括号表达式
	}
	return _finishPrefixExp(l, exp)
}

func parseParensExp(l *lexer2.Lexer) ast.Exp {
	l.NextTokenOfKind(lexer2.TOKEN_SEP_LPAREN) // (
	exp := parseExp(l)                         // exp
	l.NextTokenOfKind(lexer2.TOKEN_SEP_RPAREN) // )

	switch exp.(type) {
	// 只有这四种情况需要保留圆括号，因为圆括号会改变语义
	case *ast.VarargExp, *ast.FuncCallExp, *ast.NameExp, *ast.TableAccessExp:
		return &ast.ParensExp{exp}
	}

	// no need to keep parens
	return exp
}

func _finishPrefixExp(l *lexer2.Lexer, exp ast.Exp) ast.Exp {
	for {
		switch l.LookAhead() {
		case lexer2.TOKEN_SEP_LBRACK: // prefixexp ‘[’ exp ‘]’
			l.NextToken()                              // ‘[’
			keyExp := parseExp(l)                      // exp
			l.NextTokenOfKind(lexer2.TOKEN_SEP_RBRACK) // ‘]’
			exp = &ast.TableAccessExp{l.Line(), exp, keyExp}
		case lexer2.TOKEN_SEP_DOT: // prefixexp ‘.’ Name
			l.NextToken()                    // ‘.’
			line, name := l.NextIdentifier() // Name
			keyExp := &ast.StringExp{line, name}
			exp = &ast.TableAccessExp{line, exp, keyExp}
		case lexer2.TOKEN_SEP_COLON, // prefixexp ‘:’ Name args
			lexer2.TOKEN_SEP_LPAREN, lexer2.TOKEN_SEP_LCURLY, lexer2.TOKEN_STRING: // prefixexp args
			exp = _finishFuncCallExp(l, exp)
		default:
			return exp
		}
	}
	return exp
}

// functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
func _finishFuncCallExp(l *lexer2.Lexer, prefixExp ast.Exp) *ast.FuncCallExp {
	nameExp := _parseNameExp(l)
	line := l.Line() // todo
	args := _parseArgs(l)
	lastLine := l.Line()
	return &ast.FuncCallExp{Line: line, LastLine: lastLine, PrefixExp: prefixExp, NameExp: nameExp, Args: args}
}

func _parseNameExp(l *lexer2.Lexer) *ast.NameExp {
	if l.LookAhead() == lexer2.TOKEN_SEP_COLON {
		l.NextToken()
		line, name := l.NextIdentifier()
		return &ast.NameExp{line, name}
	}
	return nil
}

// args ::=  ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
func _parseArgs(l *lexer2.Lexer) (args []ast.Exp) {
	switch l.LookAhead() {
	case lexer2.TOKEN_SEP_LPAREN: // ‘(’ [explist] ‘)’
		l.NextToken() // TOKEN_SEP_LPAREN
		if l.LookAhead() != lexer2.TOKEN_SEP_RPAREN {
			args = parseExpList(l)
		}
		l.NextTokenOfKind(lexer2.TOKEN_SEP_RPAREN)
	case lexer2.TOKEN_SEP_LCURLY: // ‘{’ [fieldlist] ‘}’
		args = []ast.Exp{parseTableConstructorExp(l)}
	default: // LiteralString
		line, str := l.NextTokenOfKind(lexer2.TOKEN_STRING)
		args = []ast.Exp{&ast.StringExp{line, str}}
	}
	return
}
