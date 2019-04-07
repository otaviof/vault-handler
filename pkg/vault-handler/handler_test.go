package vaulthandler

import (
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	vaultAddr      = "http://127.0.0.1:8200"
	vaultRootToken = "vault-root-token"
	groupName      = "group"
	inputDir       = "/var/tmp"
	outputDir      = "/tmp"
)

var handler *Handler
var handlerManifest = &Manifest{
	Secrets: map[string]Secrets{
		groupName: {
			Path: "secret/data/test/handler/upload",
			Data: []SecretData{
				{Name: "zipped", Extension: "txt", NameAsSubPath: true, Zip: true},
				{Name: "plain", Extension: "txt", NameAsSubPath: true, Zip: false},
			},
		},
	},
}

func TestHandlerAuthenticateToken(t *testing.T) {
	log.SetLevel(log.TraceLevel)

	config := &Config{
		VaultAddr:  vaultAddr,
		VaultToken: vaultRootToken,
	}

	h, err := NewHandler(config)
	assert.Nil(t, err)

	err = h.Authenticate()
	assert.Nil(t, err)
}

func TestHandlerAuthenticateAppRole(t *testing.T) {
	var err error

	config := &Config{
		VaultAddr:     vaultAddr,
		VaultRoleID:   os.Getenv("VAULT_HANDLER_VAULT_ROLE_ID"),
		VaultSecretID: os.Getenv("VAULT_HANDLER_VAULT_SECRET_ID"),
		InputDir:      inputDir,
		OutputDir:     outputDir,
	}
	err = config.Validate()
	assert.Nil(t, err)

	handler, err = NewHandler(config)
	assert.Nil(t, err)

	err = handler.Authenticate()
	assert.Nil(t, err)
}

func TestHandlerUpload(t *testing.T) {
	var err error

	_ = os.Remove(fmtManifestFilePath(inputDir, 0))
	zipped := NewFile(groupName, "", &handlerManifest.Secrets[groupName].Data[0], []byte("zipped"))
	err = zipped.Write(inputDir)
	assert.Nil(t, err)
	assert.Equal(t, fmtManifestFilePath(inputDir, 0), zipped.FilePath(inputDir))

	_ = os.Remove(fmtManifestFilePath(inputDir, 1))
	plain := NewFile(groupName, "", &handlerManifest.Secrets[groupName].Data[1], []byte("plain"))
	err = plain.Write(inputDir)
	assert.Nil(t, err)
	assert.Equal(t, fmtManifestFilePath(inputDir, 1), plain.FilePath(inputDir))

	err = handler.Upload(handlerManifest)
	assert.Nil(t, err)
}

func fmtManifestFilePath(dir string, i int) string {
	return fmt.Sprintf("%s/%s.%s.%s",
		dir,
		groupName,
		handlerManifest.Secrets[groupName].Data[i].Name,
		handlerManifest.Secrets[groupName].Data[i].Extension)
}

func TestHandlerDownload(t *testing.T) {
	var err error

	zippedPath := fmt.Sprintf("%s/%s.zipped.txt", outputDir, groupName)
	_ = os.Remove(zippedPath)
	plainPath := fmt.Sprintf("%s/%s.plain.txt", outputDir, groupName)
	_ = os.Remove(plainPath)

	err = handler.Download(handlerManifest)
	assert.Nil(t, err)

	assert.FileExists(t, zippedPath)
	assert.Equal(t, []byte("zipped"), readFile(zippedPath))

	assert.FileExists(t, plainPath)
	assert.Equal(t, []byte("plain"), readFile(plainPath))
}
