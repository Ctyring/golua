package state

import (
	"lua/src/number"
)

func (self *luaState) Len(idx int) {
	val := self.stack.get(idx)
	if s, ok := val.(string); ok { // 是否是字符串
		self.stack.push(int64(len(s)))
	} else if result, ok := callMetamethod(val, val, "__len", self); ok { // 是否有元方法
		self.stack.push(result)
	} else if t, ok := val.(*luaTable); ok { // 如果找不到元方法，但值是表，结果就是表的长度
		self.stack.push(int64(t.len()))
	} else {
		panic("length error!")
	}
}

// 从栈顶弹出n个值进行拼接
func (self *luaState) Concat(n int) {
	if n == 0 {
		self.stack.push("")
	} else if n >= 2 {
		for i := 1; i < n; i++ {
			if self.IsString(-1) && self.IsString(-2) {
				s2 := self.ToString(-1)
				s1 := self.ToString(-2)
				self.Pop(2)
				self.stack.push(s1 + s2)
				continue
			}
			// 如果不是字符串，尝试使用元方法
			b := self.stack.pop()
			a := self.stack.pop()
			if result, ok := callMetamethod(a, b, "__concat", self); ok {
				self.stack.push(result)
				continue
			}
			panic("concatenation error!")
		}
	}
}

// 根据键获取表的下一个键值对
func (self *luaState) Next(idx int) bool {
	val := self.stack.get(idx)
	if t, ok := val.(*luaTable); ok {
		key := self.stack.pop()
		if nextKey := t.nextKey(key); nextKey != nil {
			self.stack.push(nextKey)
			self.stack.push(t.get(nextKey))
			return true
		}
		return false
	}
	panic("table expected!")
}

func (self *luaState) Error() int {
	err := self.stack.pop()
	panic(err)
}

func (self *luaState) StringToNumber(s string) bool {
	if n, ok := number.ParseInteger(s); ok {
		self.PushInteger(n)
		return true
	}
	if n, ok := number.ParseFloat(s); ok {
		self.PushNumber(n)
		return true
	}
	return false
}
