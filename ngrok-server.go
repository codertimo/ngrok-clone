package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
)

var remoteControlAddr *string = flag.String("c", "localhost:5000", "remote control address")
var remoteDataAddr *string = flag.String("d", "localhost:5001", "remote data address")

var localAddr *string = flag.String("l", ":443", "local address")
var tlsCertFilePath *string = flag.String("tc", "/etc/letsencrypt/live/scatternel.ml/fullchain.pem", "certificate file")
var tlsKeyFilePath *string = flag.String("tk", "/etc/letsencrypt/live/scatternel.ml/privkey.pem", "privated key")

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
	var remoteConn *net.Conn

	http.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		request.Proto = "HTTP/1.1"
		request.ProtoMajor = 1
		request.ProtoMinor = 1

		requestDump, err := httputil.DumpRequest(request, true)
		log.Println(string(requestDump))
		requestReader := bytes.NewReader(requestDump)
		if err != nil {
			http.Error(responseWriter, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		_, err2 := (*remoteConn).Write([]byte("o"))
		if err2 != nil {
			panic(err)
		}

		dataConn := <-dataConnChan
		defer dataConn.Close()

		closer := make(chan signal)
		go copy(closer, dataConn, requestReader)
		// go copy(closer, responseWriter, dataConn)
		response, err := http.ReadResponse(bufio.NewReader(dataConn), request)
		for key, values := range response.Header {
			for _, val := range values {
				responseWriter.Header().Add(key, val)
			}
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(response.Body)
		body_string := buf.String()
		responseWriter.Write([]byte(body_string))
		<-closer

		log.Printf("Data connection close: %s\n", dataConn.LocalAddr())
	})

	go listenTCP(remoteDataAddr, func(dataConn net.Conn) {
		log.Printf("Data connection open: %s\n", dataConn.LocalAddr())
		dataConnChan <- dataConn
	}, listen)

	go listenTCP(remoteControlAddr, func(newRemoteConn net.Conn) {
		if remoteConn != nil {
			(*remoteConn).Close()
		}
		remoteConn = &newRemoteConn

		/* when user request */
		log.Printf("Control connection open: %s\n", newRemoteConn.LocalAddr())
	}, listen)

	err := http.ListenAndServeTLS(*localAddr, *tlsCertFilePath, *tlsKeyFilePath, nil)
	log.Fatal(err)
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
