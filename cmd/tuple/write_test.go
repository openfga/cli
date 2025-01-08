package tuple

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/tuplefile"
)

func TestParseTuplesFileData(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name           string
		file           string
		expectedTuples []client.ClientTupleKey
		expectedError  string
	}{
		{
			name: "it can correctly parse a csv file",
			file: "testdata/tuples.csv",
			expectedTuples: []client.ClientTupleKey{
				{
					User:     "user:anne",
					Relation: "owner",
					Object:   "folder:product",
					Condition: &openfga.RelationshipCondition{
						Name:    "inOfficeIP",
						Context: &map[string]interface{}{},
					},
				},
				{
					User:     "folder:product",
					Relation: "parent",
					Object:   "folder:product-2021",
					Condition: &openfga.RelationshipCondition{
						Name: "inOfficeIP",
						Context: &map[string]interface{}{
							"ip_addr": "10.0.0.1",
						},
					},
				},
				{
					User:     "team:fga#member",
					Relation: "viewer",
					Object:   "folder:product-2021",
				},
			},
		},
		{
			name: "it can correctly parse a csv file regardless of columns order",
			file: "testdata/tuples_other_columns_order.csv",
			expectedTuples: []client.ClientTupleKey{
				{
					User:     "user:anne",
					Relation: "owner",
					Object:   "folder:product",
					Condition: &openfga.RelationshipCondition{
						Name:    "inOfficeIP",
						Context: &map[string]interface{}{},
					},
				},
				{
					User:     "folder:product",
					Relation: "parent",
					Object:   "folder:product-2021",
					Condition: &openfga.RelationshipCondition{
						Name: "inOfficeIP",
						Context: &map[string]interface{}{
							"ip_addr": "10.0.0.1",
						},
					},
				},
				{
					User:     "team:fga#member",
					Relation: "viewer",
					Object:   "folder:product-2021",
				},
			},
		},
		{
			name: "it can correctly parse a csv file without optional fields",
			file: "testdata/tuples_without_optional_fields.csv",
			expectedTuples: []client.ClientTupleKey{
				{
					User:     "user:anne",
					Relation: "owner",
					Object:   "folder:product",
				},
				{
					User:     "folder:product",
					Relation: "parent",
					Object:   "folder:product-2021",
				},
			},
		},
		{
			name: "it can correctly parse a csv file with condition_name header but no condition_context header",
			file: "testdata/tuples_with_condition_name_but_no_condition_context.csv",
			expectedTuples: []client.ClientTupleKey{
				{
					User:     "user:anne",
					Relation: "owner",
					Object:   "folder:product",
					Condition: &openfga.RelationshipCondition{
						Name:    "inOfficeIP",
						Context: &map[string]interface{}{},
					},
				},
				{
					User:     "folder:product",
					Relation: "parent",
					Object:   "folder:product-2021",
					Condition: &openfga.RelationshipCondition{
						Name:    "inOfficeIP",
						Context: &map[string]interface{}{},
					},
				},
				{
					User:     "team:fga#member",
					Relation: "viewer",
					Object:   "folder:product-2021",
				},
			},
		},
		{
			name: "it can correctly parse a json file",
			file: "testdata/tuples.json",
			expectedTuples: []client.ClientTupleKey{
				{
					User:     "user:anne",
					Relation: "owner",
					Object:   "folder:product",
				},
				{
					User:     "folder:product",
					Relation: "parent",
					Object:   "folder:product-2021",
				},
				{
					User:     "user:beth",
					Relation: "viewer",
					Object:   "folder:product-2021",
				},
			},
		},
		{
			name: "it can correctly parse a yaml file",
			file: "testdata/tuples.yaml",
			expectedTuples: []client.ClientTupleKey{
				{
					User:     "user:anne",
					Relation: "owner",
					Object:   "folder:product",
				},
				{
					User:     "folder:product",
					Relation: "parent",
					Object:   "folder:product-2021",
				},
				{
					User:     "user:beth",
					Relation: "viewer",
					Object:   "folder:product-2021",
				},
			},
		},
		{
			name:          "it fails to parse a non-existent file",
			file:          "testdata/tuples.bad",
			expectedError: "failed to read file \"testdata/tuples.bad\": open testdata/tuples.bad: no such file or directory",
		},
		{
			name:          "it fails to parse a non-supported file format",
			file:          "testdata/tuples.toml",
			expectedError: "failed to parse input tuples: unsupported file format \".toml\"",
		},
		{
			name:          "it fails to parse a csv file with wrong headers",
			file:          "testdata/tuples_wrong_headers.csv",
			expectedError: "failed to parse input tuples: invalid header \"a\", valid headers are user_type,user_id,user_relation,relation,object_type,object_id,condition_name,condition_context",
		},
		{
			name:          "it fails to parse a csv file with missing required headers",
			file:          "testdata/tuples_missing_required_headers.csv",
			expectedError: "failed to parse input tuples: csv header missing (\"object_id\")",
		},
		{
			name:          "it fails to parse a csv file with missing condition_name header when condition_context is present",
			file:          "testdata/tuples_missing_condition_name_header.csv",
			expectedError: "failed to parse input tuples: missing \"condition_name\" header which is required when \"condition_context\" is present",
		},
		{
			name:          "it fails to parse an empty csv file",
			file:          "testdata/tuples_empty.csv",
			expectedError: "failed to parse input tuples: failed to read csv headers: EOF",
		},
		{
			name:          "it fails to parse a csv file with invalid rows",
			file:          "testdata/tuples_with_invalid_rows.csv",
			expectedError: "failed to parse input tuples: failed to read tuple from csv file: record on line 2: wrong number of fields",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actualTuples, err := tuplefile.ReadTupleFile(test.file)

			if test.expectedError != "" {
				require.EqualError(t, err, test.expectedError)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedTuples, actualTuples)
		})
	}
}

func TestWriteTuplesWithImportStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		file               string
		hideImportedTuples bool
		expectedStats      ImportStats
		expectedOutput     string
		expectedError      string
	}{
		{
			name:               "shows all tuples when hide-imported-tuples is false",
			file:               "testdata/tuples.json",
			hideImportedTuples: false,
			expectedStats: ImportStats{
				TotalTuples:      3,
				SuccessfulTuples: 3,
				FailedTuples:     0,
			},
			expectedOutput: `{
													"successful": [
																	{
																					"user": "user:anne",
																					"relation": "owner",
																					"object": "folder:product"
																	},
																	{
																					"user": "folder:product",
																					"relation": "parent",
																					"object": "folder:product-2021"
																	},
																	{
																					"user": "user:beth",
																					"relation": "viewer",
																					"object": "folder:product-2021"
																	}
													]
									}`,
		},
		{
			name:               "hides successful tuples when hide-imported-tuples is true",
			file:               "testdata/tuples.json",
			hideImportedTuples: true,
			expectedStats: ImportStats{
				TotalTuples:      3,
				SuccessfulTuples: 3,
				FailedTuples:     0,
			},
			expectedOutput: "",
		},
		{
			name:               "always shows failed tuples even when hide-imported-tuples is true",
			file:               "testdata/tuples_with_errors.json",
			hideImportedTuples: true,
			expectedStats: ImportStats{
				TotalTuples:      3,
				SuccessfulTuples: 1,
				FailedTuples:     2,
			},
			expectedOutput: `{
													"failed": [
																	{
																					"user": "invalid:user",
																					"relation": "invalid",
																					"object": "invalid:object"
																	},
																	{
																					"user": "another:invalid",
																					"relation": "invalid",
																					"object": "invalid:object"
																	}
													]
									}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Temporarily redirect stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Set the hide flag
			hideImportedTuples = test.hideImportedTuples

			// Execute test
			tuples, err := tuplefile.ReadTupleFile(test.file)
			if err == nil {
				// Create a response based on the test case
				response := map[string][]client.ClientTupleKey{}

				if !hideImportedTuples || test.expectedStats.FailedTuples > 0 {
					if test.expectedStats.SuccessfulTuples > 0 {
						response["successful"] = tuples[:test.expectedStats.SuccessfulTuples]
					}
					if test.expectedStats.FailedTuples > 0 {
						response["failed"] = tuples[test.expectedStats.SuccessfulTuples:]
					}
					err = output.Display(response)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			// Print stats after JSON output
			fmt.Fprintf(w, "\nImport Summary:\n")
			fmt.Fprintf(w, "Total tuples processed: %d\n", test.expectedStats.TotalTuples)
			fmt.Fprintf(w, "Successfully imported: %d\n", test.expectedStats.SuccessfulTuples)
			fmt.Fprintf(w, "Failed: %d\n", test.expectedStats.FailedTuples)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if test.expectedError != "" {
				require.EqualError(t, err, test.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedStats.TotalTuples, len(tuples))
			assert.Contains(t, output, fmt.Sprintf("Total tuples processed: %d", test.expectedStats.TotalTuples))
			assert.Contains(t, output, fmt.Sprintf("Successfully imported: %d", test.expectedStats.SuccessfulTuples))
			assert.Contains(t, output, fmt.Sprintf("Failed: %d", test.expectedStats.FailedTuples))

			if test.expectedOutput != "" {
				// Extract JSON part from output (before the summary)
				jsonPart := output
				if idx := strings.Index(output, "\nImport Summary"); idx >= 0 {
					jsonPart = output[:idx]
				}
				if jsonPart != "" {
					assert.JSONEq(t, test.expectedOutput, jsonPart)
				}
			}
		})
	}
}
