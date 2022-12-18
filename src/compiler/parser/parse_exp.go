package parser

import (
	"lua/src/compiler/ast"
	lexer2 "lua/src/compiler/lexer"
	"lua/src/number"
)

// 解析返回值序列
func parseExpList(l *lexer2.Lexer) []ast.Exp {
	// 创建一个切片
	exps := make([]ast.Exp, 0, 4)
	// 解析第一个表达式并添加到切片中
	exps = append(exps, parseExp(l))
	for l.LookAhead() == lexer2.TOKEN_SEP_COMMA { // 如果下一个token是逗号，跳过逗号继续解析
		l.NextToken() // skip `,`
		exps = append(exps, parseExp(l))
	}
	return exps
}

// 运算符分为12个优先级，所以需要编写十二个函数

/*
exp ::=  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
	 prefixexp | tableconstructor | exp binop exp | unop exp
*/
/*
exp   ::= exp12
exp12 ::= exp11 {or exp11}
exp11 ::= exp10 {and exp10}
exp10 ::= exp9 {(‘<’ | ‘>’ | ‘<=’ | ‘>=’ | ‘~=’ | ‘==’) exp9}
exp9  ::= exp8 {‘|’ exp8}
exp8  ::= exp7 {‘~’ exp7}
exp7  ::= exp6 {‘&’ exp6}
exp6  ::= exp5 {(‘<<’ | ‘>>’) exp5}
exp5  ::= exp4 {‘..’ exp4}
exp4  ::= exp3 {(‘+’ | ‘-’) exp3}
exp3  ::= exp2 {(‘*’ | ‘/’ | ‘//’ | ‘%’) exp2}
exp2  ::= {(‘not’ | ‘#’ | ‘-’ | ‘~’)} exp1
exp1  ::= exp0 {‘^’ exp2}
exp0  ::= nil | false | true | Numeral | LiteralString
		| ‘...’ | functiondef | prefixexp | tableconstructor
*/
func parseExp(l *lexer2.Lexer) ast.Exp {
	return parseExp12(l)
}

// x or y
func parseExp12(l *lexer2.Lexer) ast.Exp {
	// 先解析更高优先级的运算符表达式
	exp := parseExp11(l)
	for l.LookAhead() == lexer2.TOKEN_OP_OR { // 左结合，直接for遍历
		line, op, _ := l.NextToken()
		lor := &ast.BinopExp{line, op, exp, parseExp11(l)}
		exp = optimizeLogicalOr(lor)
	}
	return exp
}

// x and y
func parseExp11(l *lexer2.Lexer) ast.Exp {
	exp := parseExp10(l)
	for l.LookAhead() == lexer2.TOKEN_OP_AND {
		line, op, _ := l.NextToken()
		land := &ast.BinopExp{line, op, exp, parseExp10(l)}
		exp = optimizeLogicalAnd(land)
	}
	return exp
}

// compare
func parseExp10(l *lexer2.Lexer) ast.Exp {
	exp := parseExp9(l)
	for {
		switch l.LookAhead() {
		case lexer2.TOKEN_OP_LT, lexer2.TOKEN_OP_GT, lexer2.TOKEN_OP_NE,
			lexer2.TOKEN_OP_LE, lexer2.TOKEN_OP_GE, lexer2.TOKEN_OP_EQ:
			line, op, _ := l.NextToken()
			exp = &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp9(l)}
		default:
			return exp
		}
	}
	return exp
}

// x | y
func parseExp9(l *lexer2.Lexer) ast.Exp {
	exp := parseExp8(l)
	for l.LookAhead() == lexer2.TOKEN_OP_BOR {
		line, op, _ := l.NextToken()
		bor := &ast.BinopExp{line, op, exp, parseExp8(l)}
		exp = optimizeBitwiseBinaryOp(bor)
	}
	return exp
}

// x ~ y
func parseExp8(l *lexer2.Lexer) ast.Exp {
	exp := parseExp7(l)
	for l.LookAhead() == lexer2.TOKEN_OP_BXOR {
		line, op, _ := l.NextToken()
		bxor := &ast.BinopExp{line, op, exp, parseExp7(l)}
		exp = optimizeBitwiseBinaryOp(bxor)
	}
	return exp
}

// x & y
func parseExp7(l *lexer2.Lexer) ast.Exp {
	exp := parseExp6(l)
	for l.LookAhead() == lexer2.TOKEN_OP_BAND {
		line, op, _ := l.NextToken()
		band := &ast.BinopExp{line, op, exp, parseExp6(l)}
		exp = optimizeBitwiseBinaryOp(band)
	}
	return exp
}

