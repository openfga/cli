package tuplefile

import (
	"bytes"
	"encoding/csv"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CSV header parsing tests

func TestReadHeaders_ValidRequiredHeaders(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		csv      string
		expected csvColumns
	}{
		{
			name: "minimal required headers",
			csv:  "user_type,user_id,relation,object_type,object_id\n",
			expected: csvColumns{
				UserType:         0,
				UserID:           1,
				UserRelation:     -1,
				Relation:         2,
				ObjectType:       3,
				ObjectID:         4,
				ConditionName:    -1,
				ConditionContext: -1,
			},
		},
		{
			name: "all headers present",
			csv:  "user_type,user_id,user_relation,relation,object_type,object_id,condition_name,condition_context\n",
			expected: csvColumns{
				UserType:         0,
				UserID:           1,
				UserRelation:     2,
				Relation:         3,
				ObjectType:       4,
				ObjectID:         5,
				ConditionName:    6,
				ConditionContext: 7,
			},
		},
		{
			name: "headers in different order",
			csv:  "object_id,relation,user_id,object_type,user_type\n",
			expected: csvColumns{
				UserType:         4,
				UserID:           2,
				UserRelation:     -1,
				Relation:         1,
				ObjectType:       3,
				ObjectID:         0,
				ConditionName:    -1,
				ConditionContext: -1,
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			reader := csv.NewReader(bytes.NewReader([]byte(testCase.csv)))
			columns, err := readHeaders(reader)
			require.NoError(t, err)
			require.NotNil(t, columns)
			assert.Equal(t, testCase.expected, *columns)

			var tuples []client.ClientTupleKey

			err = parseTuplesFromCSV([]byte(testCase.csv), &tuples)
			require.NoError(t, err)
		})
	}
}

func TestReadHeaders_MissingRequiredHeaders(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		csv         string
		expectedErr string
	}{
		{
			name:        "missing user_type",
			csv:         "user_id,relation,object_type,object_id\n",
			expectedErr: `"user_type"`,
		},
		{
			name:        "missing user_id",
			csv:         "user_type,relation,object_type,object_id\n",
			expectedErr: `"user_id"`,
		},
		{
			name:        "missing relation",
			csv:         "user_type,user_id,object_type,object_id\n",
			expectedErr: `"relation"`,
		},
		{
			name:        "missing object_type",
			csv:         "user_type,user_id,relation,object_id\n",
			expectedErr: `"object_type"`,
		},
		{
			name:        "missing object_id",
			csv:         "user_type,user_id,relation,object_type\n",
			expectedErr: `"object_id"`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var tuples []client.ClientTupleKey

			err := parseTuplesFromCSV([]byte(testCase.csv), &tuples)
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErr)
		})
	}
}

