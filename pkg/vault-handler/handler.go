package vaulthandler

import (
	log "github.com/sirupsen/logrus"
)

// Handler application primary runtime object.
type Handler struct {
	logger *log.Entry // logger
	config *Config    // configuration instance
	vault  *Vault     // vault api instance
}

// actOnSecret method that will receive a secret entry in a group, where vault-path is also shared.
type actOnSecret func(logger *log.Entry, group, vaultPath string, data SecretData) error

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

	u := NewUpload(h.vault, manifest, h.config.InputDir)
	if err = h.loop(h.logger.WithField("action", "upload"), manifest, u.Prepare); err != nil {
		return err
	}

	return u.Execute(h.config.DryRun)
}

// Download files from vault based on manifest.
func (h *Handler) Download(manifest *Manifest) error {
	var err error

	d := NewDownload(h.vault, manifest, h.config.OutputDir)
	if err = h.loop(h.logger.WithField("action", "download"), manifest, d.Prepare); err != nil {
		return err
	}

	return d.Execute(h.config.DryRun)
}

// Copy secrets from Vault into Kubernetes.
func (h *Handler) Copy(manifest *Manifest) error {
	return nil
}

// loop execute the primary manifest item loop, yielding informed method.
func (h *Handler) loop(logger *log.Entry, manifest *Manifest, fn actOnSecret) error {
	for group, secrets := range manifest.Secrets {
		for _, data := range secrets.Data {
			logger = logger.WithFields(log.Fields{
				"name":      data.Name,
				"extension": data.Extension,
				"zip":       data.Zip,
				"group":     group,
				"vaultPath": secrets.Path,
			})
			if err := fn(logger, group, secrets.Path, data); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewHandler instantiates a new application.
func NewHandler(config *Config) (*Handler, error) {
	var err error

	handler := &Handler{config: config, logger: log.WithField("type", "Handler")}
	if handler.vault, err = NewVault(config.VaultAddr); err != nil {
		return nil, err
	}

	return handler, nil
}
