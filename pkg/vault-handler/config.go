package vaulthandler

import (
	"errors"
	"fmt"
)

// Config object for vault-handler.
type Config struct {
	DryRun        bool   // dry-run flag
	OutputDir     string // output directory path
	VaultAddr     string // vault api endpoint
	VaultToken    string // vault token
	VaultRoleID   string // vault approle role-id
	VaultSecretID string // vault approle secret-id
}

// Validate configuration object.
func (c *Config) Validate() error {
	if c.VaultAddr == "" {
		return errors.New("vault-addr is not informed")
	}
	if c.VaultToken == "" && c.VaultRoleID == "" && c.VaultSecretID == "" {
		return errors.New("inform vault-token, or vault-role-id and secret-id")
	}
	if c.VaultToken != "" && (c.VaultRoleID != "" || c.VaultSecretID != "") {
		return errors.New("vault-token can't be used in combination with role-id or secret-id")
	}
	if c.OutputDir == "" {
		return errors.New("output-dir is not informed")
	}
	if !isDir(c.OutputDir) {
		return fmt.Errorf("output-dir '%s' is not found", c.OutputDir)
	}

	return nil
}
