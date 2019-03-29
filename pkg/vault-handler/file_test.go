package vaulthandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var file *File

const payload = "payload"

func TestFileNewFile(t *testing.T) {
	secretData := &SecretData{Name: "file", Extension: "text"}
	file = NewFile("test", secretData, []byte(payload))
	assert.True(t, len(file.payload) > 0)
}

func TestFileZip(t *testing.T) {
	err := file.Zip()

	assert.Nil(t, err)
	assert.True(t, len(file.payload) > 0)
}

func TestFileUnzip(t *testing.T) {
	err := file.Unzip()

	assert.Nil(t, err)
	assert.Equal(t, []byte(payload), file.payload)
}
