package main

import (
	"bufio"
	"bytes"
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

type signal = struct{}

func main() {
	flag.Parse()

	log.Println("Remote Control Address:", *remoteControlAddr)
	log.Println("Remote Data Address:", *remoteDataAddr)
	log.Println("Local Address:", *localAddr)

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
	})

	go listenTCP(remoteControlAddr, func(newRemoteConn net.Conn) {
		if remoteConn != nil {
			(*remoteConn).Close()
		}
		remoteConn = &newRemoteConn

		/* when user request */
		log.Printf("Control connection open: %s\n", newRemoteConn.LocalAddr())
	})

	err := http.ListenAndServeTLS(*localAddr, *tlsCertFilePath, *tlsKeyFilePath, nil)
	log.Fatal(err)
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
