package model

import (
	"reflect"
	"strings"
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

func TestValidateCmdWithWeightedGraphFlag(t *testing.T) {
	// Not parallel: validateCmd is shared, and the flag value persists across tests.
	t.Cleanup(func() {
		ModelCmd.SetArgs(nil)

		if err := validateCmd.Flags().Set("weighted-graph", "false"); err != nil {
			t.Errorf("failed to reset weighted-graph flag: %v", err)
		}
	})

	incompatibleModel := `model
  schema 1.1
type user
type organization
  relations
    define member: [user]
type folder
  relations
    define viewer: [user]
type document
  relations
    define parent: [organization, folder]
    define viewer: member from parent
`

	// Execute runs from the root command, so the args must be set on ModelCmd.
	ModelCmd.SetArgs([]string{"validate", "--weighted-graph", incompatibleModel})

	err := ModelCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "weighted-graph:") {
		t.Fatalf("expected error containing %q, got %v", "weighted-graph:", err)
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

			output := validate(model, false)

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

	result := validate(model, false)
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

	result := validate(model, false)
	if !result.IsValid {
		t.Fatalf("expected valid model, got error: %v", *result.Error)
	}

	if result.SizeKB == nil {
		t.Fatalf("expected SizeKB to be set")
	}
}

func TestValidateWeightedGraph(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		model         string
		expectedError string
	}{
		{
			name: "direct relation",
			model: `model
  schema 1.1
type user
type document
  relations
    define viewer: [user]
`,
		},
		{
			name: "union recursion",
			model: `model
  schema 1.1
type user
type folder
  relations
    define parent: [folder]
    define viewer: [user] or viewer from parent
`,
		},
		{
			name: "nonrecursive intersection and exclusion",
			model: `model
  schema 1.1
type user
type document
  relations
    define editor: [user]
    define allowed: [user]
    define blocked: [user]
    define can_edit: (editor and allowed) but not blocked
`,
		},
		{
			name: "parent type missing computed relation",
			model: `model
  schema 1.1
type user
type organization
type folder
  relations
    define viewer: [user]
type document
  relations
    define parent: [folder, organization]
    define viewer: viewer from parent
`,
			expectedError: "relation",
		},
		{
			name: "intersection without a common user type",
			model: `model
  schema 1.1
type user
type organization
type document
  relations
    define editor: [user]
    define publisher: [organization]
    define can_publish: editor and publisher
`,
			expectedError: "not all paths return the same type",
		},
		{
			name: "intersection in a recursive cycle",
			model: `model
  schema 1.1
type user
type document
  relations
    define allowed: [user]
    define viewer: [user, document#viewer] and allowed
`,
			expectedError: "cycle",
		},
		{
			name: "exclusion in a recursive cycle",
			model: `model
  schema 1.1
type user
type document
  relations
    define viewer: [user] but not restricted
    define restricted: [user, document#viewer]
`,
			expectedError: "cycle",
		},
		// Cases below are from https://openfga.dev/blog/weighted-graph-upcoming-changes
		{
			name: "blog: missing relation",
			model: `model
  schema 1.1
type user
type organization
  relations
    define member: [user]
type folder
  relations
    define viewer: [user]
type document
  relations
    define parent: [organization, folder]
    define viewer: member from parent
`,
			expectedError: "relation",
		},
		{
			name: "blog: missing relation fixed with alias",
			model: `model
  schema 1.1
type user
type organization
  relations
    define member: [user]
type folder
  relations
    define viewer: [user]
    define member: viewer
type document
  relations
    define parent: [organization, folder]
    define viewer: member from parent
`,
		},
		{
			name: "blog: missing relation fixed with explicit per-type relations",
			model: `model
  schema 1.1
type user
type organization
  relations
    define member: [user]
type folder
  relations
    define viewer: [user]
type document
  relations
    define organization: [organization]
    define folder: [folder]
    define viewer: member from organization or viewer from folder
`,
		},
		{
			name: "blog: tuple cycle with intersection",
			model: `model
  schema 1.1
type user
type group
  relations
    define approved: [user, group#member]
    define member: [user, group#member] and approved
`,
			expectedError: "cycle",
		},
		{
			name: "blog: tuple cycle with exclusion",
			model: `model
  schema 1.1
type user
type group
  relations
    define blocked: [user, group#member]
    define member: [user, group#member] but not blocked
`,
			expectedError: "cycle",
		},
		{
			name: "blog: tuple cycle fixed by splitting base and gated relations",
			model: `model
  schema 1.1
type user
type group
  relations
    define blocked: [user, group#member]
    define member: [user, group#member]
    define allowed_member: member but not blocked
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := authorizationmodel.AuthzModel{}
			if err := model.ReadModelFromString(test.model, "fga"); err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if result := validate(model, false); !result.IsValid {
				t.Fatalf("expected normal validation to succeed: %v", *result.Error)
			}

			result := validate(model, true)
			if result.IsValid != (test.expectedError == "") {
				t.Fatalf("unexpected validation result: %+v", result)
			}
			if test.expectedError == "" {
				if result.Error != nil {
					t.Fatalf("unexpected error: %s", *result.Error)
				}
			} else {
				if result.Error == nil {
					t.Fatal("expected an explanation for weighted graph incompatibility")
				}
				if !strings.HasPrefix(*result.Error, "weighted-graph: ") ||
					!strings.Contains(*result.Error, test.expectedError) {
					t.Fatalf("expected error containing %q, got %s", test.expectedError, *result.Error)
				}
			}
			if result.SizeKB == nil {
				t.Fatal("expected model size to be preserved")
			}
		})
	}
}

func TestValidateWeightedGraphSkipsTypeSystemValidation(t *testing.T) {
	t.Parallel()

	model := authorizationmodel.AuthzModel{}
	if err := model.ReadFromJSONString(`{"schema_version":"1.1","type_definitions":[{"type":"user"},{"type":"user"}]}`); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if result := validate(model, false); result.IsValid {
		t.Fatal("expected normal validation to reject duplicate types")
	}

	if result := validate(model, true); !result.IsValid {
		t.Fatalf("expected weighted graph validation to skip type-system validation, got: %+v", result)
	}
}
