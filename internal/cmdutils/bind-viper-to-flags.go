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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// BindViperToFlags recursively binds viper configs to cobra commands and subcommands.
func BindViperToFlags(cmd *cobra.Command, viperInstance *viper.Viper) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		configName := flag.Name

		if !flag.Changed && viperInstance.IsSet(configName) {
			value := viperInstance.Get(configName)
			setFlagFromViper(cmd, flag, value)
		}
	})

	for _, subcmd := range cmd.Commands() {
		BindViperToFlags(subcmd, viperInstance)
	}
}

func setFlagFromViper(cmd *cobra.Command, flag *pflag.Flag, value interface{}) {
	switch v := value.(type) {
	case []interface{}:
		for _, elem := range v {
			err := cmd.Flags().Set(flag.Name, fmt.Sprintf("%v", elem))
			cobra.CheckErr(err)
		}
	default:
		err := cmd.Flags().Set(flag.Name, fmt.Sprintf("%v", v))
		cobra.CheckErr(err)
	}
}
