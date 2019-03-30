package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	vh "github.com/otaviof/vault-handler/pkg/vault-handler"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var manifestFiles = []string{"../mock/manifest-1.yaml", "../mock/manifest-2.yaml"}

var config = &vh.Config{
	VaultAddr:     "http://127.0.0.1:8200",
	InputDir:      "../mock/input-dir",
	OutputDir:     "/tmp",
	VaultRoleID:   os.Getenv("VAULT_HANDLER_VAULT_ROLE_ID"),
	VaultSecretID: os.Getenv("VAULT_HANDLER_VAULT_SECRET_ID"),
}

func TestVaultHandler(t *testing.T) {
	log.SetLevel(log.TraceLevel)

	cleanUp(t)

	t.Run("DRY-RUN upload", uploadDryRun)
	t.Run("upload", upload)
	t.Run("DRY-RUN download", downloadDryRun)
	t.Run("download", download)
	t.Run("compare", compare)
}

func loopOverManifests(t *testing.T, fn func(t *testing.T, manifest *vh.Manifest)) {
	for _, manifestFile := range manifestFiles {
		manifest, err := vh.NewManifest(manifestFile)
		assert.Nil(t, err)
		fn(t, manifest)
	}
}

func loopOverGroupSecrets(
	t *testing.T, manifest *vh.Manifest, fn func(t *testing.T, group string, data *vh.SecretData),
) {
	for group, secrets := range manifest.Secrets {
		for _, data := range secrets.Data {
			fn(t, group, &data)
		}
	}
}

func readFile(t *testing.T, path string) []byte {
	fileBytes, err := ioutil.ReadFile(path)
	assert.Nil(t, err)
	return fileBytes
}

func cleanUp(t *testing.T) {
	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		loopOverGroupSecrets(t, manifest, func(t *testing.T, group string, data *vh.SecretData) {
			file := vh.NewFile(group, data, nil)
			path := file.FilePath(config.OutputDir)
			t.Logf("Excluding file: '%s'", path)
			_ = os.Remove(path)
		})
	})
}

func spinUpNewHandler(t *testing.T, dryRun bool) *vh.Handler {
	config.DryRun = dryRun
	handler, err := vh.NewHandler(config)
	assert.Nil(t, err)

	err = handler.Authenticate()
	assert.Nil(t, err)

	return handler
}

func uploadDryRun(t *testing.T) {
	handler := spinUpNewHandler(t, true)

	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		err := handler.Upload(manifest)
		assert.Nil(t, err)
	})
}

func upload(t *testing.T) {
	handler := spinUpNewHandler(t, false)

	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		err := handler.Upload(manifest)
		assert.Nil(t, err)
	})
}

func downloadDryRun(t *testing.T) {
	handler := spinUpNewHandler(t, true)

	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		err := handler.Download(manifest)
		assert.Nil(t, err)
	})
}

func download(t *testing.T) {
	handler := spinUpNewHandler(t, false)

	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		err := handler.Download(manifest)
		assert.Nil(t, err)
	})
}

func compare(t *testing.T) {
	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		loopOverGroupSecrets(t, manifest, func(t *testing.T, group string, data *vh.SecretData) {
			file := vh.NewFile(group, data, nil)
			pathIn := file.FilePath(config.InputDir)
			pathOut := file.FilePath(config.OutputDir)

			assert.FileExists(t, pathOut)
			t.Logf("Comparing files: '%s' vs. '%s'", pathIn, pathOut)
			assert.Equal(t, string(readFile(t, pathIn)), string(readFile(t, pathOut)))
		})
	})
}
