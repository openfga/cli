package query

import (
	"context"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"

	mock_client "github.com/openfga/cli/internal/mocks"
)

func TestListUsersSimpleType(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientListUsersRequestInterface(mockCtrl)

	expectedResponse := client.ClientListUsersResponse{
		Users: []openfga.User{
			{
				Object: &openfga.FgaObject{
					Type: "user",
					Id:   "anne",
				},
			},
		},
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockRequest := mock_client.NewMockSdkClientListUsersRequestInterface(mockCtrl)
	options := client.ClientListUsersOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientListUsersRequestInterface(mockCtrl)

	contextualTuples := []client.ClientContextualTupleKey{}
	userFilters := []openfga.UserTypeFilter{
		{
			Type: "user",
		},
	}

	body := client.ClientListUsersRequest{
		Object: openfga.FgaObject{
			Type: "doc",
			Id:   "doc1",
		},
		Relation:         "admin",
		UserFilters:      userFilters,
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().ListUsers(context.Background()).Return(mockBody)

	output, err := listUsers(
		mockFgaClient,
		"doc:doc1",
		"admin",
		"user",
		contextualTuples,
		queryContext,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	if output != &expectedResponse {
		t.Errorf("Expect %v but actual %v", expectedResponse, *output)
	}
}

func TestListUsersSimpleTypeAndRelation(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientListUsersRequestInterface(mockCtrl)

	expectedResponse := client.ClientListUsersResponse{
		Users: []openfga.User{
			{
				Object: &openfga.FgaObject{
					Type: "user",
					Id:   "anne",
				},
			},
		},
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockRequest := mock_client.NewMockSdkClientListUsersRequestInterface(mockCtrl)
	options := client.ClientListUsersOptions{}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientListUsersRequestInterface(mockCtrl)

	contextualTuples := []client.ClientContextualTupleKey{}
	userFilters := []openfga.UserTypeFilter{
		{
			Type:     "group",
			Relation: openfga.PtrString("member"),
		},
	}

	body := client.ClientListUsersRequest{
		Object: openfga.FgaObject{
			Type: "doc",
			Id:   "doc1",
		},
		Relation:         "admin",
		UserFilters:      userFilters,
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().ListUsers(context.Background()).Return(mockBody)

	output, err := listUsers(
		mockFgaClient,
		"doc:doc1",
		"admin",
		"group#member",
		contextualTuples,
		queryContext,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	if output != &expectedResponse {
		t.Errorf("Expect %v but actual %v", expectedResponse, *output)
	}
}

func TestListUsersWithConsistency(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientListUsersRequestInterface(mockCtrl)

	expectedResponse := client.ClientListUsersResponse{
		Users: []openfga.User{
			{
				Object: &openfga.FgaObject{
					Type: "user",
					Id:   "anne",
				},
			},
		},
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	mockRequest := mock_client.NewMockSdkClientListUsersRequestInterface(mockCtrl)
	options := client.ClientListUsersOptions{
		Consistency: openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr(),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientListUsersRequestInterface(mockCtrl)

	contextualTuples := []client.ClientContextualTupleKey{}
	userFilters := []openfga.UserTypeFilter{
		{
			Type: "user",
		},
	}

	body := client.ClientListUsersRequest{
		Object: openfga.FgaObject{
			Type: "doc",
			Id:   "doc1",
		},
		Relation:         "admin",
		UserFilters:      userFilters,
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().ListUsers(context.Background()).Return(mockBody)

	output, err := listUsers(
		mockFgaClient,
		"doc:doc1",
		"admin",
		"user",
		contextualTuples,
		queryContext,
		openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	if output != &expectedResponse {
		t.Errorf("Expect %v but actual %v", expectedResponse, *output)
	}
}
