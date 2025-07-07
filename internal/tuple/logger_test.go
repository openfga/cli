package tuple

import (
	"bytes"
	"encoding/csv"
	"os"
	"testing"

	"github.com/openfga/go-sdk/client"
	"github.com/stretchr/testify/require"
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
