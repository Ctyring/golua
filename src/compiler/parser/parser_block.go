package parser

import (
	. "lua/src/compiler/ast"
	. "lua/src/compiler/lexer"
)

// 创建Block结构体实例
func parseBlock(l *Lexer) *Block {
	return &Block{
		Stats:    parseStats(l),
		RetExps:  parseRetExps(l),
		LastLine: l.Line(),
	}
}

// 解析语句序列
func parseStats(l *Lexer) []Stat {
	stats := make([]Stat, 0, 8)
	for !_isReturnOrBlockEnd(l.LookAhead()) {
		stat := parseStat(l)
		if _, ok := stat.(*EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

// 判断块是否结束
func _isReturnOrBlockEnd(tokenKind int) bool {
	switch tokenKind {
	case TOKEN_KW_RETURN, TOKEN_KW_END, TOKEN_EOF, TOKEN_KW_ELSE, TOKEN_KW_ELSEIF, TOKEN_KW_UNTIL:
		return true
	}
	return false
}

// 解析返回值表达式
func parseRetExps(l *Lexer) []Exp {
	// 如果不是return说明没有返回值
	if l.LookAhead() != TOKEN_KW_RETURN {
		return nil
	}

	l.NextToken() // skip `return`
	switch l.LookAhead() {
	// 如果发现是分号或者块结束符号，说明没有返回值
	case TOKEN_EOF, TOKEN_KW_END, TOKEN_KW_ELSE, TOKEN_KW_ELSEIF, TOKEN_KW_UNTIL:
		return []Exp{}
	case TOKEN_SEP_SEMI:
		l.NextToken() // 跳过分号
		return []Exp{}
	default:
		// 解析返回值序列
		exps := parseExpList(l)
		if l.LookAhead() == TOKEN_SEP_SEMI {
			l.NextToken()
		}
		return exps
	}
}
