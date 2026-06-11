package output

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCSVRow struct {
	fields []string
	err    error
}

func (r fakeCSVRow) MarshalCSV() ([]string, error) {
	return r.fields, r.err
}

func rows(values [][]string) []fakeCSVRow {
	records := make([]fakeCSVRow, len(values))
	for i, v := range values {
		records[i] = fakeCSVRow{fields: v}
	}

	return records
}

func TestMarshalCSV(t *testing.T) {
	t.Parallel()

	headers := []string{"user_type", "user_id", "relation", "object_type", "object_id", "condition_context"}

	tests := []struct {
		name     string
		records  [][]string
		expected string
	}{
		{
			name:     "no records writes only headers",
			records:  nil,
			expected: "user_type,user_id,relation,object_type,object_id,condition_context\n",
		},
		{
			name: "single record",
			records: [][]string{
				{"user", "john", "writer", "document", "abc.doc", ""},
			},
			expected: "user_type,user_id,relation,object_type,object_id,condition_context\n" +
				"user,john,writer,document,abc.doc,\n",
		},
		{
			name: "multiple records",
			records: [][]string{
				{"user", "anne", "reader", "document", "x", ""},
				{"group", "eng", "owner", "repo", "y", ""},
			},
			expected: "user_type,user_id,relation,object_type,object_id,condition_context\n" +
				"user,anne,reader,document,x,\n" +
				"group,eng,owner,repo,y,\n",
		},
		{
			name: "values with commas, quotes and newlines are escaped",
			records: [][]string{
				{"user", "a,b", "say \"hi\"", "doc", "line\nbreak", `{"ip_addr":"10.0.0.1"}`},
			},
			expected: "user_type,user_id,relation,object_type,object_id,condition_context\n" +
				"user,\"a,b\",\"say \"\"hi\"\"\",doc,\"line\nbreak\",\"{\"\"ip_addr\"\":\"\"10.0.0.1\"\"}\"\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			err := MarshalCSV(rows(test.records), &buf, headers...)
			require.NoError(t, err)
			assert.Equal(t, test.expected, buf.String())
		})
	}
}

func TestMarshalCSVNoHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	err := MarshalCSV(rows([][]string{{"user", "john"}}), &buf)
	require.NoError(t, err)
	assert.Equal(t, "user,john\n", buf.String())
}

func TestMarshalCSVRecordError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("boom")

	var buf bytes.Buffer

	err := MarshalCSV([]fakeCSVRow{{err: sentinel}}, &buf, "col")
	assert.ErrorIs(t, err, sentinel)
}
