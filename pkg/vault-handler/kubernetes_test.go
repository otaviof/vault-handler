package vaulthandler

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var kube *Kubernetes

func TestKubernetesNew(t *testing.T) {
	var err error

	kubeConfig := os.Getenv("KUBECONFIG")
	t.Logf("Test kube-config: '%s'", kubeConfig)
	kube, err = NewKubernetes(kubeConfig, "", "default", false)

	assert.Nil(t, err)
}

func TestKubernetesSecretWrite(t *testing.T) {
	data := make(map[string][]byte)
	data["test"] = []byte("test")
	err := kube.SecretWrite("test", "", data)

	assert.Nil(t, err)
}

func TestKubernetesSecretExists(t *testing.T) {
	exists, err := kube.SecretExists("test")

	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestKubernetesSecretRead(t *testing.T) {
	data, err := kube.SecretRead("test")

	assert.Nil(t, err)
	assert.Equal(t, []byte("test"), data["test"])
}
