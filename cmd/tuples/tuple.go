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
package tuples

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// TupleCmd represents the tuple command.
var TupleCmd = &cobra.Command{
	Use:   "tuples",
	Short: "Interact with Relationship Tuples in a store.",
}

func init() {
	TupleCmd.AddCommand(writeCmd)
	TupleCmd.AddCommand(deleteCmd)
	TupleCmd.AddCommand(readCmd)
	TupleCmd.AddCommand(changesCmd)

	TupleCmd.PersistentFlags().String("store-id", "", "Store ID")
	TupleCmd.Flags().String("store-id", "", "Store ID")
	err := TupleCmd.MarkFlagRequired("store-id")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
