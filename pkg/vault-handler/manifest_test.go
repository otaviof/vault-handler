package vaulthandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var manifest *Manifest

func TestManifestNew(t *testing.T) {
	var err error

	manifest, err = NewManifest("../../test/manifest.yaml")

	assert.NotNil(t, manifest)
	assert.Nil(t, err)
}
