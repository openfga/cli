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

// Package store contains commands to manage OpenFGA stores.
package store

import (
	"github.com/spf13/cobra"
)

// StoreCmd represents the store command.
var StoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Interact with OpenFGA Stores",
	Long:  "Create, Get, Delete and List OpenFGA Stores",
}

func init() {
	StoreCmd.AddCommand(createCmd)
	StoreCmd.AddCommand(listCmd)
	StoreCmd.AddCommand(getCmd)
	StoreCmd.AddCommand(deleteCmd)
	StoreCmd.AddCommand(importCmd)
	StoreCmd.AddCommand(exportCmd)
}
