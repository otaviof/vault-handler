package vaulthandler

import (
	"fmt"
)

// Config object for vault-handler.
type Config struct {
	DryRun        bool   // dry-run flag
	OutputDir     string // output directory path
	InputDir      string // input directory, when uploading
	DotEnv        bool   // create a dot-env file with secrets
	VaultAddr     string // vault api endpoint
	VaultToken    string // vault token
	VaultRoleID   string // vault approle role-id
	VaultSecretID string // vault approle secret-id
	InCluster     bool   // kubernetes in-cluster
	Context       string // kubernetes context
	Namespace     string // kubernetes namespace
	KubeConfig    string // kubernetes config
}

// Validate configuration object.
func (c *Config) Validate() error {
	if c.VaultAddr == "" {
		return fmt.Errorf("vault-addr is not informed")
	}
	if c.VaultToken == "" && c.VaultRoleID == "" && c.VaultSecretID == "" {
		return fmt.Errorf("inform vault-token, or vault-role-id and secret-id")
	}
	if c.VaultToken != "" && (c.VaultRoleID != "" || c.VaultSecretID != "") {
		return fmt.Errorf("vault-token can't be used in combination with role-id or secret-id")
	}
	if c.InputDir != "" && !isDir(c.InputDir) {
		return fmt.Errorf("input-dir '%s' is not found", c.InputDir)
	}
	if c.OutputDir != "" && !isDir(c.OutputDir) {
		return fmt.Errorf("output-dir '%s' is not found", c.OutputDir)
	}
	return nil
}

// ValidateKubernetes configuration related to Kubernetes.
func (c *Config) ValidateKubernetes() error {
	if c.InCluster && c.Context != "" {
		return fmt.Errorf("configuration 'context' cannot be used in combination with 'in-cluster'")
	}
	if c.Namespace == "" {
		return fmt.Errorf("namespace is not informed")
	}
	if c.KubeConfig != "" && !FileExists(c.KubeConfig) {
		return fmt.Errorf("can't find kube-config file at '%s'", c.KubeConfig)
	}
	return nil
}
