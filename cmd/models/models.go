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
package models

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ModelsCmd represents the store command.
var ModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Write, Read and List authorization models in a store",
}

func init() {
	ModelsCmd.AddCommand(writeCmd)
	ModelsCmd.AddCommand(listCmd)
	ModelsCmd.AddCommand(getCmd)

	ModelsCmd.PersistentFlags().String("store-id", "", "Store ID")
	ModelsCmd.Flags().String("store-id", "", "Store ID")
	err := ModelsCmd.MarkFlagRequired("store-id")
	if err != nil { //nolint:wsl
		fmt.Print(err)
		os.Exit(1)
	}
}
