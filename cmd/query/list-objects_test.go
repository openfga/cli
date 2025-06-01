package query

import (
	"context"
	"errors"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"

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

	_, err := listObjects(
		t.Context(),
		mockFgaClient,
		"user:foo",
		"writer",
		"doc",
		contextualTuples,
		queryContext,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
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

	output, err := listObjects(
		t.Context(),
		mockFgaClient,
		"user:foo",
		"writer",
		"doc",
		contextualTuples,
		nil,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	if output != &expectedResponse {
		t.Errorf("Expect %v but actual %v", expectedResponse, *output)
	}
}

func TestListObjectsWithConsistency(t *testing.T) {
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
	options := client.ClientListObjectsOptions{
		Consistency: openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr(),
	}
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

	output, err := listObjects(
		t.Context(),
		mockFgaClient,
		"user:foo",
		"writer",
		"doc",
		contextualTuples,
		nil,
		openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	if output != &expectedResponse {
		t.Errorf("Expect %v but actual %v", expectedResponse, *output)
	}
}
