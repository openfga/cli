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

// Package query contains commands that run evaluations according to a particular model.
package query

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// QueryCmd represents the query command.
var QueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Run Queries",
	Long:  "Run queries (Check, ListObjects, ListRelations, Expand) that are evaluated according to a particular model.",
}

func init() {
	QueryCmd.AddCommand(checkCmd)
	QueryCmd.AddCommand(expandCmd)
	QueryCmd.AddCommand(listObjectsCmd)
	QueryCmd.AddCommand(listRelationsCmd)

	QueryCmd.PersistentFlags().String("store-id", "", "Store ID")
	QueryCmd.PersistentFlags().String("model-id", "", "Model ID")
	QueryCmd.PersistentFlags().StringArray("contextual-tuple", []string{}, `Contextual Tuple, output: "user relation object"`) //nolint:lll
	QueryCmd.PersistentFlags().String("context", "", "Query context (as a JSON string)")

	err := QueryCmd.MarkPersistentFlagRequired("store-id")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
