package p2p

import "net"

type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

type Transport interface {
	Consume() <-chan RCP
	Addr() string
	ListenAndAccept() error
	Close() error
	Dial(string) error
}
