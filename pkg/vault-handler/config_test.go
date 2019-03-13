package vaulthandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	config := &Config{}
	err := config.Validate()
	assert.NotNil(t, err)

	config.OutputDir = "../../test"
	config.VaultAddr = "http://127.0.0.1:8200"
	config.VaultToken = "token"

	err = config.Validate()
	assert.Nil(t, err)
}
