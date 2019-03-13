package vaulthandler

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
)

// File definition on how to write a secret to file-system.
type File struct {
	path    string // full path, including name and extension
	payload []byte // data payload
}

// Unzip payload.
func (f *File) Unzip() error {
	var reader io.Reader
	var bufferOut bytes.Buffer
	var err error

	bufferIn := bytes.NewBuffer(f.payload)
	if reader, err = gzip.NewReader(bufferIn); err != nil {
		return err
	}

	if _, err = bufferOut.ReadFrom(reader); err != nil {
		return err
	}

	f.payload = bufferOut.Bytes()
	return nil
}

// Write contents to file-system.
func (f *File) Write() error {
	return ioutil.WriteFile(f.path, f.payload, 0600)
}

// NewFile instance.
func NewFile(path string, payload []byte) *File {
	return &File{path: path, payload: payload}
}
