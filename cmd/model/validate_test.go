package model

import (
	"encoding/json"
	"reflect"
	"testing"

	openfga "github.com/openfga/go-sdk"

	"github.com/openfga/cli/internal/authorizationmodel"
)

func TestValidateCmdWithArgs(t *testing.T) {
	t.Parallel()
	validateCmd.SetArgs([]string{`{"schema_version":"1.1"}`})

	if err := validateCmd.Execute(); err != nil {
		t.Errorf("failed to execute validateCmd")
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	type validationTest struct {
		Name           string
		Input          string
		IsValid        bool
		ExpectedOutput validationResult
		NoPretty       bool
	}

	tests := []validationTest{
		{
			Name:    "missing schema version",
			Input:   "{",
			IsValid: false,
			ExpectedOutput: validationResult{
				IsValid: false,
				Error:   openfga.PtrString("unable to parse json input"),
			},
		},
		{
			Name:    "missing schema version",
			Input:   `{"id":"abcde","schema_version":"1.1"}`,
			IsValid: false,
			ExpectedOutput: validationResult{
				ID:      "abcde",
				IsValid: false,
				Error:   openfga.PtrString("unable to parse id: invalid ulid format"),
			},
		},
		{
			Name:    "missing schema version",
			Input:   "{}",
			IsValid: false,
			ExpectedOutput: validationResult{
				IsValid: false,
				Error:   openfga.PtrString("invalid schema version"),
			},
		},
		{
			Name:    "invalid schema version",
			Input:   `{"schema_version":"1.0"}`,
			IsValid: false,
			ExpectedOutput: validationResult{
				IsValid: false,
				Error:   openfga.PtrString("invalid schema version"),
			},
		},
		{
			Name:    "only schema",
			Input:   `{"schema_version":"1.1"}`,
			IsValid: true,
			ExpectedOutput: validationResult{
				IsValid: true,
			},
		},
		{
			Name:     "no-pretty output",
			Input:    `{"schema_version":"1.1"}`,
			IsValid:  true,
			NoPretty: true,
			ExpectedOutput: validationResult{
				IsValid: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			model := authorizationmodel.AuthzModel{}

			err := model.ReadFromJSONString(test.Input)
			if err != nil {
				return
			}

			output := validate(model)

			if test.NoPretty {
				outputJSON, _ := json.Marshal(output)
				expectedJSON, _ := json.Marshal(test.ExpectedOutput)

				if string(outputJSON) != string(expectedJSON) {
					t.Fatalf("Expect output %s actual %s", string(expectedJSON), string(outputJSON))
				}
			} else if !reflect.DeepEqual(output, test.ExpectedOutput) {
				t.Fatalf("Expect output %v actual %v", test.ExpectedOutput, output)
			}
		})
	}
}
