package main

import (
	"fmt"
	"log"

	"github.com/bibektamang7/filestorage/p2p"
)

type FileServerOpts struct {
	Root              string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	Store             *Store
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts
	qtChan chan struct{}
}

func NewServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		FileServerOpts: opts,
		qtChan:         make(chan struct{}),
	}
}

func (fs *FileServer) dialBootstrapNetwork() {
	for _, node := range fs.BootstrapNodes {
		if err := fs.Transport.Dial(node); err != nil {
			log.Printf("(%s) failed to connect", node)
		}
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

	fs.dialBootstrapNodes()
	fs.loop()
	return nil
}
