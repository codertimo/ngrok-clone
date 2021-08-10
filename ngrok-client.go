package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
)

var ngrokControlAddr *string = flag.String("c", "localhost:5000", "ngrok control address")
var ngrokDataAddr *string = flag.String("d", "localhost:5001", "ngrok data address")
var webServerAddr *string = flag.String("w", "localhost:8080", "webserver address")

var isSupportTls *bool = flag.Bool("t", false, "tls support")
var isAllowInsecure *bool = flag.Bool("k", false, "Allow insecure server connections when using SSL")

type signal = struct{}

func main() {
	flag.Parse()

	var dial func(*string) (net.Conn, error)
	if *isSupportTls {
		conf := &tls.Config{InsecureSkipVerify: *isAllowInsecure}
		dial = func(addr *string) (net.Conn, error) {
			return tls.Dial("tcp", *addr, conf)
		}
	} else {
		dial = func(addr *string) (net.Conn, error) {
			return net.Dial("tcp", *addr)
		}
	}

	ngrokControlConn, err := dial(ngrokControlAddr)
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
				go conncectToNgrokServer(dial)
			} else {
				log.Println("wrong buf", string(recvBuf[:numBytes]))
			}
		}
	}
}

func conncectToNgrokServer(dial func(*string) (net.Conn, error)) {
	ngrokDataConn, err := dial(ngrokDataAddr)
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
