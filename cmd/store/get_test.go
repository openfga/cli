package store

import (
	"errors"
	"testing"
	"time"

	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"

	"github.com/openfga/cli/internal/fga"
	mock_client "github.com/openfga/cli/internal/mocks"
)

var errMockGet = errors.New("mock error")

func TestGetError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientGetStoreRequestInterface(mockCtrl)

	var expectedResponse client.ClientGetStoreResponse

	mockExecute.EXPECT().Execute().Return(&expectedResponse, errMockGet)

	mockFgaClient.EXPECT().GetStore(t.Context()).Return(mockExecute)

	clientConfig := fga.ClientConfig{
		StoreID: "12345",
	}

	_, err := getStore(t.Context(), clientConfig, mockFgaClient)
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestGetSuccess(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientGetStoreRequestInterface(mockCtrl)
	expectedTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	expectedResponse := client.ClientGetStoreResponse{
		Id:        "12345",
		Name:      "foo",
		CreatedAt: expectedTime,
		UpdatedAt: expectedTime,
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockFgaClient.EXPECT().GetStore(t.Context()).Return(mockExecute)

	clientConfig := fga.ClientConfig{
		StoreID: "12345",
	}

	output, err := getStore(t.Context(), clientConfig, mockFgaClient)
	if err != nil {
		t.Error(err)
	}

	if *output != expectedResponse {
		t.Errorf("Expected output %v actual %v", expectedResponse, *output)
	}
}
