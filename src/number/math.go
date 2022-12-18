package number

import "math"

// 整除，lua中的整除是向负无穷取整，go是向0取整
func IFloorDiv(a, b int64) int64 {
	if a > 0 && b > 0 || a < 0 && b < 0 || a%b == 0 {
		return a / b
	} else {
		return a/b - 1
	}
}

func FFloorDiv(a, b float64) float64 {
	return math.Floor(a / b)
}

// 取模函数
func IMod(a, b int64) int64 {
	return a - IFloorDiv(a, b)*b
}

// 浮点数取模
func FMod(a, b float64) float64 {
	return a - float64(int(a/b))*b
}

// 左移
func ShiftLeft(a, n int64) int64 {
	if n >= 0 {
		return a << uint(n)
	} else {
		return ShiftRight(a, -n)
	}
}

// 右移
func ShiftRight(a, n int64) int64 {
	if n >= 0 {
		return a >> uint(n)
	} else {
		return ShiftLeft(a, -n)
	}
}

// 浮点数转整数
func FloatToInteger(f float64) (int64, bool) {
	i := int64(f)
	return i, float64(i) == f
}

func _stringToInteger(s string, base int) (int64, bool) {
	// 先判断字符串是否能转换成整数
	if i, ok := ParseInteger(s); ok {
		return i, true
	}
	if f, ok := ParseFloat(s); ok {
		return FloatToInteger(f)
	}
	return 0, false
}
