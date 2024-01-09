package query

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	mock_client "github.com/openfga/cli/internal/mocks"
)

var errMockCheck = errors.New("mock error")

func TestCheckWithError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientCheckRequestInterface(mockCtrl)

	var expectedResponse client.ClientCheckResponse

	mockExecute.EXPECT().Execute().Return(&expectedResponse, errMockCheck)

	mockRequest := mock_client.NewMockSdkClientCheckRequestInterface(mockCtrl)
	options := client.ClientCheckOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientCheckRequestInterface(mockCtrl)
	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}

	body := client.ClientCheckRequest{
		User:             "user:foo",
		Relation:         "writer",
		Object:           "doc:doc1",
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().Check(context.Background()).Return(mockBody)

	_, err := check(mockFgaClient, "user:foo", "writer", "doc:doc1", contextualTuples, queryContext)
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestCheckWithNoError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientCheckRequestInterface(mockCtrl)

	expectedResponse := client.ClientCheckResponse{
		CheckResponse: openfga.CheckResponse{
			Allowed: openfga.PtrBool(true),
		},
		HttpResponse: nil,
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockRequest := mock_client.NewMockSdkClientCheckRequestInterface(mockCtrl)
	options := client.ClientCheckOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientCheckRequestInterface(mockCtrl)

	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}
	body := client.ClientCheckRequest{
		User:             "user:foo",
		Relation:         "writer",
		Object:           "doc:doc1",
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().Check(context.Background()).Return(mockBody)

	output, err := check(mockFgaClient, "user:foo", "writer", "doc:doc1", contextualTuples, queryContext)
	if err != nil {
		t.Error(err)
	}

	if *output != expectedResponse {
		t.Errorf("Expected output %v actual %v", expectedResponse, *output)
	}
}
