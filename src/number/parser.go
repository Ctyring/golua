package number

import "strconv"

// 将字符串解析为整数
func ParseInteger(str string) (int64, bool) {
	i, err := strconv.ParseInt(str, 10, 64)
	return i, err == nil
}

// 将字符串解析为浮点数
func ParseFloat(str string) (float64, bool) {
	f, err := strconv.ParseFloat(str, 64)
	return f, err == nil
}
