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
	"fmt"
	"os"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// MaxTuplesPerWrite Limit the tuples in a single batch.
var MaxTuplesPerWrite = 20

// MaxParallelRequests Limit the parallel writes to the API.
var MaxParallelRequests = 4

type failedWriteResponse struct {
	TupleKey client.ClientTupleKey `json:"tuple_key"`
	Reason   string                `json:"reason"`
}

type importResponse struct {
	Successful []client.ClientTupleKey `json:"successful"`
	Failed     []failedWriteResponse   `json:"failed"`
}

func importTuples(
	fgaClient client.SdkClient,
	tuples []client.ClientTupleKey,
	maxTuplesPerWrite int,
	maxParallelRequests int,
) (*importResponse, error) {
	options := client.ClientWriteOptions{
		Transaction: &client.TransactionOptions{
			Disable:             true,
			MaxPerChunk:         int32(maxTuplesPerWrite),
			MaxParallelRequests: int32(maxParallelRequests),
		},
	}

	deletes := []client.ClientTupleKey{}
	body := &client.ClientWriteRequest{
		Writes:  &tuples,
		Deletes: &deletes,
	}

	response, err := fgaClient.Write(context.Background()).Body(*body).Options(options).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to import tuples due to %w", err)
	}

	successfulWrites := []client.ClientTupleKey{}
	failedWrites := []failedWriteResponse{}

	for index := 0; index < len(response.Writes); index++ {
		write := response.Writes[index]
		if write.Status == client.SUCCESS {
			successfulWrites = append(successfulWrites, write.TupleKey)
		} else {
			failedWrites = append(failedWrites, failedWriteResponse{
				TupleKey: write.TupleKey,
				Reason:   write.Error.Error(),
			})
		}
	}

	result := importResponse{Successful: successfulWrites, Failed: failedWrites}

	return &result, nil
}

// importCmd represents the import command.
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import Relationship Tuples",
	Long: "Imports Relationship Tuples to the store. " +
		"This will write the tuples in chunks and at the end will report the tuple chunks that failed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("failed to parse file name due to %w", err)
		}

		maxTuplesPerWrite, err := cmd.Flags().GetInt("max-tuples-per-write")
		if err != nil {
			return fmt.Errorf("failed to parse max tuples per write due to %w", err)
		}

		maxParallelRequests, err := cmd.Flags().GetInt("max-parallel-requests")
		if err != nil {
			return fmt.Errorf("failed to parse parallel requests due to %w", err)
		}

		tuples := []client.ClientTupleKey{}

		data, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("failed to read file %s due to %w", fileName, err)
		}

		err = yaml.Unmarshal(data, &tuples)
		if err != nil {
			return fmt.Errorf("failed to parse input tuples due to %w", err)
		}

		result, err := importTuples(fgaClient, tuples, maxTuplesPerWrite, maxParallelRequests)
		if err != nil {
			return err
		}

		return output.Display(*result) //nolint:wrapcheck
	},
}

func init() {
	importCmd.Flags().String("file", "", "Tuples file")
	importCmd.Flags().Int("max-tuples-per-write", MaxTuplesPerWrite, "Max tuples per write chunk.")
	importCmd.Flags().Int("max-parallel-requests", MaxParallelRequests, "Max number of requests to issue to the server in parallel.") //nolint:lll
}
