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

	uploadPerPath := make(map[string]map[string]interface{})
	logger := h.logger.WithField("action", "upload")

	if err = h.loop(logger, manifest, func(logger *log.Entry, group, vaultPath string, data SecretData) error {
		logger.Info("Handling file")
		file := NewFile(group, &data, []byte{})

		if err = file.Read(h.config.InputDir); err != nil {
			logger.Error("error on reading file", err)
			return err
		}

		if data.Zip {
			if err = file.Zip(); err != nil {
				logger.Error("error on zipping payload", err)
				return err
			}
		}

		// preparing map of data for the same vault path, dealing with payload as string
		vaultPath = h.composeVaultPath(data, vaultPath)
		if _, exists := uploadPerPath[vaultPath]; !exists {
			uploadPerPath[vaultPath] = make(map[string]interface{})
		}
		uploadPerPath[vaultPath][data.Name] = string(file.Payload())

		return nil
	}); err != nil {
		return err
	}

	for vaultPath, data := range uploadPerPath {
		if err = h.dispense(vaultPath, data); err != nil {
			h.logger.Error("error on writting data to vault", err)
			return err
		}
	}

	return nil
}

// Download files from vault based on manifest.
func (h *Handler) Download(manifest *Manifest) error {
	var err error

	logger := h.logger.WithField("action", "download")

	return h.loop(logger, manifest, func(logger *log.Entry, group, vaultPath string, data SecretData) error {
		var payload []byte

		vaultPath = h.composeVaultPath(data, vaultPath)
		logger = logger.WithField("vaultPath", vaultPath)

		logger.Info("Reading data from Vault")
		if payload, err = h.vault.Read(vaultPath, data.Name); err != nil {
			return err
		}

		logger.Info("Creating a file instance")
		file := NewFile(group, &data, payload)

		if data.Zip {
			if err = file.Unzip(); err != nil {
				return err
			}
		}

		logger.Info("Persisting in file-system")
		return h.persist(file)
	})
}

// loop execute the primary manifest item loop, yielding informed method.
func (h *Handler) loop(
	logger *log.Entry,
	manifest *Manifest,
	fn func(logger *log.Entry, group, vaultPath string, data SecretData) error,
) error {
	for group, secrets := range manifest.Secrets {
		logger = logger.WithFields(log.Fields{"group": group, "vaultPath": secrets.Path})
		for _, data := range secrets.Data {
			logger = logger.WithFields(log.Fields{
				"name": data.Name, "extension": data.Extension, "zip": data.Zip,
			})
			if err := fn(logger, group, secrets.Path, data); err != nil {
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

// dispense a data map to a given vault path.
func (h *Handler) dispense(vaultPath string, data map[string]interface{}) error {
	var err error
	logger := log.WithField("vaultPath", vaultPath)
	logger.Info("Uploading secrets to Vault path")

	for name, payload := range data {
		logger = logger.WithField("key", name)
		logger.Info("Uploading key")
		logger.Tracef("Payload: '%s'", payload)
	}

	if h.config.DryRun {
		logger.Infof("[DRY-RUN] File is not uploaded to Vault!")
		return nil
	}

	if err = h.vault.Write(vaultPath, data); err != nil {
		return err
	}

	return nil
}

// composeVaultPath based in the current SecretData.
func (h *Handler) composeVaultPath(data SecretData, vaultPath string) string {
	if !data.NameAsSubPath {
		return vaultPath
	}
	return path.Join(vaultPath, data.Name)
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
