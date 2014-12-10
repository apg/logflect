package logflect

import "fmt"

type Message interface {
	Field(n string) (interface{}, bool)
	String() string
}

type StrMessage string

type SyslogMessage struct {
	PrivalVersion []byte
	Time          []byte
	Hostname      []byte
	Name          []byte
	Procid        []byte
	Msgid         []byte
	Message       []byte
}

func (s StrMessage) Field(f string) (interface{}, bool) {
	return s, true
}

func (s StrMessage) String() string {
	return string(s)
}

func (s SyslogMessage) Field(f string) (interface{}, bool) {
	switch f {
	case "PrivalVersion", "privalVersion", "privalversion":
		return s.PrivalVersion, true
	case "Time", "time":
		return s.Time, true
	case "Hostname", "hostname":
		return s.Hostname, true
	case "Name", "name":
		return s.Name, true
	case "Procid", "procid":
		return s.Procid, true
	case "Msgid", "msgid", "msgId":
		return s.Msgid, true
	case "Message", "message":
		return string(s.Message), true
	default:
		return "", false
	}
}

func (s SyslogMessage) String() string {
	tmp := fmt.Sprintf("%s %s %s %s %s %s %s\n", s.PrivalVersion, s.Time, s.Hostname, s.Name, s.Procid, s.Msgid, s.Message)
	return fmt.Sprintf("%d %s\n", len(tmp), tmp)
}
