package e2e

import (
	"fmt"
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
	KubeConfig:    os.Getenv("KUBE_CONFIG"),
	Context:       "",
	Namespace:     "default",
	InCluster:     false,
}

func TestVaultHandler(t *testing.T) {
	log.SetLevel(log.TraceLevel)

	cleanUp(t)

	t.Run("DRY-RUN upload", uploadDryRun)
	t.Run("upload", upload)
	t.Run("DRY-RUN download", downloadDryRun)
	t.Run("download", download)
	t.Run("compare files", compareFiles)
	t.Run("DRY-RUN copy", copyDryRun)
	t.Run("copy", copy)
	t.Run("DRY-RUN copy having secrets", copyDryRun)
	t.Run("copy having secrets", copy)
	t.Run("compare secrets", compareSecrets)
}

type actOnManifest func(t *testing.T, manifest *vh.Manifest)

func loopOverManifests(t *testing.T, fn actOnManifest) {
	for _, manifestFile := range manifestFiles {
		manifest, err := vh.NewManifest(manifestFile)
		assert.Nil(t, err)
		fn(t, manifest)
	}
}

type actOnSecret func(t *testing.T, group string, data *vh.SecretData)

func loopOverGroupSecrets(t *testing.T, manifest *vh.Manifest, fn actOnSecret) {
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

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func cleanUp(t *testing.T) {
	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		loopOverGroupSecrets(t, manifest, func(t *testing.T, group string, data *vh.SecretData) {
			file := vh.NewFile(group, "", data, nil)
			path := file.FilePath(config.OutputDir)

			t.Logf("Excluding file: '%s'", path)

			_ = os.Remove(path)
			assert.False(t, fileExists(path))
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
	runUpload(t, spinUpNewHandler(t, true))
}

func upload(t *testing.T) {
	runUpload(t, spinUpNewHandler(t, false))
}

func runUpload(t *testing.T, handler *vh.Handler) {
	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		err := handler.Upload(manifest)
		assert.Nil(t, err)
	})
}

func downloadDryRun(t *testing.T) {
	runDownload(t, spinUpNewHandler(t, true))
}

func download(t *testing.T) {
	runDownload(t, spinUpNewHandler(t, false))
}

func runDownload(t *testing.T, handler *vh.Handler) {
	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		err := handler.Download(manifest)
		assert.Nil(t, err)
	})
}

func compareFiles(t *testing.T) {
	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		loopOverGroupSecrets(t, manifest, func(t *testing.T, group string, data *vh.SecretData) {
			file := vh.NewFile(group, "", data, nil)
			pathIn := file.FilePath(config.InputDir)
			pathOut := file.FilePath(config.OutputDir)

			assert.FileExists(t, pathOut)
			t.Logf("Comparing files: '%s' vs. '%s'", pathIn, pathOut)
			assert.Equal(t, string(readFile(t, pathIn)), string(readFile(t, pathOut)))
		})
	})
}

func copyDryRun(t *testing.T) {
	runCopy(t, spinUpNewHandler(t, true))
}

func copy(t *testing.T) {
	runCopy(t, spinUpNewHandler(t, false))
}

func runCopy(t *testing.T, handler *vh.Handler) {
	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		err := handler.Copy(manifest)
		assert.Nil(t, err)
	})
}

func compareSecrets(t *testing.T) {
	vaultSecrets := make(map[string]map[string][]byte)

	loopOverManifests(t, func(t *testing.T, manifest *vh.Manifest) {
		loopOverGroupSecrets(t, manifest, func(t *testing.T, group string, data *vh.SecretData) {
			file := vh.NewFile(group, "", data, nil)
			err := file.Read(config.OutputDir)
			assert.Nil(t, err)

			if _, exists := vaultSecrets[group]; !exists {
				vaultSecrets[group] = make(map[string][]byte)
			}
			vaultSecrets[group][file.Properties.Name] = file.Payload
		})
	})

	kube, err := vh.NewKubernetes(config.KubeConfig, config.Context, config.Namespace, config.InCluster)
	assert.Nil(t, err)

	for group, data := range vaultSecrets {
		kubeSecrets, err := kube.SecretRead(group)
		assert.Nil(t, err)
		for name, payload := range data {
			t.Logf("Comparing Kubenetes secret '%s' key '%s', %d bytes", group, name, len(payload))
			assert.Equal(t, string(payload), fmt.Sprintf("%s\n", string(kubeSecrets[name])))
		}
	}
}
