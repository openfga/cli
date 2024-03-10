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

package tuple

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

// MaxReadPagesLength Limit the tuples so that we are not paginating indefinitely.
var MaxReadPagesLength = 20

type readResponse struct {
	complete *openfga.ReadResponse
	simple   []openfga.TupleKey
}
type readResponseCSVDTO struct {
	UserType         string `csv:"user_type"`
	UserID           string `csv:"user_id"`
	UserRelation     string `csv:"user_relation,omitempty"`
	Relation         string `csv:"relation"`
	ObjectType       string `csv:"object_type"`
	ObjectID         string `csv:"object_id"`
	ConditionName    string `csv:"condition_name,omitempty"`
	ConditionContext string `csv:"condition_context,omitempty"`
}

func (r readResponse) toCsvDTO() ([]readResponseCSVDTO, error) {
	readResponseDTO := make([]readResponseCSVDTO, 0, len(r.simple))

	for _, readRes := range r.simple {
		// Handle Condition
		conditionName := ""
		conditionalContext := ""

		if readRes.Condition != nil {
			conditionName = readRes.Condition.Name

			if readRes.Condition.Context != nil {
				b, err := json.Marshal(readRes.Condition.Context)
				if err != nil {
					return nil, fmt.Errorf("failed to convert condition context to CSV: %w", err)
				}

				conditionalContext = string(b)
			}
		}
		// Split User and Object
		user := strings.Split(readRes.User, ":")
		object := strings.Split(readRes.Object, ":")

		// Append to DTO
		readResponseDTO = append(readResponseDTO, readResponseCSVDTO{
			UserType:         user[0],
			UserID:           user[1],
			Relation:         readRes.Relation,
			ObjectType:       object[0],
			ObjectID:         object[1],
			ConditionName:    conditionName,
			ConditionContext: conditionalContext,
		})
	}

	return readResponseDTO, nil
}

func baseRead(fgaClient client.SdkClient, body *client.ClientReadRequest, maxPages int) (
	*openfga.ReadResponse, error,
) {
	tuples := make([]openfga.Tuple, 0)
	continuationToken := ""
	pageIndex := 0
	options := client.ClientReadOptions{}

	for {
		options.ContinuationToken = &continuationToken

		response, err := fgaClient.Read(context.Background()).Body(*body).Options(options).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to read tuples due to %w", err)
		}

		tuples = append(tuples, response.Tuples...)
		pageIndex++

		if response.ContinuationToken == "" ||
			(maxPages != 0 && pageIndex >= maxPages) {
			break
		}

		continuationToken = response.ContinuationToken
	}

	return &openfga.ReadResponse{Tuples: tuples}, nil
}

func read(fgaClient client.SdkClient, user string, relation string, object string, maxPages int) (
	*readResponse, error,
) {
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

	response, err := baseRead(fgaClient, body, maxPages)
	if err != nil {
		return nil, err
	}

	justKeys := make([]openfga.TupleKey, 0)
	for _, tuple := range response.GetTuples() {
		justKeys = append(justKeys, tuple.Key)
	}

	res := readResponse{complete: &openfga.ReadResponse{Tuples: response.Tuples}, simple: justKeys}

	return &res, nil
}

// readCmd represents the read command.
var readCmd = &cobra.Command{
	Use:     "read",
	Short:   "Read Relationship Tuples",
	Long:    "Read relationship tuples that exist in the system (does not evaluate).",
	Example: "fga tuple read --store-id=01H0H015178Y2V4CX10C2KGHF4 --user user:anne --relation can_view --object document:roadmap", //nolint:lll
	RunE: func(cmd *cobra.Command, _ []string) error {
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
			return fmt.Errorf("failed to parse max pages due to %w", err)
		}

		response, err := read(fgaClient, user, relation, object, maxPages)
		if err != nil {
			return err
		}

		simpleOutput, _ := cmd.Flags().GetBool("simple-output")
		outputFormat, _ := cmd.Flags().GetString("output-format")
		dataPrinter := output.NewUniPrinter(outputFormat)

		if outputFormat == "csv" {
			data, _ := response.toCsvDTO()

			err := dataPrinter.Display(data)
			if err != nil {
				return fmt.Errorf("failed to display csv: %w", err)
			}

			return nil
		}
		var data any
		data = *response.complete

		if simpleOutput || outputFormat == "simple-json" {
			data = response.simple
		}

		return dataPrinter.Display(data) //nolint:wrapcheck
	},
}

func init() {
	readCmd.Flags().String("user", "", "User")
	readCmd.Flags().String("relation", "", "Relation")
	readCmd.Flags().String("object", "", "Object")
	readCmd.Flags().Int("max-pages", MaxReadPagesLength, "Max number of pages to get. Set to 0 to get all pages.")
	readCmd.Flags().String("output-format", "json", "Specifies the format for data presentation. Valid options: "+
		"json, simple-json, csv, and yaml.")
	readCmd.Flags().Bool("simple-output", false, "Output data in simpler version. (It can be used by write and delete commands)") //nolint:lll
	_ = readCmd.Flags().MarkDeprecated("simple-output", "the flag \"simple-output\" is deprecated and will be removed in future releases.\n"+
		"Please use the \"--output-format=simple-json\" flag instead.")
}
