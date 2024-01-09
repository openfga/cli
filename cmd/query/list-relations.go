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

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/fga"
	"github.com/openfga/cli/internal/output"
)

func getRelationsForType(
	clientConfig fga.ClientConfig,
	fgaClient client.SdkClient,
	object string,
) (*[]string, error) {
	var authorizationModel openfga.AuthorizationModel

	if clientConfig.AuthorizationModelID != "" {
		response, err := fgaClient.ReadAuthorizationModel(context.Background()).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to list relations due to %w", err)
		}

		authorizationModel = response.GetAuthorizationModel()
	} else {
		response, err := fgaClient.ReadLatestAuthorizationModel(context.Background()).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to list relations due to %w", err)
		}

		authorizationModel = response.GetAuthorizationModel()
	}

	typeDefs := authorizationModel.TypeDefinitions
	objectType := strings.Split(object, ":")[0]
	relations := []string{}

	for index := range typeDefs {
		if typeDefs[index].Type == objectType {
			typeDef := typeDefs[index]
			for relation := range *typeDef.Relations {
				relations = append(relations, relation)
			}

			break
		}
	}

	return &relations, nil
}

func listRelations(clientConfig fga.ClientConfig,
	fgaClient client.SdkClient,
	user string,
	object string,
	relations []string,
	contextualTuples []client.ClientContextualTupleKey,
	queryContext *map[string]interface{},
) (*client.ClientListRelationsResponse, error) {
	if len(relations) < 1 {
		relationsForType, err := getRelationsForType(clientConfig, fgaClient, object)
		if err != nil {
			return nil, fmt.Errorf("failed to list relations due to %w", err)
		}

		relations = *relationsForType

		if len(relations) < 1 {
			// there is still no relations.  This means for the model, the corresponding object's type has no relations
			return &client.ClientListRelationsResponse{
				Relations: []string{},
			}, nil
		}
	}

	body := &client.ClientListRelationsRequest{
		User:             user,
		Object:           object,
		Relations:        relations,
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	options := &client.ClientListRelationsOptions{}

	response, err := fgaClient.ListRelations(context.Background()).Body(*body).Options(*options).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list relations due to %w", err)
	}

	if response.Relations == nil {
		response.Relations = []string{}
	}

	return response, nil
}

// listRelationsCmd represents the listRelations command.
var listRelationsCmd = &cobra.Command{
	Use:     "list-relations",
	Short:   "List Relations",
	Long:    "List relations that a user has with an object.",
	Example: `fga query list-relations --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne document:roadmap --relation can_view`, //nolint:lll
	Args:    cobra.ExactArgs(2),                                                                                              //nolint:gomnd,lll
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		contextualTuples, err := cmdutils.ParseContextualTuples(cmd)
		if err != nil {
			return fmt.Errorf("error parsing contextual tuples for listRelations: %w", err)
		}

		queryContext, err := cmdutils.ParseQueryContext(cmd, "context")
		if err != nil {
			return fmt.Errorf("error parsing query context for check: %w", err)
		}

		relations, _ := cmd.Flags().GetStringArray("relation")

		response, err := listRelations(clientConfig, fgaClient, args[0], args[1], relations, contextualTuples, queryContext)
		if err != nil {
			return fmt.Errorf("failed to list relations due to %w", err)
		}

		return output.Display(*response) //nolint:wrapcheck
	},
}

func init() {
	listRelationsCmd.Flags().StringArray("relation", []string{}, "Relation")
}
