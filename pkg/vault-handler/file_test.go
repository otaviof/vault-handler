package vaulthandler

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/stretchr/testify/assert"
)

var file *File

func zipBytes(payload []byte) ([]byte, error) {
	var buffer bytes.Buffer
	var err error

	gz := gzip.NewWriter(&buffer)

	if _, err = gz.Write(payload); err != nil {
		return nil, err
	}
	if err = gz.Flush(); err != nil {
		return nil, err
	}
	if err = gz.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func TestFileNewFile(t *testing.T) {
	zipped, err := zipBytes([]byte("payload"))
	assert.Nil(t, err)
	assert.NotNil(t, zipped)
	assert.True(t, len(zipped) > 0)

	file = NewFile("/var/tmp/vault-handler.test", zipped)
	assert.Equal(t, zipped, file.payload)
}

func TestFileUnzip(t *testing.T) {
	err := file.Unzip()

	assert.Nil(t, err)
	assert.Equal(t, []byte("payload"), file.payload)
}
