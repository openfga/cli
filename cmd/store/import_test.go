package store

import (
	"testing"
	"time"

	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"

	"github.com/openfga/cli/internal/fga"
	mockclient "github.com/openfga/cli/internal/mocks"
	"github.com/openfga/cli/internal/storetest"
)

func TestImportStore(t *testing.T) {
	t.Parallel()

	expectedAssertions := []client.ClientAssertion{{
		User:        "user:anne",
		Relation:    "reader",
		Object:      "document:doc1",
		Expectation: true,
	}}

	multiUserAssertions := []client.ClientAssertion{
		{
			User:        "user:anne",
			Relation:    "reader",
			Object:      "document:doc1",
			Expectation: true,
		},
		{
			User:        "user:peter",
			Relation:    "reader",
			Object:      "document:doc1",
			Expectation: true,
		},
	}
	modelID, storeID := "model-1", "store-1"
	expectedOptions := client.ClientWriteAssertionsOptions{AuthorizationModelId: &modelID, StoreId: &storeID}

	importStoreTests := []struct {
		name                string
		mockWriteAssertions bool
		mockCreateStore     bool
		mockWriteModel      bool
		testStore           storetest.StoreData
	}{
		{
			name:                "import store with assertions",
			mockWriteAssertions: true,
			mockWriteModel:      true,
			mockCreateStore:     true,
			testStore: storetest.StoreData{
				Model: `type user
                                        type document
                                                relations
                                                        define reader: [user]`,
				Tests: []storetest.ModelTest{
					{
						Name: "Test",
						Check: []storetest.ModelTestCheck{
							{
								User:       "user:anne",
								Object:     "document:doc1",
								Assertions: map[string]bool{"reader": true},
							},
						},
					},
				},
			},
		},
		{
			name:                "import store with multi user assertions",
			mockWriteAssertions: true,
			mockWriteModel:      true,
			mockCreateStore:     true,
			testStore: storetest.StoreData{
				Model: `type user
                                        type document
                                                relations
                                                        define reader: [user]`,
				Tests: []storetest.ModelTest{
					{
						Name: "Test",
						Check: []storetest.ModelTestCheck{
							{
								Users:      []string{"user:anne", "user:peter"},
								Object:     "document:doc1",
								Assertions: map[string]bool{"reader": true},
							},
						},
					},
				},
			},
		},
		{
			name:                "create new store without assertions",
			mockWriteAssertions: false,
			mockCreateStore:     true,
			mockWriteModel:      false,
			testStore:           storetest.StoreData{Name: "test-store"},
		},
		{
			name:                "create new store without check assertions",
			mockCreateStore:     true,
			mockWriteModel:      true,
			mockWriteAssertions: false,
			testStore: storetest.StoreData{
				Model: `type user
					type document
						relations
							define reader: [user]`,
				Tests: []storetest.ModelTest{
					{
						Name: "Test",
						ListObjects: []storetest.ModelTestListObjects{
							{
								User:       "user:anne",
								Type:       "organization",
								Assertions: map[string][]string{"member": {"organization:acme"}},
							},
						},
					},
				},
			},
		},
		{
			name:                "do not write assertions if imported store does not have a model",
			mockCreateStore:     true,
			mockWriteAssertions: false,
			testStore: storetest.StoreData{
				Tests: []storetest.ModelTest{
					{Name: "Test"},
				},
			},
		},
	}

	for _, test := range importStoreTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

			defer mockCtrl.Finish()

			if test.mockWriteAssertions {
				expected := expectedAssertions
				if test.name == "import store with multi user assertions" {
					expected = multiUserAssertions
				}
				setupWriteAssertionsMock(mockCtrl, mockFgaClient, expected, expectedOptions)
			} else {
				mockFgaClient.EXPECT().WriteAssertions(gomock.Any()).Times(0)
			}

			if test.mockWriteModel {
				setupWriteModelMock(mockCtrl, mockFgaClient, modelID)
			}

			if test.mockCreateStore {
				setupCreateStoreMock(mockCtrl, mockFgaClient, storeID)
			}

			_, err := importStore(t.Context(), &fga.ClientConfig{}, mockFgaClient, &test.testStore, "", "", 10, 1, "")
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestUpdateStore(t *testing.T) {
	t.Parallel()

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

	importStoreTests := []struct {
		name                string
		mockWriteAssertions bool
		mockGetStore        bool
		mockWriteModel      bool
		testStore           storetest.StoreData
	}{
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
		},
	}

	for _, test := range importStoreTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			clientConfig := fga.ClientConfig{}

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

			defer mockCtrl.Finish()

			if test.mockWriteAssertions {
				setupWriteAssertionsMock(mockCtrl, mockFgaClient, expectedAssertions, expectedOptions)
			} else {
				mockFgaClient.EXPECT().WriteAssertions(gomock.Any()).Times(0)
			}

			if test.mockWriteModel {
				setupWriteModelMock(mockCtrl, mockFgaClient, modelID)
			}

			if test.mockGetStore {
				setupGetStoreMock(mockCtrl, mockFgaClient, storeID, sampleTime)
			}

			_, err := importStore(t.Context(), &clientConfig, mockFgaClient, &test.testStore, "", storeID, 10, 1, "")
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func setupGetStoreMock(
	mockCtrl *gomock.Controller,
	mockFgaClient *mockclient.MockSdkClient,
	storeID string,
	sampleTime time.Time,
) {
	mockGetStore := mockclient.NewMockSdkClientGetStoreRequestInterface(mockCtrl)
	mockFgaClient.EXPECT().GetStore(gomock.Any()).Return(mockGetStore)
	mockGetStore.EXPECT().Execute().Return(
		&client.ClientGetStoreResponse{Id: storeID, Name: "test-store", CreatedAt: sampleTime, UpdatedAt: sampleTime},
		nil,
	)
}

func setupCreateStoreMock(mockCtrl *gomock.Controller, mockFgaClient *mockclient.MockSdkClient, storeID string) {
	mockCreateStore := mockclient.NewMockSdkClientCreateStoreRequestInterface(mockCtrl)
	mockFgaClient.EXPECT().CreateStore(gomock.Any()).Return(mockCreateStore)
	mockCreateStore.EXPECT().Body(gomock.Any()).Return(mockCreateStore)
	mockCreateStore.EXPECT().Execute().Return(&client.ClientCreateStoreResponse{Id: storeID}, nil)
	mockFgaClient.EXPECT().SetStoreId(storeID)
}

func setupWriteModelMock(mockCtrl *gomock.Controller, mockFgaClient *mockclient.MockSdkClient, modelID string) {
	mockWriteModel := mockclient.NewMockSdkClientWriteAuthorizationModelRequestInterface(mockCtrl)
	mockFgaClient.EXPECT().WriteAuthorizationModel(gomock.Any()).Return(mockWriteModel)
	mockWriteModel.EXPECT().Body(gomock.Any()).Return(mockWriteModel)
	mockWriteModel.EXPECT().Execute().Return(
		&client.ClientWriteAuthorizationModelResponse{AuthorizationModelId: modelID},
		nil,
	)
}

func setupWriteAssertionsMock(
	mockCtrl *gomock.Controller,
	mockFgaClient *mockclient.MockSdkClient,
	expectedAssertions []client.ClientAssertion,
	expectedOptions client.ClientWriteAssertionsOptions,
) {
	mockWriteAssertions := mockclient.NewMockSdkClientWriteAssertionsRequestInterface(mockCtrl)
	mockFgaClient.EXPECT().WriteAssertions(gomock.Any()).Return(mockWriteAssertions)
	mockWriteAssertions.EXPECT().Body(expectedAssertions).Return(mockWriteAssertions)
	mockWriteAssertions.EXPECT().Options(expectedOptions).Return(mockWriteAssertions)
	mockWriteAssertions.EXPECT().Execute().Return(nil, nil)
}
