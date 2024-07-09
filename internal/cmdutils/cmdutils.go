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

// Package cmdutils contains utility and common functions that interaction with the input
// such as reading or binding flags
package cmdutils

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// CheckNoPrettyFlag checks if the --no-pretty flag is set.
func CheckNoPrettyFlag(cmd *cobra.Command) bool {
	noPretty, err := cmd.Flags().GetBool("no-pretty")
	if err != nil {
		return false
	}

	return noPretty
}

// BindViperToCobraFlags binds Viper to Cobra flags.
func BindViperToCobraFlags(cmd *cobra.Command, viperInstance *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viperInstance.IsSet(f.Name) {
			val := viperInstance.Get(f.Name)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				fmt.Printf("Error setting flag %s: %v\n", f.Name, err)
			}
		}
	})
}
