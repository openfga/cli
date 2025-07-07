package tuple

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestTupleLoggerCSVIncludesAllColumns(t *testing.T) {
	t.Parallel()

	tmp, err := os.CreateTemp(t.TempDir(), "log*.csv")
	require.NoError(t, err)
	tmp.Close()

	logger, err := NewTupleLogger(tmp.Name())
	require.NoError(t, err)
	defer logger.Close()

	key := client.ClientTupleKey{
		User:     "user:anne",
		Relation: "viewer",
		Object:   "document:1",
	}

	logger.LogSuccess(key)
	require.NoError(t, logger.Close())

	data, err := os.ReadFile(tmp.Name())
	require.NoError(t, err)

	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 2)

	header := []string{"user_type", "user_id", "user_relation", "relation", "object_type", "object_id", "condition_name", "condition_context"}
	require.Equal(t, header, records[0])

	expected := []string{"user", "anne", "", "viewer", "document", "1", "", ""}
	require.Equal(t, expected, records[1])
}

func TestTupleLoggerFormatsSuccessAndFailure(t *testing.T) {
	t.Parallel()

	condCtx := map[string]interface{}{"ip_addr": "10.0.0.1"}
	key := client.ClientTupleKey{
		User:     "user:anne10",
		Relation: "owner",
		Object:   "group:foo",
		Condition: &openfga.RelationshipCondition{
			Name:    "inOfficeIP",
			Context: &condCtx,
		},
	}

	formats := []string{".csv", ".json", ".jsonl", ".yaml"}
	for _, ext := range formats {
		ext := ext
		t.Run("success"+ext, func(t *testing.T) {
			tmp, err := os.CreateTemp(t.TempDir(), "success*"+ext)
			require.NoError(t, err)
			tmp.Close()

			logger, err := NewTupleLogger(tmp.Name())
			require.NoError(t, err)
			logger.LogSuccess(key)
			require.NoError(t, logger.Close())

			verifyLoggedTuple(t, tmp.Name(), ext, key)
		})

		t.Run("failure"+ext, func(t *testing.T) {
			tmp, err := os.CreateTemp(t.TempDir(), "failure*"+ext)
			require.NoError(t, err)
			tmp.Close()

			logger, err := NewTupleLogger(tmp.Name())
			require.NoError(t, err)
			logger.LogFailure(key)
			require.NoError(t, logger.Close())

			verifyLoggedTuple(t, tmp.Name(), ext, key)
		})
	}
}

func verifyLoggedTuple(t *testing.T, path, ext string, expected client.ClientTupleKey) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	switch ext {
	case ".csv":
		r := csv.NewReader(bytes.NewReader(data))
		records, err := r.ReadAll()
		require.NoError(t, err)
		require.Len(t, records, 2)
		header := []string{"user_type", "user_id", "user_relation", "relation", "object_type", "object_id", "condition_name", "condition_context"}
		require.Equal(t, header, records[0])
		expectedRow := []string{"user", "anne10", "", "owner", "group", "foo", "inOfficeIP", "{\"ip_addr\":\"10.0.0.1\"}"}
		require.Equal(t, expectedRow, records[1])
	case ".json":
		var got client.ClientTupleKey
		require.NoError(t, json.Unmarshal(data, &got))
		require.Equal(t, expected, got)
	case ".jsonl":
		var got client.ClientTupleKey
		require.NoError(t, json.Unmarshal(bytes.TrimSpace(data), &got))
		require.Equal(t, expected, got)
	case ".yaml", ".yml":
		var got []client.ClientTupleKey
		require.NoError(t, yaml.Unmarshal(data, &got))
		require.Len(t, got, 1)
		require.Equal(t, expected, got[0])
	default:
		t.Fatalf("unknown extension %s", ext)
	}
}
