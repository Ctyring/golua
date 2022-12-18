package ast

type Stat interface{}
type EmptyStat struct{}              // 空语句 `;`
type BreakStat struct{ Line int }    // break语句，会生成跳转指令，所以需要记录行号
type LabelStat struct{ Name string } // 标签语句 `::label::` 记录标签名
type GotoStat struct{ Name string }  // goto语句 `goto label` 记录标签名
type DoStat struct{ Block *Block }   // do语句 `do block end` 给语句块引入新的作用域，所以需要记录语句块
type FuncCallStat = FuncCallExp      // 函数调用语句 既可以是语句也可以是表达式，所以起了别名
type WhileStat struct {              // while语句 `while exp do block end` 记录条件表达式和语句块
	Exp   Exp
	Block *Block
}
type RepeatStat struct { // repeat语句 `repeat block until exp` 记录条件表达式和语句块
	Block *Block
	Exp   Exp
}
type IfStat struct { // if语句 `if exp then block {elseif exp then block} [else block] end` 可以合并为 if exp then block {elseif exp then block} end
	Exps   []Exp
	Blocks []*Block
}
type ForNumStat struct { // 数值for语句 `for Name = exp1, exp2, exp3 do block end`
	LineOfFor int    // for关键字所在行号
	LineOfDo  int    // do关键字所在行号
	VarName   string // 循环变量名
	InitExp   Exp    // 初始值表达式
	LimitExp  Exp    // 终止值表达式
	StepExp   Exp    // 步长表达式
	Block     *Block // 循环体
}
type ForInStat struct { // 泛型for语句 `for namelist in explist do block end`
	LineOfDo int      // do关键字所在行号
	NameList []string // 循环变量名列表
	ExpList  []Exp    // 迭代器函数和状态常量表达式列表
	Block    *Block   // 循环体
}
type LocalVarDeclStat struct { // 局部变量声明语句 `local namelist [= explist]`
	LastLine int      // 末尾行号
	NameList []string // 变量名列表
	ExpList  []Exp    // 表达式列表
}
type AssignStat struct { // 赋值语句 `varlist = explist`
	LastLine int   // 末尾行号
	VarList  []Exp // 变量列表
	ExpList  []Exp // 表达式列表
}
type LocalFuncDefStat struct { // 局部函数定义语句 `local function Name funcbody` 是局部变量声明语句的语法糖
	Name string
	Exp  *FuncDefExp
}
