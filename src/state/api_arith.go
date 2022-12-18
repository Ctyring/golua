package state

import (
	api2 "lua/src/api"
	"lua/src/number"
	"math"
)

// Arith方法
// 使用两个参数返回一个参数的形式表达，如果是一元运算符，忽略第二个参数即可
var (
	iadd  = func(a, b int64) int64 { return a + b }
	fadd  = func(a, b float64) float64 { return a + b }
	isub  = func(a, b int64) int64 { return a - b }
	fsub  = func(a, b float64) float64 { return a - b }
	imul  = func(a, b int64) int64 { return a * b }
	fmul  = func(a, b float64) float64 { return a * b }
	imod  = number.IMod
	fmod  = number.FMod
	pow   = math.Pow
	div   = func(a, b float64) float64 { return a / b }
	iidiv = number.IFloorDiv
	fidiv = number.FFloorDiv
	band  = func(a, b int64) int64 { return a & b }
	bor   = func(a, b int64) int64 { return a | b }
	bxor  = func(a, b int64) int64 { return a ^ b }
	shl   = number.ShiftLeft
	shr   = number.ShiftRight
	iunm  = func(a, _ int64) int64 { return -a }
	funm  = func(a, _ float64) float64 { return -a }
	bnot  = func(a, _ int64) int64 { return ^a }
)

// 容纳整数运算和浮点数运算
type operator struct {
	metamethod  string // 元方法
	integerFunc func(int64, int64) int64
	floatFunc   func(float64, float64) float64
}

var operators = []operator{
	{"__add", iadd, fadd},    // OP_ADD
	{"__sub", isub, fsub},    // OP_SUB
	{"__mul", imul, fmul},    // OP_MUL
	{"__mod", imod, fmod},    // OP_MOD
	{"__pow", nil, pow},      // OP_POW
	{"__div", nil, div},      // OP_DIV
	{"__idiv", iidiv, fidiv}, // OP_IDIV
	{"__band", band, nil},    // OP_BAND
	{"__bor", bor, nil},      // OP_BOR
	{"__bxor", bxor, nil},    // OP_BXOR
	{"__shl", shl, nil},      // OP_SHL
	{"__shr", shr, nil},      // OP_SHR
	{"__unm", iunm, funm},    // OP_UNM
	{"__bnot", bnot, nil},    // OP_BNOT
}

func (self *luaState) Arith(op api2.ArithOp) {
	var a, b luaValue
	b = self.stack.pop()
	if op != api2.LUA_OPUNM && op != api2.LUA_OPBNOT {
		a = self.stack.pop()
	} else {
		a = b
	}

	operator := operators[op]

	// 如果操作数都可以转成数字，那么进行常规的算术运算
	if result := _arith(a, b, operator); result != nil {
		self.stack.push(result)
		return
	}

	// 否则尝试调用元方法
	mm := operator.metamethod
	if result, ok := callMetamethod(a, b, mm, self); ok {
		self.stack.push(result)
		return
	}

	// 找不到对应元方法就报错
	panic("arithmetic error!")
}

// 执行计算
func _arith(a, b luaValue, op operator) luaValue {
	if op.integerFunc != nil { // 整数运算
		if x, ok := convertToInteger(a); ok {
			if y, ok := convertToInteger(b); ok {
				return op.integerFunc(x, y)
			}
		}
	}
	if op.floatFunc != nil {
		if x, ok := convertToFloat(a); ok {
			if y, ok := convertToFloat(b); ok {
				return op.floatFunc(x, y)
			}
		}
	}
	return nil
}
