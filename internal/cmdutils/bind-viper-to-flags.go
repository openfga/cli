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

package cmdutils

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// flagsDisallowedFromConfigFile lists flags that must never be set from a
// config file. They can still be set via the CLI flag or the corresponding
// environment variable (e.g. FGA_DEBUG).
var flagsDisallowedFromConfigFile = map[string]bool{
	// debug dumps full HTTP requests, including Authorization headers, so no
	// config file should be able to turn on credential logging.
	"debug": true,
}

// BindViperToFlags recursively binds viper configs to cobra commands and subcommands.
func BindViperToFlags(cmd *cobra.Command, viperInstance *viper.Viper) {
	envPrefix := viperInstance.GetEnvPrefix()

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		configName := flag.Name

		if flag.Changed {
			return
		}

		// Flag-only settings come from the environment rather than the viper
		// instance, which cannot tell a config-file value from an env one.
		if flagsDisallowedFromConfigFile[configName] {
			if val, ok := envValue(envPrefix, configName); ok {
				cobra.CheckErr(cmd.Flags().Set(flag.Name, val))
			}

			return
		}

		if viperInstance.IsSet(configName) {
			for _, strVal := range viperInstance.GetStringSlice(configName) {
				cobra.CheckErr(cmd.Flags().Set(flag.Name, strVal))
			}
		}
	})

	for _, subcmd := range cmd.Commands() {
		BindViperToFlags(subcmd, viperInstance)
	}
}

// envValue returns the value of the environment variable corresponding to the
// flag, and whether it is set to a non-empty value. It uses the same naming
// convention as viper's AutomaticEnv: PREFIX_FLAG_NAME with hyphens replaced by
// underscores (e.g. "debug" → "FGA_DEBUG").
func envValue(envPrefix, flagName string) (string, bool) {
	envKey := strings.ToUpper(envPrefix + "_" + strings.ReplaceAll(flagName, "-", "_"))

	val, present := os.LookupEnv(envKey)
	if !present || val == "" {
		return "", false
	}

	return val, true
}
