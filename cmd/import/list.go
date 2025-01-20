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

package _import

import (
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// listCmd represents the import command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the import jobs",
	Long:  "List all the import jobs",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return nil
	},
}

func init() {
}
