package main

import (
	"log"
	"strings"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"

	vaulthandler "github.com/otaviof/vault-handler/pkg/vault-handler"
)

var rootCmd = &cobra.Command{
	Use:   "vault-handler",
	Run:   runVaultHandlerCmd,
	Short: "",
	Long:  ``,
}

// init command-line flags and configuration coming from environment.
func init() {
	var err error

	flags := rootCmd.PersistentFlags()

	// setting up rules for environment variables
	viper.SetEnvPrefix("vault-handler")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// command-line flags
	flags.String("output-dir", ".", "Output directory.")
	flags.String("vault-addr", "http://127.0.0.1:8200", "Vault address")
	flags.String("vault-token", "", "Vault access token")
	flags.String("vault-role-id", "", "Vault AppRole role-id")
	flags.String("vault-secret-id", "", "Vault AppRole secret-id")
	flags.Bool("dry-run", false, "dry-run mode")

	if err = viper.BindPFlags(flags); err != nil {
		log.Fatal(err)
	}
}

// configFromEnv creates a configuration object using Viper, which brings overwritten values from
// environment variables.
func configFromEnv() *vaulthandler.Config {
	return &vaulthandler.Config{
		DryRun:        viper.GetBool("dry-run"),
		OutputDir:     viper.GetString("output-dir"),
		VaultAddr:     viper.GetString("vault-addr"),
		VaultToken:    viper.GetString("vault-token"),
		VaultRoleID:   viper.GetString("vault-role-id"),
		VaultSecretID: viper.GetString("vault-secret-id"),
	}
}

// runVaultHandlerCmd execute the primary objective of this app, to realize the manifest files.
func runVaultHandlerCmd(cmd *cobra.Command, args []string) {
	var handler *vaulthandler.Handler
	var manifest *vaulthandler.Manifest
	var err error

	config := configFromEnv()
	if err = config.Validate(); err != nil {
		log.Fatalf("[ERROR] On validating parameters: '%s'", err)
	}

	if handler, err = vaulthandler.NewHandler(config); err != nil {
		log.Fatalf("[ERROR] On instantiating Vault-API: '%s'", err)
	}

	for _, manifestFile := range args {
		log.Printf("Handling manifest file: '%s'", manifestFile)

		if manifest, err = vaulthandler.NewManifest(manifestFile); err != nil {
			log.Fatalf("[ERROR] On parsing manifest: '%s'", err)
		}
		if err = handler.Run(manifest); err != nil {
			log.Fatalf("[ERROR] During realization of manifest: '%s'", err)
		}
	}
}

func main() {
	var err error
	if err = rootCmd.Execute(); err != nil {
		log.Fatalf("[MAIN] %s", err)
	}
}
