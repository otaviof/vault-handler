package main

import (
	"log"

	vaulthandler "github.com/otaviof/vault-handler/pkg/vault-handler"
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

	log.Printf("runDownloadCmd")

	handler := bootstrap()

	for _, manifestFile := range args {
		log.Printf("[Download] Handling manifest file: '%s'", manifestFile)

		if manifest, err = vaulthandler.NewManifest(manifestFile); err != nil {
			log.Fatalf("[ERROR] On parsing manifest: '%s'", err)
		}

		if err = handler.Download(manifest); err != nil {
			log.Fatalf("[ERROR] During realization of manifest: '%s'", err)
		}
	}
}

func init() {
	flags := downloadCmd.PersistentFlags()

	flags.String("output-dir", ".", "Output directory.")

	rootCmd.AddCommand(downloadCmd)

	if err := viper.BindPFlags(flags); err != nil {
		panic(err)
	}
}
