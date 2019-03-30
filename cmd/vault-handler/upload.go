package main

import (
	"os"

	vaulthandler "github.com/otaviof/vault-handler/pkg/vault-handler"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var uploadCmd = &cobra.Command{
	Use:   "upload [manifest-files]",
	Run:   runUploadCmd,
	Short: "Realize manifest uploading secrets to Vault.",
	Long: ` # vault-handler upload

Based on manifest, it will look for files in "--input-dir" based in naming convention, and upload
data to Vault accordingly, following configuration for Vault's path and zipped contents.
`,
}

// runUploadCmd execute the actions to upload files to vault.
func runUploadCmd(cmd *cobra.Command, args []string) {
	var manifest *vaulthandler.Manifest
	var err error

	logger := log.WithField("cmd", "upload")
	logger.Info("Starting upload")

	handler := bootstrap()

	for _, manifestFile := range args {
		logger = logger.WithField("manifest", manifestFile)
		logger.Info("Handling manifest definitions")

		if manifest, err = vaulthandler.NewManifest(manifestFile); err != nil {
			logger.Fatalf("On parsing manifest: '%s'", err)
			os.Exit(1)
		}
		if err = handler.Upload(manifest); err != nil {
			logger.Fatalf("On realization of manifest: '%s'", err)
			os.Exit(1)
		}
	}
}

func init() {
	flags := uploadCmd.PersistentFlags()

	flags.String("input-dir", ".", "Input directory.")

	rootCmd.AddCommand(uploadCmd)

	if err := viper.BindPFlags(flags); err != nil {
		log.Panic(err)
	}
}
