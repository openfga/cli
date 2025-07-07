package query

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"

	mock_client "github.com/openfga/cli/internal/mocks"
)

var errMockExpand = errors.New("mock error")

func TestExpandWithError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientExpandRequestInterface(mockCtrl)

	var expectedResponse client.ClientExpandResponse

	mockExecute.EXPECT().Execute().Return(&expectedResponse, errMockExpand)

	mockRequest := mock_client.NewMockSdkClientExpandRequestInterface(mockCtrl)
	options := client.ClientExpandOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientExpandRequestInterface(mockCtrl)

	body := client.ClientExpandRequest{
		Relation: "writer",
		Object:   "doc:doc1",
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().Expand(t.Context()).Return(mockBody)

	_, err := expand(t.Context(), mockFgaClient, "writer", "doc:doc1", openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr())
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestExpandWithNoError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientExpandRequestInterface(mockCtrl)

	expandResponseTxt := `{"tree":{"root":{"name":"document:roadmap#viewer","union":{"nodes":[{"name": "document:roadmap#viewer","leaf":{"users":{"users":["user:81684243-9356-4421-8fbf-a4f8d36aa31b"]}}}]}}}}`

	expectedResponse := client.ClientExpandResponse{}
	if err := json.Unmarshal([]byte(expandResponseTxt), &expectedResponse); err != nil {
		t.Fatalf("%v", err)
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockRequest := mock_client.NewMockSdkClientExpandRequestInterface(mockCtrl)
	options := client.ClientExpandOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientExpandRequestInterface(mockCtrl)

	body := client.ClientExpandRequest{
		Relation: "writer",
		Object:   "doc:doc1",
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().Expand(t.Context()).Return(mockBody)

	output, err := expand(t.Context(), mockFgaClient, "writer", "doc:doc1", openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr())
	if err != nil {
		t.Error(err)
	}

	if !(reflect.DeepEqual(*output, expectedResponse)) {
		t.Errorf("Expect output response %v actual response %v", expandResponseTxt, *output)
	}
}

func TestExpandWithConsistency(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientExpandRequestInterface(mockCtrl)

	expandResponseTxt := `{"tree":{"root":{"name":"document:roadmap#viewer","union":{"nodes":[{"name": "document:roadmap#viewer","leaf":{"users":{"users":["user:81684243-9356-4421-8fbf-a4f8d36aa31b"]}}}]}}}}`

	expectedResponse := client.ClientExpandResponse{}
	if err := json.Unmarshal([]byte(expandResponseTxt), &expectedResponse); err != nil {
		t.Fatalf("%v", err)
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockRequest := mock_client.NewMockSdkClientExpandRequestInterface(mockCtrl)
	options := client.ClientExpandOptions{
		Consistency: openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr(),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientExpandRequestInterface(mockCtrl)

	body := client.ClientExpandRequest{
		Relation: "writer",
		Object:   "doc:doc1",
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().Expand(t.Context()).Return(mockBody)

	output, err := expand(t.Context(), mockFgaClient, "writer", "doc:doc1", openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr())
	if err != nil {
		t.Error(err)
	}

	if !(reflect.DeepEqual(*output, expectedResponse)) {
		t.Errorf("Expect output response %v actual response %v", expandResponseTxt, *output)
	}
}
