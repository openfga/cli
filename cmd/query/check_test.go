package query

import (
	"errors"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"

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

	mockFgaClient.EXPECT().Check(t.Context()).Return(mockBody)

	_, err := check(
		t.Context(),
		mockFgaClient,
		"user:foo",
		"writer",
		"doc:doc1",
		contextualTuples,
		queryContext,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
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

	mockFgaClient.EXPECT().Check(t.Context()).Return(mockBody)

	output, err := check(
		t.Context(),
		mockFgaClient,
		"user:foo",
		"writer",
		"doc:doc1",
		contextualTuples,
		queryContext,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	if *output != expectedResponse {
		t.Errorf("Expected output %v actual %v", expectedResponse, *output)
	}
}

func TestCheckWithConsistency(t *testing.T) {
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
	options := client.ClientCheckOptions{
		Consistency: openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr(),
	}
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

	mockFgaClient.EXPECT().Check(t.Context()).Return(mockBody)

	output, err := check(
		t.Context(),
		mockFgaClient,
		"user:foo",
		"writer",
		"doc:doc1",
		contextualTuples,
		queryContext,
		openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	if *output != expectedResponse {
		t.Errorf("Expected output %v actual %v", expectedResponse, *output)
	}
}
