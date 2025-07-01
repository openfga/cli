package storetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	file := filepath.Join(dir, name)

	err := os.WriteFile(file, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return file
}

func TestLoadTuples(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	tupleContent := `[
  {"user": "user:jon", "relation": "viewer", "object": "document:doc1"},
  {"user": "user:sam", "relation": "editor", "object": "document:doc2"}
]`

	tupleFile := writeTempFile(t, tempDir, "tuples1.json", tupleContent)

	tupleFile2 := writeTempFile(t, tempDir, "tuples2.json", `[
  {"user": "user:jon", "relation": "viewer", "object": "document:doc1"},
  {"user": "user:amy", "relation": "editor", "object": "document:doc3"}
]`)

	cases := []struct {
		name         string
		storeData    StoreData
		expectErr    bool
		expectTuples int
	}{
		{
			name:      "no tuple file or tuple files",
			storeData: StoreData{},
			expectErr: true,
		},
		{
			name: "single tuple_file",
			storeData: StoreData{
				TupleFile: filepath.Base(tupleFile),
			},
			expectTuples: 2,
		},
		{
			name: "multiple tuple_files no dedup",
			storeData: StoreData{
				TupleFiles: []string{filepath.Base(tupleFile), filepath.Base(tupleFile2)},
			},
			expectTuples: 4,
		},
		{
			name: "combined tuple_file and tuple_files",
			storeData: StoreData{
				TupleFile:  filepath.Base(tupleFile),
				TupleFiles: []string{filepath.Base(tupleFile2)},
			},
			expectTuples: 4,
		},
		{
			name: "test-level tuple file with error",
			storeData: StoreData{
				TupleFile: filepath.Base(tupleFile),
				Tests: []ModelTest{{
					Name:      "test1",
					TupleFile: "invalid.json",
				}},
			},
			expectErr: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.storeData.LoadTuples(tempDir)

			if testCase.expectErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if got := len(testCase.storeData.Tuples); got != testCase.expectTuples {
					t.Errorf("expected %d tuples, got %d", testCase.expectTuples, got)
				}
			}
		})
	}


func TestStoreDataValidate(t *testing.T) {
	t.Parallel()

	validSingle := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Object: "doc:1", Assertions: map[string]bool{"read": true},
	}}}}}
	assert.NoError(t, validSingle.Validate())

	validUsers := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		Users: []string{"user:1", "user:2"}, Object: "doc:1", Assertions: map[string]bool{"read": true},
	}}}}}
	assert.NoError(t, validUsers.Validate())

	validObjects := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Objects: []string{"doc:1", "doc:2"}, Assertions: map[string]bool{"read": true},
	}}}}}
	assert.NoError(t, validObjects.Validate())

	invalidBoth := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Users: []string{"user:2"}, Object: "doc:1", Assertions: map[string]bool{"read": true},
	}}}}}
	require.Error(t, invalidBoth.Validate())

	invalidObjectBoth := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Object: "doc:1", Objects: []string{"doc:2"}, Assertions: map[string]bool{"read": true},
	}}}}}
	require.Error(t, invalidObjectBoth.Validate())

	invalidNone := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Assertions: map[string]bool{"read": true},
	}}}}}
	require.Error(t, invalidNone.Validate())
}

func TestModelTestCheckStructTagsOmitEmpty(t *testing.T) {
	t.Parallel()

	// Test that struct can handle single user/object
	checkSingle := ModelTestCheck{
		User:       "user:anne",
		Object:     "folder:product-2021",
		Assertions: map[string]bool{"can_view": true},
	}
	assert.NotEmpty(t, checkSingle.User)
	assert.Empty(t, checkSingle.Users)
	assert.NotEmpty(t, checkSingle.Object)
	assert.Empty(t, checkSingle.Objects)

	// Test that struct can handle multiple users
	checkUsers := ModelTestCheck{
		Users:      []string{"user:anne", "user:bob"},
		Object:     "folder:product-2021",
		Assertions: map[string]bool{"can_view": true},
	}
	assert.Empty(t, checkUsers.User)
	assert.NotEmpty(t, checkUsers.Users)
	assert.Len(t, checkUsers.Users, 2)

	// Test that struct can handle multiple objects
	checkObjects := ModelTestCheck{
		User:       "user:anne",
		Objects:    []string{"folder:product-2021", "folder:product-2022"},
		Assertions: map[string]bool{"can_view": true},
	}
	assert.NotEmpty(t, checkObjects.User)
	assert.Empty(t, checkObjects.Object)
	assert.NotEmpty(t, checkObjects.Objects)
	assert.Len(t, checkObjects.Objects, 2)
}
