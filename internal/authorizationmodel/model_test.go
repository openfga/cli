package authorizationmodel_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/openfga/pkg/typesystem"

	"github.com/openfga/cli/internal/authorizationmodel"
)

const (
	modelID        = "01GVKXGDCV2SMG6TRE9NMBQ2VG"
	typeName       = "user"
	modelCreatedAt = "2023-03-16 00:35:51.835 +0000 UTC"
)

func TestReadingInvalidModelFromInvalidJSON(t *testing.T) {
	t.Parallel()

	modelString := "{bad_json"

	model := authorizationmodel.AuthzModel{}

	err := model.ReadFromJSONString(modelString)
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestReadingValidModelFromJSON(t *testing.T) {
	t.Parallel()

	modelString := `{"id":"01GVKXGDCV2SMG6TRE9NMBQ2VG","schema_version":"1.1","type_definitions":[{"type":"user"}]}`

	model := authorizationmodel.AuthzModel{}

	err := model.ReadFromJSONString(modelString)
	if err != nil {
		t.Errorf("Got error when reading a valid model %v", err)
	}

	if model.GetSchemaVersion() != typesystem.SchemaVersion1_1 {
		t.Errorf("Expected %v to equal %v", model.GetSchemaVersion(), typesystem.SchemaVersion1_1)
	}

	if model.GetID() != modelID {
		t.Errorf("Expected %v to equal %v", model.GetID(), modelID)
	}

	if model.CreatedAt.String() != modelCreatedAt {
		t.Errorf("Expected %v to equal %v", model.GetCreatedAt().String(), modelCreatedAt)
	}

	if model.GetTypeDefinitions()[0].GetType() != typeName {
		t.Errorf("Expected %v to equal %v", model.GetTypeDefinitions()[0].GetType(), typeName)
	}
}

func TestReadingValidModelFromDSL(t *testing.T) {
	t.Parallel()

	model := authorizationmodel.AuthzModel{}
	if err := model.ReadFromDSLString(`model
  schema 1.1

type user
`); err != nil {
		t.Errorf("Got error when parsing a valid model %v", err)
	}

	if model.GetSchemaVersion() != typesystem.SchemaVersion1_1 {
		t.Errorf("Expected %v to equal %v", model.GetSchemaVersion(), typesystem.SchemaVersion1_1)
	}

	if model.GetTypeDefinitions()[0].GetType() != typeName {
		t.Errorf("Expected %v to equal %v", model.GetTypeDefinitions()[0].GetType(), typeName)
	}
}

func TestDisplayAsJsonWithFields(t *testing.T) {
	t.Parallel()

	typeDefs := []openfga.TypeDefinition{{
		Type: typeName,
	}}
	model := authorizationmodel.AuthzModel{
		SchemaVersion:   openfga.PtrString(typesystem.SchemaVersion1_1),
		ID:              openfga.PtrString(modelID),
		TypeDefinitions: &typeDefs,
	}

	jsonModel1 := model.DisplayAsJSON([]string{"model", "id", "created_at"})
	if jsonModel1.GetSchemaVersion() != typesystem.SchemaVersion1_1 {
		t.Errorf("Expected %v to equal %v", jsonModel1.GetSchemaVersion(), typesystem.SchemaVersion1_1)
	}

	if jsonModel1.GetID() != modelID {
		t.Errorf("Expected %v to equal %v", jsonModel1.GetID(), modelID)
	}

	if jsonModel1.GetCreatedAt().String() != modelCreatedAt {
		t.Errorf("Expected %v to equal %v", jsonModel1.GetCreatedAt().String(), modelCreatedAt)
	}

	if jsonModel1.GetTypeDefinitions()[0].GetType() != typeName {
		t.Errorf("Expected %v to equal %v", jsonModel1.GetTypeDefinitions()[0].GetType(), typeName)
	}

	jsonModel2 := model.DisplayAsJSON([]string{"id", "created_at"})
	if jsonModel2.GetSchemaVersion() != "" {
		t.Errorf("Expected %v to be empty", jsonModel2.GetSchemaVersion())
	}

	if jsonModel2.GetID() != modelID {
		t.Errorf("Expected %v to equal %v", jsonModel2.GetID(), modelID)
	}

	if jsonModel1.GetCreatedAt().String() != modelCreatedAt {
		t.Errorf("Expected %v to equal %v", jsonModel2.GetCreatedAt().String(), modelCreatedAt)
	}

	if jsonModel2.GetTypeDefinitions() != nil {
		t.Errorf("Expected %v to equal nil", jsonModel2.GetTypeDefinitions())
	}
}

func TestAuthzModel_WildcardSupport(t *testing.T) {
	t.Parallel()

	typeFragment := func(typeName string) string {
		return fmt.Sprintf(`type %s
  relations
    define owner: [user]`, typeName)
	}

	type testCase struct {
		name             string
		setupFiles       func(t *testing.T, dir string)
		modContent       string
		expectError      bool
		expectedErrMsg   string
		expectedTypeDefs []string
	}

	tests := []testCase{
		{
			name: "Wildcard expansion loads all module files",
			setupFiles: func(t *testing.T, dir string) {
				subdir := filepath.Join(dir, "subdir")
				require.NoError(t, os.MkdirAll(subdir, 0755))
				writeModuleFragment(t, subdir, "foo", typeFragment("foo"))
				writeModuleFragment(t, subdir, "bar", typeFragment("bar"))
				writeModuleFragment(t, subdir, "baz", typeFragment("baz"))
			},
			modContent: `schema: "1.2"
contents:
  - "subdir/*.fga"
`,
			expectError:      false,
			expectedTypeDefs: []string{"foo", "bar", "baz"},
		},
		{
			name: "Wildcard pattern matches no files",
			setupFiles: func(t *testing.T, dir string) {
				// No files created
			},
			modContent: `schema: "1.2"
contents:
  - "subdir/*.fga"
`,
			expectError:    true,
			expectedErrMsg: "no files matched",
		},
		{
			name: "Wildcard combined with direct file reference",
			setupFiles: func(t *testing.T, dir string) {
				subdir := filepath.Join(dir, "subdir")
				require.NoError(t, os.MkdirAll(subdir, 0755))
				writeModuleFragment(t, subdir, "foo", typeFragment("foo"))
				writeModuleFragment(t, dir, "core", typeFragment("core"))
			},
			modContent: `schema: "1.2"
contents:
  - "core.fga"
  - "subdir/*.fga"
`,
			expectError:      false,
			expectedTypeDefs: []string{"core", "foo"},
		},
		{
			name: "No files referenced at all",
			setupFiles: func(t *testing.T, dir string) {
				// No files created
			},
			modContent: `schema: "1.2"
contents: []
`,
			expectError:    false,
			expectedErrMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			tc.setupFiles(t, dir)

			modPath := filepath.Join(dir, "fga.mod")
			require.NoError(t, os.WriteFile(modPath, []byte(tc.modContent), 0644))

			model := authorizationmodel.AuthzModel{}
			err := model.ReadModelFromModFGA(modPath)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
				return
			}

			require.NoError(t, err)

			types := model.GetTypeDefinitions()
			got := map[string]bool{}
			for _, td := range types {
				got[td.Type] = true
			}

			for _, want := range tc.expectedTypeDefs {
				assert.True(t, got[want], "expected type %q to be loaded", want)
			}
			// Also check that no unexpected types are loaded
			assert.Equal(t, len(tc.expectedTypeDefs), len(got), "unexpected number of types loaded")
		})
	}
}

func writeModuleFragment(t *testing.T, dir, name, content string) {
	t.Helper()
	fullPath := filepath.Join(dir, name+".fga")
	fragment := "model\n  schema 1.2\n\n" + strings.TrimSpace(content) + "\n"
	require.NoError(t, os.WriteFile(fullPath, []byte(fragment), 0644))
}
