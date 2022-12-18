package parser

import (
	ast2 "lua/src/compiler/ast"
	lexer2 "lua/src/compiler/lexer"
)

// 前瞻一个token，根据类型调用相应的函数解析语句
func parseStat(l *lexer2.Lexer) ast2.Stat {
	switch l.LookAhead() {
	case lexer2.TOKEN_SEP_SEMI:
		return parseEmptyStat(l)
	case lexer2.TOKEN_KW_BREAK:
		return parseBreakStat(l)
	case lexer2.TOKEN_SEP_LABEL:
		return parseLabelStat(l)
	case lexer2.TOKEN_KW_GOTO:
		return parseGotoStat(l)
	case lexer2.TOKEN_KW_DO:
		return parseDoStat(l)
	case lexer2.TOKEN_KW_WHILE:
		return parseWhileStat(l)
	case lexer2.TOKEN_KW_REPEAT:
		return parseRepeatStat(l)
	case lexer2.TOKEN_KW_IF:
		return parseIfStat(l)
	case lexer2.TOKEN_KW_FOR:
		return parseForStat(l)
	case lexer2.TOKEN_KW_FUNCTION:
		return parseFuncDefStat(l)
	case lexer2.TOKEN_KW_LOCAL:
		return parseLocalAssignOrFuncDefStat(l)
	default:
		return parseAssignOrFuncCallStat(l)
	}
}

// 空语句：分号 跳过
func parseEmptyStat(l *lexer2.Lexer) *ast2.EmptyStat {
	l.NextTokenOfKind(lexer2.TOKEN_SEP_SEMI) // skip `;`
	return &ast2.EmptyStat{}
}

// break语句 记录行号
func parseBreakStat(l *lexer2.Lexer) *ast2.BreakStat {
	l.NextTokenOfKind(lexer2.TOKEN_KW_BREAK) // skip `break`
	return &ast2.BreakStat{Line: l.Line()}
}

// label语句 跳过分隔符并记录标签名
func parseLabelStat(l *lexer2.Lexer) *ast2.LabelStat {
	l.NextTokenOfKind(lexer2.TOKEN_SEP_LABEL)             // skip `::`
	_, name := l.NextTokenOfKind(lexer2.TOKEN_IDENTIFIER) // name
	l.NextTokenOfKind(lexer2.TOKEN_SEP_LABEL)             // skip `::`
	return &ast2.LabelStat{Name: name}
}

// goto语句 跳过关键字并记录标签名
func parseGotoStat(l *lexer2.Lexer) *ast2.GotoStat {
	l.NextTokenOfKind(lexer2.TOKEN_KW_GOTO)               // skip `goto`
	_, name := l.NextTokenOfKind(lexer2.TOKEN_IDENTIFIER) // name
	return &ast2.GotoStat{Name: name}
}

// do语句 跳过关键字并解析块
func parseDoStat(l *lexer2.Lexer) *ast2.DoStat {
	l.NextTokenOfKind(lexer2.TOKEN_KW_DO) // skip `do`
	block := parseBlock(l)
	l.NextTokenOfKind(lexer2.TOKEN_KW_END) // skip `end`
	return &ast2.DoStat{Block: block}
}

// while语句 跳过关键字并解析条件和块
func parseWhileStat(l *lexer2.Lexer) *ast2.WhileStat {
	l.NextTokenOfKind(lexer2.TOKEN_KW_WHILE) // skip `while`
	exp := parseExp(l)
	l.NextTokenOfKind(lexer2.TOKEN_KW_DO) // skip `do`
	block := parseBlock(l)
	l.NextTokenOfKind(lexer2.TOKEN_KW_END) // skip `end`
	return &ast2.WhileStat{Exp: exp, Block: block}
}

// repeat语句 跳过关键字并解析块和条件
func parseRepeatStat(l *lexer2.Lexer) *ast2.RepeatStat {
	l.NextTokenOfKind(lexer2.TOKEN_KW_REPEAT) // skip `repeat`
	block := parseBlock(l)
	l.NextTokenOfKind(lexer2.TOKEN_KW_UNTIL) // skip `until`
	exp := parseExp(l)
	return &ast2.RepeatStat{Block: block, Exp: exp}
}

