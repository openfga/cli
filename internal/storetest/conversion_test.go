package storetest

import (
	"testing"

	pb "github.com/openfga/api/proto/openfga/v1"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertPbUsersToStrings(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *pb.User
		expected string
	}{
		"User_Object": {
			input:    &pb.User{User: &pb.User_Object{Object: &pb.Object{Type: "user", Id: "anne"}}},
			expected: "user:anne",
		},
		"User_Userset": {
			input:    &pb.User{User: &pb.User_Userset{Userset: &pb.UsersetUser{Type: "group", Id: "fga", Relation: "member"}}},
			expected: "group:fga#member",
		},
		"User_Wildcard": {
			input:    &pb.User{User: &pb.User_Wildcard{Wildcard: &pb.TypedWildcard{Type: "user"}}},
			expected: "user:*",
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := convertPbUsersToStrings([]*pb.User{testcase.input})

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
		expectedPBObject  *pb.Object
	}{
		"Converts object": {
			input:             "document:roadmap",
			expectedFGAObject: openfga.FgaObject{Type: "document", Id: "roadmap"},
			expectedPBObject:  &pb.Object{Type: "document", Id: "roadmap"},
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
		expected []*pb.TupleKey
	}{
		"User_Object": {
			input: []client.ClientContextualTupleKey{
				{User: "user:anne", Relation: "owner", Object: "folder:product"},
			},
			expected: []*pb.TupleKey{
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

func TestConvertPbObjectOrUsersetToStrings(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *pb.ObjectOrUserset
		expected string
	}{
		"User_Object": {
			input: &pb.ObjectOrUserset{
				User: &pb.ObjectOrUserset_Object{
					Object: &pb.Object{
						Type: "user",
						Id:   "anne",
					},
				},
			},
			expected: "user:anne",
		},
		"User_Userset": {
			input: &pb.ObjectOrUserset{
				User: &pb.ObjectOrUserset_Userset{
					Userset: &pb.UsersetUser{
						Type:     "group",
						Id:       "fga",
						Relation: "member",
					},
				},
			},
			expected: "group:fga#member",
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := convertPbObjectOrUsersetToStrings([]*pb.ObjectOrUserset{testcase.input})

			assert.Equal(t, []string{testcase.expected}, got)
		})
	}
}

func TestConvertObjectOrUsersetToStrings(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    openfga.ObjectOrUserset
		expected string
	}{
		"User_Object": {
			input:    openfga.ObjectOrUserset{Object: &openfga.FgaObject{Type: "user", Id: "anne"}},
			expected: "user:anne",
		},
		"User_Userset": {
			input:    openfga.ObjectOrUserset{Userset: &openfga.UsersetUser{Type: "group", Id: "fga", Relation: "member"}},
			expected: "group:fga#member",
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := convertOpenfgaObjectOrUserset([]openfga.ObjectOrUserset{testcase.input})

			assert.Equal(t, []string{testcase.expected}, got)
		})
	}
}
