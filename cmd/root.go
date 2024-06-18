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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/openfga/cli/cmd/generate"
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

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fga.yaml)")

	// 'server-url' is deprecated in favor of 'api-url' for consistency with the SDKs,
	// it is still kept here for backward compatibility
	rootCmd.PersistentFlags().String("server-url", "http://localhost:8080", "OpenFGA API URI e.g. https://api.fga.example:8080") //nolint:lll
	rootCmd.PersistentFlags().String("api-url", "http://localhost:8080", "OpenFGA API URI e.g. https://api.fga.example:8080")    //nolint:lll

	rootCmd.PersistentFlags().String("api-token", "", "API Token. Will be sent in as a Bearer in the Authorization header")
	rootCmd.PersistentFlags().String("api-token-issuer", "", "API Token Issuer. API responsible for issuing the API Token. Used in the Client Credentials flow") //nolint:lll
	rootCmd.PersistentFlags().String("api-audience", "", "API Audience. Used when performing the Client Credentials flow")
	rootCmd.PersistentFlags().String("client-id", "", "Client ID. Sent to the Token Issuer during the Client Credentials flow")                            //nolint:lll
	rootCmd.PersistentFlags().String("client-secret", "", "Client Secret. Sent to the Token Issuer during the Client Credentials flow")                    //nolint:lll
	rootCmd.PersistentFlags().StringArray("api-scopes", []string{}, "API Scopes (repeat option for multiple values). Used in the Client Credentials flow") //nolint:lll

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
	rootCmd.AddCommand(generate.GenerateCmd)
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

		// Search for .fga.yaml in:
		// 1- XDG_CONFIG/User Config Directory
		// 2- fga directory under User Config Directory
		// 3- Home directory
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
	}

	viperInstance.SetEnvPrefix("FGA")
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viperInstance.AutomaticEnv() // read in environment variables that match

	cmdutils.BindViperToFlags(rootCmd, viperInstance)
}
