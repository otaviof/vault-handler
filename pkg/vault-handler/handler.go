package vaulthandler

import (
	"fmt"
	"log"
	"path"
)

// Handler application primary runtime object.
type Handler struct {
	config *Config
	vault  *Vault
}

// persist a slice of bytes to file-system.
func (h *Handler) persist(groupName string, data *SecretData, payload []byte) error {
	var err error

	fileName := fmt.Sprintf("%s.%s.%s", groupName, data.Name, data.Extension)
	filePath := path.Join(h.config.OutputDir, fileName)
	log.Printf("[Handler] Writting '%s'", filePath)

	file := NewFile(filePath, payload)

	if data.Unzip {
		log.Print("[Handler] Extracting ZIP payload.")
		if err = file.Unzip(); err != nil {
			return err
		}
	}

	if h.config.DryRun {
		log.Printf("[DRY-RUN] File '%s' is not written to file-system!", filePath)
	} else {
		if err = file.Write(); err != nil {
			return err
		}
	}

	return nil
}

// Run app main loop.
func (h *Handler) Run(manifest *Manifest) error {
	var err error

	if h.config.VaultToken != "" {
		log.Printf("[Handler] Using token based authentication against Vault.")
		h.vault.TokenAuth(h.config.VaultToken)
	} else {
		log.Printf("[Handler] Using AppRole based authentication against Vault.")
		if err = h.vault.AppRoleAuth(h.config.VaultRoleID, h.config.VaultSecretID); err != nil {
			return err
		}
	}

	for groupName, secrets := range manifest.Secrets {
		log.Printf("[Handler] Handling secrets for '%s' group...", groupName)
		log.Printf("[Handler] [%s] Vault path '%s'", groupName, secrets.Path)

		for _, data := range secrets.Data {
			log.Printf("[Handler] [%s] Reading data from Vault '%s.%s' (unzip: %v)",
				groupName, data.Name, data.Extension, data.Unzip)

			// putting together the secret path in vault
			vaultPath := secrets.Path
			if data.NameAsSubPath {
				vaultPath = path.Join(vaultPath, data.Name)
			}
			log.Printf("[Handler] [%s] '%s' path in Vault: '%s'", data.Name, groupName, vaultPath)

			// loading secret from vault
			payload := []byte{}
			if payload, err = h.vault.Read(vaultPath, data.Name); err != nil {
				return err
			}

			// saving data to disk
			if err = h.persist(groupName, &data, payload); err != nil {
				return err
			}
		}
	}

	return nil
}

// NewHandler instantiates a new application.
func NewHandler(config *Config) (*Handler, error) {
	var err error

	handler := &Handler{config: config}
	if handler.vault, err = NewVault(config.VaultAddr); err != nil {
		return nil, err
	}

	return handler, nil
}
