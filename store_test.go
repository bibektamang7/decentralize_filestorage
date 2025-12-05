package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunc(t *testing.T) {
	key := "hello world"
	expectedFilename := "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
	expectedPathname := "2aae6/c35c9/4fcfb/415db/e95f4/08b9c/e91ee/846ed"
	pathkey := CASPathTransformFunc(key)
	assert.Equal(t, expectedPathname, pathkey.PathName)
	assert.Equal(t, expectedFilename, pathkey.FileName)
}

func TestReadStore(t *testing.T) {
	opts := StoreOpts{
		pathTransformFunc: CASPathTransformFunc,
		Root:              "fileStorage",
	}
	s := Store{
		StoreOpts: opts,
	}

	key := "mypictures"

	data := []byte("some of my old pictures")
	size, _, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, size, len(data))
}

func TestDeleteFile(t *testing.T) {
	opts := StoreOpts{
		pathTransformFunc: CASPathTransformFunc,
		Root:              "fileStorage",
	}
	s := Store{
		StoreOpts: opts,
	}

	key := "mypictures"

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		pathTransformFunc: CASPathTransformFunc,
		Root:              "fileStorage",
	}
	s := Store{
		StoreOpts: opts,
	}

	key := "mypictures"
	data := []byte("some of my old pictures")

	if _, err := s.Write(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if !s.Has(key) {
		t.Error("key not exists")
	}
}
