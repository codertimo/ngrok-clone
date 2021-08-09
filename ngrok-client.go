package main

import (
	"flag"
	"io"
	"log"
	"net"
)

var ngrokControlAddr *string = flag.String("", "localhost:5000", "ngrok control address")
var ngrokDataAddr *string = flag.String("l", "localhost:5001", "ngrok data address")
var webServerAddr *string = flag.String("r", "localhost:8080", "webserver address")

type signal = struct{}

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
				log.Printf("connection is closed from client; %v", ngrokControlConn.RemoteAddr())
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

	log.Printf("data proxy start %s -> %s\n", ngrokDataConn.RemoteAddr(), webServerConn.RemoteAddr())
	closer := make(chan signal, 2)
	go copy(closer, ngrokDataConn, webServerConn)
	go copy(closer, webServerConn, ngrokDataConn)
	<-closer
	webServerConn.Close()
	ngrokDataConn.Close()
	log.Printf("data proxy closed %s -> %s\n", ngrokDataConn.RemoteAddr(), webServerConn.RemoteAddr())
}

func copy(closer chan signal, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- signal{}
}
