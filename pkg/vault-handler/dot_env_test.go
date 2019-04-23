package vaulthandler

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/subosito/gotenv"
)

var dotEnv *DotEnv
var dotEnvBaseDir = "/var/tmp"
var dotEnvFullPath = path.Join(dotEnvBaseDir, ".env")

func TestDotEnvNew(t *testing.T) {
	_ = os.Remove(dotEnvFullPath)

	file := NewFile("dotenv", "", &SecretData{Name: "dotenv", Extension: "txt"}, []byte("dotenv"))
	files := []*File{file}
	dotEnv = NewDotEnv(dotEnvBaseDir, files)
}

func TestDotEnvPrepare(t *testing.T) {
	err := dotEnv.Prepare()
	assert.Nil(t, err)
}

func TestDotEnvWrite(t *testing.T) {
	err := dotEnv.Write()
	assert.Nil(t, err)
}

func TestDotEnvPrepareWithExistingData(t *testing.T) {
	TestDotEnvPrepare(t)
}

func TestDataEnvWriteWithExistingData(t *testing.T) {
	var data map[string]string

	TestDotEnvWrite(t)

	f, err := os.Open(dotEnvFullPath)
	defer f.Close()
	assert.Nil(t, err)

	data = gotenv.Parse(f)
	assert.Equal(t, map[string]string{"DOTENV_DOTENV_TXT": "dotenv"}, data)
}
