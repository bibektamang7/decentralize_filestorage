package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

type Pathkey struct {
	PathName string
	FileName string
}

type PathTransformFunc func(string) Pathkey

func defaultTransformFunc(key string) Pathkey {
	return Pathkey{
		PathName: key,
		FileName: key,
	}
}

func CASPathTransformFunc(key string) Pathkey {
	byteHash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(byteHash[:])
	blockSize := 5

	sliceLen := len(hashStr) / blockSize

	parts := make([]string, sliceLen)

	for i := range sliceLen {
		from, to := i*blockSize, blockSize+i*blockSize
		parts[i] = hashStr[from:to]
	}

	return Pathkey{
		PathName: strings.Join(parts, "/"),
		FileName: hashStr,
	}
}

type StoreOpts struct {
	Root              string
	pathTransformFunc PathTransformFunc
}
type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if len(opts.Root) == 0 {
		opts.Root = "defaultStoreRoot"
	}
	if opts.pathTransformFunc == nil {
		opts.pathTransformFunc = defaultTransformFunc
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Read(key string, buf []byte) (int, error) {
	return s.readStream(key, buf)
}

func (s *Store) readStream(key string, buf []byte) (int, error) {
	pathkey := s.pathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.PathName)
	fileNameWithFullPath := fmt.Sprintf("%s/%s", pathNameWithRoot, pathkey.FileName)
	f, err := os.Open(fileNameWithFullPath)
	if err != nil {
		return 0, err
	}
	return f.Read(buf)
}

func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathkey := s.pathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return err
	}
	fileNameWithFullPath := fmt.Sprintf("%s/%s", pathNameWithRoot, pathkey.FileName)
	f, err := os.Create(fileNameWithFullPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}
