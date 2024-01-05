package model

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/fga"
	mock_client "github.com/openfga/cli/internal/mocks"
)

var errMockGet = errors.New("mock error")

// Test the case where get model is called without auth model ID is not specified.
func TestGetModelNoAuthModelID(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadLatestAuthorizationModelRequestInterface(mockCtrl)

	var expectedResponse client.ClientReadAuthorizationModelResponse

	modelJSON := `{"authorization_model":{"id":"01GXSA8YR785C4FYS3C0RTG7B1","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}]}}` //nolint:all
	if err := json.Unmarshal([]byte(modelJSON), &expectedResponse); err != nil {
		t.Fatalf("%v", err)
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockRequest := mock_client.NewMockSdkClientReadLatestAuthorizationModelRequestInterface(mockCtrl)
	options := client.ClientReadLatestAuthorizationModelOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)
	mockFgaClient.EXPECT().ReadLatestAuthorizationModel(context.Background()).Return(mockRequest)

	var clientConfig fga.ClientConfig

	output, err := getModel(clientConfig, mockFgaClient)
	if err != nil {
		t.Fatalf("%v", err)
	} else if *output != expectedResponse {
		t.Fatalf("Expect output %v actual %v", modelJSON, *output)
	}
}

// Test the case where get model is called without auth model ID is specified.
func TestGetModelAuthModelID(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadAuthorizationModelRequestInterface(mockCtrl)

	var expectedResponse client.ClientReadAuthorizationModelResponse

	modelJSON := `{"authorization_model":{"id":"01GXSA8YR785C4FYS3C0RTG7B1","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}]}}` //nolint:all
	if err := json.Unmarshal([]byte(modelJSON), &expectedResponse); err != nil {
		t.Fatalf("%v", err)
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockRequest := mock_client.NewMockSdkClientReadAuthorizationModelRequestInterface(mockCtrl)
	options := client.ClientReadAuthorizationModelOptions{
		AuthorizationModelId: openfga.PtrString("01GXSA8YR785C4FYS3C0RTG7B1"),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)
	mockFgaClient.EXPECT().ReadAuthorizationModel(context.Background()).Return(mockRequest)

	clientConfig := fga.ClientConfig{
		AuthorizationModelID: "01GXSA8YR785C4FYS3C0RTG7B1",
	}

	output, err := getModel(clientConfig, mockFgaClient)
	if err != nil {
		t.Fatalf("%v", err)
	} else if *output != expectedResponse {
		t.Fatalf("Expect output %v actual %v", modelJSON, *output)
	}
}

// Test the case where get model is called, but it returns error.
func TestGetModelNoAuthModelIDError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadLatestAuthorizationModelRequestInterface(mockCtrl)

	var expectedResponse client.ClientReadAuthorizationModelResponse

	mockExecute.EXPECT().Execute().Return(&expectedResponse, errMockGet)

	mockRequest := mock_client.NewMockSdkClientReadLatestAuthorizationModelRequestInterface(mockCtrl)
	options := client.ClientReadLatestAuthorizationModelOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)
	mockFgaClient.EXPECT().ReadLatestAuthorizationModel(context.Background()).Return(mockRequest)

	var clientConfig fga.ClientConfig

	_, err := getModel(clientConfig, mockFgaClient)
	if err == nil {
		t.Fatalf("Expect error but there is none")
	}
}
