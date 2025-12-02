package main

import "github.com/bibektamang7/filestorage/p2p"

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		Decoder:       p2p.DefaultDecoder{},
		HandshakeFunc: p2p.NoHandShakeFunc,
	}
	tcp := p2p.NewTCPTransport(tcpOpts)
	tcp.ListenAndAccept()
	select {}
}
