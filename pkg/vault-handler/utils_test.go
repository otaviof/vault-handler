package vaulthandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtilsFileExists(t *testing.T) {
	found := fileExists("../../test/manifest.yaml")
	assert.True(t, found)
}

func TestUtilsReadFile(t *testing.T) {
	fileBytes := readFile("../../test/manifest.yaml")
	assert.True(t, len(fileBytes) > 0)
}

func TestUtilsIsDir(t *testing.T) {
	found := isDir("../../test")
	assert.True(t, found)
}
