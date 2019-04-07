package vaulthandler

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// Upload data to Vault, by realizing a manifest against Vault.
type Upload struct {
	logger        *log.Entry                        // logger
	vault         *Vault                            // vault api instance
	inputDir      string                            // input directory path
	uploadPerPath map[string]map[string]interface{} // map of vault-paths with another for secrets
}

// Prepare by reading secrets and letting them ready for next step of uploading.
func (u *Upload) Prepare(logger *log.Entry, group, secretType, vaultPath string, data SecretData) error {
	var err error

	logger.Info("Handling file")
	file := NewFile(group, secretType, &data, []byte{})

	if data.FromEnv != "" {
		logger.Infof("Reading payload from environment-variable '%s'", data.FromEnv)
		payload := os.Getenv(data.FromEnv)
		if payload == "" {
			return fmt.Errorf("can't find environment variable '%s'", data.FromEnv)
		}
		file.Payload = []byte(payload)
	} else {
		logger.Infof("Reading payload from file-system.")
		if err = file.Read(u.inputDir); err != nil {
			logger.Error("error on reading file", err)
			return err
		}
	}
	if data.Zip {
		if err = file.Zip(); err != nil {
			logger.Error("error on zipping payload", err)
			return err
		}
	}

	// preparing map of data for the same vault path, dealing with payload as string
	vaultPath = u.vault.composePath(data, vaultPath)
	if _, exists := u.uploadPerPath[vaultPath]; !exists {
		u.uploadPerPath[vaultPath] = make(map[string]interface{})
	}
	u.uploadPerPath[vaultPath][data.Name] = string(file.Payload)

	return nil
}

// Execute upload secrets to Vault per vault path.
func (u *Upload) Execute(dryRun bool) error {
	var err error

	for vaultPath, data := range u.uploadPerPath {
		if err = u.vaultWrite(vaultPath, data, dryRun); err != nil {
			u.logger.Error("error on writing data to vault", err)
			return err
		}
	}

	return nil
}

// vaultWrite data to vault path, or just print things out in dry-run mode.
func (u *Upload) vaultWrite(vaultPath string, data map[string]interface{}, dryRun bool) error {
	logger := log.WithField("vaultPath", vaultPath)
	logger.Info("Uploading secrets to Vault path")

	for name, payload := range data {
		logger = logger.WithField("key", name)
		logger.Info("Uploading key")
		logger.Tracef("Payload: '%s'", payload)
	}
	if dryRun {
		logger.Infof("[DRY-RUN] File is not uploaded to Vault!")
		return nil
	}

	return u.vault.Write(vaultPath, data)
}

// NewUpload creates a new instance of Upload.
func NewUpload(vault *Vault, inputDir string) *Upload {
	return &Upload{
		logger:        log.WithField("type", "upload"),
		vault:         vault,
		inputDir:      inputDir,
		uploadPerPath: make(map[string]map[string]interface{}),
	}
}
