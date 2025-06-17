package storetest

import (
	"os"
	"path/filepath"
	"testing"
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
}
