package codegen

import (
	. "lua/src/compiler/ast"
	. "lua/src/compiler/lexer"
	. "lua/src/vm"
)

var arithAndBitwiseBinops = map[int]int{
	TOKEN_OP_ADD:  OP_ADD,
	TOKEN_OP_SUB:  OP_SUB,
	TOKEN_OP_MUL:  OP_MUL,
	TOKEN_OP_MOD:  OP_MOD,
	TOKEN_OP_POW:  OP_POW,
	TOKEN_OP_DIV:  OP_DIV,
	TOKEN_OP_IDIV: OP_IDIV,
	TOKEN_OP_BAND: OP_BAND,
	TOKEN_OP_BOR:  OP_BOR,
	TOKEN_OP_BXOR: OP_BXOR,
	TOKEN_OP_SHL:  OP_SHL,
	TOKEN_OP_SHR:  OP_SHR,
}

type funcInfo struct {
	constants  map[interface{}]int    // 常量表
	usedRegs   int                    // 已分配的寄存器数量
	maxRegs    int                    // 最大寄存器数量
	scopeLv    int                    // 作用域层级
	locVars    []*locVarInfo          // 局部变量表
	locNames   map[string]*locVarInfo // 局部变量名表
	breaks     [][]int                // 记录break指令的跳转位置
	parent     *funcInfo              // 父函数
	upvalues   map[string]upvalInfo   // Upvalue表
	insts      []uint32               // 指令表
	subFuncs   []*funcInfo            // 子函数表
	numParams  int                    // 参数数量
	isVararg   bool                   // 是否是可变参数
	upvalNames []string               // Upvalue名表
}

func newFuncInfo(parent *funcInfo, fd *FuncDefExp) *funcInfo {
	return &funcInfo{
		parent:     parent,
		subFuncs:   []*funcInfo{},
		constants:  map[interface{}]int{},
		upvalues:   map[string]upvalInfo{},
		locNames:   map[string]*locVarInfo{},
		locVars:    make([]*locVarInfo, 0, 8),
		breaks:     make([][]int, 1),
		insts:      make([]uint32, 1, 8),
		isVararg:   fd.IsVararg,
		numParams:  len(fd.ParList),
		upvalNames: make([]string, 0, 8),
	}
}

type locVarInfo struct {
	prev     *locVarInfo // 前一个局部变量，实现单向链表
	name     string      // 变量名
	scopeLv  int         // 变量的作用域层级
	slot     int         // 变量的寄存器索引
	captured bool        // 是否被闭包捕获
}

type upvalInfo struct {
	locVarSlot int // 如果Upvalue捕获的是直接外围函数的局部变量，则该字段记录该局部变量所占用的寄存器索引
	upvalIndex int // 否则Upvalue已经被外围函数捕获，该字段记录该Upvalue在外围函数的Upvalue表中的索引
	index      int // 记录Upvalue在函数中出现的顺序
}

// 返回常量在表中的索引
func (self *funcInfo) indexOfConstant(k interface{}) int {
	if idx, found := self.constants[k]; found {
		return idx
	}
	idx := len(self.constants)
	self.constants[k] = idx
	return idx
}

// 判断名字是否已经和某个Upvalue绑定，如果是，则返回其索引，否则尝试绑定并返回索引，如果失败返回-1
func (self *funcInfo) indexOfUpval(name string) int {
	if upval, found := self.upvalues[name]; found {
		return upval.index
	}
	if self.parent != nil {
		if locVar, found := self.parent.locNames[name]; found { // 如果是在外围函数中定义的局部变量
			idx := len(self.upvalues)
			self.upvalues[name] = upvalInfo{locVar.slot, -1, idx}
			self.upvalNames = append(self.upvalNames, name)
			locVar.captured = true
			return idx
		}
		if idx := self.parent.indexOfUpval(name); idx >= 0 { // 如果是在外围函数的Upvalue表中(不用捕获)
			self.upvalues[name] = upvalInfo{-1, idx, idx}
			self.upvalNames = append(self.upvalNames, name)
			return idx
		}
	}
	return -1
}

