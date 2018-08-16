package main

import (
	"flag"

	"gitlab.com/patsak/gpircd/server"
)

func main() {
	port := flag.Int("p", 6667, "listen port")
	name := flag.String("name", "irc.example.net", "server hostname")
	flag.Parse()

	server.Start(*port, *name)
}
