package main

import (
	"log"
	"strings"

	vaulthandler "github.com/otaviof/vault-handler/pkg/vault-handler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "vault-handler",
	Short: "",
	Long:  ``,
}

// configFromEnv creates a configuration object using Viper, which brings overwritten values from
// environment variables.
func configFromEnv() *vaulthandler.Config {
	return &vaulthandler.Config{
		DryRun:        viper.GetBool("dry-run"),
		OutputDir:     viper.GetString("output-dir"),
		InputDir:      viper.GetString("input-dir"),
		VaultAddr:     viper.GetString("vault-addr"),
		VaultToken:    viper.GetString("vault-token"),
		VaultRoleID:   viper.GetString("vault-role-id"),
		VaultSecretID: viper.GetString("vault-secret-id"),
	}
}

// bootstrap creates connection with vault, by instantiating Handler.
func bootstrap() *vaulthandler.Handler {
	var handler *vaulthandler.Handler
	var err error

	config := configFromEnv()
	if err = config.Validate(); err != nil {
		log.Fatalf("[ERROR] On validating parameters: '%s'", err)
	}

	if handler, err = vaulthandler.NewHandler(config); err != nil {
		log.Fatalf("[ERROR] On instantiating Vault-API: '%s'", err)
	}

	if err = handler.Authenticate(); err != nil {
		log.Fatalf("[ERROR] On authenticating against Vault: '%s'", err)
	}

	return handler
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
	flags.String("vault-addr", "http://127.0.0.1:8200", "Vault address")
	flags.String("vault-token", "", "Vault access token")
	flags.String("vault-role-id", "", "Vault AppRole role-id")
	flags.String("vault-secret-id", "", "Vault AppRole secret-id")
	flags.Bool("dry-run", false, "dry-run mode")

	if err = viper.BindPFlags(flags); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error

	if err = rootCmd.Execute(); err != nil {
		log.Fatalf("[MAIN] %s", err)
	}
}
