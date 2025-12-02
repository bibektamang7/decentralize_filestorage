package p2p

type Peer interface {
}

type Transport interface {
	Consume() <-chan RCP
	ListenAndAccept() error
	Close() error
}
