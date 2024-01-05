package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/openfga/go-sdk/client"

	mock_client "github.com/openfga/cli/internal/mocks"
)

var errMockCreate = errors.New("mock error")

func TestCreateError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientCreateStoreRequestInterface(mockCtrl)

	var expectedResponse client.ClientCreateStoreResponse

	mockExecute.EXPECT().Execute().Return(&expectedResponse, errMockCreate)

	mockBody := mock_client.NewMockSdkClientCreateStoreRequestInterface(mockCtrl)

	body := client.ClientCreateStoreRequest{
		Name: "foo",
	}
	mockBody.EXPECT().Body(body).Return(mockExecute)

	mockFgaClient.EXPECT().CreateStore(context.Background()).Return(mockBody)

	_, err := create(mockFgaClient, "foo")
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestCreateSuccess(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientCreateStoreRequestInterface(mockCtrl)
	expectedTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	expectedResponse := client.ClientCreateStoreResponse{
		Id:        "12345",
		Name:      "foo",
		CreatedAt: expectedTime,
		UpdatedAt: expectedTime,
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockBody := mock_client.NewMockSdkClientCreateStoreRequestInterface(mockCtrl)

	body := client.ClientCreateStoreRequest{
		Name: "foo",
	}
	mockBody.EXPECT().Body(body).Return(mockExecute)

	mockFgaClient.EXPECT().CreateStore(context.Background()).Return(mockBody)

	output, err := create(mockFgaClient, "foo")
	if err != nil {
		t.Error(err)
	}

	if *output != expectedResponse {
		t.Errorf("Expected output %v actual %v", expectedResponse, *output)
	}
}
