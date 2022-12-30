package ast

type Exp interface{}

// 简单表达式
type NilExp struct{ Line int }    // nil
type TrueExp struct{ Line int }   // true
type FalseExp struct{ Line int }  // false
type VarargExp struct{ Line int } // ...
type IntegerExp struct {
	Line int
	Val  int64
} // 整数
type FloatExp struct {
	Line int
	Val  float64
} // 浮点数
type StringExp struct {
	Line int
	Str  string
} // 字符串
type NameExp struct {
	Line int
	Name string
} // 变量名

// 运算符表达式
type UnopExp struct { // 一元运算符表达式
	Line int
	Op   int
	Exp  Exp
}

type BinopExp struct { // 二元运算符表达式
	Line int
	Op   int
	Exp1 Exp
	Exp2 Exp
}

type ConcatExp struct { // 字符串连接表达式
	Line int
	Exps []Exp
}

// 表构造表达式
type TableConstructorExp struct {
	Line     int
	LastLine int
	KeyExps  []Exp
	ValExps  []Exp
}

type FuncDefExp struct { // 函数定义表达式
	Line     int
	LastLine int
	ParList  []string
	IsVararg bool
	Block    *Block
}

// 前缀表达式(可以作为表访问表达式、记录访问表达式、函数调用表达式的前缀)
// 包括 var表达式 函数调用表达式 圆括号表达式
// var表达式又包括名字表达式，表访问表达式，记录访问表达式
// prefixexp ::= Name | '(' exp ')' | prefixexp '[' exp ']' | prefixexp '.' Name | prefixexp [':' Name] args

type ParensExp struct { // 圆括号表达式 用途：改变运算符的优先级或者结合性
	Exp Exp
}

type TableAccessExp struct { // 表访问表达式
	LastLine  int // `]`所在行号
	PrefixExp Exp
	KeyExp    Exp
}

type FuncCallExp struct { // 函数调用表达式
	Line      int // `(`所在行号
	LastLine  int // `)`所在行号
	PrefixExp Exp
	NameExp   *StringExp
	Args      []Exp
}
