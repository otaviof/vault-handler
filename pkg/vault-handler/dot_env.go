package vaulthandler

import (
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/subosito/gotenv"
	shellescape "gopkg.in/alessio/shellescape.v1"
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
	if !FileExists(d.fullPath) {
		d.logger.Info("Dot-env file is not found.")
		return nil
	}

	d.logger.Info("Reading existing dot-env file contents.")
	return d.readExisting()
}

// Write down dot-env file.
func (d *DotEnv) Write() error {
	var f *os.File
	var err error

	d.loadFiles()

	if f, err = os.OpenFile(d.fullPath, os.O_CREATE|os.O_RDWR, 0600); err != nil {
		return err
	}
	defer f.Close()

	for k, v := range d.data {
		if _, err = f.WriteString(fmt.Sprintf("%s=%s\n", k, shellescape.Quote(v))); err != nil {
			return err
		}
	}
	return f.Sync()
}

// readExisting dot-env file, parsing out variable names and values.
func (d *DotEnv) readExisting() error {
	var f *os.File
	var err error

	if f, err = os.Open(d.fullPath); err != nil {
		return err
	}
	defer f.Close()
	d.data = gotenv.Parse(f)

	return nil
}

// envVarName format a variable name based on a File instance.
func (d *DotEnv) envVarName(file *File) string {
	name := fmt.Sprintf("%s_%s_%s", file.Group, file.Properties.Name, file.Properties.Extension)
	return strings.ToUpper(name)
}

// loadFiles loop over array of Files, load contents
func (d *DotEnv) loadFiles() {
	for _, file := range d.files {
		k := d.envVarName(file)
		if _, found := d.data[k]; found {
			d.logger.Warnf("Variable '%s' is already preset on '%s'!", k, d.fullPath)
		}
		v := string(file.Payload)
		d.logger.Tracef("Adding entry on dot-env: '%s'='%s'", k, v)
		d.data[k] = v
	}
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
