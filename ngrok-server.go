package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var webClientAddr *string = flag.String("l", "localhost:3000", "webclient connection address")
var ngrokClientAddr *string = flag.String("r", "localhost:5000", "ngrok-client connection address")

func main() {
	flag.Parse()
	fmt.Printf("Listening: %v\nProxying: %v\n\n", *webClientAddr, *ngrokClientAddr)

	ngrokClientListener, err := net.Listen("tcp", *ngrokClientAddr)
	if err != nil {
		panic(err)
	}
	defer ngrokClientListener.Close()

	webClientListener, err := net.Listen("tcp", *webClientAddr)
	if err != nil {
		panic(err)
	}
	defer webClientListener.Close()

	for {
		ngrokClientConn, err := ngrokClientListener.Accept()
		log.Println("New ngrok-client connection", ngrokClientConn.RemoteAddr())
		if err != nil {
			log.Println("error accepting connection", err)
		}
		go ngrokClientConnectionHandler(ngrokClientConn, webClientListener)
	}
}

func ngrokClientConnectionHandler(ngrokClientConn net.Conn, webClientListener net.Listener) {
	for {
		webClientConn, err := webClientListener.Accept()
		if err != nil {
			panic(err)
		}
		log.Println("New web-client connection", webClientConn.RemoteAddr())
		go webClientConnectionHandler(ngrokClientConn, webClientConn)
	}
}

func webClientConnectionHandler(ngrokClientConn, webClientConn net.Conn) {
	defer webClientConn.Close()
	webConnCloser := make(chan struct{}, 2)
	go copy(webConnCloser, webClientConn, ngrokClientConn, "web2client")
	go copy(webConnCloser, ngrokClientConn, webClientConn, "client2web")
	<-webConnCloser
	log.Println("Connection complete", webClientConn.RemoteAddr())
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader, tag string) {
	log.Println("start copying", tag)
	_, _ = io.Copy(dst, src)
	log.Println("done copying", tag)
	closer <- struct{}{}
}
