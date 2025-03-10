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
	"errors"
	"fmt"
	"os"

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

// deleteCmd represents the delete command.
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete Relationship Tuples",
	Args:    ExactArgsOrFlag(3, "file"), //nolint:mnd
	Long:    "Delete relationship tuples from the store.",
	Example: "fga tuple delete --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap",
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
		if fileName != "" {
			var tuplesWithoutCondition []client.ClientTupleKeyWithoutCondition

			data, err := os.ReadFile(fileName)
			if err != nil {
				return fmt.Errorf("failed to read file %s due to %w", fileName, err)
			}

			err = yaml.Unmarshal(data, &tuplesWithoutCondition)
			if err != nil {
				return fmt.Errorf("failed to parse input tuples due to %w", err)
			}

			// Convert ClientTupleKeyWithoutCondition to ClientTupleKey
			tuples := make([]client.ClientTupleKey, len(tuplesWithoutCondition))
			for i, t := range tuplesWithoutCondition {
				tuples[i] = client.ClientTupleKey{
					User:     t.User,
					Relation: t.Relation,
					Object:   t.Object,
				}
			}

			maxTuplesPerWrite, err := cmd.Flags().GetInt32("max-tuples-per-write")
			if err != nil {
				return fmt.Errorf("failed to parse max tuples per write due to %w", err)
			}

			maxParallelRequests, err := cmd.Flags().GetInt32("max-parallel-requests")
			if err != nil {
				return fmt.Errorf("failed to parse parallel requests due to %w", err)
			}

			// Extract RPS control parameters
			minRPS, err := cmd.Flags().GetInt32("min-rps")
			if err != nil {
				return fmt.Errorf("failed to parse min-rps: %w", err)
			}

			maxRPS, err := cmd.Flags().GetInt32("max-rps")
			if err != nil {
				return fmt.Errorf("failed to parse max-rps: %w", err)
			}

			rampupPeriod, err := cmd.Flags().GetInt32("rampup-period-in-sec")
			if err != nil {
				return fmt.Errorf("failed to parse rampup-period-in-sec: %w", err)
			}

			// Validate RPS parameters - if one is provided, all three should be required
			if minRPS > 0 || maxRPS > 0 || rampupPeriod > 0 {
				if minRPS <= 0 || maxRPS <= 0 || rampupPeriod <= 0 {
					return errors.New("if any of min-rps, max-rps, or rampup-period-in-sec is provided, all three must be provided with positive values") //nolint:goerr113
				}

				if minRPS > maxRPS {
					return errors.New("min-rps cannot be greater than max-rps") //nolint:goerr113
				}
			}

			deleteRequest := client.ClientWriteRequest{
				Deletes: tuples,
			}
			response, err := ImportTuples(fgaClient, deleteRequest, maxTuplesPerWrite, maxParallelRequests, minRPS, maxRPS, rampupPeriod)
			if err != nil {
				return err
			}

			return output.Display(*response)
		}

		// Create a ClientTupleKey from the arguments
		tupleKey := client.ClientTupleKey{
			User:     args[0],
			Relation: args[1],
			Object:   args[2],
		}

		// Create a delete request with the tuple
		deleteRequest := client.ClientWriteRequest{
			Deletes: []client.ClientTupleKey{tupleKey},
		}

		// Extract RPS control parameters
		minRPS, err := cmd.Flags().GetInt32("min-rps")
		if err != nil {
			return fmt.Errorf("failed to parse min-rps: %w", err)
		}

		maxRPS, err := cmd.Flags().GetInt32("max-rps")
		if err != nil {
			return fmt.Errorf("failed to parse max-rps: %w", err)
		}

		rampupPeriod, err := cmd.Flags().GetInt32("rampup-period-in-sec")
		if err != nil {
			return fmt.Errorf("failed to parse rampup-period-in-sec: %w", err)
		}

		// Validate RPS parameters - if one is provided, all three should be required
		if minRPS > 0 || maxRPS > 0 || rampupPeriod > 0 {
			if minRPS <= 0 || maxRPS <= 0 || rampupPeriod <= 0 {
				return errors.New("if any of min-rps, max-rps, or rampup-period-in-sec is provided, all three must be provided with positive values") //nolint:goerr113
			}

			if minRPS > maxRPS {
				return errors.New("min-rps cannot be greater than max-rps") //nolint:goerr113
			}
		}

		maxTuplesPerWrite, err := cmd.Flags().GetInt32("max-tuples-per-write")
		if err != nil {
			return fmt.Errorf("failed to parse max tuples per write due to %w", err)
		}

		maxParallelRequests, err := cmd.Flags().GetInt32("max-parallel-requests")
		if err != nil {
			return fmt.Errorf("failed to parse parallel requests due to %w", err)
		}

		response, err := ImportTuples(fgaClient, deleteRequest, maxTuplesPerWrite, maxParallelRequests, minRPS, maxRPS, rampupPeriod)
		if err != nil {
			return err
		}

		return output.Display(*response)
	},
}

func init() {
	deleteCmd.Flags().String("file", "", "Tuples file")
	deleteCmd.Flags().String("model-id", "", "Model ID")
	deleteCmd.Flags().Int32("max-tuples-per-write", MaxTuplesPerWrite, "Max tuples per write chunk.")
	deleteCmd.Flags().Int32("max-parallel-requests", MaxParallelRequests, "Max number of requests to issue to the server in parallel.") //nolint:lll
	deleteCmd.Flags().Int32("min-rps", 0, "Minimum requests per second for writes")
	deleteCmd.Flags().Int32("max-rps", 0, "Maximum requests per second for writes")
	deleteCmd.Flags().Int32("rampup-period-in-sec", 0, "Period in seconds to ramp up from min-rps to max-rps")
}

func ExactArgsOrFlag(n int, flag string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n && !cmd.Flags().Changed(flag) {
			return fmt.Errorf("at least %d arg(s) are required OR the flag --%s", n, flag) //nolint:goerr113
		}

		return nil
	}
}
