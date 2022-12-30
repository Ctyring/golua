package parser

import (
	. "lua/src/compiler/ast"
	. "lua/src/compiler/lexer"
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
func parsePrefixExp(l *Lexer) Exp {
	var exp Exp
	if l.LookAhead() == TOKEN_IDENTIFIER { // 先前瞻一个token看是不是标识符
		line, name := l.NextIdentifier() // Name
		exp = &NameExp{line, name}
	} else { // ‘(’ exp ‘)’
		exp = parseParensExp(l) // 圆括号表达式
	}
	return _finishPrefixExp(l, exp)
}

func parseParensExp(l *Lexer) Exp {
	l.NextTokenOfKind(TOKEN_SEP_LPAREN) // (
	exp := parseExp(l)                  // exp
	l.NextTokenOfKind(TOKEN_SEP_RPAREN) // )

	switch exp.(type) {
	// 只有这四种情况需要保留圆括号，因为圆括号会改变语义
	case *VarargExp, *FuncCallExp, *NameExp, *TableAccessExp:
		return &ParensExp{exp}
	}

	// no need to keep parens
	return exp
}

func _finishPrefixExp(l *Lexer, exp Exp) Exp {
	for {
		switch l.LookAhead() {
		case TOKEN_SEP_LBRACK: // prefixexp ‘[’ exp ‘]’
			l.NextToken()                       // ‘[’
			keyExp := parseExp(l)               // exp
			l.NextTokenOfKind(TOKEN_SEP_RBRACK) // ‘]’
			exp = &TableAccessExp{l.Line(), exp, keyExp}
		case TOKEN_SEP_DOT: // prefixexp ‘.’ Name
			l.NextToken()                    // ‘.’
			line, name := l.NextIdentifier() // Name
			keyExp := &StringExp{line, name}
			exp = &TableAccessExp{line, exp, keyExp}
		case TOKEN_SEP_COLON, // prefixexp ‘:’ Name args
			TOKEN_SEP_LPAREN, TOKEN_SEP_LCURLY, TOKEN_STRING: // prefixexp args
			exp = _finishFuncCallExp(l, exp)
		default:
			return exp
		}
	}
	return exp
}

// functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
func _finishFuncCallExp(lexer *Lexer, prefixExp Exp) *FuncCallExp {
	nameExp := _parseNameExp(lexer)
	line := lexer.Line() // todo
	args := _parseArgs(lexer)
	lastLine := lexer.Line()
	return &FuncCallExp{line, lastLine, prefixExp, nameExp, args}
}

func _parseNameExp(lexer *Lexer) *StringExp {
	if lexer.LookAhead() == TOKEN_SEP_COLON {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		return &StringExp{line, name}
	}
	return nil
}

// args ::=  ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
func _parseArgs(l *Lexer) (args []Exp) {
	switch l.LookAhead() {
	case TOKEN_SEP_LPAREN: // ‘(’ [explist] ‘)’
		l.NextToken() // TOKEN_SEP_LPAREN
		if l.LookAhead() != TOKEN_SEP_RPAREN {
			args = parseExpList(l)
		}
		l.NextTokenOfKind(TOKEN_SEP_RPAREN)
	case TOKEN_SEP_LCURLY: // ‘{’ [fieldlist] ‘}’
		args = []Exp{parseTableConstructorExp(l)}
	default: // LiteralString
		line, str := l.NextTokenOfKind(TOKEN_STRING)
		args = []Exp{&StringExp{line, str}}
	}
	return
}
