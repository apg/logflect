package logflect

type Message interface {
	Field(n string) (interface{}, bool)
}

type StrMessage string

func (s StrMessage) Field(f string) (interface{}, bool) {
	return s, true
}
