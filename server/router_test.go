package server

import (
	"testing"
)

var user = User{
	inbox:     make(chan Message),
	mode:      Mode{},
	info:      UserInfo{},
	channels:  make(map[string]bool),
	connected: true,
}

func TestNewRouter(t *testing.T) {
	router = NewRouter("new")
	if router.name != "new" {
		t.Error("incorrect setup router name")
	}
}

func TestHandlePing(t *testing.T) {
	router = NewRouter("new")

	go router.handlePing(&Message{
		Cmd:    PING,
		Name:   "PING",
		Args:   []string{"new"},
		Prefix: Prefix{},
		User:   &user,
	})
	msg := <-user.inbox
	if msg.Cmd != PONG {
		t.Error("server must response with pong on ping request")
	}
}

func TestHandleJoin(t *testing.T) {
	router = NewRouter("new")

	go router.handleJoin(&Message{
		Cmd:    JOIN,
		Name:   "JOIN",
		Args:   []string{"#test"},
		Prefix: Prefix{},
		User:   &user,
	})

	msg := <-user.inbox
	if msg.Cmd != JOIN {
		t.Error("server must response with join on join request")
	}
	if _, ok := msg.User.channels["#test"]; !ok {
		t.Error("user doesn't join in channel")
	}

	if _, ok := router.channel["#test"]; !ok {
		t.Error("channel not exists")
	}
}
