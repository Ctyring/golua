package parser

import (
	ast2 "lua/src/compiler/ast"
	lexer2 "lua/src/compiler/lexer"
)

// 创建Block结构体实例
func parseBlock(l *lexer2.Lexer) *ast2.Block {
	return &ast2.Block{
		Stats:    parseStats(l),
		RetExps:  parseRetExps(l),
		LastLine: l.Line(),
	}
}

// 解析语句序列
func parseStats(l *lexer2.Lexer) []ast2.Stat {
	stats := make([]ast2.Stat, 0, 8)
	for !_isReturnOrBlockEnd(l.LookAhead()) {
		stat := parseStat(l)
		if _, ok := stat.(*ast2.EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

// 判断块是否结束
func _isReturnOrBlockEnd(tokenKind int) bool {
	switch tokenKind {
	case lexer2.TOKEN_KW_RETURN, lexer2.TOKEN_KW_END, lexer2.TOKEN_EOF, lexer2.TOKEN_KW_ELSE, lexer2.TOKEN_KW_ELSEIF, lexer2.TOKEN_KW_UNTIL:
		return true
	}
	return false
}

// 解析返回值表达式
func parseRetExps(l *lexer2.Lexer) []ast2.Exp {
	// 如果不是return说明没有返回值
	if l.LookAhead() != lexer2.TOKEN_KW_RETURN {
		return nil
	}

	l.NextToken() // skip `return`
	switch l.LookAhead() {
	// 如果发现是分号或者块结束符号，说明没有返回值
	case lexer2.TOKEN_EOF, lexer2.TOKEN_KW_END, lexer2.TOKEN_KW_ELSE, lexer2.TOKEN_KW_ELSEIF, lexer2.TOKEN_KW_UNTIL:
		return []ast2.Exp{}
	case lexer2.TOKEN_SEP_SEMI:
		l.NextToken() // 跳过分号
		return []ast2.Exp{}
	default:
		// 解析返回值序列
		exps := parseExpList(l)
		if l.LookAhead() == lexer2.TOKEN_SEP_SEMI {
			l.NextToken()
		}
		return exps
	}
}
