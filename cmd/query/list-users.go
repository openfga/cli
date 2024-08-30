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
	"os"
	"strings"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

func parseUserFilters(rawUserFilter string) []openfga.UserTypeFilter {
	userFilters := []openfga.UserTypeFilter{}

	if rawUserFilter != "" {
		if strings.Contains(rawUserFilter, "#") {
			splitFilter := strings.Split(rawUserFilter, "#")
			userFilters = append(userFilters, openfga.UserTypeFilter{
				Type:     splitFilter[0],
				Relation: &splitFilter[1],
			})
		} else {
			userFilters = append(userFilters, openfga.UserTypeFilter{
				Type: rawUserFilter,
			})
		}
	}

	return userFilters
}

func parseObject(rawObject string) openfga.FgaObject {
	splitObject := strings.Split(rawObject, ":")

	return openfga.FgaObject{
		Type: splitObject[0],
		Id:   splitObject[1],
	}
}

func listUsers(
	fgaClient client.SdkClient,
	rawObject string,
	relation string,
	rawUserFilter string,
	contextualTuples []client.ClientContextualTupleKey,
	queryContext *map[string]interface{},
	consistency *openfga.ConsistencyPreference,
) (*client.ClientListUsersResponse, error) {
	body := &client.ClientListUsersRequest{
		Object:           parseObject(rawObject),
		Relation:         relation,
		UserFilters:      parseUserFilters(rawUserFilter),
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	options := &client.ClientListUsersOptions{}

	// Don't set if UNSPECIFIED has been provided, it's the default anyway
	if *consistency != openfga.CONSISTENCYPREFERENCE_UNSPECIFIED {
		options.Consistency = consistency
	}

	response, err := fgaClient.ListUsers(context.Background()).Body(*body).Options(*options).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list users due to %w", err)
	}

	return response, nil
}

var listUsersCmd = &cobra.Command{
	Use:     "list-users",
	Short:   "List users",
	Long:    "List all users that have a certain relation with a particular object",
	Example: `fga query list-users --store-id=01H0H015178Y2V4CX10C2KGHF4 --object document:roadmap --relation can_view`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		contextualTuples, err := cmdutils.ParseContextualTuples(cmd)
		if err != nil {
			return fmt.Errorf("error parsing contextual tuples: %w", err)
		}

		queryContext, err := cmdutils.ParseQueryContext(cmd, "context")
		if err != nil {
			return fmt.Errorf("error parsing query context: %w", err)
		}

		consistency, err := cmdutils.ParseConsistencyFromCmd(cmd)
		if err != nil {
			return fmt.Errorf("error parsing consistency for check: %w", err)
		}

		userFilter, _ := cmd.Flags().GetString("user-filter")
		object, _ := cmd.Flags().GetString("object")
		relation, _ := cmd.Flags().GetString("relation")

		response, err := listUsers(fgaClient, object, relation, userFilter, contextualTuples, queryContext, consistency)
		if err != nil {
			return fmt.Errorf("failed to list users due to %w", err)
		}

		return output.Display(*response)
	},
}

func init() {
	listUsersCmd.Flags().String("object", "", "Object to list users for")
	listUsersCmd.Flags().String("relation", "", "Relation to evaluate on")
	listUsersCmd.Flags().String("user-filter", "", "Filter the responses can be in the formats <type> (to filter objects and typed public bound access) or <type>#<relation> (to filter usersets)") //nolint:lll

	if err := listUsersCmd.MarkFlagRequired("object"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/query/list-users", err)
		os.Exit(1)
	}

	if err := listUsersCmd.MarkFlagRequired("relation"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/query/list-users", err)
		os.Exit(1)
	}

	if err := listUsersCmd.MarkFlagRequired("user-filter"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/query/list-users", err)
		os.Exit(1)
	}
}
