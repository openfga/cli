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
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/openfga/cli/internal/clierrors"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
)

const writeCommandArgumentsCount = 3

// writeCmd represents the write command.
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Create Relationship Tuples",
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
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize fga client: %w", err)
		}

		if len(args) == writeCommandArgumentsCount {
			return writeTuplesFromArgs(cmd, args, fgaClient)
		}

		return writeTuplesFromFile(cmd.Flags(), fgaClient)
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

func writeTuplesFromFile(flags *flag.FlagSet, fgaClient *client.OpenFgaClient) error {
	fileName, err := flags.GetString("file")
	if err != nil {
		return fmt.Errorf("failed to parse file name: %w", err)
	}

	if fileName == "" {
		return fmt.Errorf("file name cannot be empty") //nolint:goerr113
	}

	maxTuplesPerWrite, err := flags.GetInt("max-tuples-per-write")
	if err != nil {
		return fmt.Errorf("failed to parse max tuples per write: %w", err)
	}

	maxParallelRequests, err := flags.GetInt("max-parallel-requests")
	if err != nil {
		return fmt.Errorf("failed to parse parallel requests: %w", err)
	}

	tuples, err := parseTuplesFileData(fileName)
	if err != nil {
		return err
	}

	writeRequest := client.ClientWriteRequest{
		Writes: tuples,
	}

	response, err := ImportTuples(fgaClient, writeRequest, maxTuplesPerWrite, maxParallelRequests)
	if err != nil {
		return err
	}

	return output.Display(response) //nolint:wrapcheck
}

func parseTuplesFileData(fileName string) ([]client.ClientTupleKey, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", fileName, err)
	}

	var tuples []client.ClientTupleKey

	switch path.Ext(fileName) {
	case ".json", ".yaml", ".yml":
		err = yaml.Unmarshal(data, &tuples)
	case ".csv":
		err = parseTuplesFromCSV(data, &tuples)
	default:
		err = fmt.Errorf("unsupported file format %q", path.Ext(fileName)) //nolint:goerr113
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse input tuples: %w", err)
	}

	return tuples, nil
}

func parseTuplesFromCSV(data []byte, tuples *[]client.ClientTupleKey) error {
	reader := csv.NewReader(bytes.NewReader(data))

	columns, err := readHeaders(reader)
	if err != nil {
		return err
	}

	for index := 0; true; index++ {
		tuple, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to read tuple from csv file: %w", err)
		}

		tupleUserKey := tuple[columns.UserType] + ":" + tuple[columns.UserID]
		if columns.UserRelation != -1 && tuple[columns.UserRelation] != "" {
			tupleUserKey += "#" + tuple[columns.UserRelation]
		}

		condition, err := parseConditionColumnsForRow(columns, tuple, index)
		if err != nil {
			return err
		}

		tupleKey := client.ClientTupleKey{
			User:      tupleUserKey,
			Relation:  tuple[columns.Relation],
			Object:    tuple[columns.ObjectType] + ":" + tuple[columns.ObjectID],
			Condition: condition,
		}

		*tuples = append(*tuples, tupleKey)
	}

	return nil
}

func parseConditionColumnsForRow(columns *csvColumns, tuple []string, index int) (*openfga.RelationshipCondition, error) {
	var condition *openfga.RelationshipCondition

	if columns.ConditionName != -1 && tuple[columns.ConditionName] != "" {
		conditionContext := &(map[string]interface{}{})

		if columns.ConditionContext != -1 {
			var err error

			conditionContext, err = cmdutils.ParseQueryContextInner(tuple[columns.ConditionContext])
			if err != nil {
				return nil, fmt.Errorf("failed to read condition context on line %d: %w", index, err)
			}
		}

		condition = &openfga.RelationshipCondition{
			Name:    tuple[columns.ConditionName],
			Context: conditionContext,
		}
	}

	return condition, nil
}

type csvColumns struct {
	UserType         int
	UserID           int
	UserRelation     int
	Relation         int
	ObjectType       int
	ObjectID         int
	ConditionName    int
	ConditionContext int
}

func (columns *csvColumns) setHeaderIndex(headerName string, index int) error {
	switch headerName {
	case "user_type":
		columns.UserType = index
	case "user_id":
		columns.UserID = index
	case "user_relation":
		columns.UserRelation = index
	case "relation":
		columns.Relation = index
	case "object_type":
		columns.ObjectType = index
	case "object_id":
		columns.ObjectID = index
	case "condition_name":
		columns.ConditionName = index
	case "condition_context":
		columns.ConditionContext = index
	default:
		return fmt.Errorf("invalid header %q, valid headers are user_type,user_id,user_relation,relation,object_type,object_id,condition_name,condition_context", headerName) //nolint:goerr113
	}

	return nil
}

func (columns *csvColumns) validate() error {
	if columns.UserType == -1 {
		return clierrors.MissingRequiredCsvHeaderError("user_type") //nolint:wrapcheck
	}

	if columns.UserID == -1 {
		return clierrors.MissingRequiredCsvHeaderError("user_id") //nolint:wrapcheck
	}

	if columns.Relation == -1 {
		return clierrors.MissingRequiredCsvHeaderError("relation") //nolint:wrapcheck
	}

	if columns.ObjectType == -1 {
		return clierrors.MissingRequiredCsvHeaderError("object_type") //nolint:wrapcheck
	}

	if columns.ObjectID == -1 {
		return clierrors.MissingRequiredCsvHeaderError("object_id") //nolint:wrapcheck
	}

	if columns.ConditionContext != -1 && columns.ConditionName == -1 {
		return errors.New("missing \"condition_name\" header which is required when \"condition_context\" is present") //nolint:goerr113
	}

	return nil
}

func readHeaders(reader *csv.Reader) (*csvColumns, error) {
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv headers: %w", err)
	}

	columns := &csvColumns{
		UserType:         -1,
		UserID:           -1,
		UserRelation:     -1,
		Relation:         -1,
		ObjectType:       -1,
		ObjectID:         -1,
		ConditionName:    -1,
		ConditionContext: -1,
	}
	for index, header := range headers {
		err = columns.setHeaderIndex(strings.TrimSpace(header), index)
		if err != nil {
			return nil, err
		}
	}

	return columns, columns.validate()
}

func init() {
	writeCmd.Flags().String("model-id", "", "Model ID")
	writeCmd.Flags().String("file", "", "Tuples file")
	writeCmd.Flags().String("condition-name", "", "Condition Name")
	writeCmd.Flags().String("condition-context", "", "Condition Context (as a JSON string)")
	writeCmd.Flags().Int("max-tuples-per-write", MaxTuplesPerWrite, "Max tuples per write chunk.")
	writeCmd.Flags().Int("max-parallel-requests", MaxParallelRequests, "Max number of requests to issue to the server in parallel.")
}
