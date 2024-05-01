package store

import (
	"fmt"
	"github.com/openfga/cli/cmd/tuple"
	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/fga"
	"github.com/openfga/cli/internal/storetest"
	"github.com/openfga/go-sdk/client"
	"path"
	"reflect"
	"testing"
	"time"
)

func TestImportStore(t *testing.T) {
	t.Run("Must create store and modelID when there's no store configured", func(t *testing.T) {
		// Arrange
		clientConfig := fga.ClientConfig{ApiUrl: "https://localhost:8080"}
		storeData := &storetest.StoreData{}
		authorizationModelID := "01HWJGBQQNNQATBQ661SH6585Y"
		storeID := "Test-001"

		ioAggregator := importStoreIODependencies{
			importTuples: func(fgaClient client.SdkClient, body client.ClientWriteRequest, maxTuplesPerWrite, maxParallelRequests int) (*tuple.ImportResponse, error) {
				return nil, nil
			},
			createStoreWithModel: func(clientConfig fga.ClientConfig, storeName, inputModel string, inputFormat authorizationmodel.ModelFormat) (*CreateStoreAndModelResponse, error) {
				return &CreateStoreAndModelResponse{
					Store: &client.ClientCreateStoreResponse{
						Id:        "01HWJGBQQHZZJHQEQZ6MCBC3B0",
						Name:      storeID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Model: &client.ClientWriteAuthorizationModelResponse{
						AuthorizationModelId: authorizationModelID,
					},
				}, nil
			},
		}

		// Act
		response, err := importStore(clientConfig, storeData, "", "", 2, 2, ioAggregator)

		// Assert
		if err != nil {
			t.Error(err)
		}
		if response.Model.AuthorizationModelId != authorizationModelID {
			t.Fatalf("expected: %s\nreturned: %s", authorizationModelID, response.Model.AuthorizationModelId)
		}
		if response.Store.Name != storeID {
			t.Fatalf("expected: %s\nreturned: %s", storeID, response.Store.Name)
		}
	})

	t.Run("Must return the Model ID and an empty Store when importing into an existing store", func(t *testing.T) {
		clientConfig := fga.ClientConfig{ApiUrl: "https://localhost:8080"}
		storeData := &storetest.StoreData{}
		authorizationModelID := "01HWJGBQQNNQATBQ661SH6585Y"

		ioAggregator := importStoreIODependencies{
			importTuples: func(fgaClient client.SdkClient, body client.ClientWriteRequest, maxTuplesPerWrite, maxParallelRequests int) (*tuple.ImportResponse, error) {
				return nil, nil
			},
			modelWrite: func(fgaClient client.SdkClient, inputModel authorizationmodel.AuthzModel) (*client.ClientWriteAuthorizationModelResponse, error) {
				return &client.ClientWriteAuthorizationModelResponse{
					AuthorizationModelId: authorizationModelID,
				}, nil
			},
			createStoreWithModel: func(clientConfig fga.ClientConfig, storeName, inputModel string, inputFormat authorizationmodel.ModelFormat) (*CreateStoreAndModelResponse, error) {
				return &CreateStoreAndModelResponse{
					Model: &client.ClientWriteAuthorizationModelResponse{
						AuthorizationModelId: authorizationModelID,
					},
				}, nil
			},
		}

		// Act
		response, err := importStore(clientConfig, storeData, "", "01HWJGBQQHZZJHQEQZ6MCBC3B0", 2, 2, ioAggregator)

		// Assert
		if err != nil {
			t.Error(err)
		}
		if response.Store != nil {
			t.Fatalf("Expected: null\nReturn: %v", response.Store)
		}
		if response.Model.AuthorizationModelId != authorizationModelID {
			t.Fatalf("expected: %s\nreturned: %s", authorizationModelID, response.Model.AuthorizationModelId)
		}
	})

	t.Run("Must returns the modelID, a null store object, and a list of failed/successfully imported tuples when tuples containing unregistered types are provided as input", func(t *testing.T) {
		clientConfig := fga.ClientConfig{ApiUrl: "https://localhost:8080"}
		fileName := "./.test-data/failed-store-import-001.fga.yaml"
		format, storeData, err := storetest.ReadFromFile(fileName, path.Dir(fileName))
		authorizationModelID := "01HWJGBQQNNQATBQ661SH6585Y"

		successfulTupleImport := storeData.Tuples[0]
		failedTupleImport := storeData.Tuples[1]
		failureReason := fmt.Sprintf("error message: Invalid tuple '%s#%s@%s'. Reason: type 'document' not found",
			failedTupleImport.Object,
			failedTupleImport.Relation,
			failedTupleImport.User,
		)
		ioAggregator := importStoreIODependencies{
			importTuples: func(fgaClient client.SdkClient, body client.ClientWriteRequest, maxTuplesPerWrite, maxParallelRequests int) (*tuple.ImportResponse, error) {
				return &tuple.ImportResponse{
					Successful: []client.ClientTupleKey{
						successfulTupleImport,
					},
					Failed: []tuple.FailedWriteResponse{
						{
							TupleKey: failedTupleImport,
							Reason:   failureReason,
						},
					},
				}, nil
			},
			modelWrite: func(fgaClient client.SdkClient, inputModel authorizationmodel.AuthzModel) (*client.ClientWriteAuthorizationModelResponse, error) {
				return &client.ClientWriteAuthorizationModelResponse{
					AuthorizationModelId: authorizationModelID,
				}, nil
			},
			createStoreWithModel: func(clientConfig fga.ClientConfig, storeName, inputModel string, inputFormat authorizationmodel.ModelFormat) (*CreateStoreAndModelResponse, error) {
				return &CreateStoreAndModelResponse{
					Model: &client.ClientWriteAuthorizationModelResponse{
						AuthorizationModelId: authorizationModelID,
					},
				}, nil
			},
		}

		// Act
		importResponse, err := importStore(clientConfig, storeData, format, "01HWJGBQQHZZJHQEQZ6MCBC3B0", 2, 2, ioAggregator)
		if err != nil {
			t.Error(err)
		}
		if len(importResponse.Tuple.Successful) != 1 {
			t.Fatalf("expected: %d\nreturned: %d", 1, len(importResponse.Tuple.Successful))
		}
		if len(importResponse.Tuple.Failed) != 1 {
			t.Fatalf("expected: %d\nreturned: %d", 1, len(importResponse.Tuple.Failed))
		}
		if !reflect.DeepEqual(importResponse.Tuple.Successful[0], successfulTupleImport) {
			t.Fatalf("expected: %v\nreturned: %v", successfulTupleImport, importResponse.Tuple.Successful[0])
		}
		if !reflect.DeepEqual(importResponse.Tuple.Failed[0].TupleKey, failedTupleImport) {
			t.Fatalf("expected: %v\nreturned: %v", failedTupleImport, importResponse.Tuple.Failed[0])
		}
		if failureReason != importResponse.Tuple.Failed[0].Reason {
			t.Fatalf("expected: %s\nreturned: %s", failureReason, importResponse.Tuple.Failed[0].Reason)
		}
	})
}
