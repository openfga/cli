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
		Name           string
		Input          string
		IsValid        bool
		ExpectedOutput validationResult
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
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			model := authorizationmodel.AuthzModel{}

			err := model.ReadFromJSONString(test.Input)
			if err != nil {
				return
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
