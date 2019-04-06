package vaulthandler

import (
	log "github.com/sirupsen/logrus"
)

// Download represents the actions needed to download data from Vault, based in the manifest.
type Download struct {
	logger    *log.Entry // logger
	vault     *Vault     // vault api instance
	manifest  *Manifest  // manifest instance
	outputDir string     // output directory
	files     []*File    // list of downloaded files
}

// Prepare files by downloading them from vault, and keeping them aside for later write.
func (d *Download) Prepare(logger *log.Entry, group, vaultPath string, data SecretData) error {
	var payload []byte
	var err error

	vaultPath = d.vault.composePath(data, vaultPath)
	logger = logger.WithField("vaultPath", vaultPath)

	logger.Info("Reading data from Vault")
	if payload, err = d.vault.Read(vaultPath, data.Name); err != nil {
		return err
	}

	logger.Info("Creating a file instance")
	file := NewFile(group, &data, payload)

	if data.Zip {
		if err = file.Unzip(); err != nil {
			return err
		}
	}

	d.files = append(d.files, file)
	return nil
}

// Execute save data to file-system, or just print out in dry-run mode.
func (d *Download) Execute(dryRun bool) error {
	var err error

	d.logger.Info("Persisting in file-system")
	for _, file := range d.files {
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
func NewDownload(vault *Vault, manifest *Manifest, outputDir string) *Download {
	return &Download{
		logger:    log.WithField("type", "download"),
		vault:     vault,
		manifest:  manifest,
		outputDir: outputDir,
	}
}
