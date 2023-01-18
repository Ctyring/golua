package parser

import (
	. "lua/src/compiler/ast"
	. "lua/src/compiler/lexer"
)

// 前瞻一个token，根据类型调用相应的函数解析语句
func parseStat(l *Lexer) Stat {
	switch l.LookAhead() {
	case TOKEN_SEP_SEMI:
		return parseEmptyStat(l)
	case TOKEN_KW_BREAK:
		return parseBreakStat(l)
	case TOKEN_SEP_LABEL:
		return parseLabelStat(l)
	case TOKEN_KW_GOTO:
		return parseGotoStat(l)
	case TOKEN_KW_DO:
		return parseDoStat(l)
	case TOKEN_KW_WHILE:
		return parseWhileStat(l)
	case TOKEN_KW_REPEAT:
		return parseRepeatStat(l)
	case TOKEN_KW_IF:
		return parseIfStat(l)
	case TOKEN_KW_FOR:
		return parseForStat(l)
	case TOKEN_KW_FUNCTION:
		return parseFuncDefStat(l)
	case TOKEN_KW_LOCAL:
		return parseLocalAssignOrFuncDefStat(l)
	default:
		return parseAssignOrFuncCallStat(l)
	}
}

// 空语句：分号 跳过
func parseEmptyStat(l *Lexer) *EmptyStat {
	l.NextTokenOfKind(TOKEN_SEP_SEMI) // skip `;`
	return &EmptyStat{}
}

// break语句 记录行号
func parseBreakStat(l *Lexer) *BreakStat {
	l.NextTokenOfKind(TOKEN_KW_BREAK) // skip `break`
	return &BreakStat{Line: l.Line()}
}

// label语句 跳过分隔符并记录标签名
func parseLabelStat(l *Lexer) *LabelStat {
	l.NextTokenOfKind(TOKEN_SEP_LABEL)             // skip `::`
	_, name := l.NextTokenOfKind(TOKEN_IDENTIFIER) // name
	l.NextTokenOfKind(TOKEN_SEP_LABEL)             // skip `::`
	return &LabelStat{Name: name}
}

// goto语句 跳过关键字并记录标签名
func parseGotoStat(l *Lexer) *GotoStat {
	l.NextTokenOfKind(TOKEN_KW_GOTO)               // skip `goto`
	_, name := l.NextTokenOfKind(TOKEN_IDENTIFIER) // name
	return &GotoStat{Name: name}
}

// do语句 跳过关键字并解析块
func parseDoStat(l *Lexer) *DoStat {
	l.NextTokenOfKind(TOKEN_KW_DO) // skip `do`
	block := parseBlock(l)
	l.NextTokenOfKind(TOKEN_KW_END) // skip `end`
	return &DoStat{Block: block}
}

// while语句 跳过关键字并解析条件和块
func parseWhileStat(l *Lexer) *WhileStat {
	l.NextTokenOfKind(TOKEN_KW_WHILE) // skip `while`
	exp := parseExp(l)
	l.NextTokenOfKind(TOKEN_KW_DO) // skip `do`
	block := parseBlock(l)
	l.NextTokenOfKind(TOKEN_KW_END) // skip `end`
	return &WhileStat{Exp: exp, Block: block}
}

// repeat语句 跳过关键字并解析块和条件
func parseRepeatStat(l *Lexer) *RepeatStat {
	l.NextTokenOfKind(TOKEN_KW_REPEAT) // skip `repeat`
	block := parseBlock(l)
	l.NextTokenOfKind(TOKEN_KW_UNTIL) // skip `until`
	exp := parseExp(l)
	return &RepeatStat{Block: block, Exp: exp}
}

