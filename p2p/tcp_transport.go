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
	listener net.Listener
	rcpChan  chan RCP
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rcpChan:          make(chan RCP),
	}
}

func (t *TCPTransport) Consume() <-chan RCP {
	return t.rcpChan
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dial(remoteAddr string) error {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}
	fmt.Printf("(connected to) : %s\n", remoteAddr)
	go t.handleConnection(conn, true)
	return nil
}

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
		go t.handleConnection(conn, false)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn, outbound bool) {
	defer func() {
		fmt.Printf("(%s) connection closed with (%s)", conn.RemoteAddr().String(), t.ListenAddr)
		conn.Close()
	}()
	peer := TCPPeer{Conn: conn, outbound: outbound}

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
		t.rcpChan <- rcp
	}
}