// 分配一个寄存器
func (self *funcInfo) allocReg() int {
	self.usedRegs++
	if self.usedRegs >= 255 {
		panic("function or expression needs too many registers")
	}
	if self.usedRegs > self.maxRegs { // 必要时更新最大寄存器数量
		self.maxRegs = self.usedRegs
	}
	return self.usedRegs - 1 // 寄存器索引从0开始
}

// 分配n个寄存器,返回第一个寄存器的索引
func (self *funcInfo) allocRegs(n int) int {
	for i := 0; i < n; i++ {
		self.allocReg()
	}
	return self.usedRegs - n
}

// 释放一个寄存器
func (self *funcInfo) freeReg() {
	self.usedRegs--
}

// 释放n个寄存器
func (self *funcInfo) freeRegs(n int) {
	self.usedRegs -= n
}

// 进入新的作用域
func (self *funcInfo) enterScope(breakable bool) {
	self.scopeLv++
	if breakable {
		self.breaks = append(self.breaks, []int{}) // 循环块
	} else {
		self.breaks = append(self.breaks, nil) // 非循环块
	}
}

// 在当前作用域中添加一个局部变量，返回其分配的寄存器索引
func (self *funcInfo) addLocVar(name string) int {
	newVar := &locVarInfo{
		prev:    self.locNames[name],
		name:    name,
		scopeLv: self.scopeLv,
		slot:    self.allocReg(),
	}
	self.locVars = append(self.locVars, newVar)
	self.locNames[name] = newVar
	return newVar.slot
}

// 检查局部变量名是否已经和某个寄存器绑定，如果是则返回其寄存器索引，否则返回-1
func (self *funcInfo) slotOfLocVar(name string) int {
	if locVar, found := self.locNames[name]; found {
		return locVar.slot
	}
	return -1
}

// 退出当前作用域
func (self *funcInfo) exitScope() {
	pendingBreakJmps := self.breaks[len(self.breaks)-1] // 获取末尾元素
	self.breaks = self.breaks[:len(self.breaks)-1]      // 删除末尾元素
	a := self.getJmpArgA()                              // 是否需要关闭Upvalue
	for _, pc := range pendingBreakJmps {               // 遍历末尾元素
		sBx := self.pc() - pc                     // 计算跳转偏移量
		i := (sBx+MAXARG_sBx)<<14 | a<<6 | OP_JMP // 组装指令
		self.insts[pc] = uint32(i)                // 修改指令(break的时候会生成指令，但不能确定跳转偏移量，所以先用0占位)
	}
	self.scopeLv--
	for _, locVar := range self.locNames { // 遍历并判断变量的作用域层级
		if locVar.scopeLv > self.scopeLv {
			self.removeLocVar(locVar)
		}
	}
}

// 移除一个局部变量:解绑局部变量名，回收寄存器
func (self *funcInfo) removeLocVar(locVar *locVarInfo) {
	self.freeReg() // 回收寄存器
	if locVar.prev == nil {
		delete(self.locNames, locVar.name) // 解绑局部变量名
	} else if locVar.prev.scopeLv == locVar.scopeLv {
		self.removeLocVar(locVar.prev) // 递归删除前一个局部变量
	} else {
		self.locNames[locVar.name] = locVar.prev // 更新局部变量名表
	}
}

// 把break语句对应的跳转指令添加到最近的循环块内
func (self *funcInfo) addBreakJmp(pc int) {
	for i := self.scopeLv; i >= 0; i-- {
		if self.breaks[i] != nil {
			self.breaks[i] = append(self.breaks[i], pc)
			return
		}
	}
	panic("<break> at line ? not inside a loop!")
}

// 获取JMP指令的A操作数，操作数A决定了Upvalue的数量
func (self *funcInfo) getJmpArgA() int {
	hasCapturedLocVars := false
	minSlotOfLocVars := self.maxRegs
	for _, locVar := range self.locNames { // 遍历局部变量名表
		if locVar.scopeLv == self.scopeLv { // 作用域层级相同
			for v := locVar; v != nil && v.scopeLv == self.scopeLv; v = v.prev { // 遍历同名局部变量
				if v.captured { // 是否被捕获
					hasCapturedLocVars = true
				}
				if v.slot < minSlotOfLocVars && v.name[0] != '(' { // 获取到最小的local变量寄存器索引
					minSlotOfLocVars = v.slot
				}
			}
		}
	}
	if hasCapturedLocVars {
		return minSlotOfLocVars + 1
	} else {
		return 0
	}
}

