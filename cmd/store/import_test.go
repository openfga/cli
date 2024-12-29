package store

import (
	"context"
	"github.com/openfga/cli/internal/fga"
	mockclient "github.com/openfga/cli/internal/mocks"
	"github.com/openfga/cli/internal/storetest"
	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestImportStore(t *testing.T) {
	t.Parallel()

	t.Run("imports assertions into store", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)
		clientConfig := fga.ClientConfig{}

		mockWriteAssertions := mockclient.NewMockSdkClientWriteAssertionsRequestInterface(mockCtrl)
		mockCreateStore := mockclient.NewMockSdkClientCreateStoreRequestInterface(mockCtrl)
		mockWriteModel := mockclient.NewMockSdkClientWriteAuthorizationModelRequestInterface(mockCtrl)

		expectedAssertions := []client.ClientAssertion{{
			User:        "user:anne",
			Relation:    "reader",
			Object:      "document:doc1",
			Expectation: true,
		}}

		modelID := "model-1"
		storeID := "store-1"
		expectedOptions := client.ClientWriteAssertionsOptions{
			AuthorizationModelId: &modelID,
			StoreId:              &storeID,
		}

		mockFgaClient.EXPECT().CreateStore(context.Background()).Return(mockCreateStore)
		mockCreateStore.EXPECT().Body(gomock.Any()).Return(mockCreateStore)
		mockCreateStore.EXPECT().Execute().Return(&client.ClientCreateStoreResponse{Id: "store-1"}, nil)

		mockFgaClient.EXPECT().SetStoreId("store-1")

		mockFgaClient.EXPECT().WriteAuthorizationModel(context.Background()).Return(mockWriteModel)
		mockWriteModel.EXPECT().Body(gomock.Any()).Return(mockWriteModel)
		mockWriteModel.EXPECT().Execute().Return(&client.ClientWriteAuthorizationModelResponse{AuthorizationModelId: "model-1"}, nil)

		mockFgaClient.EXPECT().WriteAssertions(context.Background()).Return(mockWriteAssertions)
		mockWriteAssertions.EXPECT().Body(expectedAssertions).Return(mockWriteAssertions)
		mockWriteAssertions.EXPECT().Options(expectedOptions).Return(mockWriteAssertions)
		mockWriteAssertions.EXPECT().Execute().Return(nil, nil)

		testStore := &storetest.StoreData{
			Name: "test-store",
			Model: `type user
					type document
						relations
							define reader: [user]`,
			Tests: []storetest.ModelTest{
				{
					Name: "Test",
					Check: []storetest.ModelTestCheck{
						{
							User:   "user:anne",
							Object: "document:doc1",
							Assertions: map[string]bool{
								"reader": true,
							},
						},
					},
				},
			},
		}
		_, err := importStore(&clientConfig, mockFgaClient, testStore, "", "", 1, 1, "")

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
