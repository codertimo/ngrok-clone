package mainsadfasf

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var ngrokControlAddr *string = flag.String("l", "localhost:5001", "local address")
var ngrokDataAddr *string = flag.String("l", "localhost:5002", "local address")
var webServerAddr *string = flag.String("r", "localhost:8080", "remote address")

func main() {
	flag.Parse()
	fmt.Printf("Listening: %v\nProxying: %v\n\n", *ngrokControlAddr, *webServerAddr)

	controlConnection, err := net.Dial("tcp", *ngrokControlAddr)
	if err != nil {
		panic(err)
	}
	defer controlConnection.Close()

	log.Println("New ngrok control connection", controlConnection.RemoteAddr())
	recvBuf := make([]byte, 4096)
	for {
		n, err := controlConnection.Read(recvBuf)
		if nil != err {
			if io.EOF == err {
				log.Printf("connection is closed from client; %v", controlConnection.RemoteAddr().String())
				break
			}
			log.Printf("fail to receive data; err: %v", err)
			break
		}
		if 0 < n {
			go handleDataTransfer()
		}
	}
}

func handleDataTransfer() {
	ngrokServerDataConn, err := net.Dial("tcp", *ngrokServerAddr)
	if err != nil {
		panic(err)
	}
	defer ngrokServerDataConn.Close()

	webServerDataConn, err := net.Dial("tcp", *webServerAddr)
	if err != nil {
		panic(err)
	}
	defer webServerDataConn.Close()

	closer := make(chan struct{}, 2)
	go copy(closer, ngrokServerDataConn, webServerDataConn, "server2webserver")
	go copy(closer, webServerDataConn, ngrokServerDataConn, "webserver2server")
	<-closer
	log.Println("Connection complete", ngrokServerDataConn.RemoteAddr())
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader, tag string) {
	log.Println("start copying", tag)
	_, _ = io.Copy(dst, src)
	log.Println("done copying", tag)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}
