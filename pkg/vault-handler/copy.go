package vaulthandler

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
)

// Copy secrets to Kubernetes, by receiving Vault files, and comparing with existing data.
type Copy struct {
	logger     *log.Entry                   // logger
	kube       *Kubernetes                  // kubernetes api-client instance
	download   *Download                    // downloaded data from Vault
	secretType map[string]string            // kubernetes secret-type per group
	data       map[string]map[string][]byte // group as first key, filename as second key
}

// Prepare by looking at Kubernetes secrets and checking if they are different from whats downloaded
// from Vault, the ones that are different, are stored to be persisted later.
func (c *Copy) Prepare() error {
	var err error

	// organizing files by groups, group is the kubernetes secret name
	data := make(map[string][]*File)
	for _, file := range c.download.Files {
		data[file.Group] = append(data[file.Group], file)
	}

	for group, files := range data {
		if len(files) > 0 {
			c.secretType[group] = c.download.Files[0].SecretType
			c.logger.Infof("Setting secret type as '%s'", c.secretType[group])
		}
		if err = c.compare(group, files); err != nil {
			return err
		}
	}

	return nil
}

// Execute inspect collected data during Prepare and create a kubernetes secret.
func (c *Copy) Execute(dryRun bool) error {
	var err error

	for group, data := range c.data {
		c.logger.Infof("Creating Kubernetes secret '%s'", group)
		if dryRun {
			c.logger.Infof("[DRY-RUN] Kubernetes secret '%s'", group)
			continue
		}
		if err = c.kube.SecretWrite(group, c.secretType[group], data); err != nil {
			return err
		}
	}
	return nil
}

// compare with secret present in kubernetes, saving the non-existing of different entries.
func (c *Copy) compare(group string, files []*File) error {
	var kubeSecrets map[string][]byte
	var vaultSecrets = make(map[string][]byte)
	var exists bool
	var err error

	logger := c.logger.WithField("group", group)

	logger.Info("Organizing Vault secrets in the same way than Kubernetes")
	for _, file := range files {
		var exists bool

		if _, exists = vaultSecrets[file.Properties.Name]; exists {
			return fmt.Errorf("name '%s' was found more than once", file.Properties.Name)
		}
		vaultSecrets[file.Properties.Name] = file.Payload
	}

	logger.Info("Checking if secret exists in Kubernetes")
	if exists, err = c.kube.SecretExists(group); err != nil {
		return err
	}
	if !exists {
		logger.Info("Secret does not exist in Kubernetes, yet.")
		c.data[group] = vaultSecrets
		return nil
	}

	logger.Info("Reading Kubernetes secret...")
	if kubeSecrets, err = c.kube.SecretRead(group); err != nil {
		return err
	}

	logger.Info("Commparing Vault secrets with Kubernetes...")
	if !reflect.DeepEqual(vaultSecrets, kubeSecrets) {
		logger.Info("Secrets are different!")
		c.data[group] = vaultSecrets
	}
	logger.Info("Secrets are the same!")

	return nil
}

// NewCopy creates a new Copy instance.
func NewCopy(kube *Kubernetes, download *Download) *Copy {
	return &Copy{
		logger:     log.WithField("type", "copy"),
		kube:       kube,
		download:   download,
		secretType: make(map[string]string),
		data:       make(map[string]map[string][]byte),
	}
}