func TestReadHeaders_InvalidHeader(t *testing.T) {
	t.Parallel()

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte("user_type,user_id,relation,object_type,object_id,unknown_col\n"), &tuples)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid header "unknown_col"`)
}

func TestReadHeaders_EmptyInput(t *testing.T) {
	t.Parallel()

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(""), &tuples)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read csv headers")
}

func TestReadHeaders_ConditionContextWithoutConditionName(t *testing.T) {
	t.Parallel()

	csv := "user_type,user_id,relation,object_type,object_id,condition_context\n"

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(csv), &tuples)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `missing "condition_name" header`)
}

// CSV row parsing tests

func TestParseTuplesFromCSV_ValidRows(t *testing.T) {
	t.Parallel()

	csv := "user_type,user_id,relation,object_type,object_id\n" +
		"user,anne,can_view,document,roadmap\n" +
		"user,beth,editor,folder,finance\n"

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(csv), &tuples)
	require.NoError(t, err)
	require.Len(t, tuples, 2)

	assert.Equal(t, "user:anne", tuples[0].User)
	assert.Equal(t, "can_view", tuples[0].Relation)
	assert.Equal(t, "document:roadmap", tuples[0].Object)

	assert.Equal(t, "user:beth", tuples[1].User)
	assert.Equal(t, "editor", tuples[1].Relation)
	assert.Equal(t, "folder:finance", tuples[1].Object)
}

func TestParseTuplesFromCSV_WithUserRelation(t *testing.T) {
	t.Parallel()

	csv := "user_type,user_id,user_relation,relation,object_type,object_id\n" +
		"group,eng,member,can_view,document,roadmap\n"

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(csv), &tuples)
	require.NoError(t, err)
	require.Len(t, tuples, 1)
	assert.Equal(t, "group:eng#member", tuples[0].User)
}

func TestParseTuplesFromCSV_EmptyUserRelationIgnored(t *testing.T) {
	t.Parallel()

	csv := "user_type,user_id,user_relation,relation,object_type,object_id\n" +
		"user,anne,,can_view,document,roadmap\n"

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(csv), &tuples)
	require.NoError(t, err)
	require.Len(t, tuples, 1)
	assert.Equal(t, "user:anne", tuples[0].User)
}

func TestParseTuplesFromCSV_WithCondition(t *testing.T) {
	t.Parallel()

	csv := "user_type,user_id,relation,object_type,object_id,condition_name,condition_context\n" +
		`user,anne,can_view,document,roadmap,inOffice,"{""ip"":""10.0.0.1""}"` + "\n"

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(csv), &tuples)
	require.NoError(t, err)
	require.Len(t, tuples, 1)
	require.NotNil(t, tuples[0].Condition)
	assert.Equal(t, "inOffice", tuples[0].Condition.Name)
	assert.Equal(t, "10.0.0.1", (*tuples[0].Condition.Context)["ip"])
}

func TestParseTuplesFromCSV_ConditionNameWithoutContext(t *testing.T) {
	t.Parallel()

	csv := "user_type,user_id,relation,object_type,object_id,condition_name\n" +
		"user,anne,can_view,document,roadmap,inOffice\n"

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(csv), &tuples)
	require.NoError(t, err)
	require.Len(t, tuples, 1)
	require.NotNil(t, tuples[0].Condition)
	assert.Equal(t, "inOffice", tuples[0].Condition.Name)
}

func TestParseTuplesFromCSV_EmptyConditionNameSkipsCondition(t *testing.T) {
	t.Parallel()

	csv := "user_type,user_id,relation,object_type,object_id,condition_name\n" +
		"user,anne,can_view,document,roadmap,\n"

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(csv), &tuples)
	require.NoError(t, err)
	require.Len(t, tuples, 1)
	assert.Nil(t, tuples[0].Condition)
}

func TestParseTuplesFromCSV_InvalidConditionContext(t *testing.T) {
	t.Parallel()

	csv := "user_type,user_id,relation,object_type,object_id,condition_name,condition_context\n" +
		"user,anne,can_view,document,roadmap,inOffice,not-json\n"

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(csv), &tuples)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read condition context")
}

func TestParseTuplesFromCSV_HeadersOnlyNoRows(t *testing.T) {
	t.Parallel()

	csv := "user_type,user_id,relation,object_type,object_id\n"

	var tuples []client.ClientTupleKey

	err := parseTuplesFromCSV([]byte(csv), &tuples)
	require.NoError(t, err)
	assert.Empty(t, tuples)
}

// JSONL parsing tests

func TestParseTuplesFromJSONL_ValidLines(t *testing.T) {
	t.Parallel()

	jsonl := `{"user": "user:anne", "relation": "can_view", "object": "document:roadmap"}
{"user": "user:beth", "relation": "editor", "object": "folder:finance"}
`

	var tuples []client.ClientTupleKey

	err := parseTuplesFromJSONL([]byte(jsonl), &tuples)
	require.NoError(t, err)
	require.Len(t, tuples, 2)

	assert.Equal(t, "user:anne", tuples[0].User)
	assert.Equal(t, "can_view", tuples[0].Relation)
	assert.Equal(t, "document:roadmap", tuples[0].Object)

	assert.Equal(t, "user:beth", tuples[1].User)
	assert.Equal(t, "editor", tuples[1].Relation)
	assert.Equal(t, "folder:finance", tuples[1].Object)
}

func TestParseTuplesFromJSONL_SkipsBlankLines(t *testing.T) {
	t.Parallel()

	jsonl := `{"user": "user:anne", "relation": "can_view", "object": "document:roadmap"}

