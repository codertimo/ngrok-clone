package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var remoteAddr *string = flag.String("r", "localhost:5000", "remote address")
var localAddr *string = flag.String("l", "localhost:4000", "local address")

func main() {
	flag.Parse()

	fmt.Printf("Listening: %v\nProxying: %v\n\n", *localAddr, *remoteAddr)

	listener, err := net.Listen("tcp", *localAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error accepting connection", err)
			continue
		}
		go func() {
			log.Println("Server Connection")
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
