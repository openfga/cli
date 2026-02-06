/*
Copyright © 2023 OpenFGA

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package cmd
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/openfga/cli/cmd/model"
	"github.com/openfga/cli/cmd/query"
	"github.com/openfga/cli/cmd/store"
	"github.com/openfga/cli/cmd/tuple"
	"github.com/openfga/cli/internal/cmdutils"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "fga",
	Short:        "OpenFGA CLI",
	Long:         ``,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// isDebugMode checks if debug mode is enabled via --debug flag or FGA_DEBUG env var.
// We are not following cobra's built-in flag checking here because we want to determine
// debug mode status before cobra parses flags (to control logging during initConfig).
// The precedence is:
// 1. Command-line flag --debug
// 2. Environment variable FGA_DEBUG
// Other areas in the code should parse the flag using cobra after initialization rather
// than rely on this function.
func isDebugMode() bool {
	// Command-line flag takes precedence
	for _, arg := range os.Args {
		if arg == "--debug=true" {
			return true
		}
		if arg == "--debug=false" {
			return false
		}
	}

	// Check environment variable first
	if os.Getenv("FGA_DEBUG") == "true" {
		return true
	}

	return false
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fga.yaml)")

	// 'server-url' is deprecated in favor of 'api-url' for consistency with the SDKs,
	// it is still kept here for backward compatibility
	rootCmd.PersistentFlags().String("server-url", "http://localhost:8080", "OpenFGA API URI e.g. https://api.fga.example:8080") //nolint:lll
	rootCmd.PersistentFlags().String("api-url", "http://localhost:8080", "OpenFGA API URI e.g. https://api.fga.example:8080")    //nolint:lll
	_ = rootCmd.PersistentFlags().MarkHidden("server-url")

	rootCmd.PersistentFlags().String("api-token", "", "API Token. Will be sent in as a Bearer in the Authorization header")
	rootCmd.PersistentFlags().String("api-token-issuer", "", "API Token Issuer. API responsible for issuing the API Token. Used in the Client Credentials flow") //nolint:lll
	rootCmd.PersistentFlags().String("api-audience", "", "API Audience. Used when performing the Client Credentials flow")
	rootCmd.PersistentFlags().String("client-id", "", "Client ID. Sent to the Token Issuer during the Client Credentials flow")                            //nolint:lll
	rootCmd.PersistentFlags().String("client-secret", "", "Client Secret. Sent to the Token Issuer during the Client Credentials flow")                    //nolint:lll
	rootCmd.PersistentFlags().StringArray("api-scopes", []string{}, "API Scopes (repeat option for multiple values). Used in the Client Credentials flow") //nolint:lll
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode - can print more detailed information for debugging")

	_ = rootCmd.Flags().MarkHidden("debug")
	rootCmd.MarkFlagsRequiredTogether(
		"api-token-issuer",
		"client-id",
		"client-secret",
	)

	rootCmd.Version = versionStr
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(manCmd)
	rootCmd.AddCommand(store.StoreCmd)
	rootCmd.AddCommand(model.ModelCmd)
	rootCmd.AddCommand(tuple.TupleCmd)
	rootCmd.AddCommand(query.QueryCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viperInstance := viper.New()

	if cfgFile != "" {
		// Use config file from the flag.
		viperInstance.SetConfigFile(cfgFile)
	} else {
		// Find config directory.
		configDir, err := os.UserConfigDir()
		cobra.CheckErr(err)

		// Find config directory.
		homeDir, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search for .fga.yml or .fga.yaml in:
		// 1- The current working directory
		// 2- The user-specific config directory
		// 3- The fga subdirectory under the user-specific config directory
		// 4- The current user's home directory
		viperInstance.AddConfigPath(".")
		viperInstance.AddConfigPath(configDir)
		viperInstance.AddConfigPath(configDir + "/" + "fga")
		viperInstance.AddConfigPath(homeDir)
		viperInstance.SetConfigType("yml")
		viperInstance.SetConfigType("yaml")
		viperInstance.SetConfigName(".fga")
	}

	// If a config file is found, read it in.
	if err := viperInstance.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viperInstance.ConfigFileUsed())
	} else {
		// Check if error is due to config file not found (this is OK, we continue silently)
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			// Config file exists but failed to parse - show warning
			if isDebugMode() {
				fmt.Fprintf(os.Stderr, "Warning: Failed to load config file %s: %v\n",
					viperInstance.ConfigFileUsed(), err)
			} else {
				fmt.Fprintln(os.Stderr, "Warning: Failed to load config file. Use --debug=true or set FGA_DEBUG=true for details.")
			}
		}
	}

	viperInstance.SetEnvPrefix("FGA")
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viperInstance.AutomaticEnv() // read in environment variables that match

	cmdutils.BindViperToFlags(rootCmd, viperInstance)
}
