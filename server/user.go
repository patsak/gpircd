package server

type Mode struct {
}

type UserInfo struct {
	Nick     string
	Server   string
	Host     string
	User     string
	RealName string
}

type User struct {
	inbox     chan Message
	mode      Mode
	info      UserInfo
	channels  map[string]bool
	connected bool
}

func (user *User) Put(msg Message) {

	user.inbox <- msg
}

func (user *User) Name() string {
	return user.info.Nick
}

func RunUserLoop(from, to chan Message, router *Router) {
	var user = User{}
	user.inbox = to
	user.channels = make(map[string]bool)

	go func() {
		for {
			select {
			case m := <-from:
				m.User = &user
				m.Prefix.Host = user.info.Host
				m.Prefix.Nick = user.info.Nick
				m.Prefix.User = user.info.User

				router.sendMessage(m)

				if m.Cmd == QUIT {
					close(to)
					break
				}
			}

		}
	}()

}
