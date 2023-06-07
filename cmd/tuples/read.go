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
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openfga/fga-cli/lib/cmd-utils"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// readCmd represents the read command.
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read Relationship Tuples",
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}
		user, _ := cmd.Flags().GetString("user")
		relation, _ := cmd.Flags().GetString("relation")
		object, _ := cmd.Flags().GetString("object")
		body := &client.ClientReadRequest{}
		if user != "" {
			body.User = &user
		}
		if relation != "" {
			body.Relation = &relation
		}
		if object != "" {
			body.Object = &object
		}
		options := &client.ClientReadOptions{}
		tuples, err := fgaClient.Read(context.Background()).Body(*body).Options(*options).Execute()
		if err != nil {
			fmt.Printf("Failed to read tuples due to %v", err)
			os.Exit(1)
		}

		tuplesJSON, err := json.Marshal(tuples)
		if err != nil {
			fmt.Printf("Failed to read tuples due to %v", err)
			os.Exit(1)
		}
		fmt.Print(string(tuplesJSON))
	},
}

func init() {
	readCmd.Flags().String("user", "", "User")
	readCmd.Flags().String("relation", "", "Relation")
	readCmd.Flags().String("object", "", "Object")
}
