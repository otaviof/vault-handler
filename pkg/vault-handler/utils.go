package vaulthandler

import (
	"io/ioutil"
	"log"
	"os"
)

// fileExists Check if path exists, boolean return.
func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// readFile Wrap up a ioutil call, using fatal log in case of error.
func readFile(path string) []byte {
	log.Printf("[Utils] Reading file: '%s'", path)

	if !fileExists(path) {
		log.Fatalf("Can't find file: '%s'", path)
	}

	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
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
