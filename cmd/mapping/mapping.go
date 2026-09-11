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

// Package mapping implements the fga mapping command group.
package mapping

import "github.com/spf13/cobra"

// MappingCmd is the root of the fga mapping command group.
var MappingCmd = &cobra.Command{
	Use:   "mapping",
	Short: "Manage JSON-to-tuple mappings",
	Long:  "Validate, test, and run JSON-to-tuple mapping files.",
}

func init() {
	MappingCmd.AddCommand(validateCmd)
	MappingCmd.AddCommand(testCmd)
}
