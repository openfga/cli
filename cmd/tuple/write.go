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
	"time"

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/tuple"
	"github.com/openfga/cli/internal/tuplefile"
	"github.com/openfga/cli/internal/utils"
)

const writeCommandArgumentsCount = 3

var hideImportedTuples bool

type ImportStats struct {
	TotalTuples      int
	SuccessfulTuples int
	FailedTuples     int
}

// writeCmd represents the write command.
var writeCmd = &cobra.Command{
	Use:     "write",
	Aliases: []string{"import"},
	Short:   "Create Relationship Tuples",
	Long: "Add relationship tuples to the store. This command allows for the creation of " +
		"relationship tuples either through direct command line arguments or by specifying a " +
		"file. The file can be in JSON, YAML, or CSV format.\n\n" +
		"When using a CSV file, the file must adhere to a specific header structure for the " +
		"command to correctly interpret the data. The required CSV header structure is as " +
		"follows:\n" +
		"- \"user_type\":         Specifies the type of the user in the relationship tuple. (e.g. \"team\")\n" +
		"- \"user_id\":           The unique identifier of the user. (e.g. \"marketing\")\n" +
		"- \"user_relation\":     Defines the user relation forming a userset. (optional) (e.g. \"member\")\n" +
		"- \"relation\":          Defines the tuple relation. (e.g. \"viewer\")\n" +
		"- \"object_type\":       Specifies the type of the object in the relationship tuple. (e.g. \"document\")\n" +
		"- \"object_id\":         The unique identifier of the object. (e.g. \"roadmap\")\n" +
		"- \"condition_name\":    The name of the condition. (optional) (e.g. \"inOfficeIP\")\n" +
		"- \"condition_context\": The context of the condition as a json object. " +
		"(optional) (e.g. \"{\"\"ip_addr\"\":\"\"10.0.0.1\"\"}\")\n\n" +
		"For example, a valid CSV file might start with a row like:\n" +
		"user_type,user_id,user_relation,relation,object_type,object_id,condition_name,condition_context\n\n" +
		"This command is flexible in accepting data inputs, making it easier to add multiple " +
		"relationship tuples in various convenient formats.",
	Args: ExactArgsOrFlag(writeCommandArgumentsCount, "file"),
	Example: `  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap --condition-name inOffice --condition-context '{"office_ip":"10.0.1.10"}'
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.json
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.yaml
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.csv
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.csv --max-tuples-per-write 10 --max-parallel-requests 5
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.csv --max-rps 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize fga client: %w", err)
		}

		if len(args) == writeCommandArgumentsCount {
			return writeTuplesFromArgs(cmd, args, fgaClient)
		}

		return writeTuplesFromFile(cmd.Context(), cmd.Flags(), fgaClient)
	},
}

func writeTuplesFromArgs(cmd *cobra.Command, args []string, fgaClient *client.OpenFgaClient) error {
	condition, err := cmdutils.ParseTupleCondition(cmd)
	if err != nil {
		return err //nolint:wrapcheck
	}

	body := client.ClientWriteTuplesBody{
		client.ClientTupleKey{
			User:      args[0],
			Relation:  args[1],
			Object:    args[2],
			Condition: condition,
		},
	}

	_, err = fgaClient.
		WriteTuples(context.Background()).
		Body(body).
		Options(client.ClientWriteOptions{}).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to write tuple: %w", err)
	}

	return output.Display( //nolint:wrapcheck
		map[string]client.ClientWriteTuplesBody{
			"successful": body,
		},
	)
}

func validateWriteFlags(flags *flag.FlagSet, maxTuplesPerWrite, maxParallelRequests, maxRPS, rampUpPeriodInSec int) error {
	if flags.Changed("max-tuples-per-write") && maxTuplesPerWrite <= 0 {
		return errors.New("max-tuples-per-write must be greater than zero") //nolint:err113
	}

	if flags.Changed("max-parallel-requests") && maxParallelRequests <= 0 {
		return errors.New("max-parallel-requests must be greater than zero") //nolint:err113
	}

	if flags.Changed("max-rps") && maxRPS <= 0 {
		return errors.New("max-rps must be greater than zero") //nolint:err113
	}

	if flags.Changed("rampup-period-in-sec") && rampUpPeriodInSec <= 0 {
		return errors.New("rampup-period-in-sec must be greater than zero") //nolint:err113
	}

	return nil
}

func applyWriteDefaults(flags *flag.FlagSet, maxTuplesPerWrite, maxParallelRequests, maxRPS, rampUpPeriodInSec int) (int, int, int, int) {
	if maxRPS > 0 && !flags.Changed("rampup-period-in-sec") {
		rampUpPeriodInSec = maxRPS * tuple.RPSToRampupPeriodMultiplier
	}

	if maxRPS > 0 && !flags.Changed("max-parallel-requests") {
		defaultParallel := maxRPS / tuple.RPSToParallelRequestsDivisor

		if defaultParallel < 1 {
			defaultParallel = 1
		}

		maxParallelRequests = defaultParallel
	}

	if maxRPS > 0 && !flags.Changed("max-tuples-per-write") {
		maxTuplesPerWrite = tuple.DefaultMaxTuplesPerWriteWithRPS
	}

	return maxTuplesPerWrite, maxParallelRequests, maxRPS, rampUpPeriodInSec
}

func writeTuplesFromFile(ctx context.Context, flags *flag.FlagSet, fgaClient *client.OpenFgaClient) error { //nolint:cyclop
	startTime := time.Now()

	fileName, err := flags.GetString("file")
	if err != nil {
		return fmt.Errorf("failed to parse file name: %w", err)
	}

	if fileName == "" {
		return errors.New("file name cannot be empty") //nolint:err113
	}

	maxTuplesPerWrite, err := flags.GetInt("max-tuples-per-write")
	if err != nil {
		return fmt.Errorf("failed to parse max-tuples-per-write due to %w", err)
	}

	maxParallelRequests, err := flags.GetInt("max-parallel-requests")
	if err != nil {
		return fmt.Errorf("failed to parse max-parallel-requests due to %w", err)
	}

	maxRPS, err := flags.GetInt("max-rps")
	if err != nil {
		return fmt.Errorf("failed to parse max-rps due to %w", err)
	}

	rampUpPeriodInSec, err := flags.GetInt("rampup-period-in-sec")
	if err != nil {
		return fmt.Errorf("failed to parse parallel requests due to %w", err)
	}

	if err := validateWriteFlags(flags, maxTuplesPerWrite, maxParallelRequests, maxRPS, rampUpPeriodInSec); err != nil {
		return err
	}

	maxTuplesPerWrite, maxParallelRequests, maxRPS, rampUpPeriodInSec = applyWriteDefaults(
		flags, maxTuplesPerWrite, maxParallelRequests, maxRPS, rampUpPeriodInSec,
	)

	debug, err := flags.GetBool("debug")
	if err != nil {
		return fmt.Errorf("failed to parse debug flag due to %w", err)
	}

	tuples, err := tuplefile.ReadTupleFile(fileName)
	if err != nil {
		return err //nolint:wrapcheck
	}

	writeRequest := client.ClientWriteRequest{
		Writes: tuples,
	}

	newCtx := utils.WithDebugContext(ctx, debug)

	response, err := tuple.ImportTuples(
		newCtx, fgaClient,
		tuple.DefaultMinRPS, maxRPS, rampUpPeriodInSec, maxTuplesPerWrite, maxParallelRequests,
		writeRequest)
	if err != nil {
		return err //nolint:wrapcheck
	}

	duration := time.Since(startTime)
	timeSpent := duration.String()

	outputResponse := make(map[string]interface{})

	if !hideImportedTuples && len(response.Successful) > 0 {
		outputResponse["successful"] = response.Successful
	}

	if len(response.Failed) > 0 {
		outputResponse["failed"] = response.Failed
	}

	outputResponse["total_count"] = len(tuples)
	outputResponse["successful_count"] = len(response.Successful)
	outputResponse["failed_count"] = len(response.Failed)
	outputResponse["time_spent"] = timeSpent

	return output.Display(outputResponse) //nolint:wrapcheck
}

func init() {
	writeCmd.Flags().String("model-id", "", "Model ID")
	writeCmd.Flags().String("file", "", "Tuples file")
	writeCmd.Flags().String("condition-name", "", "Condition Name")
	writeCmd.Flags().String("condition-context", "", "Condition Context (as a JSON string)")
	writeCmd.Flags().Int("max-tuples-per-write", tuple.MaxTuplesPerWrite, "Max tuples per write chunk.")
	writeCmd.Flags().Int("max-parallel-requests", tuple.MaxParallelRequests, "Max number of requests to issue to the server in parallel.")

	writeCmd.Flags().Int("max-rps", 0, "The maximum requests per second.")
	writeCmd.Flags().Int("rampup-period-in-sec", 0, "The period over which to ramp up the request rate.")

	writeCmd.Flags().BoolVar(&hideImportedTuples, "hide-imported-tuples", false, "Hide successfully imported tuples from output")
}