// if语句
func parseIfStat(l *Lexer) *IfStat {
	exps := make([]Exp, 0, 4)
	blocks := make([]*Block, 0, 4)

	l.NextTokenOfKind(TOKEN_KW_IF)         // skip `if`
	exps = append(exps, parseExp(l))       // exp
	l.NextTokenOfKind(TOKEN_KW_THEN)       // skip `then`
	blocks = append(blocks, parseBlock(l)) // block

	for l.LookAhead() == TOKEN_KW_ELSEIF { // {
		l.NextToken()                          // skip `elseif`
		exps = append(exps, parseExp(l))       // exp
		l.NextTokenOfKind(TOKEN_KW_THEN)       // skip `then`
		blocks = append(blocks, parseBlock(l)) // block
	}

	if l.LookAhead() == TOKEN_KW_ELSE { // {
		l.NextToken()                           // skip `else`
		exps = append(exps, &TrueExp{l.Line()}) // exp
		blocks = append(blocks, parseBlock(l))  // block
	}

	l.NextTokenOfKind(TOKEN_KW_END) // skip `end`
	return &IfStat{Exps: exps, Blocks: blocks}
}

// for语句
func parseForStat(l *Lexer) Stat {
	lineOfFor, _ := l.NextTokenOfKind(TOKEN_KW_FOR) // skip `for`
	_, name := l.NextIdentifier()
	if l.LookAhead() == TOKEN_OP_ASSIGN { // 前瞻下一个token 如果是等号，按照数值for循环来解析
		return _finishForNumStat(l, lineOfFor, name)
	} else {
		return _finishForInStat(l, name)
	}
}

