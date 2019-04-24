package vaulthandler

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	shellescape "gopkg.in/alessio/shellescape.v1"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/shell"
)

// DotEnv represents a .env file
type DotEnv struct {
	logger   *log.Entry        // logger
	fullPath string            // dot-env full path
	files    []*File           // list of downloaded files
	data     map[string]string // doe-env data.
}

// Prepare by checking if dot-env (".env") file already exists, read it's contents.
func (d *DotEnv) Prepare() error {
	d.loadFiles()

	if !FileExists(d.fullPath) {
		d.logger.Info("Dot-env file is not found.")
		return nil
	}
	if err := d.readExisting(); err != nil {
		return err
	}
	return nil
}

// Write down dot-env file.
func (d *DotEnv) Write(dryRun bool) error {
	var f *os.File
	var escaped []string
	var err error

	for k, v := range d.data {
		d.logger.Infof("Adding key '%s' to dot-env file.", k)
		escaped = append(escaped, fmt.Sprintf("%s=%s\n", k, shellescape.Quote(v)))
	}

	if dryRun {
		d.logger.Info("[DRY-RUN] Skipping writting dot-env file.")
		return nil
	}

	if f, err = os.OpenFile(d.fullPath, os.O_CREATE|os.O_RDWR, 0600); err != nil {
		return err
	}
	defer f.Close()
	for _, s := range escaped {
		if _, err = f.WriteString(s); err != nil {
			return err
		}
	}
	return f.Sync()
}

// readExisting dot-env file, parsing out variable names and values.
func (d *DotEnv) readExisting() error {
	var existing map[string]expand.Variable
	var err error

	if existing, err = shell.SourceFile(context.TODO(), d.fullPath); err != nil {
		return err
	}

	for k, v := range existing {
		d.logger.Infof("Already existing dot-env variable '%s'", k)
		d.put(k, v.String())
	}
	return nil
}

// loadFiles loop over array of Files, load contents
func (d *DotEnv) loadFiles() {
	for _, file := range d.files {
		k := d.envVarName(file)
		v := string(file.Payload)
		d.logger.Tracef("Adding entry on dot-env: '%s'='%s'", k, v)
		d.put(k, v)
	}
}

// envVarName format a variable name based on a File instance.
func (d *DotEnv) envVarName(file *File) string {
	name := fmt.Sprintf("%s_%s_%s", file.Group, file.Properties.Name, file.Properties.Extension)
	return strings.ToUpper(name)
}

func (d *DotEnv) put(k, v string) {
	if _, found := d.data[k]; found {
		d.logger.Warnf("Key '%s' is being overwritten!", k)
	}
	d.data[k] = v
}

// NewDotEnv creates a new instance.
func NewDotEnv(outputDir string, files []*File) *DotEnv {
	fullPath := path.Join(outputDir, ".env")
	return &DotEnv{
		logger: log.WithFields(log.Fields{
			"type":      "dotEnv",
			"outputDir": outputDir,
			"fullPath":  fullPath,
		}),
		fullPath: fullPath,
		files:    files,
		data:     make(map[string]string),
	}
}
