package vaulthandler

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	foo = "bar"
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

	t.Logf("Role-ID: '%s', Secret-ID: '%s'", roleID, secretID)
	if roleID == "" || secretID == "" {
		t.Fatalf("Can't find role-id ('%s'), secret-id ('%s') in the environment", roleID, secretID)
	}

	err := vault.AppRoleAuth(roleID, secretID)

	assert.Nil(t, err)
}

func TestVaultWrite(t *testing.T) {
	err := vault.Write("secret/data/foo/bar/baz", map[string]interface{}{"foo": foo})

	assert.Nil(t, err)
}

func TestVaultRead(t *testing.T) {
	out, err := vault.Read("secret/data/foo/bar/baz", "foo")

	assert.Nil(t, err)
	assert.Equal(t, string(out), foo)
}

func TestVaultComposePath(t *testing.T) {
	path := vault.composePath(SecretData{Name: "name", Extension: "ext"},
		handlerManifest.Secrets[groupName].Path)
	assert.Equal(t, handlerManifest.Secrets[groupName].Path, path)

	data := SecretData{Name: "name", Extension: "ext", NameAsSubPath: true}
	path = vault.composePath(data, handlerManifest.Secrets[groupName].Path)
	expect := fmt.Sprintf("%s/%s", handlerManifest.Secrets[groupName].Path, data.Name)
	assert.Equal(t, expect, path)
}
