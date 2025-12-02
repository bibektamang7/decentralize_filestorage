package main

import (
	"log"
	"time"

	"github.com/bibektamang7/filestorage/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NoHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fsOpts := FileServerOpts{
		StorageRoot:       listenAddr,
		PathTransformFunc: CASPathTransformFunc,
		BootstrapNodes:    nodes,
		Transport:         tcpTransport,
	}
	fs := NewFileServer(fsOpts)
	return fs
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(1000 * time.Millisecond)
	s2.Start()

	select {}
}
