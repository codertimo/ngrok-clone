package main

import (
	"flag"
	"io"
	"log"
	"net"
)

var remoteControlAddr *string = flag.String("c", "localhost:5000", "remote control address")
var remoteDataAddr *string = flag.String("d", "localhost:5001", "remote data address")

var localAddr *string = flag.String("l", "localhost:4000", "local address")

type signal = struct{}

func main() {
	flag.Parse()

	// fmt.Printf("Listening: %v\nProxying: %v\n\n", *localAddr, *remoteAddr)
	dataConnChan := make(chan net.Conn)
	go listenTCP(remoteDataAddr, func(dataConn net.Conn) {
		log.Printf("Data connection open: %s\n", dataConn.RemoteAddr())
		dataConnChan <- dataConn
	})

	var remoteConn *net.Conn

	go listenTCP(localAddr, func(localConn net.Conn) {
		log.Printf("Local request: %s\n", localConn.RemoteAddr())

		/* when developer request */
		defer localConn.Close()

		_, err := (*remoteConn).Write([]byte("o"))
		if err != nil {
			panic(err)
		}

		dataConn := <-dataConnChan
		defer dataConn.Close()

		closer := make(chan signal)
		go copy(closer, dataConn, localConn)
		go copy(closer, localConn, dataConn)
		<-closer
	})

	listenTCP(remoteControlAddr, func(newRemoteConn net.Conn) {
		if remoteConn != nil {
			(*remoteConn).Close()
		}
		remoteConn = &newRemoteConn

		/* when user request */
		log.Printf("Control connection open: %s\n", newRemoteConn.RemoteAddr())
	})
}

func copy(closer chan signal, dst io.Writer, src io.Reader) {
	io.Copy(dst, src)
	closer <- signal{}
}

func listenTCP(addr *string, handler func(net.Conn)) {
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error accepting connection", err)
			continue
		}

		go handler(conn)
	}
}
