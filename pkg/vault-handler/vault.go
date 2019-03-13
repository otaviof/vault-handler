package vaulthandler

import (
	"errors"
	"fmt"
	"log"

	vaultapi "github.com/hashicorp/vault/api"
)

// Vault represent Vault server and the actions it can receive.
type Vault struct {
	client *vaultapi.Client
	token  string
}

// extractKey coming from Read method, where the user can choose one key to be taken out of the data
// read from Vault.
func (v *Vault) extractKey(payload map[string]interface{}, key string) ([]byte, error) {
	var data string
	var exists bool

	if _, exists = payload["data"]; exists {
		log.Print("[Vault] Using V2 API style, extracting 'data' from payload.")
		payload = payload["data"].(map[string]interface{})
	}

	if data, exists = payload[key].(string); !exists {
		return nil, fmt.Errorf("cannot extract key '%s' from vault payload", key)
	}

	dataAsBytes := []byte(data)
	log.Printf("Obtained '%d' bytes from key '%s'", len(dataAsBytes), key)
	return dataAsBytes, nil
}

// Read data from a given vault path and key name, and returning a slice of bytes with payload.
func (v *Vault) Read(path, key string) ([]byte, error) {
	var secret *vaultapi.Secret
	var err error

	log.Printf("[Vault] Reading data from path '%s', looking for key '%s'", path, key)
	headers := map[string][]string{"X-Vault-Token": []string{v.token}}
	v.client.SetHeaders(headers)
	v.client.SetToken(v.token)

	if secret, err = v.client.Logical().Read(path); err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil || len(secret.Data) == 0 {
		return nil, fmt.Errorf("no data found on path '%s'", path)
	}

	return v.extractKey(secret.Data, key)
}

// AppRoleAuth execute approle authentication.
func (v *Vault) AppRoleAuth(roleID, secretID string) error {
	var secret *vaultapi.Secret
	var err error

	log.Printf("[Vault] Starting AppRole authentication.")
	authData := map[string]interface{}{"role_id": roleID, "secret_id": secretID}
	if secret, err = v.client.Logical().Write("auth/approle/login", authData); err != nil {
		return err
	}
	if secret.Auth == nil || secret.Auth.ClientToken == "" {
		return errors.New("no authentication data is returned from vault")
	}

	log.Printf("[Vault] Obtained a token via AppRole.")
	// saving token for next API calls.
	v.token = secret.Auth.ClientToken

	return nil
}

// TokenAuth execute token based authentication.
func (v *Vault) TokenAuth(token string) {
	v.token = token
}

// NewVault creates a Vault instance, by bootstrapping it's API client.
func NewVault(addr string) (*Vault, error) {
	var err error

	log.Printf("[Vault] Instantiating Vault API client on address '%s'", addr)
	vault := &Vault{}
	config := &vaultapi.Config{Address: addr}

	if vault.client, err = vaultapi.NewClient(config); err != nil {
		return nil, err
	}

	return vault, nil
}