// 返回已经生成的最后一条指令的程序计数器
func (self *funcInfo) pc() int {
	return len(self.insts) - 1
}

// 填充指令中的sBx字段
func (self *funcInfo) fixSbx(pc, sBx int) {
	i := self.insts[pc]
	i = i << 18 >> 18                 // 清除sBx字段
	i |= uint32(sBx+MAXARG_sBx) << 14 // 重新设置sBx字段
	self.insts[pc] = i
}

// 关闭未关闭的upvalue
func (self *funcInfo) closeOpenUpvals() {
	a := self.getJmpArgA()
	if a > 0 {
		self.emitJmp(a, 0) // sBx == 0 也就是不跳转，只是关闭upvalue
	}
}

// 四种编码生成
// ABC
func (self *funcInfo) emitABC(op, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | op
	self.insts = append(self.insts, uint32(i))
}

// ABx
func (self *funcInfo) emitABx(op, a, bx int) {
	i := bx<<14 | a<<6 | op
	self.insts = append(self.insts, uint32(i))
}

// AsBx
func (self *funcInfo) emitAsBx(op, a, sbx int) {
	i := (sbx+MAXARG_sBx)<<14 | a<<6 | op
	self.insts = append(self.insts, uint32(i))
}

// Ax
func (self *funcInfo) emitAx(op, ax int) {
	i := ax<<6 | op
	self.insts = append(self.insts, uint32(i))
}

// r[a] = r[b]
func (self *funcInfo) emitMove(a, b int) {
	self.emitABC(OP_MOVE, a, b, 0)
}

// r[a], r[a+1], ..., r[a+b] = nil
func (self *funcInfo) emitLoadNil(a, n int) {
	self.emitABC(OP_LOADNIL, a, n-1, 0)
}

// r[a] = (bool)b; if (c) pc++
func (self *funcInfo) emitLoadBool(a, b, c int) {
	self.emitABC(OP_LOADBOOL, a, b, c)
}

