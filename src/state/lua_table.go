package state

import (
	"lua/src/number"
	"math"
)

type luaTable struct {
	metatable *luaTable             // 元表
	arr       []luaValue            // 数组
	_map      map[luaValue]luaValue // map
	keys      map[luaValue]luaValue // key集合
	changed   bool                  // 是否改变
}

// 创建一个空的表，接受两个参数来预估表的用途和容量。
func newLuaTable(nArr, nRec int) *luaTable {
	t := &luaTable{}
	// 数组
	if nArr > 0 {
		t.arr = make([]luaValue, 0, nArr)
	}
	// map
	if nRec > 0 {
		t._map = make(map[luaValue]luaValue, nRec)
	}
	return t
}

// 获取表中指定键的值
func (self *luaTable) get(key luaValue) luaValue {
	// 如果能转成整数(或者本身是整数) 并且索引在数组范围内 则从数组中取值
	key = _floatToInteger(key)
	if idx, ok := key.(int64); ok {
		if idx >= 1 && idx <= int64(len(self.arr)) {
			return self.arr[idx-1]
		}
	}
	// 否则从map中取值
	return self._map[key]
}

func _floatToInteger(key luaValue) luaValue {
	if i, ok := key.(float64); ok {
		if i, ok := number.FloatToInteger(i); ok {
			return i
		}
	}
	return key
}

func (self *luaTable) put(key, val luaValue) {
	// 先判断是否是nil或nan
	if key == nil {
		panic("table index is nil!")
	}
	if f, ok := key.(float64); ok && math.IsNaN(f) {
		panic("table index is NaN!")
	}
	key = _floatToInteger(key)
	if idx, ok := key.(int64); ok {
		arrLne := int64(len(self.arr))
		// 如果索引在数组范围内 则放入数组
		if idx >= 1 && idx <= arrLne {
			self.arr[idx-1] = val
			// 如果nil在数组的末尾，删除末尾全部的nil
			if idx == arrLne && val == nil {
				self._shrinkArray()
			}
			return
		}
		// 如果索引在数组范围外 则放入map
		if idx == arrLne+1 {
			delete(self._map, key)
			if val != nil {
				self.arr = append(self.arr, val)
				// 动态扩展数组
				self._expandArray()
			}
			return
		}
	}
	// 如果值不是nil，写入，否则删除，节约空间
	if val != nil {
		if self._map == nil {
			self._map = make(map[luaValue]luaValue, 8)
		}
		self._map[key] = val
	} else {
		delete(self._map, key)
	}
}

// 删除末尾的nil
func (self *luaTable) _shrinkArray() {
	for i := len(self.arr) - 1; i >= 0; i-- {
		if self.arr[i] != nil {
			break
		}
		self.arr = self.arr[0:i]
	}
}

// 动态扩展数组
func (self *luaTable) _expandArray() {
	for idx := int64(len(self.arr)) + 1; true; idx++ {
		if val, found := self._map[idx]; found {
			self.arr = append(self.arr, val)
			delete(self._map, idx)
		} else {
			break
		}
	}
}

// 长度
func (self *luaTable) len() int {
	return len(self.arr)
}

func (self *luaTable) hasMetafield(fieldName string) bool {
	return self.metatable != nil &&
		self.metatable.get(fieldName) != nil
}

// 根据传入键返回表的下一个键
func (self *luaTable) nextKey(key luaValue) luaValue {
	if self.keys == nil || key == nil {
		self.initKeys()
		self.changed = false
	}
	return self.keys[key]
}

func (self *luaTable) initKeys() {
	// 记录表的键和下一个键的关系
	self.keys = make(map[luaValue]luaValue, len(self.arr)+len(self._map))
	var prevKey luaValue = nil
	for i, v := range self.arr { // 数组部分
		if v != nil {
			self.keys[prevKey] = int64(i + 1)
			prevKey = int64(i + 1)
		}
	}
	for k, v := range self._map { // 表部分
		if v != nil {
			self.keys[prevKey] = k
			prevKey = k
		}
	}
}