// shift
func parseExp6(l *lexer2.Lexer) ast.Exp {
	exp := parseExp5(l)
	for {
		switch l.LookAhead() {
		case lexer2.TOKEN_OP_SHL, lexer2.TOKEN_OP_SHR:
			line, op, _ := l.NextToken()
			shx := &ast.BinopExp{line, op, exp, parseExp5(l)}
			exp = optimizeBitwiseBinaryOp(shx)
		default:
			return exp
		}
	}
	return exp
}

// a .. b
func parseExp5(l *lexer2.Lexer) ast.Exp {
	exp := parseExp4(l)
	if l.LookAhead() != lexer2.TOKEN_OP_CONCAT {
		return exp
	}

	line := 0
	exps := []ast.Exp{exp}
	for l.LookAhead() == lexer2.TOKEN_OP_CONCAT {
		line, _, _ = l.NextToken()
		exps = append(exps, parseExp4(l))
	}
	return &ast.ConcatExp{line, exps}
}

// x +/- y
func parseExp4(l *lexer2.Lexer) ast.Exp {
	exp := parseExp3(l)
	for {
		switch l.LookAhead() {
		case lexer2.TOKEN_OP_ADD, lexer2.TOKEN_OP_SUB:
			line, op, _ := l.NextToken()
			arith := &ast.BinopExp{line, op, exp, parseExp3(l)}
			exp = optimizeArithBinaryOp(arith)
		default:
			return exp
		}
	}
	return exp
}

// *, %, /, //
func parseExp3(l *lexer2.Lexer) ast.Exp {
	exp := parseExp2(l)
	for {
		switch l.LookAhead() {
		case lexer2.TOKEN_OP_MUL, lexer2.TOKEN_OP_MOD, lexer2.TOKEN_OP_DIV, lexer2.TOKEN_OP_IDIV:
			line, op, _ := l.NextToken()
			arith := &ast.BinopExp{line, op, exp, parseExp2(l)}
			exp = optimizeArithBinaryOp(arith)
		default:
			return exp
		}
	}
	return exp
}

// unary
func parseExp2(l *lexer2.Lexer) ast.Exp {
	switch l.LookAhead() {
	case lexer2.TOKEN_OP_UNM, lexer2.TOKEN_OP_BNOT, lexer2.TOKEN_OP_LEN, lexer2.TOKEN_OP_NOT:
		line, op, _ := l.NextToken()
		exp := &ast.UnopExp{line, op, parseExp2(l)}
		return optimizeUnaryOp(exp)
	}
	return parseExp1(l) // 递归调用实现右结合性
}

// x ^ y
func parseExp1(l *lexer2.Lexer) ast.Exp { // pow is right associative
	exp := parseExp0(l)
	if l.LookAhead() == lexer2.TOKEN_OP_POW { // 乘方具有右结合性，需要递归调用自己解析后面的乘方运算符表达式(这里使用if)
		line, op, _ := l.NextToken()
		exp = &ast.BinopExp{line, op, exp, parseExp2(l)}
	}
	return optimizePow(exp)
}

func parseExp0(l *lexer2.Lexer) ast.Exp {
	switch l.LookAhead() {
	case lexer2.TOKEN_VARARG: // ...
		line, _, _ := l.NextToken()
		return &ast.VarargExp{line}
	case lexer2.TOKEN_KW_NIL: // nil
		line, _, _ := l.NextToken()
		return &ast.NilExp{line}
	case lexer2.TOKEN_KW_TRUE: // true
		line, _, _ := l.NextToken()
		return &ast.TrueExp{line}
	case lexer2.TOKEN_KW_FALSE: // false
		line, _, _ := l.NextToken()
		return &ast.FalseExp{line}
	case lexer2.TOKEN_STRING: // LiteralString
		line, _, token := l.NextToken()
		return &ast.StringExp{line, token}
	case lexer2.TOKEN_NUMBER: // Numeral
		return parseNumberExp(l)
	case lexer2.TOKEN_SEP_LCURLY: // tableconstructor
		return parseTableConstructorExp(l)
	case lexer2.TOKEN_KW_FUNCTION: // functiondef
		l.NextToken()
		return parseFuncDefExp(l)
	default: // prefixexp
		return parsePrefixExp(l)
	}
}

