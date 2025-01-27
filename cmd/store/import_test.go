package store

import (
	"context"
	"github.com/openfga/cli/internal/fga"
	mockclient "github.com/openfga/cli/internal/mocks"
	"github.com/openfga/cli/internal/storetest"
	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestImportStore(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)
	clientConfig := fga.ClientConfig{}

	expectedAssertions := []client.ClientAssertion{{
		User:        "user:anne",
		Relation:    "reader",
		Object:      "document:doc1",
		Expectation: true,
	}}

	modelID := "model-1"
	storeID := "store-1"
	sampleTime := time.Now()
	expectedOptions := client.ClientWriteAssertionsOptions{
		AuthorizationModelId: &modelID,
		StoreId:              &storeID,
	}

	defaultStore := storetest.StoreData{
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

	importStoreTests := []struct {
		name                string
		mockWriteAssertions bool
		mockGetStore        bool
		mockCreateStore     bool
		mockWriteModel      bool
		testStore           storetest.StoreData
		storeId             string
	}{
		{
			name:                "import store with assertions",
			mockWriteAssertions: true,
			mockGetStore:        false,
			mockWriteModel:      true,
			mockCreateStore:     true,
			testStore:           defaultStore,
			storeId:             "",
		},
		{
			name:                "create new store without assertions",
			mockWriteAssertions: false,
			mockCreateStore:     true,
			mockGetStore:        false,
			mockWriteModel:      false,
			testStore: storetest.StoreData{
				Name: "test-store",
			},
			storeId: "",
		},
		{
			name:                "create new store without check assertions",
			mockCreateStore:     true,
			mockWriteModel:      true,
			mockWriteAssertions: false,
			testStore: storetest.StoreData{
				Name: "test-store",
				Model: `type user
					type document
						relations
							define reader: [user]`,
				Tests: []storetest.ModelTest{
					{
						Name: "Test",
						ListObjects: []storetest.ModelTestListObjects{
							{
								User: "user:anne",
								Type: "organization",
								Assertions: map[string][]string{
									"member": {"organization:acme"},
								},
							},
						},
					},
				},
			},
			storeId: "",
		},
		{
			name:                "do not write assertions if imported store does not have a model",
			mockCreateStore:     true,
			mockWriteAssertions: false,
			testStore: storetest.StoreData{
				Name: "test-store",
				Tests: []storetest.ModelTest{
					{
						Name: "Test",
						ListObjects: []storetest.ModelTestListObjects{
							{
								User: "user:anne",
								Type: "organization",
								Assertions: map[string][]string{
									"member": {"organization:acme"},
								},
							},
						},
					},
				},
			},
			storeId: "",
		},
		{
			name:                "update store with assertions",
			mockWriteAssertions: true,
			mockGetStore:        true,
			mockWriteModel:      true,
			testStore: storetest.StoreData{
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
			},
			storeId: storeID,
		},
		{
			name:                "update store without assertions",
			mockWriteAssertions: false,
			mockGetStore:        true,
			mockWriteModel:      true,
			testStore: storetest.StoreData{
				Name: "test-store",
				Model: `type user
					type document
						relations
							define reader: [user]`,
			},
			storeId: storeID,
		},
	}

	for _, test := range importStoreTests {

		t.Run(test.name, func(t *testing.T) {

			if test.mockWriteAssertions {
				mockWriteAssertions := mockclient.NewMockSdkClientWriteAssertionsRequestInterface(mockCtrl)
				mockFgaClient.EXPECT().WriteAssertions(context.Background()).Return(mockWriteAssertions)
				mockWriteAssertions.EXPECT().Body(expectedAssertions).Return(mockWriteAssertions)
				mockWriteAssertions.EXPECT().Options(expectedOptions).Return(mockWriteAssertions)
				mockWriteAssertions.EXPECT().Execute().Return(nil, nil)
			} else {
				mockFgaClient.EXPECT().WriteAssertions(context.Background()).Times(0)
			}

			if test.mockWriteModel {
				mockWriteModel := mockclient.NewMockSdkClientWriteAuthorizationModelRequestInterface(mockCtrl)
				mockFgaClient.EXPECT().WriteAuthorizationModel(context.Background()).Return(mockWriteModel)
				mockWriteModel.EXPECT().Body(gomock.Any()).Return(mockWriteModel)
				mockWriteModel.EXPECT().Execute().Return(&client.ClientWriteAuthorizationModelResponse{AuthorizationModelId: modelID}, nil)
			}

			if test.mockCreateStore {
				mockCreateStore := mockclient.NewMockSdkClientCreateStoreRequestInterface(mockCtrl)
				mockFgaClient.EXPECT().CreateStore(context.Background()).Return(mockCreateStore)
				mockCreateStore.EXPECT().Body(gomock.Any()).Return(mockCreateStore)
				mockCreateStore.EXPECT().Execute().Return(&client.ClientCreateStoreResponse{Id: storeID}, nil)
				mockFgaClient.EXPECT().SetStoreId(storeID)
			}

			if test.mockGetStore {
				mockGetStore := mockclient.NewMockSdkClientGetStoreRequestInterface(mockCtrl)
				mockFgaClient.EXPECT().GetStore(context.Background()).Return(mockGetStore)
				mockGetStore.EXPECT().Execute().Return(&client.ClientGetStoreResponse{Id: storeID, Name: "test-store", CreatedAt: sampleTime, UpdatedAt: sampleTime}, nil)
			}

			var err error
			if storeID != "" {
				_, err = importStore(&clientConfig, mockFgaClient, &test.testStore, "", test.storeId, 1, 1, "")
			} else {
				_, err = importStore(&clientConfig, mockFgaClient, &test.testStore, "", "", 1, 1, "")
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

		})
	}
}
