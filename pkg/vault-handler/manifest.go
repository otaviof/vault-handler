package vaulthandler

import (
	yaml "gopkg.in/yaml.v2"
)

// Manifest to be applied against Vault, define secrets.
type Manifest struct {
	Secrets map[string]Secrets `yaml:"secrets"`
}

// Secrets map with group-name, metadata and secrets list.
type Secrets struct {
	Path string       `yaml:"path"`
	Type string       `yaml:"type"`
	Data []SecretData `yaml:"data"`
}

// SecretData define a single secret in Vault, mapping to a regular file.
type SecretData struct {
	Name          string `yaml:"name"`                    // file name
	Extension     string `yaml:"extension"`               // file extension
	Unzip         bool   `yaml:"unzip,omitempty"`         // unzip data after download
	NameAsSubPath bool   `yaml:"nameAsSubPath,omitempty"` // employ name as part of the path
}

// NewManifest by parsing informed manifest file.
func NewManifest(file string) (*Manifest, error) {
	var err error

	manifest := Manifest{}
	if err = yaml.Unmarshal(readFile(file), &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}