// if语句
func parseIfStat(l *lexer2.Lexer) *ast2.IfStat {
	exps := make([]ast2.Exp, 0, 4)
	blocks := make([]*ast2.Block, 0, 4)

	l.NextTokenOfKind(lexer2.TOKEN_KW_IF)   // skip `if`
	exps = append(exps, parseExp(l))        // exp
	l.NextTokenOfKind(lexer2.TOKEN_KW_THEN) // skip `then`
	blocks = append(blocks, parseBlock(l))  // block

	for l.LookAhead() == lexer2.TOKEN_KW_ELSEIF { // {
		l.NextToken()                           // skip `elseif`
		exps = append(exps, parseExp(l))        // exp
		l.NextTokenOfKind(lexer2.TOKEN_KW_THEN) // skip `then`
		blocks = append(blocks, parseBlock(l))  // block
	}

	if l.LookAhead() == lexer2.TOKEN_KW_ELSE { // {
		l.NextToken()                                // skip `else`
		exps = append(exps, &ast2.TrueExp{l.Line()}) // exp
		blocks = append(blocks, parseBlock(l))       // block
	}

	l.NextTokenOfKind(lexer2.TOKEN_KW_END) // skip `end`
	return &ast2.IfStat{Exps: exps, Blocks: blocks}
}

// for语句
func parseForStat(l *lexer2.Lexer) ast2.Stat {
	lineOfFor, _ := l.NextTokenOfKind(lexer2.TOKEN_KW_FOR) // skip `for`
	_, name := l.NextIdentifier()
	if l.LookAhead() == lexer2.TOKEN_OP_ASSIGN { // 前瞻下一个token 如果是等号，按照数值for循环来解析
		return _finishForNumStat(l, lineOfFor, name)
	} else {
		return _finishForInStat(l, name)
	}
}

// 数值for循环
func _finishForNumStat(l *lexer2.Lexer, lineOfFor int, varName string) *ast2.ForNumStat {
	l.NextTokenOfKind(lexer2.TOKEN_OP_ASSIGN) // skip `=`
	initExp := parseExp(l)
	l.NextTokenOfKind(lexer2.TOKEN_SEP_COMMA) // skip `,`
	limitExp := parseExp(l)

	var stepExp ast2.Exp
	if l.LookAhead() == lexer2.TOKEN_SEP_COMMA { // `,`
		l.NextToken() // skip `,`
		stepExp = parseExp(l)
	} else {
		stepExp = &ast2.IntegerExp{Line: l.Line(), Val: 1} // 默认步长为1
	}

	lineOfDo, _ := l.NextTokenOfKind(lexer2.TOKEN_KW_DO) // skip `do`
	block := parseBlock(l)
	l.NextTokenOfKind(lexer2.TOKEN_KW_END) // skip `end`

	return &ast2.ForNumStat{
		LineOfFor: lineOfFor,
		LineOfDo:  lineOfDo,
		VarName:   varName,
		InitExp:   initExp,
		LimitExp:  limitExp,
		StepExp:   stepExp,
		Block:     block,
	}
}

// 泛型for循环
func _finishForInStat(l *lexer2.Lexer, name0 string) *ast2.ForInStat {
	name := _finishNameList(l, name0)
	l.NextTokenOfKind(lexer2.TOKEN_KW_IN) // skip `in`
	expList := parseExpList(l)
	lineOfDo, _ := l.NextTokenOfKind(lexer2.TOKEN_KW_DO) // skip `do`
	block := parseBlock(l)
	l.NextTokenOfKind(lexer2.TOKEN_KW_END) // skip `end`
	return &ast2.ForInStat{LineOfDo: lineOfDo, NameList: name, ExpList: expList, Block: block}
}

// 解析循环变量名列表
func _finishNameList(l *lexer2.Lexer, name0 string) []string {
	names := []string{name0}
	for l.LookAhead() == lexer2.TOKEN_SEP_COMMA { // `,`
		l.NextToken() // skip `,`
		_, name := l.NextIdentifier()
		names = append(names, name)
	}
	return names
}

// 局部变量声明和局部函数定义
func parseLocalAssignOrFuncDefStat(l *lexer2.Lexer) ast2.Stat {
	l.NextTokenOfKind(lexer2.TOKEN_KW_LOCAL)       // skip `local`
	if l.LookAhead() == lexer2.TOKEN_KW_FUNCTION { // `function`
		return _finishLocalFuncDefStat(l)
	} else {
		return _finishLocalAssignStat(l)
	}
}

// 局部函数定义
func _finishLocalFuncDefStat(l *lexer2.Lexer) *ast2.LocalFuncDefStat {
	l.NextTokenOfKind(lexer2.TOKEN_KW_FUNCTION) // skip `function`
	_, name := l.NextIdentifier()
	fdExp := parseFuncDefExp(l)
	return &ast2.LocalFuncDefStat{Name: name, Exp: fdExp}
}

