package vaulthandler

import (
	"path"

	log "github.com/sirupsen/logrus"
)

// Handler application primary runtime object.
type Handler struct {
	logger *log.Entry // logger
	config *Config    // configuration instance
	vault  *Vault     // vault api instance
}

// Authenticate against vault either via token directly or via AppRole, must be invoked before other
// actions using the API.
func (h *Handler) Authenticate() error {
	var err error

	if h.config.VaultToken != "" {
		h.logger.Info("Using token based authentication")
		h.vault.TokenAuth(h.config.VaultToken)
	} else {
		h.logger.Info("Using AppRole based authentication")
		if err = h.vault.AppRoleAuth(h.config.VaultRoleID, h.config.VaultSecretID); err != nil {
			return err
		}
	}

	return nil
}

// Upload files to Vault, accordingly to the manifest.
func (h *Handler) Upload(manifest *Manifest) error {
	var err error

	h.logger.Info("Uploading secrets...")
	for group, secrets := range manifest.Secrets {
		logger := h.logger.WithFields(log.Fields{
			"action":    "upload",
			"inputDir":  h.config.InputDir,
			"group":     group,
			"vaultPath": secrets.Path,
		})
		logger.Info("Handling secrets for group")

		for _, data := range secrets.Data {
			logger = logger.WithFields(log.Fields{
				"name":      data.Name,
				"extension": data.Extension,
				"unzip":     data.Unzip,
			})
			logger.Info("Handling file")

			file := NewFile(group, &data, []byte{})

			if err = file.Read(h.config.InputDir); err != nil {
				logger.Error("error on reading file", err)
				return err
			}

			if data.Unzip {
				if err = file.Zip(); err != nil {
					logger.Error("error on zipping payload", err)
					return err
				}
			}

			if err = h.dispense(file, h.composeVaultPath(secrets, data)); err != nil {
				logger.Error("error on writting data to vault", err)
				return err
			}
		}
	}

	return nil
}

// Download files from vault based on manifest.
func (h *Handler) Download(manifest *Manifest) error {
	var err error

	for group, secrets := range manifest.Secrets {
		logger := h.logger.WithFields(log.Fields{"group": group, "vaultPath": secrets.Path})
		logger.Info("Handling secrets")

		for _, data := range secrets.Data {
			var vaultPath = h.composeVaultPath(secrets, data)
			var payload []byte

			logger.WithFields(log.Fields{
				"name":      data.Name,
				"extension": data.Extension,
				"unzip":     data.Unzip,
				"vaultPath": vaultPath,
			}).Info("Reading data from Vault")

			if payload, err = h.vault.Read(vaultPath, data.Name); err != nil {
				return err
			}
			file := NewFile(group, &data, payload)

			if data.Unzip {
				if err = file.Unzip(); err != nil {
					return err
				}
			}

			if err = h.persist(file); err != nil {
				return err
			}
		}
	}

	return nil
}

// persist a slice of bytes to file-system.
func (h *Handler) persist(file *File) error {
	if h.config.DryRun {
		log.WithField("path", file.FilePath(h.config.OutputDir)).
			Info("[DRY-RUN] File is not written to file-system!")
	} else {
		if err := file.Write(h.config.OutputDir); err != nil {
			return err
		}
	}
	return nil
}

// dispense a file payload to Vault server.
func (h *Handler) dispense(file *File, vaultPath string) error {
	var data = make(map[string]interface{})
	var err error

	logger := log.WithFields(log.Fields{"name": file.Name(), "vaultPath": vaultPath})

	if h.config.DryRun {
		logger.Infof("[DRY-RUN] File is not uploaded to Vault!")
		return nil
	}

	data[file.Name()] = string(file.Payload())
	logger.Tracef("data: '%#v'", data)
	if err = h.vault.Write(vaultPath, data); err != nil {
		return err
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

// NewHandler instantiates a new application.
func NewHandler(config *Config) (*Handler, error) {
	var err error

	handler := &Handler{
		config: config,
		logger: log.WithFields(log.Fields{"type": "Handler"}),
	}

	if handler.vault, err = NewVault(config.VaultAddr); err != nil {
		return nil, err
	}

	return handler, nil
}
