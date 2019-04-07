package vaulthandler

import (
	log "github.com/sirupsen/logrus"
)

// Download represents the actions needed to download data from Vault, based in the manifest.
type Download struct {
	logger    *log.Entry // logger
	vault     *Vault     // vault api instance
	outputDir string     // output directory
	Files     []*File    // list of downloaded files
}

// Prepare files by downloading them from vault, and keeping them aside for later write.
func (d *Download) Prepare(logger *log.Entry, group, secretType, vaultPath string, data SecretData) error {
	var keyName string
	var payload []byte
	var err error

	vaultPath = d.vault.composePath(data, vaultPath)
	logger = logger.WithField("vaultPath", vaultPath)

	if data.Key != "" {
		keyName = data.Key
	} else {
		keyName = data.Name
	}

	logger.Infof("Reading data from Vault, key '%s'", keyName)
	if payload, err = d.vault.Read(vaultPath, keyName); err != nil {
		return err
	}

	logger.Info("Creating a file instance")
	file := NewFile(group, secretType, &data, payload)
	if data.Zip {
		if err = file.Unzip(); err != nil {
			return err
		}
	}

	d.Files = append(d.Files, file)
	return nil
}

// Execute save data to file-system, or just print out in dry-run mode.
func (d *Download) Execute(dryRun bool) error {
	var err error

	d.logger.Info("Persisting in file-system")
	for _, file := range d.Files {
		if dryRun {
			d.logger.WithField("path", file.FilePath(d.outputDir)).
				Info("[DRY-RUN] File is not written to file-system!")
			continue
		}
		if err = file.Write(d.outputDir); err != nil {
			return err
		}
	}
	return nil
}

// NewDownload creates a new Download instance.
func NewDownload(vault *Vault, outputDir string) *Download {
	return &Download{
		logger:    log.WithField("type", "download"),
		vault:     vault,
		outputDir: outputDir,
	}
}
