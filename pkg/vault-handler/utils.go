package vaulthandler

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

// FileExists Check if path exists, boolean return.
func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// readFile Wrap up a ioutil call, using fatal log in case of error.
func readFile(path string) []byte {
	var fileBytes []byte
	var err error

	logger := log.WithField("path", path)
	logger.Infof("Reading file bytes")

	if !FileExists(path) {
		logger.Fatal("Can't find file")
	}
	if fileBytes, err = ioutil.ReadFile(path); err != nil {
		logger.Fatalf("Error on read file: '%s'", err)
	}
	return fileBytes
}

// isDir Check if informed path is a directory, boolean return.
func isDir(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	return stat.IsDir()
}
