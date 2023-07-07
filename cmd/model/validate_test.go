package model

import (
	"testing"
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
		ExpectedOutput string
	}

	tests := []validationTest{
		{
			Name:           "missing schema version",
			Input:          "{",
			IsValid:        false,
			ExpectedOutput: `{"is_valid":false,"error":"unable to parse json input"}`,
		},
		{
			Name:           "missing schema version",
			Input:          `{"id":"abcde","schema_version":"1.1"}`,
			IsValid:        false,
			ExpectedOutput: `{"id":"abcde","is_valid":false,"error":"unable to parse id: invalid ulid format"}`,
		},
		{
			Name:           "missing schema version",
			Input:          "{}",
			IsValid:        false,
			ExpectedOutput: `{"is_valid":false,"error":"invalid schema version"}`,
		},
		{
			Name:           "invalid schema version",
			Input:          `{"schema_version":"1.0"}`,
			IsValid:        false,
			ExpectedOutput: `{"is_valid":false,"error":"invalid schema version"}`,
		},
		{
			Name:           "only schema",
			Input:          `{"schema_version":"1.1"}`,
			IsValid:        true,
			ExpectedOutput: `{"is_valid":true}`,
		},
	}

	for index := 0; index < len(tests); index++ {
		test := tests[index]

		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			output, err := validate(test.Input)
			if err != nil {
				t.Fatalf("%v", err)
			}

			if output != test.ExpectedOutput {
				t.Fatalf("Expect output %v actual %v", test.ExpectedOutput, output)
			}
		})
	}
}
