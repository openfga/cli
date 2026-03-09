package storetest

import (
	"testing"

	"github.com/openfga/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openfga/cli/internal/authorizationmodel"
)

func TestRunLocalListObjectsTest(t *testing.T) {
	t.Parallel()

	const modelDSL = `model
  schema 1.1

type user

type document
  relations
    define viewer: [user]`

	tests := map[string]struct {
		tuples         []client.ClientContextualTupleKey
		listObjectTest ModelTestListObjects
		expectedGot    []string
		expectedPass   bool
		expectError    bool
	}{
		"returns_matching_objects_when_tuples_exist": {
			tuples: []client.ClientContextualTupleKey{
				{User: "user:anne", Relation: "viewer", Object: "document:roadmap"},
				{User: "user:anne", Relation: "viewer", Object: "document:budget"},
			},
			listObjectTest: ModelTestListObjects{
				User: "user:anne",
				Type: "document",
				Assertions: map[string][]string{
					"viewer": {"document:roadmap", "document:budget"},
				},
			},
			expectedGot:  []string{"document:roadmap", "document:budget"},
			expectedPass: true,
			expectError:  false,
		},
		"returns_empty_slice_when_no_matching_tuples": {
			tuples: []client.ClientContextualTupleKey{},
			listObjectTest: ModelTestListObjects{
				User: "user:anne",
				Type: "document",
				Assertions: map[string][]string{
					"viewer": {},
				},
			},
			expectedGot:  []string{},
			expectedPass: true,
			expectError:  false,
		},
		"returns_empty_slice_when_user_has_no_access": {
			tuples: []client.ClientContextualTupleKey{
				{User: "user:bob", Relation: "viewer", Object: "document:roadmap"},
			},
			listObjectTest: ModelTestListObjects{
				User: "user:anne",
				Type: "document",
				Assertions: map[string][]string{
					"viewer": {},
				},
			},
			expectedGot:  []string{},
			expectedPass: true,
			expectError:  false,
		},
		"test_fails_when_expected_objects_do_not_match": {
			tuples: []client.ClientContextualTupleKey{
				{User: "user:anne", Relation: "viewer", Object: "document:roadmap"},
			},
			listObjectTest: ModelTestListObjects{
				User: "user:anne",
				Type: "document",
				Assertions: map[string][]string{
					"viewer": {"document:roadmap", "document:budget"},
				},
			},
			expectedGot:  []string{"document:roadmap"},
			expectedPass: false,
			expectError:  false,
		},
		"test_fails_when_got_is_empty_but_expected_is_not": {
			tuples: []client.ClientContextualTupleKey{},
			listObjectTest: ModelTestListObjects{
				User: "user:anne",
				Type: "document",
				Assertions: map[string][]string{
					"viewer": {"document:roadmap"},
				},
			},
			expectedGot:  []string{},
			expectedPass: false,
			expectError:  false,
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			storeData := &StoreData{Model: modelDSL}
			config := LocalServerConfig{MaxTypesPerAuthorizationModel: 100}

			fgaServer, authModel, stopFn, err := getLocalServerModelAndTuples(storeData, authorizationmodel.ModelFormatDefault, config)
			require.NoError(t, err)

			defer stopFn()

			storeID, modelID, err := initLocalStore(t.Context(), fgaServer, authModel.GetProtoModel(), testCase.tuples)
			require.NoError(t, err)

			options := ModelTestOptions{
				StoreID: storeID,
				ModelID: modelID,
			}

			results := RunLocalListObjectsTest(
				t.Context(),
				fgaServer,
				testCase.listObjectTest,
				testCase.tuples,
				options,
			)

			require.Len(t, results, len(testCase.listObjectTest.Assertions))

			for _, result := range results {
				if testCase.expectError {
					require.Error(t, result.Error)
					assert.Nil(t, result.Got)
				} else {
					require.NoError(t, result.Error)
					require.NotNil(t, result.Got)
					assert.ElementsMatch(t, testCase.expectedGot, result.Got)
					assert.Equal(t, testCase.expectedPass, result.TestResult)
				}
			}
		})
	}
}

func TestRunLocalListObjectsTest_GotIsNeverNilOnSuccess(t *testing.T) {
	t.Parallel()

	const modelDSL = `model
  schema 1.1

type user

type document
  relations
    define viewer: [user]`

	storeData := &StoreData{Model: modelDSL}
	config := LocalServerConfig{MaxTypesPerAuthorizationModel: 100}

	fgaServer, authModel, stopFn, err := getLocalServerModelAndTuples(storeData, authorizationmodel.ModelFormatDefault, config)
	require.NoError(t, err)

	defer stopFn()

	storeID, modelID, err := initLocalStore(t.Context(), fgaServer, authModel.GetProtoModel(), nil)
	require.NoError(t, err)

	options := ModelTestOptions{
		StoreID: storeID,
		ModelID: modelID,
	}

	listObjectsTest := ModelTestListObjects{
		User: "user:anne",
		Type: "document",
		Assertions: map[string][]string{
			"viewer": {},
		},
	}

	results := RunLocalListObjectsTest(
		t.Context(),
		fgaServer,
		listObjectsTest,
		nil,
		options,
	)

	require.Len(t, results, 1)

	result := results[0]
	require.NoError(t, result.Error)
	require.NotNil(t, result.Got, "Got should be a non-nil empty slice, not nil")
	assert.Empty(t, result.Got)
	assert.True(t, result.TestResult, "test should pass when expected and got are both empty")
}
