package p2p

import (
	"fmt"
	"log"
	"net"
)

type TCPPeer struct {
	net.Conn
	outbound bool
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}
type TCPTransport struct {
	TCPTransportOpts
	rcvChan chan RCP
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rcvChan:          make(chan RCP),
	}
}

func (t *TCPTransport) Consume() <-chan RCP {
	return t.rcvChan
}

// func (t *TCPTransport) Close() error {
// }

func (t *TCPTransport) ListenAndAccept() error {
	listener, err := net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	fmt.Println("TCP coonection is open in port : ", t.ListenAddr)
	go t.handleAccept(listener)
	return nil
}

func (t *TCPTransport) handleAccept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("new connection failed")
		}
		go t.handleConnection(conn)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
	}()
	peer := TCPPeer{Conn: conn, outbound: false}

	if err := t.HandshakeFunc(peer); err != nil {
		log.Println("handshake failed")
		return
	}

	rcp := RCP{}
	for {
		if err := t.Decoder.Decode(conn, rcp); err != nil {
			log.Println("connection message decoding failed")
			continue
		}
		rcp.From = conn.RemoteAddr().Network()
		t.rcvChan <- rcp
	}
}