func parseNumberExp(l *lexer2.Lexer) ast.Exp {
	line, _, token := l.NextToken()
	if i, ok := number.ParseInteger(token); ok {
		return &ast.IntegerExp{line, i}
	} else if f, ok := number.ParseFloat(token); ok {
		return &ast.FloatExp{line, f}
	} else { // todo
		panic("not a number: " + token)
	}
}

// functiondef ::= function funcbody
// funcbody ::= ‘(’ [parlist] ‘)’ block end
func parseFuncDefExp(l *lexer2.Lexer) *ast.FuncDefExp {
	line := l.Line()                                      // function
	l.NextTokenOfKind(lexer2.TOKEN_SEP_LPAREN)            // (
	parList, isVararg := _parseParList(l)                 // [parlist]
	l.NextTokenOfKind(lexer2.TOKEN_SEP_RPAREN)            // )
	block := parseBlock(l)                                // block
	lastLine, _ := l.NextTokenOfKind(lexer2.TOKEN_KW_END) // end
	return &ast.FuncDefExp{line, lastLine, parList, isVararg, block}
}

// [parlist]
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
func _parseParList(l *lexer2.Lexer) (names []string, isVararg bool) {
	switch l.LookAhead() { //前瞻
	case lexer2.TOKEN_SEP_RPAREN: // ) 无参数
		return nil, false
	case lexer2.TOKEN_VARARG: // ... 变长参数且无固定参数
		l.NextToken()
		return nil, true
	}

	_, name := l.NextIdentifier()
	names = append(names, name)
	for l.LookAhead() == lexer2.TOKEN_SEP_COMMA {
		l.NextToken()
		if l.LookAhead() == lexer2.TOKEN_IDENTIFIER {
			_, name := l.NextIdentifier()
			names = append(names, name)
		} else {
			l.NextTokenOfKind(lexer2.TOKEN_VARARG)
			isVararg = true
			break
		}
	}
	return
}

// tableconstructor ::= ‘{’ [fieldlist] ‘}’
func parseTableConstructorExp(l *lexer2.Lexer) *ast.TableConstructorExp {
	line := l.Line()
	l.NextTokenOfKind(lexer2.TOKEN_SEP_LCURLY) // {
	keyExps, valExps := _parseFieldList(l)     // [fieldlist]
	l.NextTokenOfKind(lexer2.TOKEN_SEP_RCURLY) // }
	lastLine := l.Line()
	return &ast.TableConstructorExp{line, lastLine, keyExps, valExps}
}

// fieldlist ::= field {fieldsep field} [fieldsep]
// 解析字段列表
func _parseFieldList(l *lexer2.Lexer) (ks, vs []ast.Exp) {
	if l.LookAhead() != lexer2.TOKEN_SEP_RCURLY {
		k, v := _parseField(l) // 解析字段
		ks = append(ks, k)
		vs = append(vs, v)

		for _isFieldSep(l.LookAhead()) {
			l.NextToken()
			if l.LookAhead() != lexer2.TOKEN_SEP_RCURLY {
				k, v := _parseField(l)
				ks = append(ks, k)
				vs = append(vs, v)
			} else {
				break
			}
		}
	}
	return
}

// fieldsep ::= ‘,’ | ‘;’
func _isFieldSep(tokenKind int) bool {
	return tokenKind == lexer2.TOKEN_SEP_COMMA || tokenKind == lexer2.TOKEN_SEP_SEMI
}

// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
// 解析字段
func _parseField(l *lexer2.Lexer) (k, v ast.Exp) {
	if l.LookAhead() == lexer2.TOKEN_SEP_LBRACK { // [exp] = exp
		l.NextToken()                              // [
		k = parseExp(l)                            // exp
		l.NextTokenOfKind(lexer2.TOKEN_SEP_RBRACK) // ]
		l.NextTokenOfKind(lexer2.TOKEN_OP_ASSIGN)  // =
		v = parseExp(l)                            // exp
		return
	}

	// Name = exp
	exp := parseExp(l)
	if nameExp, ok := exp.(*ast.NameExp); ok {
		if l.LookAhead() == lexer2.TOKEN_OP_ASSIGN {
			// Name ‘=’ exp => ‘[’ LiteralString ‘]’ = exp
			l.NextToken()
			k = &ast.StringExp{nameExp.Line, nameExp.Name}
			v = parseExp(l)
			return
		}
	}

	return nil, exp
}