// 局部变量声明
func _finishLocalAssignStat(l *lexer2.Lexer) *ast2.LocalVarDeclStat {
	_, name0 := l.NextIdentifier()
	names := _finishNameList(l, name0)
	var exps []ast2.Exp = nil
	if l.LookAhead() == lexer2.TOKEN_OP_ASSIGN { // `=`
		l.NextToken() // skip `=`
		exps = parseExpList(l)
	}
	lastLine := l.Line()
	return &ast2.LocalVarDeclStat{LastLine: lastLine, NameList: names, ExpList: exps}
}

// 赋值和函数调用语句
func parseAssignOrFuncCallStat(l *lexer2.Lexer) ast2.Stat {
	// 先解析前缀表达式
	prefixExp := parsePrefixExp(l)
	if fc, ok := prefixExp.(*ast2.FuncCallExp); ok { // 如果解析出来的前缀表达式时是函数调用表达式
		return fc
	} else { // 否则是var表达式
		return parseAssignStat(l, prefixExp)
	}
}

// 解析赋值语句
func parseAssignStat(l *lexer2.Lexer, var0 ast2.Exp) *ast2.AssignStat {
	varList := _finishVarList(l, var0)        // 解析变量列表
	l.NextTokenOfKind(lexer2.TOKEN_OP_ASSIGN) // skip `=`
	expList := parseExpList(l)                // 解析表达式列表
	lastLine := l.Line()
	return &ast2.AssignStat{LastLine: lastLine, VarList: varList, ExpList: expList}
}

// 解析变量列表
func _finishVarList(l *lexer2.Lexer, var0 ast2.Exp) []ast2.Exp {
	vars := []ast2.Exp{_checkVar(l, var0)}        // 检查变量是否合法并添加到变量列表中
	for l.LookAhead() == lexer2.TOKEN_SEP_COMMA { // `,`
		l.NextToken()                          // skip `,`
		exp := parsePrefixExp(l)               // 解析前缀表达式
		vars = append(vars, _checkVar(l, exp)) // 检查变量是否合法并添加到变量列表中
	}
	return vars
}

// 检查是否是变量
// var ::=  Name | prefixexp `[´ exp `]´ | prefixexp `.´ Name
func _checkVar(l *lexer2.Lexer, exp ast2.Exp) ast2.Exp {
	switch exp.(type) {
	case *ast2.NameExp, *ast2.TableAccessExp:
		return exp
	default:
		l.NextTokenOfKind(-1) // 报错
		return nil
	}
}

// 非局部函数定义语句
func parseFuncDefStat(l *lexer2.Lexer) ast2.Stat {
	l.NextTokenOfKind(lexer2.TOKEN_KW_FUNCTION) // skip `function`
	fnExp, hasColon := _finishFuncName(l)       // 解析函数名
	fdExp := parseFuncDefExp(l)                 // 解析函数定义表达式
	if hasColon {                               // 如果函数名是以冒号开头的 `foo:bar()`
		fdExp.ParList = append(fdExp.ParList, "") // 添加一个空的参数
		copy(fdExp.ParList[1:], fdExp.ParList)    // 将参数列表向后移动一位 `foo:bar(a, b, c)` => `foo:bar("", a, b, c)`
		fdExp.ParList[0] = "self"                 // 将第一个参数设置为 `self` `foo:bar(a, b, c)` => `foo:bar("self", a, b, c)`
	}

	// 最终将非局部函数语句转换为赋值语句
	return &ast2.AssignStat{
		LastLine: fdExp.LastLine,
		VarList:  []ast2.Exp{fnExp},
		ExpList:  []ast2.Exp{fdExp},
	}
}

// 解析函数名
func _finishFuncName(l *lexer2.Lexer) (exp ast2.Exp, hasColon bool) {
	line, name := l.NextIdentifier() // 获取下一个标识符
	exp = &ast2.NameExp{Line: line, Name: name}
	for l.LookAhead() == lexer2.TOKEN_SEP_DOT { // 不断取点
		l.NextToken()                    // skip `.`
		line, name := l.NextIdentifier() // 获取下一个标识符
		idx := &ast2.StringExp{Line: line, Str: name}
		exp = &ast2.TableAccessExp{PrefixExp: exp, KeyExp: idx} // 生成表达式 `a.b.c` => `a["b"]["c"]`
	}

	if l.LookAhead() == lexer2.TOKEN_SEP_COLON { // 如果有冒号
		l.NextToken() // skip `:`
		line, name := l.NextIdentifier()
		idx := &ast2.StringExp{Line: line, Str: name}
		exp = &ast2.TableAccessExp{PrefixExp: exp, KeyExp: idx} // 生成表达式 `a:b()` => `a["b"]`
		hasColon = true                                         // 标记函数名是以冒号开头的
	}

	return
}
