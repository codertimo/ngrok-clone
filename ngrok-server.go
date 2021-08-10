package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
)

var remoteControlAddr *string = flag.String("c", "localhost:5000", "remote control address")
var remoteDataAddr *string = flag.String("d", "localhost:5001", "remote data address")

var localAddr *string = flag.String("l", "localhost:4000", "local address")

var certFile *string = flag.String("cert", "", "certificate file path")
var keyFile *string = flag.String("key", "", "private key for TLS")

type signal = struct{}

func main() {
	flag.Parse()

	log.Println("Remote Control Address:", *remoteControlAddr)
	log.Println("Remote Data Address:", *remoteDataAddr)
	log.Println("Local Address:", *localAddr)

	listen := createConnectionMaker()

	dataConnChan := make(chan net.Conn)
	go listenTCP(remoteDataAddr, func(dataConn net.Conn) {
		log.Printf("Data connection open: %s\n", dataConn.LocalAddr())
		dataConnChan <- dataConn
	}, listen)

	var remoteConn *net.Conn

	go listenTCP(localAddr, func(localConn net.Conn) {
		log.Printf("Local request: %s\n", localConn.LocalAddr())

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
	}, listenRawTCP)

	listenTCP(remoteControlAddr, func(newRemoteConn net.Conn) {
		if remoteConn != nil {
			(*remoteConn).Close()
		}
		remoteConn = &newRemoteConn

		/* when user request */
		log.Printf("Control connection open: %s\n", newRemoteConn.LocalAddr())
	}, listen)
}

func copy(closer chan signal, dst io.Writer, src io.Reader) {
	io.Copy(dst, src)
	closer <- signal{}
}

type connectionMaker = func(*string) (net.Listener, error)

func listenRawTCP(addr *string) (net.Listener, error) {
	return net.Listen("tcp", *addr)
}

func createConnectionMaker() connectionMaker {
	if *certFile == "" && *keyFile == "" {
		return listenRawTCP
	}

	cer, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		panic(err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	return func(addr *string) (net.Listener, error) {
		return tls.Listen("tcp", *addr, config)
	}
}

func listenTCP(addr *string, handler func(net.Conn), listen connectionMaker) {
	listener, err := listen(addr)
	if err != nil {
		log.Fatal(err)
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
