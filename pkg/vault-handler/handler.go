package vaulthandler

import (
	"log"
	"path"
)

// Handler application primary runtime object.
type Handler struct {
	config *Config // configuration instance
	vault  *Vault  // vault api instance
}

// persist a slice of bytes to file-system.
func (h *Handler) persist(group string, data *SecretData, payload []byte) error {
	var err error

	file := NewFile(group, data, payload)

	if data.Unzip {
		log.Print("[Handler] Extracting ZIP payload.")
		if err = file.Unzip(); err != nil {
			return err
		}
	}

	if h.config.DryRun {
		log.Printf("[DRY-RUN] File '%s' is not written to file-system!",
			file.FilePath(h.config.OutputDir))
	} else {
		if err = file.Write(h.config.OutputDir); err != nil {
			return err
		}
	}

	return nil
}

// Authenticate against vault either via token directly or via AppRole, must be invoked before other
// actions using the API.
func (h *Handler) Authenticate() error {
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

	return nil
}

// composeVaultPath based in the current SecretData.
func (h *Handler) composeVaultPath(secrets Secrets, data SecretData) string {
	if !data.NameAsSubPath {
		return secrets.Path
	}
	return path.Join(secrets.Path, data.Name)
}

// Download files from vault based on manifest.
func (h *Handler) Download(manifest *Manifest) error {
	var err error

	for group, secrets := range manifest.Secrets {
		log.Printf("[Handler/Download] Handling secrets for '%s' group...", group)
		log.Printf("[Handler/Download] [%s] Vault path '%s'", group, secrets.Path)

		for _, data := range secrets.Data {
			log.Printf("[Handler/Download] [%s] Reading data from Vault '%s.%s' (unzip: %v)",
				group, data.Name, data.Extension, data.Unzip)

			vaultPath := h.composeVaultPath(secrets, data)
			log.Printf("[Handler/Download] [%s] '%s' path in Vault: '%s'", data.Name, group, vaultPath)

			// loading secret from vault
			payload := []byte{}
			if payload, err = h.vault.Read(vaultPath, data.Name); err != nil {
				return err
			}

			// saving data to disk
			if err = h.persist(group, &data, payload); err != nil {
				return err
			}
		}
	}

	return nil
}

// Upload files to Vault, accordingly to the manifest.
func (h *Handler) Upload(manifest *Manifest) error {
	var err error

	log.Printf("[Handler/Upload] Input directory: '%s'", h.config.InputDir)

	for group, secrets := range manifest.Secrets {
		log.Printf("[Handler/Upload] Handling secrets for '%s' group...", group)
		log.Printf("[Handler/Upload] [%s] Vault path '%s'", group, secrets.Path)

		for _, data := range secrets.Data {
			file := NewFile(group, &data, []byte{})

			if err = file.Read(h.config.InputDir); err != nil {
				log.Printf("[Handler/Upload] error on reading file: '%s'", err)
				return err
			}

			if data.Unzip {
				if err = file.Zip(); err != nil {
					log.Printf("[Handler/Upload] error on zipping file payload: '%s'", err)
					return err
				}
			}

			if err = h.dispense(file, h.composeVaultPath(secrets, data)); err != nil {
				log.Printf("[Handler/Upload] Error on writting data to Vault: '%s'", err)
				return err
			}
		}
	}

	return nil
}

func (h *Handler) dispense(file *File, vaultPath string) error {
	var data = make(map[string]interface{})
	var err error

	if h.config.DryRun {
		log.Printf("[DRY-RUN] File '%s' is not uploaded to Vault path '%s'", file.Name(), vaultPath)
		return nil
	}

	data[file.Name()] = file.Payload()
	if err = h.vault.Write(vaultPath, data); err != nil {
		return err
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
