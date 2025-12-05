package p2p

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type TCPPeer struct {
	net.Conn
	outbound bool

	wg *sync.WaitGroup
}

func (p TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}
func (p TCPPeer) CloseStream() {
	p.wg.Done()
}
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
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

func (t *TCPTransport) Addr() string {
	return t.ListenAddr
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
		fmt.Println("new connection", conn.RemoteAddr().String())
		go t.handleConnection(conn, false)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn, outbound bool) {
	defer func() {
		fmt.Printf("(%s) connection closed with (%s)\n", conn.RemoteAddr().String(), t.ListenAddr)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, outbound)

	if err := t.HandshakeFunc(peer); err != nil {
		log.Println("handshake failed")
		return
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			log.Println("failed to handle OnPeer")
			return
		}
	}

	for {
		rcp := RCP{}
		if err := t.Decoder.Decode(conn, &rcp); err != nil {
			log.Println("connection message decoding failed")
			return
		}

		rcp.From = conn.RemoteAddr().String()
		if rcp.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting ...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}
		t.rcpChan <- rcp
	}
}
