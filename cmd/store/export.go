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

package store

import (
	"fmt"
	"os"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/spf13/cobra"
)

func exportStore() {}

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Export store data",
	Long:    `Export a store to the export file format`,
	Example: "fga store export",
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fileName, _ := cmd.Flags().GetString("output-file")

		if fileName != "" {
			fmt.Printf("Printing to %s\n", fileName)
		}

		fmt.Println(clientConfig.StoreID)

		return output.Display(output.EmptyStruct{})
	},
}

func init() {
	fmt.Println("Initting exportCmd")

	exportCmd.Flags().String("output-file", "", "name of the file to export the store to")
	exportCmd.Flags().String("store-id", "", "store ID")

	err := exportCmd.MarkFlagRequired("store-id")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
