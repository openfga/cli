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

// Package model contains commands that interact with Authorization Models.
package model

import (
	"github.com/spf13/cobra"
)

// ModelCmd represents the store command.
var ModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Interact with Authorization Models",
	Long:  "Write, read, list and validate authorization models.",
}

func init() {
	ModelCmd.AddCommand(writeCmd)
	ModelCmd.AddCommand(listCmd)
	ModelCmd.AddCommand(getCmd)
	ModelCmd.AddCommand(validateCmd)
	ModelCmd.AddCommand(transformCmd)
	ModelCmd.AddCommand(modelTestCmd)
	ModelCmd.PersistentFlags().String("store-id", "", "Store ID")
}
