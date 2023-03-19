package state

type userdata struct {
	val       interface{}
	metatable *luaTable
}

func newUserdata(val interface{}) *userdata {
	return &userdata{val: val}
}
