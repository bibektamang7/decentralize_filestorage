package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bibektamang7/filestorage/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	fsOpts := FileServerOpts{
		StorageRoot:       "network" + strings.ReplaceAll(listenAddr, ":", "_"),
		PathTransformFunc: CASPathTransformFunc,
		BootstrapNodes:    nodes,
	}
	fs := NewFileServer(fsOpts)

	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NoHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransportOpts.OnPeer = fs.onPeer
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fs.Transport = tcpTransport

	return fs
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	s3 := makeServer(":5000", ":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()
	go func() {
		log.Fatal(s2.Start())
	}()
	time.Sleep(1000 * time.Millisecond)
	go s3.Start()

	time.Sleep(200 * time.Millisecond)

	key := "helloMyPro"
	// data := []byte("This is a data to be stored with the key helloMyPro.")

	// if err := s3.StoreFile(key, bytes.NewReader(data)); err != nil {
	// 	fmt.Println("failed to store file", err)
	// }

	r, err := s3.Get(key)

	if err != nil {
		log.Fatal("failed to read data with key", err)
	}

	buf := make([]byte, 1020)
	n, err := r.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("at least we read it", string(buf[:n]))

	select {}
}
