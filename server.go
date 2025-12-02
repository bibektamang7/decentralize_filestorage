package main

import (
	"fmt"
	"io"
	"log"

	"github.com/bibektamang7/filestorage/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	store  *Store
	qtChan chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		FileServerOpts: opts,
		qtChan:         make(chan struct{}),
	}
}

func (fs *FileServer) Store(key, r io.Reader) error {

	return nil
}

func (fs *FileServer) dialBootstrapNetwork() {
	for _, node := range fs.BootstrapNodes {
		if len(node) == 0 {
			continue
		}
		go func(node string) {
			if err := fs.Transport.Dial(node); err != nil {
				log.Printf("(%s) failed to connect", node)
			}
		}(node)
	}
}
func (fs *FileServer) loop() {
	defer func() {
		fs.Transport.Close()
	}()
	for {
		select {
		case rpc := <-fs.Transport.Consume():
			fmt.Println(rpc)
		case <-fs.qtChan:
			return
		}
	}
}

func (fs *FileServer) Start() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.dialBootstrapNetwork()
	fs.loop()
	return nil
}