// r[a] = kst[bx]
func (self *funcInfo) emitLoadK(a int, k interface{}) {
	idx := self.indexOfConstant(k)
	if idx < (1 << 18) {
		self.emitABx(OP_LOADK, a, idx)
	} else {
		self.emitABx(OP_LOADKX, a, 0)
		self.emitAx(OP_EXTRAARG, idx)
	}
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (self *funcInfo) emitVararg(a, n int) {
	self.emitABC(OP_VARARG, a, n+1, 0)
}

// r[a] = emitClosure(proto[bx])
func (self *funcInfo) emitClosure(a, bx int) {
	self.emitABx(OP_CLOSURE, a, bx)
}

// r[a] = {}
func (self *funcInfo) emitNewTable(a, nArr, nRec int) {
	self.emitABC(OP_NEWTABLE,
		a, Int2fb(nArr), Int2fb(nRec)) // 使用浮点字节编码
}

// r[a][(c-1)*FPF+i] := r[a+i], 1 <= i <= b
func (self *funcInfo) emitSetList(a, b, c int) {
	self.emitABC(OP_SETLIST, a, b, c)
}

// r[a] := r[b][rk(c)]
func (self *funcInfo) emitGetTable(a, b, c int) {
	self.emitABC(OP_GETTABLE, a, b, c)
}

// r[a][rk(b)] = rk(c)
func (self *funcInfo) emitSetTable(a, b, c int) {
	self.emitABC(OP_SETTABLE, a, b, c)
}

// r[a] = upval[b]
func (self *funcInfo) emitGetUpval(a, b int) {
	self.emitABC(OP_GETUPVAL, a, b, 0)
}

// upval[b] = r[a]
func (self *funcInfo) emitSetUpval(a, b int) {
	self.emitABC(OP_SETUPVAL, a, b, 0)
}

// r[a] = upval[b][rk(c)]
func (self *funcInfo) emitGetTabUp(a, b, c int) {
	self.emitABC(OP_GETTABUP, a, b, c)
}

// upval[a][rk(b)] = rk(c)
func (self *funcInfo) emitSetTabUp(a, b, c int) {
	self.emitABC(OP_SETTABUP, a, b, c)
}

// r[a], ..., r[a+c-2] = r[a](r[a+1], ..., r[a+b-1])
func (self *funcInfo) emitCall(a, nArgs, nRet int) {
	self.emitABC(OP_CALL, a, nArgs+1, nRet+1)
}

// return r[a](r[a+1], ... ,r[a+b-1])
func (self *funcInfo) emitTailCall(a, nArgs int) {
	self.emitABC(OP_TAILCALL, a, nArgs+1, 0)
}

// return r[a], ... ,r[a+b-2]
// a代表寄存器索引，b代表返回值个数，b==-1说明返回所有值
func (self *funcInfo) emitReturn(a, n int) {
	self.emitABC(OP_RETURN, a, n+1, 0)
}

// r[a+1] := r[b]; r[a] := r[b][rk(c)]
func (self *funcInfo) emitSelf(a, b, c int) {
	self.emitABC(OP_SELF, a, b, c)
}

// pc+=sBx; if (a) close all upvalues >= r[a - 1]
func (self *funcInfo) emitJmp(a, sBx int) int {
	self.emitAsBx(OP_JMP, a, sBx)
	return len(self.insts) - 1
}

// if not (r[a] <=> c) then pc++
func (self *funcInfo) emitTest(a, c int) {
	self.emitABC(OP_TEST, a, 0, c)
}

// if (r[b] <=> c) then r[a] := r[b] else pc++
func (self *funcInfo) emitTestSet(a, b, c int) {
	self.emitABC(OP_TESTSET, a, b, c)
}

func (self *funcInfo) emitForPrep(a, sBx int) int {
	self.emitAsBx(OP_FORPREP, a, sBx)
	return len(self.insts) - 1
}

func (self *funcInfo) emitForLoop(a, sBx int) int {
	self.emitAsBx(OP_FORLOOP, a, sBx)
	return len(self.insts) - 1
}

func (self *funcInfo) emitTForCall(a, c int) {
	self.emitABC(OP_TFORCALL, a, 0, c)
}

func (self *funcInfo) emitTForLoop(a, sBx int) {
	self.emitAsBx(OP_TFORLOOP, a, sBx)
}

// r[a] = op r[b]
func (self *funcInfo) emitUnaryOp(op, a, b int) {
	switch op {
	case TOKEN_OP_NOT:
		self.emitABC(OP_NOT, a, b, 0)
	case TOKEN_OP_BNOT:
		self.emitABC(OP_BNOT, a, b, 0)
	case TOKEN_OP_LEN:
		self.emitABC(OP_LEN, a, b, 0)
	case TOKEN_OP_UNM:
		self.emitABC(OP_UNM, a, b, 0)
	}
}

// r[a] = rk[b] op rk[c]
// arith & bitwise & relational
func (self *funcInfo) emitBinaryOp(op, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		self.emitABC(opcode, a, b, c)
	} else {
		switch op { // 处理比较运算符
		case TOKEN_OP_EQ:
			self.emitABC(OP_EQ, 1, b, c)
		case TOKEN_OP_NE:
			self.emitABC(OP_EQ, 0, b, c)
		case TOKEN_OP_LT:
			self.emitABC(OP_LT, 1, b, c)
		case TOKEN_OP_GT:
			self.emitABC(OP_LT, 1, c, b)
		case TOKEN_OP_LE:
			self.emitABC(OP_LE, 1, b, c)
		case TOKEN_OP_GE:
			self.emitABC(OP_LE, 1, c, b)
		}
		self.emitJmp(0, 1)
		self.emitLoadBool(a, 0, 1)
		self.emitLoadBool(a, 1, 0)
	}
}
