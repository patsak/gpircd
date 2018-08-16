package server

import (
	"reflect"
	"testing"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage("ignore")
	if msg.Cmd != UNKNOWN {
		t.Error("illegal message non nil")
	}
	msg = NewMessage("join a")
	if !reflect.DeepEqual(msg.Args, []string{"a"}) {
		t.Error("illegal command arguments")
	}
}
