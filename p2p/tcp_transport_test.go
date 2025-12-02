package p2p

import "testing"

func TestTCPTransport(t *testing.T) {
	tcpOpts := TCPTransportOpts{
		ListenAddr: ":3000",
	}
	tcp := NewTCPTransport(tcpOpts)
	tcp.ListenAndAccept()
}
