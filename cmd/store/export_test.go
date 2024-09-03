package store

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	openfga "github.com/openfga/go-sdk"

	"github.com/openfga/cli/internal/fga"
	"github.com/openfga/cli/internal/storetest"
	"github.com/openfga/cli/internal/tuple"

	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"

	mock_client "github.com/openfga/cli/internal/mocks"
)

//nolint:funlen,cyclop
func TestExportSuccess(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	clientConfig := fga.ClientConfig{
		StoreID:              "12345",
		AuthorizationModelID: "01GXSA8YR785C4FYS3C0RTG7B1",
	}

	// Mocking Store GET...
	expectedTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	storeResponse := client.ClientGetStoreResponse{
		Id:        "12345",
		Name:      "Test store",
		CreatedAt: expectedTime,
		UpdatedAt: expectedTime,
	}

	mockExecute := mock_client.NewMockSdkClientGetStoreRequestInterface(mockCtrl)
	mockExecute.EXPECT().Execute().Return(&storeResponse, nil)
	mockFgaClient.EXPECT().GetStore(context.Background()).Return(mockExecute)

	// Mocking Authorization model GET...
	var modelResponse client.ClientReadAuthorizationModelResponse

	mockGetModelRequest := mock_client.NewMockSdkClientReadAuthorizationModelRequestInterface(mockCtrl)
	modelJSON := `{
	  "authorization_model": {
		"id": "01GXSA8YR785C4FYS3C0RTG7B1",
		"schema_version": "1.1",
		"type_definitions": [
		  {
			"type": "user"
		  },
		  {
			"type": "github-repo",        
			"relations": {
			  "viewer": {
				"this": {}
			  },
			  "admin": {
				"this": {}
			  }
			},
			"metadata": {
				"relations": {
				  "viewer": {
					"directly_related_user_types": [
					  {
						"type": "user"
					  }
					]
				  },
				  "admin": {
					"directly_related_user_types": [
					  {
						"type": "user"
					  }
					]
				  }
				}
			  }
		  }
		]
	  }
	}`

	if err := json.Unmarshal([]byte(modelJSON), &modelResponse); err != nil {
		t.Fatalf("%v", err)
	}

	getModelOptions := client.ClientReadAuthorizationModelOptions{
		AuthorizationModelId: openfga.PtrString("01GXSA8YR785C4FYS3C0RTG7B1"),
	}

	mockGetModelRequest.EXPECT().Options(getModelOptions).Return(mockGetModelRequest)
	mockGetModelRequest.EXPECT().Execute().Return(&modelResponse, nil)
	mockFgaClient.EXPECT().ReadAuthorizationModel(context.Background()).Return(mockGetModelRequest)

	// Mocking Tuples GET...
	readResponse := client.ClientReadResponse{
		Tuples: []openfga.Tuple{
			{
				Key: openfga.TupleKey{
					User:     "user:user-1",
					Relation: "viewer",
					Object:   "github-repo:demo",
				},
			},
			{
				Key: openfga.TupleKey{
					User:     "user:user-2",
					Relation: "viewer",
					Object:   "github-repo:demo",
				},
			},
			{
				Key: openfga.TupleKey{
					User:     "user:user-2",
					Relation: "admin",
					Object:   "github-repo:demo",
				},
			},
		},
	}

	readRequest := client.ClientReadRequest{}
	readOptions := client.ClientReadOptions{
		PageSize:          openfga.PtrInt32(tuple.DefaultReadPageSize),
		ContinuationToken: openfga.PtrString(""),
	}

	mockReadRequest := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)
	mockReadRequest.EXPECT().Body(readRequest).Return(mockReadRequest)
	mockReadRequest.EXPECT().Options(readOptions).Return(mockReadRequest)
	mockReadRequest.EXPECT().Execute().Return(&readResponse, nil)
	mockFgaClient.EXPECT().Read(context.Background()).Return(mockReadRequest)

	// Mocking assertions GET...
	assertionsResponse := client.ClientReadAssertionsResponse{
		Assertions: &[]openfga.Assertion{
			{
				TupleKey:    openfga.AssertionTupleKey{User: "user:user-1", Relation: "viewer", Object: "github-repo:demo"},
				Expectation: true,
			},
			{
				TupleKey:    openfga.AssertionTupleKey{User: "user:user-2", Relation: "viewer", Object: "github-repo:demo"},
				Expectation: true,
			},
			{
				TupleKey:    openfga.AssertionTupleKey{User: "user:user-2", Relation: "admin", Object: "github-repo:demo"},
				Expectation: true,
			},
			{
				TupleKey:    openfga.AssertionTupleKey{User: "user:user-1", Relation: "admin", Object: "github-repo:demo"},
				Expectation: false,
			},
		},
	}

	readAssertionsOptions := client.ClientReadAssertionsOptions{
		AuthorizationModelId: openfga.PtrString("01GXSA8YR785C4FYS3C0RTG7B1"),
	}

	mockAssertionsRequest := mock_client.NewMockSdkClientReadAssertionsRequestInterface(mockCtrl)
	mockAssertionsRequest.EXPECT().Options(readAssertionsOptions).Return(mockAssertionsRequest)
	mockAssertionsRequest.EXPECT().Execute().Return(&assertionsResponse, nil)
	mockFgaClient.EXPECT().ReadAssertions(context.Background()).Return(mockAssertionsRequest)

	// Execute
	output, err := buildStoreData(clientConfig, mockFgaClient, 50)
	// Expect
	if err != nil {
		t.Error(err)
	}

	expectedResponse := storetest.StoreData{
		Name: "Test store",
		Model: `model
  schema 1.1

type user

type github-repo
  relations
    define admin: [user]
    define viewer: [user]

`,
		Tuples: []openfga.TupleKey{
			{
				User:     "user:user-1",
				Relation: "viewer",
				Object:   "github-repo:demo",
			},
			{
				User:     "user:user-2",
				Relation: "viewer",
				Object:   "github-repo:demo",
			},
			{
				User:     "user:user-2",
				Relation: "admin",
				Object:   "github-repo:demo",
			},
		},
		Tests: []storetest.ModelTest{
			{
				Name: "Tests",
				Check: []storetest.ModelTestCheck{
					{
						User:   "user:user-1",
						Object: "github-repo:demo",
						Assertions: map[string]bool{
							"admin":  false,
							"viewer": true,
						},
					},
					{
						User:   "user:user-2",
						Object: "github-repo:demo",
						Assertions: map[string]bool{
							"admin":  true,
							"viewer": true,
						},
					},
				},
			},
		},
	}

	if output.Name != expectedResponse.Name {
		t.Errorf("Expected name %s got %s", expectedResponse.Name, output.Name)
	}

	if strings.TrimSpace(output.Model) != strings.TrimSpace(expectedResponse.Model) {
		t.Errorf("Expected model %s\n\ngot\n\n%s", expectedResponse.Model, output.Model)
	}

	if !reflect.DeepEqual(output.Tuples, expectedResponse.Tuples) {
		t.Errorf("Expected tuples %v\n\ngot\n\n%v", expectedResponse.Tuples, output.Tuples)
	}

	if len(output.Tests) != 1 {
		t.Errorf("Expected 1 output test, got %d", len(output.Tests))
	}

	for _, tst := range expectedResponse.Tests {
		for _, expectedCheck := range tst.Check {
			found := false

			for _, outputCheck := range output.Tests[0].Check {
				if reflect.DeepEqual(expectedCheck, outputCheck) {
					found = true

					break
				}
			}

			if !found {
				t.Errorf("Expected check %v not found in output", expectedCheck)
			}
		}
	}
}
