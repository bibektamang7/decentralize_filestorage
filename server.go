package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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

	peerLock sync.Mutex
	peers    map[string]p2p.Peer
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		pathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		qtChan:         make(chan struct{}),
		peers:          map[string]p2p.Peer{},
		store:          NewStore(storeOpts),
	}
}

func (fs *FileServer) Stop() {
	close(fs.qtChan)
}

func (fs *FileServer) onPeer(p p2p.Peer) error {
	fs.peerLock.Lock()
	defer fs.peerLock.Unlock()
	fs.peers[p.RemoteAddr().String()] = p
	fmt.Printf("(%s) connected with remote %s\n", fs.Transport.Addr(), p.RemoteAddr().String())
	return nil
}

type Message struct {
	From    string
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}
type MessageGetFile struct {
	Key string
}

func (fs *FileServer) StoreFile(key string, r io.Reader) error {
	buf := new(bytes.Buffer)

	tee := io.TeeReader(r, buf)
	size, err := fs.store.Write(key, tee)
	if err != nil {
		return err
	}

	// pr, pw := io.Pipe()
	// go func() {
	// 	defer pw.Close()
	// 	n, err := fs.store.Write(key, io.TeeReader(r, pw))
	// 	sizeChan <- n
	// 	if err != nil {
	// 		log.Println("failed to write file to disk", err)
	// 	}
	// }()
	// log.Println("is here ")
	// size := <-sizeChan
	msg := Message{
		From: fs.Transport.Addr(),
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	if err := fs.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	peersWriters := []io.Writer{}
	for _, peer := range fs.peers {
		// peer.Send([]byte{p2p.IncomingStream})
		// data := buf.Bytes()
		// n, err := io.Copy(peer, bytes.NewReader(data))
		// if err != nil {
		// 	log.Println("peer failed to get data: ", addr)
		// 	return err
		// }
		// fmt.Printf("written %d to peer %s\n", n, addr)
		peersWriters = append(peersWriters, peer)
	}

	mw := io.MultiWriter(peersWriters...)
	mw.Write([]byte{p2p.IncomingStream})
	if _, err := io.Copy(mw, buf); err != nil {
		return fmt.Errorf("failed to stream file to prees: %w", err)
	}
	return nil
}

func (fs *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for addr, peer := range fs.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			log.Printf("sending failed to %s\n", addr)
			continue
		}
	}
	return nil
}
func (fs *FileServer) handleMessages(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return fs.handleMessageStoreFile(from, v)
	case MessageGetFile:
		fs.handleMessageGetFile(from, v.Key)

	}
	return nil
}
func (fs *FileServer) Get(key string) (io.Reader, error) {
	if fs.store.Has(key) {
		fmt.Printf("%s severing file %s from local disk\n", fs.Transport.Addr(), key)
		_, r, err := fs.store.Read(key)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	fmt.Printf("[%s] don't have file (%s) locally, fetching from newtokr...\n", fs.Transport.Addr(), key)
	msg := Message{
		From: fs.Transport.Addr(),
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := fs.broadcast(&msg); err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond * 500)
	for _, peer := range fs.peers {
		var fileSize int64
		if err := binary.Read(peer, binary.LittleEndian, &fileSize); err != nil {
			continue
		}
		if fs.store.Has(key) {
			fmt.Println("already exist")
			break
		}
		n, err := fs.store.Write(key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] received (%d) bytes over the network from (%s) ", fs.Transport.Addr(), n, peer.RemoteAddr())
		peer.CloseStream()
	}
	_, r, err := fs.store.Read(key)
	return r, err
}

func (fs *FileServer) handleMessageGetFile(from string, key string) error {
	if !fs.store.Has(key) {
		return fmt.Errorf("%s doesn't have file with key %s", fs.Transport.Addr(), key)
	}
	size, r, err := fs.store.Read(key)
	if err != nil {
		return err
	}
	defer r.Close()

	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found in %s", from, fs.Transport.Addr())
	}

	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, size)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("(%s) written %d bytes over the network to %s\n", fs.Transport.Addr(), n, from)

	return nil
}
func (fs *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be fond in the peer list", from)
	}

	n, err := fs.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	fmt.Printf("(%s) written %d bytes to disk\n", fs.Transport.Addr(), n)
	// peer.CloseStream()
	return nil
}

func (fs *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to error or user quite action")
		fs.Transport.Close()
	}()
	for {
		select {
		case rpc := <-fs.Transport.Consume():
			var msg Message

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoder failed: ", err)
			}
			if err := fs.handleMessages(rpc.From, &msg); err != nil {
				log.Println("handle message error: ", err)
			}
		case <-fs.qtChan:
			log.Println("close change occur")
			return
		}
	}
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

func (fs *FileServer) Start() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.dialBootstrapNetwork()
	fs.loop()
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
