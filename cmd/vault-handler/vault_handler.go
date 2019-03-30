package main

import (
	"os"
	"strings"

	vaulthandler "github.com/otaviof/vault-handler/pkg/vault-handler"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "vault-handler",
	Short: "Hashicorp-Vault companion, to upload/download contents from Vault based on a manifest.",
	Long: `
#
# vault-handler
#

Hashicorp-Vault companion to upload and download secrets. It can be used as a Kubernetes
init-container or a command-line application, where you can define manifest files to define how
secrets are placed in Vault, corresponding how secret-files are organized in the file-system.

## Environment Variables

Command-line arguments can be expressed inline, or by exporting environment variables. For
example, the argument "--vault-addr" becomes "VAULT_HANDLER_VAULT_ADDR" in environment. Note the
prefix "VAULT_HANDLER_" in front of the actual argument value, also the capitalization and
replacement of dashes ("-") by underscores ("_").

## Manifest Files

YAML based manifest files are the last argument in "vault-handler" command-line. They represent the
layout of files in the file-system, and will drive the reflection of this data in Vault. Please
consider the GitHub project page for manifest documentation:

	https://github.com/otaviof/vault-handler

## Example

First you may want to export configuration in the environment:

	$ export VAULT_HANDLER_VAULT_ADDR="http://127.0.0.1:8200"
	$ export VAULT_HANDLER_VAULT_ROLE_ID="role-id"
	$ export VAULT_HANDLER_VAULT_SECRET_ID="secret-id"

And later call "vault-handler" with additional arguments, and the manifest files:

	$ vault-handler upload --input-dir /var/tmp --dry-run /path/to/manifest.yaml
	$ vault-handler download --output-dir /tmp --dry-run /path/to/manifest.yaml

## Command-Line
`,
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
	var level log.Level
	var handler *vaulthandler.Handler
	var err error

	if level, err = log.ParseLevel(viper.GetString("log-level")); err != nil {
		log.Fatalf("[ERROR] On parsing log-level: '%s'", err)
	}
	log.SetLevel(level)

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

	log.SetOutput(os.Stdout)

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
	flags.String("log-level", "debug", "dry-run mode")

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
