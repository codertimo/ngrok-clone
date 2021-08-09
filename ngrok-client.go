package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var localAddr *string = flag.String("l", "localhost:5000", "local address")
var remoteAddr *string = flag.String("r", "localhost:8080", "remote address")

func main() {
	flag.Parse()

	fmt.Printf("Listening: %v\nProxying: %v\n\n", *localAddr, *remoteAddr)

	listener, err := net.Listen("tcp", *localAddr)
	if err != nil {
		panic(err)
	}
	for {

		log.Println("Client Connection")
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error accepting connection", err)
			continue
		}
		go func() {
			conn2, err := net.Dial("tcp", *remoteAddr)
			if err != nil {
				log.Println("error dialing remote addr", err)
				return
			}
			go io.Copy(conn2, conn)
			io.Copy(conn, conn2)
			conn2.Close()
			conn.Close()
		}()
	}
}
