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

	cmdutils "github.com/openfga/cli/lib/cmd-utils"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

func read(fgaClient client.SdkClient, user string, relation string, object string, maxPages int) (string, error) {
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

	tuples := make([]openfga.Tuple, 0)

	continuationToken := ""
	pageIndex := 0
	options := client.ClientReadOptions{}

	for {
		options.ContinuationToken = &continuationToken

		response, err := fgaClient.Read(context.Background()).Body(*body).Options(options).Execute()
		if err != nil {
			return "", fmt.Errorf("failed to read tuples due to %w", err)
		}

		tuples = append(tuples, *response.Tuples...)
		pageIndex++

		if response.ContinuationToken == nil || *response.ContinuationToken == "" || pageIndex >= maxPages {
			break
		}

		continuationToken = *response.ContinuationToken
	}

	tuplesJSON, err := json.Marshal(openfga.ReadResponse{Tuples: &tuples})
	if err != nil {
		return "", fmt.Errorf("failed to read tuples due to %w", err)
	}

	return string(tuplesJSON), nil
}

// readCmd represents the read command.
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read Relationship Tuples",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		user, _ := cmd.Flags().GetString("user")
		relation, _ := cmd.Flags().GetString("relation")
		object, _ := cmd.Flags().GetString("object")

		maxPages, _ := cmd.Flags().GetInt("max-pages")
		if err != nil {
			return fmt.Errorf("failed to read tuples due to %w", err)
		}

		output, err := read(fgaClient, user, relation, object, maxPages)
		if err != nil {
			return err
		}

		fmt.Print(output)

		return nil
	},
}

func init() {
	readCmd.Flags().String("user", "", "User")
	readCmd.Flags().String("relation", "", "Relation")
	readCmd.Flags().String("object", "", "Object")
	readCmd.Flags().Int("max-pages", MaxReadChangesPagesLength, "Max number of pages to get.")
}
