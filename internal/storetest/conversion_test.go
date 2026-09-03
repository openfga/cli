package storetest

import (
	"testing"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertPbUsersToStrings(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *openfgav1.User
		expected string
	}{
		"User_Object": {
			input:    &openfgav1.User{User: &openfgav1.User_Object{Object: &openfgav1.Object{Type: "user", Id: "anne"}}},
			expected: "user:anne",
		},
		"User_Userset": {
			input:    &openfgav1.User{User: &openfgav1.User_Userset{Userset: &openfgav1.UsersetUser{Type: "group", Id: "fga", Relation: "member"}}},
			expected: "group:fga#member",
		},
		"User_Wildcard": {
			input:    &openfgav1.User{User: &openfgav1.User_Wildcard{Wildcard: &openfgav1.TypedWildcard{Type: "user"}}},
			expected: "user:*",
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := convertPbUsersToStrings([]*openfgav1.User{testcase.input})

			assert.Equal(t, []string{testcase.expected}, got)
		})
	}
}

func TestConvertOpenfgaUsersToStrings(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    openfga.User
		expected string
	}{
		"User_Object": {
			input:    openfga.User{Object: &openfga.FgaObject{Type: "user", Id: "anne"}},
			expected: "user:anne",
		},
		"User_Userset": {
			input:    openfga.User{Userset: &openfga.UsersetUser{Type: "group", Id: "fga", Relation: "member"}},
			expected: "group:fga#member",
		},
		"User_Wildcard": {
			input:    openfga.User{Wildcard: &openfga.TypedWildcard{Type: "user"}},
			expected: "user:*",
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := convertOpenfgaUsers([]openfga.User{testcase.input})

			assert.Equal(t, []string{testcase.expected}, got)
		})
	}
}

func TestConvertStoreObjectToObject(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input             string
		expectedFGAObject openfga.FgaObject
		expectedPBObject  *openfgav1.Object
	}{
		"Converts object": {
			input:             "document:roadmap",
			expectedFGAObject: openfga.FgaObject{Type: "document", Id: "roadmap"},
			expectedPBObject:  &openfgav1.Object{Type: "document", Id: "roadmap"},
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fgaObject, pbObject := convertStoreObjectToObject(testcase.input)

			assert.Equal(t, testcase.expectedFGAObject, fgaObject)
			assert.Equal(t, testcase.expectedPBObject, pbObject)
		})
	}
}

func TestConvertClientTupleKeysToProtoTupleKeys(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    []client.ClientContextualTupleKey
		expected []*openfgav1.TupleKey
	}{
		"User_Object": {
			input: []client.ClientContextualTupleKey{
				{User: "user:anne", Relation: "owner", Object: "folder:product"},
			},
			expected: []*openfgav1.TupleKey{
				{User: "user:anne", Relation: "owner", Object: "folder:product"},
			},
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tuples, err := convertClientTupleKeysToProtoTupleKeys(testcase.input)

			require.NoError(t, err)
			assert.Equal(t, testcase.expected, tuples)
		})
	}
}
