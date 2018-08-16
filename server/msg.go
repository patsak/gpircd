package server

import (
	"fmt"
	"strings"
)

type Command int
type CommandName string
type Code int

const (
	ERR_NORECIPIENT      = 411
	ERR_NOSUCHNICK       = 401
	ERR_UNKNOWNCOMMAND   = 421
	ERR_NEEDMOREPARAMS   = 461
	RPL_NAMEREPLY        = 353
	RPL_ENDOFNAMES       = 366
	ERR_ALREADYREGISTRED = 462
	RPL_WELCOME          = 001
	RPL_TOPIC            = 332
)

const (
	USER    Command = 1
	NICK    Command = 2
	PRIVMSG Command = 3
	JOIN    Command = 4
	INFO    Command = 5
	QUIT    Command = 6
	NAMES   Command = 7
	PING    Command = 8
	PONG    Command = 9

	UNKNOWN Command = 1000
)

var commands = map[string]Command{
	"NICK":    NICK,
	"USER":    USER,
	"PRIVMSG": PRIVMSG,
	"NAMES":   NAMES,
	"JOIN":    JOIN,
	"QUIT":    QUIT,
	"PING":    PING,
	"PONG":    PONG,
}

func GetCommand(cmdstr string) Command {
	if v, ok := commands[strings.ToUpper(cmdstr)]; ok {
		return v
	} else {
		return UNKNOWN
	}
}

type Prefix struct {
	Nick   string
	Server string
	Host   string
	User   string
}

type Message struct {
	Cmd    Command
	Name   string
	Args   []string
	Prefix Prefix
	User   *User
}

type Numeric struct {
	Code Code
	Args []string
}

func NewMessage(raw string) *Message {
	raw = strings.TrimSpace(raw)

	msgTokens := strings.Split(raw, " ")
	if len(msgTokens) < 1 {
		return nil
	}
	command := Message{}
	commandToken := 0
	if len(msgTokens[0]) == 0 {
		return nil
	}
	if msgTokens[0][0] == ':' { // parse prefix
		commandToken = 1
		prefix := Prefix{}
		nickOrServerName := msgTokens[0][1:]
		var nick, user, host string
		var inNick, inUser, inHost = 0, -1, -1
		for pos, char := range nickOrServerName {

			if char == '!' && inNick >= 0 {
				nick = nickOrServerName[0:pos]
				inUser = pos + 1
			}
			if char == '@' && inUser > 0 {
				user = nickOrServerName[inUser:pos]
				inHost = pos + 1
				inUser = -1
			}
		}
		if nick == "" {
			nick = nickOrServerName
		}

		if inUser > 0 {
			user = nickOrServerName[inUser:]
		}
		if inHost > 0 {
			host = nickOrServerName[inHost:]
		}
		prefix.User = user
		prefix.Nick = nick
		prefix.Host = host
		command.Prefix = prefix
	}
	commandCode := GetCommand(msgTokens[commandToken])

	command.Cmd = commandCode
	command.Name = msgTokens[commandToken]

	if commandToken < len(msgTokens)-1 {
		trailinToken := -1
		for i, token := range msgTokens[commandToken+1:] {
			if token[0] == ':' {
				trailinToken = commandToken + 1 + i
				msgTokens[trailinToken] = token[1:]
			}
		}
		var args []string = msgTokens[commandToken+1:]

		if trailinToken != -1 {

			trail := strings.Join(msgTokens[trailinToken:], " ")
			args = append(msgTokens[commandToken+1:trailinToken], trail)

		}
		command.Args = args
	}
	return &command
}

func (prefix *Prefix) GetPrefixAddr() string {
	if prefix.Nick != "" {
		return prefix.Nick
	} else {
		return prefix.Server
	}
}

func (prefix *Prefix) GetNick() string {
	return prefix.Nick
}

func (msg *Message) ToWire() []byte {

	prefix := ""
	if msg.Prefix.Nick != "" {
		prefix = ":" + msg.Prefix.Nick
	}
	if msg.Prefix.User != "" {
		prefix += "!" + msg.Prefix.User
	}
	if msg.Prefix.Host != "" {
		prefix += "@" + msg.Prefix.Host
	}
	if prefix != "" {
		prefix += " "
	}
	return []byte(prefix + msg.Name + " " + strings.Join(msg.Args, " ") + "\r\n")
}

func Nosuchnick(nick string) Numeric {
	return Numeric{
		ERR_NOSUCHNICK,
		[]string{nick, ":No such nick/channel"},
	}
}

func NeedMoreParams(command string) Numeric {
	return Numeric{
		ERR_NEEDMOREPARAMS,
		[]string{command, ":Not enough parameters"},
	}
}
func TopicReply(nick string, channel string, topic string) Numeric {
	return Numeric{
		RPL_TOPIC,
		[]string{nick, channel, fmt.Sprintf(":%s", topic)},
	}
}

func NameReply(nick string, channel string, nicks []string) Numeric {
	resp := strings.Builder{}
	resp.WriteByte('=')
	resp.WriteByte(' ')
	resp.WriteString(channel)
	resp.WriteByte(' ')
	resp.WriteByte(':')
	for i, nick := range nicks {
		resp.WriteString(nick)
		if i < len(nicks)-1 {
			resp.WriteByte(' ')
		}
	}
	return Numeric{
		RPL_NAMEREPLY,
		[]string{resp.String()},
	}
}

func EnfOfNamesReply(channel string) Numeric {
	return Numeric{
		RPL_ENDOFNAMES,
		[]string{channel, ":End of /NAMES list"},
	}

}
func AlreadyRegistered() Numeric {
	return Numeric{
		ERR_ALREADYREGISTRED,
		[]string{":You may not reregister"},
	}
}
func WelcomeReply(nick string) Numeric {
	return Numeric{
		RPL_WELCOME,
		[]string{nick, fmt.Sprintf(":Welcome to the Internet Relay Network %s", nick)},
	}
}

func UnknownCommand(command string) Numeric {
	return Numeric{
		ERR_UNKNOWNCOMMAND,
		[]string{fmt.Sprintf("%s :Unknown command", command)},
	}
}
