package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var ngrokServerAddr *string = flag.String("l", "localhost:5000", "local address")
var webServerAddr *string = flag.String("r", "localhost:8080", "remote address")

func main() {
	flag.Parse()
	fmt.Printf("Listening: %v\nProxying: %v\n\n", *ngrokServerAddr, *webServerAddr)

	ngrokServerConn, err := net.Dial("tcp", *ngrokServerAddr)
	if err != nil {
		panic(err)
	}

	log.Println("New ngrok-server connection", ngrokServerConn.RemoteAddr())
	ngrokServerCloser := make(chan struct{}, 2)
	handleNgrokServerConn(ngrokServerCloser, ngrokServerConn)
	<-ngrokServerCloser
}

func handleNgrokServerConn(ngrokServerCloser chan struct{}, ngrokServerConn net.Conn) {
	defer ngrokServerConn.Close()
	webServerConn, err := net.Dial("tcp", *webServerAddr)
	if err != nil {
		log.Println("error dialing remote addr", err)
		return
	}
	defer webServerConn.Close()

	closer := make(chan struct{}, 2)
	go copy(closer, ngrokServerConn, webServerConn, "server2webserver")
	go copy(closer, webServerConn, ngrokServerConn, "webserver2server")
	<-closer
	log.Println("Connection complete", webServerConn.RemoteAddr())
	ngrokServerCloser <- struct{}{}
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader, tag string) {
	log.Println("start copying", tag)
	_, _ = io.Copy(dst, src)
	log.Println("done copying", tag)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}
