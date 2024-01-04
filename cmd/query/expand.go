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

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

func expand(fgaClient client.SdkClient, relation string, object string) (*client.ClientExpandResponse, error) {
	body := &client.ClientExpandRequest{
		Relation: relation,
		Object:   object,
	}

	tuples, err := fgaClient.Expand(context.Background()).Body(*body).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to expand tuples due to %w", err)
	}

	return tuples, nil
}

// expandCmd represents the expand command.
var expandCmd = &cobra.Command{
	Use:     "expand",
	Short:   "Expand",
	Long:    "Expands the relationships in userset tree format.",
	Example: `fga query expand --store-id="01H4P8Z95KTXXEP6Z03T75Q984" can_view document:roadmap`,
	Args:    cobra.ExactArgs(2), //nolint:gomnd
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		response, err := expand(fgaClient, args[0], args[1])
		if err != nil {
			return err
		}

		return output.Display(*response) //nolint:wrapcheck
	},
}

func init() {
}
