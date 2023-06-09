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
	"strings"

	"github.com/openfga/fga-cli/lib/cmd-utils"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// listRelationsCmd represents the listRelations command.
var listRelationsCmd = &cobra.Command{
	Use:   "list-relations",
	Short: "List Relations",
	Long:  "ListRelations if a user has a particular relation with an object.",
	Args:  cobra.ExactArgs(2), //nolint:gomnd
	Run: func(cmd *cobra.Command, args []string) {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)
			os.Exit(1)
		}

		var authorizationModel openfga.AuthorizationModel
		if clientConfig.AuthorizationModelID != "" {
			response, err := fgaClient.ReadAuthorizationModel(context.Background()).Execute()
			if err != nil {
				fmt.Printf("Failed to list relations due to %v", err)
				os.Exit(1)
			}
			authorizationModel = *response.AuthorizationModel
		} else {
			response, err := fgaClient.ReadLatestAuthorizationModel(context.Background()).Execute()
			if err != nil {
				fmt.Printf("Failed to list relations due to %v", err)
				os.Exit(1)
			}
			authorizationModel = *response.AuthorizationModel
		}
		typeDefs := *(authorizationModel.TypeDefinitions)
		objectType := strings.Split(args[1], ":")[0]
		var relations []string

		for index := range typeDefs {
			if typeDefs[index].Type == objectType {
				typeDef := typeDefs[index]
				for relation := range *typeDef.Relations {
					relations = append(relations, relation)
				}

				break
			}
		}
		body := &client.ClientListRelationsRequest{
			User:      args[0],
			Object:    args[1],
			Relations: relations,
		}
		options := &client.ClientListRelationsOptions{}

		response, err := fgaClient.ListRelations(context.Background()).Body(*body).Options(*options).Execute()
		if err != nil {
			fmt.Printf("Failed to list relations due to %v", err)
			os.Exit(1)
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			fmt.Printf("Failed to list relations due to %v", err)
			os.Exit(1)
		}
		fmt.Print(string(responseJSON))
	},
}

func init() {
}
