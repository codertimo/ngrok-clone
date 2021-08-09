package main

import (
	"flag"
	"io"
	"log"
	"net"
	"sync"
)

var remoteControlAddr *string = flag.String("c", "localhost:5000", "remote control address")
var remoteDataAddr *string = flag.String("d", "localhost:5001", "remote data address")

var localAddr *string = flag.String("l", "localhost:4000", "local address")

var remoteConnection net.Conn = nil

type signal = struct{}

func main() {
	flag.Parse()

	// fmt.Printf("Listening: %v\nProxying: %v\n\n", *localAddr, *remoteAddr)

	listenTCP(remoteControlAddr, func(remoteConn net.Conn) {
		defer remoteConn.Close()

		/* when user request */
		log.Printf("Remote connection open: %v\n", *remoteControlAddr)

		dataConnChan := make(chan net.Conn)

		var wg sync.WaitGroup
		wg.Add(2)

		go (func() {
			defer wg.Done()
			listenTCP(localAddr, func(localConn net.Conn) {
				/* when developer request */
				defer localConn.Close()

				_, err := remoteConn.Write([]byte("o"))
				if err != nil {
					panic(err)
				}

				dataConn := <-dataConnChan
				defer dataConn.Close()

				closer := make(chan signal)
				go copy(closer, dataConn, localConn)
				go copy(closer, localConn, dataConn)
				<-closer
				<-closer
			})
		})()

		go (func() {
			defer wg.Done()
			listenTCP(remoteDataAddr, func(dataConn net.Conn) {
				dataConnChan <- dataConn
			})
		})()

		wg.Wait()
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
