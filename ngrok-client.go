package main

import (
	"flag"
	"io"
	"log"
	"net"
)

var ngrokControlAddr *string = flag.String("l", "localhost:5000", "ngrok")
var ngrokDataAddr *string = flag.String("l", "localhost:5001", "local address")
var webServerAddr *string = flag.String("r", "localhost:8080", "remote address")

func main() {
	flag.Parse()

	ngrokControlConn, err := net.Dial("tcp", *ngrokControlAddr)
	if err != nil {
		log.Println("error dialing ngrok server addr", err)
		return
	}
	recvBuf := make([]byte, 4096)
	for {
		numBytes, err := ngrokControlConn.Read(recvBuf)
		if err != nil {
			if io.EOF == err {
				log.Printf("connection is closed from client; %v", ngrokControlConn.RemoteAddr().String())
				return
			}
			log.Printf("fail to receive data; err: %v", err)
			return
		}
		if numBytes > 0 {
			if string(recvBuf[:numBytes]) == "o" {
				go conncectToNgrokServer()
			}
		}
	}
}

func conncectToNgrokServer() {
	ngrokDataConn, err := net.Dial("tcp", *ngrokDataAddr)
	if err != nil {
		log.Println("error dialing ngrok server addr", err)
		return
	}
	webServerConn, err := net.Dial("tcp", *webServerAddr)
	if err != nil {
		log.Println("error dialing remote addr", err)
		return
	}
	closer := make(chan struct{}, 2)
	go copy(closer, ngrokDataConn, webServerConn)
	go copy(closer, webServerConn, ngrokDataConn)
	<-closer
	webServerConn.Close()
	ngrokDataConn.Close()
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}