// 数值for循环
func _finishForNumStat(l *Lexer, lineOfFor int, varName string) *ForNumStat {
	l.NextTokenOfKind(TOKEN_OP_ASSIGN) // skip `=`
	initExp := parseExp(l)
	l.NextTokenOfKind(TOKEN_SEP_COMMA) // skip `,`
	limitExp := parseExp(l)

	var stepExp Exp
	if l.LookAhead() == TOKEN_SEP_COMMA { // `,`
		l.NextToken() // skip `,`
		stepExp = parseExp(l)
	} else {
		stepExp = &IntegerExp{Line: l.Line(), Val: 1} // 默认步长为1
	}

	lineOfDo, _ := l.NextTokenOfKind(TOKEN_KW_DO) // skip `do`
	block := parseBlock(l)
	l.NextTokenOfKind(TOKEN_KW_END) // skip `end`

	return &ForNumStat{
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
func _finishForInStat(l *Lexer, name0 string) *ForInStat {
	name := _finishNameList(l, name0)
	l.NextTokenOfKind(TOKEN_KW_IN) // skip `in`
	expList := parseExpList(l)
	lineOfDo, _ := l.NextTokenOfKind(TOKEN_KW_DO) // skip `do`
	block := parseBlock(l)
	l.NextTokenOfKind(TOKEN_KW_END) // skip `end`
	return &ForInStat{LineOfDo: lineOfDo, NameList: name, ExpList: expList, Block: block}
}

// 解析循环变量名列表
func _finishNameList(l *Lexer, name0 string) []string {
	names := []string{name0}
	for l.LookAhead() == TOKEN_SEP_COMMA { // `,`
		l.NextToken() // skip `,`
		_, name := l.NextIdentifier()
		names = append(names, name)
	}
	return names
}

// 局部变量声明和局部函数定义
func parseLocalAssignOrFuncDefStat(l *Lexer) Stat {
	l.NextTokenOfKind(TOKEN_KW_LOCAL)       // skip `local`
	if l.LookAhead() == TOKEN_KW_FUNCTION { // `function`
		return _finishLocalFuncDefStat(l)
	} else {
		return _finishLocalAssignStat(l)
	}
}

// 局部函数定义
func _finishLocalFuncDefStat(l *Lexer) *LocalFuncDefStat {
	l.NextTokenOfKind(TOKEN_KW_FUNCTION) // skip `function`
	_, name := l.NextIdentifier()
	fdExp := parseFuncDefExp(l)
	return &LocalFuncDefStat{Name: name, Exp: fdExp}
}

// 局部变量声明
func _finishLocalAssignStat(l *Lexer) *LocalVarDeclStat {
	_, name0 := l.NextIdentifier()
	names := _finishNameList(l, name0)
	var exps []Exp = nil
	if l.LookAhead() == TOKEN_OP_ASSIGN { // `=`
		l.NextToken() // skip `=`
		exps = parseExpList(l)
	}
	lastLine := l.Line()
	return &LocalVarDeclStat{LastLine: lastLine, NameList: names, ExpList: exps}
}

// 赋值和函数调用语句
func parseAssignOrFuncCallStat(l *Lexer) Stat {
	// 先解析前缀表达式
	prefixExp := parsePrefixExp(l)
	if fc, ok := prefixExp.(*FuncCallExp); ok { // 如果解析出来的前缀表达式时是函数调用表达式
		return fc
	} else { // 否则是var表达式
		return parseAssignStat(l, prefixExp)
	}
}

// 解析赋值语句
func parseAssignStat(l *Lexer, var0 Exp) *AssignStat {
	varList := _finishVarList(l, var0) // 解析变量列表
	l.NextTokenOfKind(TOKEN_OP_ASSIGN) // skip `=`
	expList := parseExpList(l)         // 解析表达式列表
	lastLine := l.Line()
	return &AssignStat{LastLine: lastLine, VarList: varList, ExpList: expList}
}

// 解析变量列表
func _finishVarList(l *Lexer, var0 Exp) []Exp {
	vars := []Exp{_checkVar(l, var0)}      // 检查变量是否合法并添加到变量列表中
	for l.LookAhead() == TOKEN_SEP_COMMA { // `,`
		l.NextToken()                          // skip `,`
		exp := parsePrefixExp(l)               // 解析前缀表达式
		vars = append(vars, _checkVar(l, exp)) // 检查变量是否合法并添加到变量列表中
	}
	return vars
}

// 检查是否是变量
// var ::=  Name | prefixexp `[´ exp `]´ | prefixexp `.´ Name
func _checkVar(l *Lexer, exp Exp) Exp {
	switch exp.(type) {
	case *NameExp, *TableAccessExp:
		return exp
	default:
		l.NextTokenOfKind(-1) // 报错
		return nil
	}
}

// 非局部函数定义语句
func parseFuncDefStat(l *Lexer) Stat {
	l.NextTokenOfKind(TOKEN_KW_FUNCTION)  // skip `function`
	fnExp, hasColon := _finishFuncName(l) // 解析函数名
	fdExp := parseFuncDefExp(l)           // 解析函数定义表达式
	if hasColon {                         // 如果函数名是以冒号开头的 `foo:bar()`
		fdExp.ParList = append(fdExp.ParList, "") // 添加一个空的参数
		copy(fdExp.ParList[1:], fdExp.ParList)    // 将参数列表向后移动一位 `foo:bar(a, b, c)` => `foo:bar("", a, b, c)`
		fdExp.ParList[0] = "self"                 // 将第一个参数设置为 `self` `foo:bar(a, b, c)` => `foo:bar("self", a, b, c)`
	}

	// 最终将非局部函数语句转换为赋值语句
	return &AssignStat{
		LastLine: fdExp.LastLine,
		VarList:  []Exp{fnExp},
		ExpList:  []Exp{fdExp},
	}
}

// 解析函数名
func _finishFuncName(l *Lexer) (exp Exp, hasColon bool) {
	line, name := l.NextIdentifier() // 获取下一个标识符
	exp = &NameExp{Line: line, Name: name}
	for l.LookAhead() == TOKEN_SEP_DOT { // 不断取点
		l.NextToken()                    // skip `.`
		line, name := l.NextIdentifier() // 获取下一个标识符
		idx := &StringExp{Line: line, Str: name}
		exp = &TableAccessExp{PrefixExp: exp, KeyExp: idx} // 生成表达式 `a.b.c` => `a["b"]["c"]`
	}

	if l.LookAhead() == TOKEN_SEP_COLON { // 如果有冒号
		l.NextToken() // skip `:`
		line, name := l.NextIdentifier()
		idx := &StringExp{Line: line, Str: name}
		exp = &TableAccessExp{PrefixExp: exp, KeyExp: idx} // 生成表达式 `a:b()` => `a["b"]`
		hasColon = true                                    // 标记函数名是以冒号开头的
	}

	return
}
