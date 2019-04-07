package vaulthandler

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // makeing sure gcp plugin is present
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Kubernetes api client instance
type Kubernetes struct {
	logger     *log.Entry            // logger
	clientset  *kubernetes.Clientset // kubernetes api client
	kubeConfig string                // kube-config path
	context    string                // kubernetes context
	namespace  string                // kubernetes namespace
}

// SecretWrite write a secret to kubernetes, based in a secret type and map with data.
func (k *Kubernetes) SecretWrite(name, secretType string, data map[string][]byte) error {
	var exists bool
	var err error

	if exists, err = k.SecretExists(name); err != nil {
		return err
	}
	if exists {
		if err = k.SecretDelete(name); err != nil {
			return err
		}
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Type:       corev1.SecretType(secretType),
		Data:       data,
	}
	if _, err = k.clientset.CoreV1().Secrets(k.namespace).Create(secret); err != nil {
		return err
	}
	return nil
}

// SecretRead reads a secret from Kubernetes, returns a map with it's contents key-value style.
func (k *Kubernetes) SecretRead(name string) (map[string][]byte, error) {
	var secret *corev1.Secret
	var err error

	getOpts := metav1.GetOptions{}
	data := make(map[string][]byte)

	k.logger.Infof("Kubernetes, reading secret '%s'", name)
	if secret, err = k.clientset.CoreV1().Secrets(k.namespace).Get(name, getOpts); err != nil {
		return nil, err
	}

	for filename, byteArray := range secret.Data {
		// executing the same treatment than in vault
		byteArray = bytes.TrimRight(byteArray, "\n")
		k.logger.Infof("Kubernetes-Secret: '%s' ('%d' bytes)", filename, len(byteArray))
		data[filename] = byteArray
	}

	return data, nil
}

// SecretExists check if a given secret exists.
func (k *Kubernetes) SecretExists(name string) (bool, error) {
	var secretList *corev1.SecretList
	var err error

	listOpts := metav1.ListOptions{}

	k.logger.Infof("Checking if secret '%s' exists...", name)
	if secretList, err = k.clientset.CoreV1().Secrets(k.namespace).List(listOpts); err != nil {
		return false, err
	}

	for _, secret := range secretList.Items {
		if secret.Name == name {
			k.logger.Infof("Kubernetes-Secret: '%s' (found)", name)
			return true, nil
		}
	}
	k.logger.Infof("Kubernetes-Secret: '%s' (NOT-found)", name)

	return false, nil
}

// SecretDelete deletes a secret.
func (k *Kubernetes) SecretDelete(name string) error {
	return k.clientset.CoreV1().Secrets(k.namespace).Delete(name, &metav1.DeleteOptions{})
}

// localConfig read kube-config from home, or alternative path.
func (k *Kubernetes) localConfig() (*rest.Config, error) {
	if k.kubeConfig == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			return nil, errors.New("environment HOME is empty, can't find '~/.kube/config' file")
		}
		k.kubeConfig = filepath.Join(homeDir, ".kube", "config")
		k.logger.Info("Using default Kubernetes config file!")
	}
	k.logger.Infof("Using kubernetes configuration: '%s'", k.kubeConfig)

	if !fileExists(k.kubeConfig) {
		return nil, fmt.Errorf("can't find kube-config file at: '%s'", k.kubeConfig)
	}

	return clientcmd.BuildConfigFromFlags(k.context, k.kubeConfig)
}

// NewKubernetes instantiate object by checking if local or in-cluster configuration first.
func NewKubernetes(kubeConfig, context, namespace string, inCluster bool) (*Kubernetes, error) {
	var cfg *rest.Config
	var err error

	logger := log.WithFields(log.Fields{
		"kubeConfig": kubeConfig, "context": context, "namespace": namespace, "inCluster": inCluster,
	})

	k := &Kubernetes{logger: logger, kubeConfig: kubeConfig, context: context, namespace: namespace}

	if inCluster {
		logger.Info("Using in-cluster Kubernetes client...")
		if cfg, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	} else {
		logger.Info("Using local kube-config")
		if cfg, err = k.localConfig(); err != nil {
			return nil, err
		}
	}

	if k.clientset, err = kubernetes.NewForConfig(cfg); err != nil {
		return nil, err
	}

	return k, nil
}
