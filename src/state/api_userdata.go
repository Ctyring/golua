package state

func (self *luaState) NewUserdata(data interface{}) {
	ud := newUserdata(data)
	self.stack.push(ud)
}
