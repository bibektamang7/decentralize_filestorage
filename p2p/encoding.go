package p2p

import (
	"encoding/gob"
	"fmt"
	"io"
)

type Decoder interface {
	Decode(io.Reader, RCP) error
}

type GOBDecoder struct{}

func (g GOBDecoder) Decode(r io.Reader, msg RCP) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (d DefaultDecoder) Decode(r io.Reader, msg RCP) error {
	buf := make([]byte, 1024)

	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	fmt.Printf("read from connection: %s", string(buf[:n]))
	return nil
}
