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
	"fmt"
	"strings"

	"github.com/openfga/cli/lib/cmd-utils"
	"github.com/openfga/cli/lib/fga"
	"github.com/openfga/cli/lib/output"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

func listRelations(clientConfig fga.ClientConfig,
	fgaClient client.SdkClient,
	user string,
	objects string,
	contextualTuples []client.ClientTupleKey,
) (*client.ClientListRelationsResponse, error) {
	var authorizationModel openfga.AuthorizationModel

	if clientConfig.AuthorizationModelID != "" {
		// note that the auth model id is already configured in the fgaClient.
		response, err := fgaClient.ReadAuthorizationModel(context.Background()).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to list relations due to %w", err)
		}

		authorizationModel = *response.AuthorizationModel
	} else {
		response, err := fgaClient.ReadLatestAuthorizationModel(context.Background()).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to list relations due to %w", err)
		}

		authorizationModel = *response.AuthorizationModel
	}

	typeDefs := *(authorizationModel.TypeDefinitions)
	objectType := strings.Split(objects, ":")[0]

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
		User:             user,
		Object:           objects,
		Relations:        relations,
		ContextualTuples: &contextualTuples,
	}
	options := &client.ClientListRelationsOptions{}

	response, err := fgaClient.ListRelations(context.Background()).Body(*body).Options(*options).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list relations due to %w", err)
	}

	return response, nil
}

// listRelationsCmd represents the listRelations command.
var listRelationsCmd = &cobra.Command{
	Use:   "list-relations",
	Short: "List Relations",
	Long:  "ListRelations if a user has a particular relation with an object.",
	Args:  cobra.ExactArgs(2), //nolint:gomnd
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			fmt.Printf("Failed to initialize FGA Client due to %v", err)

			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		contextualTuples, err := cmdutils.ParseContextualTuples(cmd)
		if err != nil {
			return fmt.Errorf("error parsing contextual tuples for listRelations: %w", err)
		}

		response, err := listRelations(clientConfig, fgaClient, args[0], args[1], contextualTuples)
		if err != nil {
			return fmt.Errorf("error listing relations: %w", err)
		}

		return output.Display(cmd, *response) //nolint:wrapcheck
	},
}

func init() {
}
