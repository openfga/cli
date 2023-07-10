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

	"github.com/openfga/cli/lib/cmd-utils"
	"github.com/openfga/cli/lib/output"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

// listObjects in the internal function for calling SDK for list objects.
func listObjects(
	fgaClient client.SdkClient,
	user string,
	relation string,
	objectType string,
	contextualTuples []client.ClientTupleKey,
) (*client.ClientListObjectsResponse, error) {
	body := &client.ClientListObjectsRequest{
		User:             user,
		Relation:         relation,
		Type:             objectType,
		ContextualTuples: &contextualTuples,
	}
	options := &client.ClientListObjectsOptions{}

	response, err := fgaClient.ListObjects(context.Background()).Body(*body).Options(*options).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list objects due to %w", err)
	}

	return response, nil
}

// listObjectsCmd represents the listObjects command.
var listObjectsCmd = &cobra.Command{
	Use:   "list-objects",
	Short: "List Objects",
	Long:  "List the relations a user has with an object.",
	Args:  cobra.ExactArgs(3), //nolint:gomnd
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		contextualTuples, err := cmdutils.ParseContextualTuples(cmd)
		if err != nil {
			return fmt.Errorf("error parsing contextual tuples for listObjects: %w", err)
		}

		response, err := listObjects(fgaClient, args[0], args[1], args[2], contextualTuples)
		if err != nil {
			return fmt.Errorf("failed to list objects due to %w", err)
		}

		return output.Display(*response) //nolint:wrapcheck
	},
}

func init() {
}
