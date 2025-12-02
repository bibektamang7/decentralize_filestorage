package p2p

type HandshakeFunc func(Peer) error

func NoHandShakeFunc(Peer) error { return nil }
