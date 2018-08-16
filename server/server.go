package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var logger = log.New(os.Stdout, "server: ", 0)
var router = NewRouter("localhost")

func handleConnection(conn net.Conn) {

	var reader = bufio.NewReader(conn)

	var writer = bufio.NewWriter(conn)
	var from = make(chan Message)
	var to = make(chan Message, 1)

	RunUserLoop(from, to, &router)

	close := make(chan bool)
	go func() {
		for {
			var msg, err = reader.ReadString('\r')
			reader.ReadByte() // read \n

			if err != nil {
				if err == io.EOF {
					return
				}
				close <- true
				logger.Printf("message reading error")
				return
			}

			command := NewMessage(strings.TrimSpace(msg))
			if command != nil {
				from <- *command
			}
		}
	}()

	go func() {
		for {
			msg, more := <-to
			if !more {
				close <- true
				return
			}
			var _, err = writer.Write(msg.ToWire())
			writer.Flush()
			if err == io.EOF {
				close <- true
				return
			}
		}
	}()

	<-close
	conn.Close()

}

func Start(port int, serverName string) {
	var ln, error = net.Listen("tcp", ":"+strconv.Itoa(port))

	if error != nil {
		logger.Fatalf("can't listent port %d", port)
	}
	defer ln.Close()

	router.name = serverName
	go router.run()
	for {
		var conn, err = ln.Accept()

		if err != nil {
			logger.Println("can't accept connection")
		}
		fmt.Println("Accept connection")
		go handleConnection(conn)
	}

}
