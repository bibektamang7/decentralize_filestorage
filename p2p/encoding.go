package p2p

import (
	"encoding/gob"
	"io"
	"log"
)

type Decoder interface {
	Decode(io.Reader, *RCP) error
}

type GOBDecoder struct{}

func (g GOBDecoder) Decode(r io.Reader, msg *RCP) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (d DefaultDecoder) Decode(r io.Reader, msg *RCP) error {
	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil {
		log.Println(" is it this")
		return err
	}
	stream := peekBuf[0] == IncomingStream
	if stream {
		msg.Stream = stream
		return nil
	}
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	msg.Payload = buf[:n]
	return nil
}
