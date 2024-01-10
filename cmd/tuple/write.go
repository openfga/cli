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
		"- \"user_type\":     Specifies the type of the user in the relationship tuple.\n" +
		"- \"user_id\":       The unique identifier of the user.\n" +
		"- \"user_relation\": Defines the user relation forming a userset.\n" +
		"- \"relation\":      Defines the tuple relation.\n" +
		"- \"object_type\":   Specifies the type of the object in the relationship tuple.\n" +
		"- \"object_id\":     The unique identifier of the object.\n\n" +
		"For example, a valid CSV file might start with a row like:\n" +
		"user_type,user_id,user_relation,relation,object_type,object_id\n\n" +
		"This command is flexible in accepting data inputs, making it easier to add multiple " +
		"relationship tuples in various convenient formats.",
	Args: ExactArgsOrFlag(writeCommandArgumentsCount, "file"),
	Example: `  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap
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
			return writeTuplesFromArgs(args, fgaClient)
		}

		return writeTuplesFromFile(cmd.Flags(), fgaClient)
	},
}

func writeTuplesFromArgs(args []string, fgaClient *client.OpenFgaClient) error {
	body := client.ClientWriteTuplesBody{
		client.ClientTupleKey{
			User:     args[0],
			Relation: args[1],
			Object:   args[2],
		},
	}

	_, err := fgaClient.
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

	for i := 0; true; i++ {
		if i == 0 {
			if err := guardAgainstInvalidHeaderWithinCSV(reader); err != nil {
				return err
			}

			continue
		}

		tuple, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to read tuple from csv file: %w", err)
		}

		const (
			UserType = iota
			UserID
			UserRelation
			Relation
			ObjectType
			ObjectID
		)

		tupleUserKey := tuple[UserType] + ":" + tuple[UserID]
		if tuple[UserRelation] != "" {
			tupleUserKey += "#" + tuple[UserRelation]
		}

		tupleKey := client.ClientTupleKey{
			User:     tupleUserKey,
			Relation: tuple[Relation],
			Object:   tuple[ObjectType] + ":" + tuple[ObjectID],
		}

		*tuples = append(*tuples, tupleKey)
	}

	return nil
}

func guardAgainstInvalidHeaderWithinCSV(reader *csv.Reader) error {
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read csv headers: %w", err)
	}

	headerMap := make(map[string]bool)
	for _, header := range headers {
		headerMap[strings.TrimSpace(header)] = true
	}

	requiredHeaders := []string{"user_type", "user_id", "user_relation", "relation", "object_type", "object_id"}

	if len(headerMap) != len(requiredHeaders) {
		return fmt.Errorf( //nolint:goerr113
			"csv file must have exactly these headers in order: %q",
			strings.Join(requiredHeaders, ","),
		)
	}

	for _, header := range requiredHeaders {
		if _, ok := headerMap[header]; !ok {
			return fmt.Errorf("required csv header %q not found", header) //nolint:goerr113
		}
	}

	return nil
}

func init() {
	writeCmd.Flags().String("model-id", "", "Model ID")
	writeCmd.Flags().String("file", "", "Tuples file")
	writeCmd.Flags().String("condition-name", "", "Condition Name")
	writeCmd.Flags().String("condition-context", "", "Condition Context (as a JSON string)")
	writeCmd.Flags().Int("max-tuples-per-write", MaxTuplesPerWrite, "Max tuples per write chunk.")
	writeCmd.Flags().Int("max-parallel-requests", MaxParallelRequests, "Max number of requests to issue to the server in parallel.") //nolint:lll
}
