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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// BindViperToFlags recursively binds viper configs to cobra commands and subcommands.
func BindViperToFlags(cmd *cobra.Command, viperInstance *viper.Viper) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		configName := flag.Name

		if !flag.Changed && viperInstance.IsSet(configName) {
			for _, strVal := range viperInstance.GetStringSlice(configName) {
				cobra.CheckErr(cmd.Flags().Set(flag.Name, strVal))
			}
		}
	})

	for _, subcmd := range cmd.Commands() {
		BindViperToFlags(subcmd, viperInstance)
	}
}
