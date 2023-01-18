package ast

// 代码块
type Block struct {
	LastLine int    // 末尾行号
	Stats    []Stat // 语句列表
	RetExps  []Exp  // 返回值表达式列表
}
