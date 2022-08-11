package main

import (
	"log"
	"net"
)

func main() {
	server()
}

func server() {
	listen, err := net.Listen("tcp4", "192.168.3.100:445")
	if err != nil {
		log.Fatalln(err)
	}
	for {
		_, _ = listen.Accept()
	}
}

func client() {
	_, err := net.Dial("tcp4", "127.0.0.2:445")
	if err != nil {
		log.Fatal(err)
	}
}
