package vaulthandler

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var vault *Vault

func TestVaultNewVault(t *testing.T) {
	var err error

	vault, err = NewVault("http://127.0.0.1:8200")

	assert.Nil(t, err)
	assert.NotNil(t, vault)
}

func TestVaultAppRoleAuth(t *testing.T) {
	roleID := os.Getenv("VAULT_HANDLER_VAULT_ROLE_ID")
	secretID := os.Getenv("VAULT_HANDLER_VAULT_SECRET_ID")

	log.Printf("Role-ID: '%s', Secret-ID: '%s'", roleID, secretID)
	err := vault.AppRoleAuth(roleID, secretID)

	assert.Nil(t, err)
}

func TestVaultRead(t *testing.T) {
	out, err := vault.Read("secret/data/foo/bar/baz", "foo")

	assert.Nil(t, err)
	log.Printf("out: '%#v'", out)
}
