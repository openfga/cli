package query

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/openfga/go-sdk/client"

	mock_client "github.com/openfga/cli/internal/mocks"
)

var errMockListObjects = errors.New("mock error")

func TestListObjectsWithError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientListObjectsRequestInterface(mockCtrl)

	var expectedResponse client.ClientListObjectsResponse

	mockExecute.EXPECT().Execute().Return(&expectedResponse, errMockListObjects)

	mockRequest := mock_client.NewMockSdkClientListObjectsRequestInterface(mockCtrl)
	options := client.ClientListObjectsOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientListObjectsRequestInterface(mockCtrl)

	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}
	body := client.ClientListObjectsRequest{
		User:             "user:foo",
		Relation:         "writer",
		Type:             "doc",
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().ListObjects(context.Background()).Return(mockBody)

	_, err := listObjects(mockFgaClient, "user:foo", "writer", "doc", contextualTuples, queryContext)
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestListObjectsWithNoError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientListObjectsRequestInterface(mockCtrl)

	expectedResponse := client.ClientListObjectsResponse{
		Objects: []string{"doc:doc1", "doc:doc2"},
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockRequest := mock_client.NewMockSdkClientListObjectsRequestInterface(mockCtrl)
	options := client.ClientListObjectsOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientListObjectsRequestInterface(mockCtrl)

	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}
	body := client.ClientListObjectsRequest{
		User:             "user:foo",
		Relation:         "writer",
		Type:             "doc",
		ContextualTuples: contextualTuples,
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().ListObjects(context.Background()).Return(mockBody)

	output, err := listObjects(mockFgaClient, "user:foo", "writer", "doc", contextualTuples, nil)
	if err != nil {
		t.Error(err)
	}

	if output != &expectedResponse {
		t.Errorf("Expect %v but actual %v", expectedResponse, *output)
	}
}
