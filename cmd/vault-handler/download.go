package main

import (
	"os"

	vaulthandler "github.com/otaviof/vault-handler/pkg/vault-handler"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Run:   runDownloadCmd,
	Short: "",
	Long:  ``,
}

// runDownloadCmd execute the download of secrets from Vault.
func runDownloadCmd(cmd *cobra.Command, args []string) {
	var manifest *vaulthandler.Manifest
	var err error

	logger := log.WithField("cmd", "download")
	logger.Info("Starting download")

	handler := bootstrap()

	for _, manifestFile := range args {
		logger = logger.WithField("manifest", manifestFile)
		logger.Info("Handling manifest definitions")

		if manifest, err = vaulthandler.NewManifest(manifestFile); err != nil {
			logger.Fatalf("On parsing manifest: '%s'", err)
			os.Exit(1)
		}
		if err = handler.Download(manifest); err != nil {
			logger.Fatalf("On realization of manifest: '%s'", err)
			os.Exit(1)
		}
	}
}

func init() {
	flags := downloadCmd.PersistentFlags()

	flags.String("output-dir", ".", "Output directory.")

	rootCmd.AddCommand(downloadCmd)

	if err := viper.BindPFlags(flags); err != nil {
		log.Panic(err)
	}
}
