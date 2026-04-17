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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

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
			for _, strVal := range viperValueToStrings(value) {
				cobra.CheckErr(cmd.Flags().Set(flag.Name, strVal))
			}
		}
	})

	for _, subcmd := range cmd.Commands() {
		BindViperToFlags(subcmd, viperInstance)
	}
}

// viperValueToStrings converts a viper config value to a slice of strings for
// pflag.Set calls. Slice values (e.g. from YAML lists) produce one string per
// element. Scalar strings that look like a JSON array (start with "[" and end
// with "]") are parsed as JSON to support multiple values via env vars, e.g.
// FGA_CUSTOM_HEADERS='["X-Foo: bar","X-Baz: qux"]'. Other scalars produce a
// single-element slice.
func viperValueToStrings(value any) []string {
	reflectValue := reflect.ValueOf(value)

	if reflectValue.Kind() == reflect.Slice || reflectValue.Kind() == reflect.Array {
		result := make([]string, 0, reflectValue.Len())
		for i := range reflectValue.Len() {
			result = append(result, fmt.Sprintf("%v", reflectValue.Index(i).Interface()))
		}

		return result
	}

	str := fmt.Sprintf("%v", value)
	if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
		var parsed []string
		if err := json.Unmarshal([]byte(str), &parsed); err == nil {
			return parsed
		}
	}

	return []string{str}
}
