package p2p

import (
	"fmt"
	"net"
)

type TCPTransport struct {
	listenAddr string
	listner    net.Listener
}

func NewTCPTransport() *TCPTransport {
	return &TCPTransport{
		listenAddr: ":3000",
	}
}
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listner, err = net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	go t.StartAcceptLoop()
	return nil
}

func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.listner.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		fmt.Printf("New incomming connection %+v\n", conn)
		go t.HandleConnection(conn)
	}
}

func (t *TCPTransport) HandleConnection(conn net.Conn) {
	fmt.Printf("TCP CONNECTED")
}
