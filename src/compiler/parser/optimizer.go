package parser

import (
	"lua/src/compiler/ast"
	"lua/src/compiler/lexer"
	"lua/src/number"
	"math"
)

func optimizeLogicalOr(exp *ast.BinopExp) ast.Exp {
	if isTrue(exp.Exp1) {
		return exp.Exp1 // true or x => true
	}
	if isFalse(exp.Exp1) && !isVarargOrFuncCall(exp.Exp2) {
		return exp.Exp2 // false or x => x
	}
	return exp
}

func optimizeLogicalAnd(exp *ast.BinopExp) ast.Exp {
	if isFalse(exp.Exp1) {
		return exp.Exp1 // false and x => false
	}
	if isTrue(exp.Exp1) && !isVarargOrFuncCall(exp.Exp2) {
		return exp.Exp2 // true and x => x
	}
	return exp
}

func optimizeBitwiseBinaryOp(exp *ast.BinopExp) ast.Exp {
	if i, ok := castToInt(exp.Exp1); ok {
		if j, ok := castToInt(exp.Exp2); ok {
			switch exp.Op {
			case lexer.TOKEN_OP_BAND:
				return &ast.IntegerExp{exp.Line, i & j}
			case lexer.TOKEN_OP_BOR:
				return &ast.IntegerExp{exp.Line, i | j}
			case lexer.TOKEN_OP_BXOR:
				return &ast.IntegerExp{exp.Line, i ^ j}
			case lexer.TOKEN_OP_SHL:
				return &ast.IntegerExp{exp.Line, number.ShiftLeft(i, j)}
			case lexer.TOKEN_OP_SHR:
				return &ast.IntegerExp{exp.Line, number.ShiftRight(i, j)}
			}
		}
	}
	return exp
}

func optimizeArithBinaryOp(exp *ast.BinopExp) ast.Exp {
	if x, ok := exp.Exp1.(*ast.IntegerExp); ok {
		if y, ok := exp.Exp2.(*ast.IntegerExp); ok {
			switch exp.Op {
			case lexer.TOKEN_OP_ADD:
				return &ast.IntegerExp{exp.Line, x.Val + y.Val}
			case lexer.TOKEN_OP_SUB:
				return &ast.IntegerExp{exp.Line, x.Val - y.Val}
			case lexer.TOKEN_OP_MUL:
				return &ast.IntegerExp{exp.Line, x.Val * y.Val}
			case lexer.TOKEN_OP_IDIV:
				if y.Val != 0 {
					return &ast.IntegerExp{exp.Line, number.IFloorDiv(x.Val, y.Val)}
				}
			case lexer.TOKEN_OP_MOD:
				if y.Val != 0 {
					return &ast.IntegerExp{exp.Line, number.IMod(x.Val, y.Val)}
				}
			}
		}
	}
	if f, ok := castToFloat(exp.Exp1); ok {
		if g, ok := castToFloat(exp.Exp2); ok {
			switch exp.Op {
			case lexer.TOKEN_OP_ADD:
				return &ast.FloatExp{exp.Line, f + g}
			case lexer.TOKEN_OP_SUB:
				return &ast.FloatExp{exp.Line, f - g}
			case lexer.TOKEN_OP_MUL:
				return &ast.FloatExp{exp.Line, f * g}
			case lexer.TOKEN_OP_DIV:
				if g != 0 {
					return &ast.FloatExp{exp.Line, f / g}
				}
			case lexer.TOKEN_OP_IDIV:
				if g != 0 {
					return &ast.FloatExp{exp.Line, number.FFloorDiv(f, g)}
				}
			case lexer.TOKEN_OP_MOD:
				if g != 0 {
					return &ast.FloatExp{exp.Line, number.FMod(f, g)}
				}
			case lexer.TOKEN_OP_POW:
				return &ast.FloatExp{exp.Line, math.Pow(f, g)}
			}
		}
	}
	return exp
}

func optimizePow(exp ast.Exp) ast.Exp {
	if binop, ok := exp.(*ast.BinopExp); ok {
		if binop.Op == lexer.TOKEN_OP_POW {
			binop.Exp2 = optimizePow(binop.Exp2)
		}
		return optimizeArithBinaryOp(binop)
	}
	return exp
}

func optimizeUnaryOp(exp *ast.UnopExp) ast.Exp {
	switch exp.Op {
	case lexer.TOKEN_OP_UNM:
		return optimizeUnm(exp)
	case lexer.TOKEN_OP_NOT:
		return optimizeNot(exp)
	case lexer.TOKEN_OP_BNOT:
		return optimizeBnot(exp)
	default:
		return exp
	}
}

func optimizeUnm(exp *ast.UnopExp) ast.Exp {
	switch x := exp.Exp.(type) { // number?
	case *ast.IntegerExp:
		x.Val = -x.Val
		return x
	case *ast.FloatExp:
		if x.Val != 0 {
			x.Val = -x.Val
			return x
		}
	}
	return exp
}

func optimizeNot(exp *ast.UnopExp) ast.Exp {
	switch exp.Exp.(type) {
	case *ast.NilExp, *ast.FalseExp: // false
		return &ast.TrueExp{exp.Line}
	case *ast.TrueExp, *ast.IntegerExp, *ast.FloatExp, *ast.StringExp: // true
		return &ast.FalseExp{exp.Line}
	default:
		return exp
	}
}

func optimizeBnot(exp *ast.UnopExp) ast.Exp {
	switch x := exp.Exp.(type) { // number?
	case *ast.IntegerExp:
		x.Val = ^x.Val
		return x
	case *ast.FloatExp:
		if i, ok := number.FloatToInteger(x.Val); ok {
			return &ast.IntegerExp{x.Line, ^i}
		}
	}
	return exp
}

func isFalse(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.FalseExp, *ast.NilExp:
		return true
	default:
		return false
	}
}

func isTrue(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.TrueExp, *ast.IntegerExp, *ast.FloatExp, *ast.StringExp:
		return true
	default:
		return false
	}
}

// todo
func isVarargOrFuncCall(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.VarargExp, *ast.FuncCallExp:
		return true
	}
	return false
}

func castToInt(exp ast.Exp) (int64, bool) {
	switch x := exp.(type) {
	case *ast.IntegerExp:
		return x.Val, true
	case *ast.FloatExp:
		return number.FloatToInteger(x.Val)
	default:
		return 0, false
	}
}

func castToFloat(exp ast.Exp) (float64, bool) {
	switch x := exp.(type) {
	case *ast.IntegerExp:
		return float64(x.Val), true
	case *ast.FloatExp:
		return x.Val, true
	default:
		return 0, false
	}
}