{"user": "user:beth", "relation": "editor", "object": "folder:finance"}
`

	var tuples []client.ClientTupleKey

	err := parseTuplesFromJSONL([]byte(jsonl), &tuples)
	require.NoError(t, err)
	assert.Len(t, tuples, 2)
}

func TestParseTuplesFromJSONL_WithCondition(t *testing.T) {
	t.Parallel()

	jsonl := `{"user": "user:anne", "relation": "can_view", "object": "document:roadmap", "condition": {"name": "inOffice", "context": {"ip": "10.0.0.1"}}}
`

	var tuples []client.ClientTupleKey

	err := parseTuplesFromJSONL([]byte(jsonl), &tuples)
	require.NoError(t, err)
	require.Len(t, tuples, 1)
	require.NotNil(t, tuples[0].Condition)
	assert.Equal(t, "inOffice", tuples[0].Condition.Name)

	ctx := tuples[0].Condition.Context
	require.NotNil(t, ctx)
	assert.Equal(t, "10.0.0.1", (*ctx)["ip"])
}

func TestParseTuplesFromJSONL_InvalidJSON(t *testing.T) {
	t.Parallel()

	jsonl := `{"user": "user:anne", "relation": "can_view"
`

	var tuples []client.ClientTupleKey

	err := parseTuplesFromJSONL([]byte(jsonl), &tuples)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read tuple from jsonl file on line 1")
}

func TestParseTuplesFromJSONL_EmptyInput(t *testing.T) {
	t.Parallel()

	var tuples []client.ClientTupleKey

	err := parseTuplesFromJSONL([]byte(""), &tuples)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestParseTuplesFromJSONL_OnlyBlankLines(t *testing.T) {
	t.Parallel()

	var tuples []client.ClientTupleKey

	err := parseTuplesFromJSONL([]byte("\n\n  \n"), &tuples)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// parseConditionColumnsForRow tests

func TestParseConditionColumnsForRow_NoConditionColumns(t *testing.T) {
	t.Parallel()

	columns := &csvColumns{
		ConditionName:    -1,
		ConditionContext: -1,
	}
	tuple := []string{"user", "anne", "can_view", "document", "roadmap"}

	condition, err := parseConditionColumnsForRow(columns, tuple, 0)
	require.NoError(t, err)
	assert.Nil(t, condition)
}

func TestParseConditionColumnsForRow_WithConditionNameAndContext(t *testing.T) {
	t.Parallel()

	columns := &csvColumns{
		ConditionName:    5,
		ConditionContext: 6,
	}
	tuple := []string{"user", "anne", "can_view", "document", "roadmap", "inOffice", `{"ip":"10.0.0.1"}`}

	condition, err := parseConditionColumnsForRow(columns, tuple, 0)
	require.NoError(t, err)
	require.NotNil(t, condition)
	assert.Equal(t, "inOffice", condition.Name)

	expected := &openfga.RelationshipCondition{
		Name:    "inOffice",
		Context: &map[string]any{"ip": "10.0.0.1"},
	}
	assert.Equal(t, expected.Name, condition.Name)
	assert.Equal(t, (*expected.Context)["ip"], (*condition.Context)["ip"])
}

func TestParseConditionColumnsForRow_EmptyConditionName(t *testing.T) {
	t.Parallel()

	columns := &csvColumns{
		ConditionName:    5,
		ConditionContext: -1,
	}
	tuple := []string{"user", "anne", "can_view", "document", "roadmap", ""}

	condition, err := parseConditionColumnsForRow(columns, tuple, 0)
	require.NoError(t, err)
	assert.Nil(t, condition)
}

// csvColumns.validate tests

func TestCsvColumnsValidate_ConditionContextWithoutName(t *testing.T) {
	t.Parallel()

	columns := &csvColumns{
		UserType:         0,
		UserID:           1,
		UserRelation:     -1,
		Relation:         2,
		ObjectType:       3,
		ObjectID:         4,
		ConditionName:    -1,
		ConditionContext: 5,
	}

	err := columns.validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `missing "condition_name" header`)
}
