package model

import (
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
		Name             string
		Input            string
		ExpectParseError bool
		ExpectedOutput   validationResult
	}

	tests := []validationTest{
		{
			Name:             "invalid json",
			Input:            "{",
			ExpectParseError: true,
		},
		{
			Name:  "invalid model id",
			Input: `{"id":"abcde","schema_version":"1.1"}`,
			ExpectedOutput: validationResult{
				ID:      "abcde",
				IsValid: false,
				Error:   openfga.PtrString("unable to parse id: invalid ulid format"),
			},
		},
		{
			Name:  "missing schema version",
			Input: "{}",
			ExpectedOutput: validationResult{
				IsValid: false,
				Error:   openfga.PtrString("invalid schema version"),
			},
		},
		{
			Name:  "invalid schema version",
			Input: `{"schema_version":"1.0"}`,
			ExpectedOutput: validationResult{
				IsValid: false,
				Error:   openfga.PtrString("invalid schema version"),
			},
		},
		{
			Name:  "only schema",
			Input: `{"schema_version":"1.1"}`,
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
			if test.ExpectParseError {
				if err == nil {
					t.Fatalf("Expected parse error for input %q, got none", test.Input)
				}

				return
			}

			if err != nil {
				t.Fatalf("Unexpected parse error for input %q: %v", test.Input, err)
			}

			expected := test.ExpectedOutput
			size := model.GetSizeInKB()
			expected.SizeKB = &size

			output := validate(model)

			if !reflect.DeepEqual(output, expected) {
				t.Fatalf("Expect output %v actual %v", expected, output)
			}
		})
	}
}

func TestValidateReportsSize(t *testing.T) {
	t.Parallel()

	model := authorizationmodel.AuthzModel{}
	if err := model.ReadFromJSONString(`{"schema_version":"1.1"}`); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := validate(model)
	if result.SizeKB == nil {
		t.Fatalf("expected SizeKB to be set")
	}

	if *result.SizeKB != model.GetSizeInKB() {
		t.Errorf("expected %v to equal %v", *result.SizeKB, model.GetSizeInKB())
	}
}

func TestValidateWithValidIDReportsSize(t *testing.T) {
	t.Parallel()

	model := authorizationmodel.AuthzModel{}
	if err := model.ReadFromJSONString(`{"id":"01GVKXGDCV2SMG6TRE9NMBQ2VG","schema_version":"1.1"}`); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := validate(model)
	if !result.IsValid {
		t.Fatalf("expected valid model, got error: %v", *result.Error)
	}

	if result.SizeKB == nil {
		t.Fatalf("expected SizeKB to be set")
	}
}
