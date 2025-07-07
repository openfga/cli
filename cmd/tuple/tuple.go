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

// Package tuple contains commands that interact directly with stored relationship tuples.
package tuple

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/flags"
)

// TupleCmd represents the tuple command.
var TupleCmd = &cobra.Command{
	Use:   "tuple",
	Short: "Interact with Relationship Tuples",
	Long:  "Read, write, delete, import and listen to changes in relationship tuples in a store.",
}

func init() {
	TupleCmd.AddCommand(changesCmd)
	TupleCmd.AddCommand(readCmd)
	TupleCmd.AddCommand(writeCmd)
	TupleCmd.AddCommand(deleteCmd)

	TupleCmd.PersistentFlags().String("store-id", "", "Store ID")

	if err := flags.SetFlagRequired(TupleCmd, "store-id", "cmd/tuple/tuple", true); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
