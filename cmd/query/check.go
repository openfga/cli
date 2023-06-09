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
package query

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openfga/fga-cli/lib/cmd-utils"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command.
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check",
	Long:  "Check if a user has a particular relation with an object.",
	Args:  cobra.ExactArgs(3), //nolint:gomnd
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}

		body := &client.ClientCheckRequest{
			User:     args[0],
			Relation: args[1],
			Object:   args[2],
		}
		options := &client.ClientCheckOptions{}

		response, err := fgaClient.Check(context.Background()).Body(*body).Options(*options).Execute()
		if err != nil {
			fmt.Printf("Failed to check due to %v", err)
			os.Exit(1)
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			fmt.Printf("Failed to check due to %v", err)
			os.Exit(1)
		}
		fmt.Print(string(responseJSON))
	},
}

func init() {}
