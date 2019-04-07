package main

import (
	"os"

	vh "github.com/otaviof/vault-handler/pkg/vault-handler"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var copyCmd = &cobra.Command{
	Use:   "copy [manifest-files]",
	Run:   runCopyCmd,
	Short: `Copy secrets from Vault to Kubernetes, based in manifest`,
	Long: `# vault-handler copy

Copy secrets from Vault into Kubernetes, following the manifest. If the secret is already present in
Kubernetes it will be updated, if different than in Vault.

The manifest file defines which type of secret will be created in Kubernetes, and based in the Secret
type, certain keys will be mandatory, so be aware about setting up mandatory items.
`,
}

func runCopyCmd(cmd *cobra.Command, args []string) {
	logger := log.WithField("cmd", "copy")
	logger.Info("Starting copy")

	h := bootstrap()
	if err := config.ValidateKubernetes(); err != nil {
		log.Fatalf("[ERROR] On validating parameters: '%s'", err)
	}

	loopManifests(logger, args, func(logger *log.Entry, m *vh.Manifest) {
		if err := h.Copy(m); err != nil {
			logger.Fatalf("On realization of manifest: '%s'", err)
			os.Exit(1)
		}
	})
}

func init() {
	flags := copyCmd.PersistentFlags()

	flags.String("context", "", "Kubernetes context")
	flags.String("namespace", "", "Kubernetes namespace")
	flags.String("kube-config", "", "Kubernetes '~/.kube/config' alternative path")
	flags.Bool("in-cluster", false, "Peek is running inside Kubernetes")

	rootCmd.AddCommand(copyCmd)

	if err := viper.BindPFlags(flags); err != nil {
		log.Panic(err)
	}
}
