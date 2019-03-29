package vaulthandler

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	log "github.com/sirupsen/logrus"
)

// File definition on how to write a secret to file-system.
type File struct {
	logger     *log.Entry  // logger
	group      string      // name of the group
	properties *SecretData // using SecretData as file properties
	payload    []byte      // data payload
}

// Zip file payload with gzip.
func (f *File) Zip() error {
	var buffer bytes.Buffer
	var err error

	f.logger.WithField("bytes", len(f.payload)).Info("Zipping file payload")
	gz := gzip.NewWriter(&buffer)

	if _, err = gz.Write(f.payload); err != nil {
		return err
	}
	if err = gz.Flush(); err != nil {
		return err
	}
	if err = gz.Close(); err != nil {
		return err
	}
	f.payload = buffer.Bytes()
	f.logger.WithField("bytes", len(f.payload)).Info("Zipped file payload")

	return nil
}

// Unzip payload.
func (f *File) Unzip() error {
	var reader io.Reader
	var bufferOut bytes.Buffer
	var err error

	f.logger.WithField("bytes", len(f.payload)).Info("Unzipping file payload")

	bufferIn := bytes.NewBuffer(f.payload)
	if reader, err = gzip.NewReader(bufferIn); err != nil {
		return err
	}
	if _, err = bufferOut.ReadFrom(reader); err != nil {
		return err
	}

	f.payload = bufferOut.Bytes()
	f.logger.WithField("bytes", len(f.payload)).Info("Unzipped file payload")

	return nil
}

// Read payload from file-system.
func (f *File) Read(baseDir string) error {
	var err error

	fullPath := f.FilePath(baseDir)
	if !fileExists(fullPath) {
		return fmt.Errorf("can't find file '%s'", fullPath)
	}
	if f.payload, err = ioutil.ReadFile(fullPath); err != nil {
		return err
	}
	f.logger.WithFields(log.Fields{"path": fullPath, "bytes": len(f.payload)}).
		Info("Reading file content")

	return nil
}

// Write contents to file-system.
func (f *File) Write(baseDir string) error {
	f.logger.WithFields(log.Fields{"name": f.fileName(), "bytes": len(f.payload)}).
		Info("Reading file content")
	return ioutil.WriteFile(f.FilePath(baseDir), f.payload, 0600)
}

// Name exposes the file name from properties.
func (f *File) Name() string {
	return f.properties.Name
}

// fileName compose file name based on group and SecretData settings.
func (f *File) fileName() string {
	return fmt.Sprintf("%s.%s.%s", f.group, f.properties.Name, f.properties.Extension)
}

// FilePath joins the infomed base directory with file name.
func (f *File) FilePath(baseDir string) string {
	return path.Join(baseDir, f.fileName())
}

// Payload returns the file payload as slice of bytes.
func (f *File) Payload() []byte {
	return f.payload
}

// NewFile instance.
func NewFile(group string, properties *SecretData, payload []byte) *File {
	return &File{
		logger: log.WithFields(log.Fields{
			"type":      "File",
			"name":      properties.Name,
			"extension": properties.Extension,
		}),
		group:      group,
		properties: properties,
		payload:    payload,
	}
}
