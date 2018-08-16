package server

import (
	"fmt"
	"strings"
)

type UserInbox interface {
	Name() string
	Put(msg Message)
}

type Router struct {
	direct  map[string]UserInbox
	channel map[string](map[string]UserInbox)
	inbox   chan Message
	name    string
	user    User
}

func NewRouter(serverName string) Router {
	inbox := make(chan Message)
	return Router{
		make(map[string]UserInbox),
		make(map[string](map[string]UserInbox)),
		inbox,
		serverName,
		User{
			inbox:     inbox,
			mode:      Mode{},
			channels:  make(map[string]bool),
			connected: true,
		},
	}
}
func (router *Router) handlePrivmsg(msg *Message) {
	var user = msg.Args[0]
	if user[0] == '#' {
		if inboxes, ok := router.channel[user]; ok {
			for _, inbox := range inboxes {
				if inbox.Name() != msg.User.Name() {
					inbox.Put(*msg)
				}
			}
			return
		}
	}

	if user[0] != '#' {
		if direct, ok := router.direct[user]; ok {
			direct.Put(*msg)
			return
		}
	}
	reply := Nosuchnick(user)
	replyMessage := router.toMessage(&reply)
	msg.User.Put(replyMessage)

}

func (router *Router) toMessage(numeric *Numeric) Message {
	ret := Message{
		Command(numeric.Code),
		fmt.Sprintf("%03d", int(numeric.Code)),
		numeric.Args,
		Prefix{
			Nick: router.name,
		},
		nil,
	}
	return ret

}

func (router *Router) handleNames(command *Message) {

	if len(command.Args) > 0 {
		if users, ok := router.channel[command.Args[0]]; ok {
			keys := make([]string, 0, len(users))
			for k := range users {
				keys = append(keys, k)
			}
			reply := NameReply(command.User.Name(), command.Args[0], keys)
			command.User.Put(router.toMessage(&reply))
		}
		reply := EnfOfNamesReply(command.Args[0])
		command.User.Put(router.toMessage(&reply))
	} else {
		reply := EnfOfNamesReply("")
		command.User.Put(router.toMessage(&reply))
	}

}
func (router *Router) handleUser(command *Message) {
	fmt.Println("Handle user registration " + string(command.ToWire()))
	if command.Cmd == NICK {
		if command.User.connected {
			delete(router.direct, command.User.Name())
			for k := range command.User.channels {
				delete(router.channel[k], command.User.Name())
				router.channel[k][command.Args[0]] = command.User
			}
		}
		command.User.info.Nick = command.Args[0]

	}
	if command.Cmd == USER {
		if len(command.Args) < 4 {
			response := NeedMoreParams(command.Name)
			command.User.Put(router.toMessage(&response))
			return
		}
		command.User.info.Host = command.Args[1]
		command.User.info.User = command.Args[0]
		command.User.info.Server = command.Args[2]
		command.User.info.RealName = command.Args[3]
	}

	if command.User.Name() != "" && command.User.info.User != "" {
		if _, ok := router.direct[command.User.Name()]; ok {
			reply := AlreadyRegistered()
			command.User.Put(router.toMessage(&reply))
		}

		router.direct[command.User.Name()] = command.User
		if !command.User.connected {
			reply := WelcomeReply(command.User.Name())

			command.User.Put(router.toMessage(&reply))
			command.User.connected = true
		}
	}
}
func (router *Router) handlePing(command *Message) {
	fmt.Printf("Handle ping %s\n", string(command.ToWire()))
	if len(command.Args) > 0 {
		if command.Args[0] == router.name {
			command.User.Put(Message{
				Cmd:    PONG,
				Name:   "PONG",
				Args:   command.Args,
				Prefix: Prefix{},
				User:   nil,
			})
		}
	}

}
func (router *Router) handleJoin(command *Message) {
	fmt.Println("Handle join " + string(command.ToWire()))

	channelName := command.Args[0]
	user := command.User.Name()
	if _, ok := router.direct[user]; !ok {
		fmt.Println("Unknown user")
	}

	if _, ok := router.channel[channelName]; !ok {
		router.channel[channelName] = make(map[string]UserInbox)
	}
	channeldict := router.channel[channelName]
	channeldict[command.User.Name()] = command.User
	command.User.channels[channelName] = true
	command.User.Put(*command)
	replyTopic := TopicReply(command.User.Name(), channelName, channelName)
	command.User.Put(router.toMessage(&replyTopic))
	if users, ok := router.channel[channelName]; ok {
		keys := make([]string, 0, len(users))
		for k := range users {
			keys = append(keys, k)
		}
		reply := NameReply(command.User.Name(), channelName, keys)
		command.User.Put(router.toMessage(&reply))
	}
	reply := EnfOfNamesReply(command.Args[0])
	command.User.Put(router.toMessage(&reply))
}

func (router *Router) run() {

	for {
		select {
		case broadcastMsg := <-router.inbox:
			switch cmd := broadcastMsg.Cmd; cmd {
			case PRIVMSG:
				router.handlePrivmsg(&broadcastMsg)

			case USER:
				router.handleUser(&broadcastMsg)

			case NICK:
				router.handleUser(&broadcastMsg)

			case JOIN:
				router.handleJoin(&broadcastMsg)

			case NAMES:
				router.handleNames(&broadcastMsg)
			case PING:
				router.handlePing(&broadcastMsg)

			case QUIT:
				delete(router.direct, broadcastMsg.User.Name())
				for c := range broadcastMsg.User.channels {
					fmt.Println("Remove user from channel " + c + " " + broadcastMsg.User.Name())
					delete(router.channel[c], broadcastMsg.User.Name())
				}
			default:
				fmt.Printf("Message %s not handled", strings.TrimSpace(string(broadcastMsg.ToWire())))

			}

		}
	}
}

func (router *Router) sendMessage(msg Message) {
	router.inbox <- msg
}
