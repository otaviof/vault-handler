package vaulthandler

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	log "github.com/sirupsen/logrus"
)

// File definition on how to write a secret to file-system.
type File struct {
	logger     *log.Entry  // logger
	Group      string      // name of the group
	SecretType string      // secret file type (for kubernetes)
	Properties *SecretData // using SecretData as file properties
	Payload    []byte      // data payload
}

// Zip file payload with gzip.
func (f *File) Zip() error {
	var buffer bytes.Buffer
	var err error

	f.logger.WithField("bytes", len(f.Payload)).Info("Zipping file payload")

	buffer64 := base64.NewEncoder(base64.StdEncoding, &buffer)
	gz := gzip.NewWriter(buffer64)

	if _, err = gz.Write(f.Payload); err != nil {
		return err
	}
	if err = gz.Flush(); err != nil {
		return err
	}
	if err = gz.Close(); err != nil {
		return err
	}
	if err = buffer64.Close(); err != nil {
		return err
	}

	f.Payload = buffer.Bytes()
	f.logger.WithField("bytes", len(f.Payload)).Info("Zipped file payload")
	return nil
}

// Unzip payload.
func (f *File) Unzip() error {
	var buffer *bytes.Buffer
	var reader io.Reader
	var err error

	f.logger.WithField("bytes", len(f.Payload)).Info("Unzipping file payload")

	buffer = bytes.NewBuffer(f.Payload)
	if reader, err = gzip.NewReader(base64.NewDecoder(base64.StdEncoding, buffer)); err != nil {
		return err
	}
	if f.Payload, err = ioutil.ReadAll(reader); err != nil {
		return err
	}

	f.logger.WithField("bytes", len(f.Payload)).Info("Unzipped file payload")
	return nil
}

// Read payload from file-system.
func (f *File) Read(baseDir string) error {
	var err error

	fullPath := f.FilePath(baseDir)
	if !FileExists(fullPath) {
		return fmt.Errorf("can't find file '%s'", fullPath)
	}
	if f.Payload, err = ioutil.ReadFile(fullPath); err != nil {
		return err
	}
	logger := f.logger.WithFields(log.Fields{"path": fullPath, "bytes": len(f.Payload)})
	logger.Info("Reading file content")
	logger.Tracef("Payload: '%s'", f.Payload)

	return nil
}

// Write contents to file-system.
func (f *File) Write(baseDir string) error {
	f.logger.WithFields(log.Fields{
		"name":    f.fileName(),
		"bytes":   len(f.Payload),
		"baseDir": baseDir,
	}).Info("Writing file content")
	return ioutil.WriteFile(f.FilePath(baseDir), f.Payload, 0600)
}

// fileName compose file name based on group and SecretData settings.
func (f *File) fileName() string {
	return fmt.Sprintf("%s.%s.%s", f.Group, f.Properties.Name, f.Properties.Extension)
}

// FilePath joins the infomed base directory with file name.
func (f *File) FilePath(baseDir string) string {
	return path.Join(baseDir, f.fileName())
}

// NewFile instance.
func NewFile(group, secretType string, properties *SecretData, payload []byte) *File {
	return &File{
		logger: log.WithFields(log.Fields{
			"type":      "File",
			"name":      properties.Name,
			"extension": properties.Extension,
		}),
		Group:      group,
		SecretType: secretType,
		Properties: properties,
		Payload:    payload,
	}
}
