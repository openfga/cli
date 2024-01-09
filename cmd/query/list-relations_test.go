package query

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/fga"
	mock_client "github.com/openfga/cli/internal/mocks"
)

var errMockGet = errors.New("mock get model error")

var errMockListRelations = errors.New("mock error")

var queryContext, _ = cmdutils.ParseQueryContextInner(`{"x": 1}`)

func TestListRelationsLatestAuthModelError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadLatestAuthorizationModelRequestInterface(mockCtrl)

	var expectedResponse client.ClientReadAuthorizationModelResponse

	mockExecute.EXPECT().Execute().Return(&expectedResponse, errMockGet)
	mockFgaClient.EXPECT().ReadLatestAuthorizationModel(context.Background()).Return(mockExecute)

	var clientConfig fga.ClientConfig

	relations := []string{}
	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}
	_, err := listRelations(clientConfig, mockFgaClient, "user:foo", "doc:doc1", relations, contextualTuples, queryContext)

	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestListRelationsAuthModelSpecifiedError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadAuthorizationModelRequestInterface(mockCtrl)

	var expectedResponse client.ClientReadAuthorizationModelResponse

	mockExecute.EXPECT().Execute().Return(&expectedResponse, errMockGet)
	mockFgaClient.EXPECT().ReadAuthorizationModel(context.Background()).Return(mockExecute)

	clientConfig := fga.ClientConfig{
		AuthorizationModelID: "01GXSA8YR785C4FYS3C0RTG7B1",
	}

	relations := []string{}
	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}
	_, err := listRelations(clientConfig, mockFgaClient, "user:foo", "doc:doc1", relations, contextualTuples, queryContext)

	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestListRelationsLatestAuthModelListError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadLatestAuthorizationModelRequestInterface(mockCtrl)

	var expectedResponse client.ClientReadAuthorizationModelResponse

	modelJSON := `{"authorization_model":{"id":"01GXSA8YR785C4FYS3C0RTG7B1","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"doc"}]}}` //nolint:all
	if err := json.Unmarshal([]byte(modelJSON), &expectedResponse); err != nil {
		t.Fatalf("%v", err)
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	// after reading the latest auth model, expect to call list relations but failure

	mockListRelationsExecute := mock_client.NewMockSdkClientListRelationsRequestInterface(mockCtrl)

	var expectedListRelationsResponse client.ClientListRelationsResponse

	mockListRelationsExecute.EXPECT().Execute().Return(&expectedListRelationsResponse, errMockListRelations)

	mockListRelationsRequest := mock_client.NewMockSdkClientListRelationsRequestInterface(mockCtrl)
	listRelationsOptions := client.ClientListRelationsOptions{}
	mockListRelationsRequest.EXPECT().Options(listRelationsOptions).Return(mockListRelationsExecute)

	mockBody := mock_client.NewMockSdkClientListRelationsRequestInterface(mockCtrl)

	relations := []string{}
	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}
	body := client.ClientListRelationsRequest{
		User:             "user:foo",
		Relations:        []string{"viewer"},
		Object:           "doc:doc1",
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	mockBody.EXPECT().Body(body).Return(mockListRelationsRequest)
	gomock.InOrder(
		mockFgaClient.EXPECT().ReadLatestAuthorizationModel(context.Background()).Return(mockExecute),
		mockFgaClient.EXPECT().ListRelations(context.Background()).Return(mockBody),
	)

	var clientConfig fga.ClientConfig

	_, err := listRelations(clientConfig, mockFgaClient, "user:foo", "doc:doc1", relations, contextualTuples, queryContext)
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestListRelationsLatestAuthModelEmpty(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadLatestAuthorizationModelRequestInterface(mockCtrl)

	var expectedResponse client.ClientReadAuthorizationModelResponse

	modelJSON := `{"authorization_model":{"id":"01GXSA8YR785C4FYS3C0RTG7B1","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"doc"},{"relations":{},"type":"user"}]}}` //nolint:all
	if err := json.Unmarshal([]byte(modelJSON), &expectedResponse); err != nil {
		t.Fatalf("%v", err)
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)
	mockFgaClient.EXPECT().ReadLatestAuthorizationModel(context.Background()).Return(mockExecute)

	relations := []string{}
	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}

	var clientConfig fga.ClientConfig

	expectedListRelationsResponse := client.ClientListRelationsResponse{
		Relations: []string{},
	}

	response, err := listRelations(clientConfig, mockFgaClient, "doc:doc1", "user:foo", relations,
		contextualTuples, queryContext)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*response, expectedListRelationsResponse) {
		t.Errorf("Expect response %v actual %v", expectedListRelationsResponse, *response)
	}
}

func TestListRelationsLatestAuthModelList(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadLatestAuthorizationModelRequestInterface(mockCtrl)

	var expectedResponse client.ClientReadAuthorizationModelResponse

	modelJSON := `{"authorization_model":{"id":"01GXSA8YR785C4FYS3C0RTG7B1","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"doc"}]}}` //nolint:all
	if err := json.Unmarshal([]byte(modelJSON), &expectedResponse); err != nil {
		t.Fatalf("%v", err)
	}

	mockExecute.EXPECT().Execute().Return(&expectedResponse, nil)

	// after reading the latest auth model, expect to call list relations but failure

	mockListRelationsExecute := mock_client.NewMockSdkClientListRelationsRequestInterface(mockCtrl)

	expectedListRelationsResponse := client.ClientListRelationsResponse{
		Relations: []string{"viewer"},
	}

	mockListRelationsExecute.EXPECT().Execute().Return(&expectedListRelationsResponse, nil)

	mockListRelationsRequest := mock_client.NewMockSdkClientListRelationsRequestInterface(mockCtrl)
	listRelationsOptions := client.ClientListRelationsOptions{}
	mockListRelationsRequest.EXPECT().Options(listRelationsOptions).Return(mockListRelationsExecute)

	mockBody := mock_client.NewMockSdkClientListRelationsRequestInterface(mockCtrl)

	relations := []string{}
	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}
	body := client.ClientListRelationsRequest{
		User:             "user:foo",
		Relations:        []string{"viewer"},
		Object:           "doc:doc1",
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	mockBody.EXPECT().Body(body).Return(mockListRelationsRequest)
	gomock.InOrder(
		mockFgaClient.EXPECT().ReadLatestAuthorizationModel(context.Background()).Return(mockExecute),
		mockFgaClient.EXPECT().ListRelations(context.Background()).Return(mockBody),
	)

	var clientConfig fga.ClientConfig

	output, err := listRelations(clientConfig, mockFgaClient, "user:foo", "doc:doc1", relations,
		contextualTuples, queryContext)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*output, expectedListRelationsResponse) {
		t.Errorf("Expect output %v actual %v", expectedListRelationsResponse, *output)
	}
}

func TestListRelationsMultipleRelations(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	// after reading the latest auth model, expect to call list relations but failure

	mockListRelationsExecute := mock_client.NewMockSdkClientListRelationsRequestInterface(mockCtrl)

	expectedListRelationsResponse := client.ClientListRelationsResponse{
		Relations: []string{"viewer"},
	}

	mockListRelationsExecute.EXPECT().Execute().Return(&expectedListRelationsResponse, nil)

	mockListRelationsRequest := mock_client.NewMockSdkClientListRelationsRequestInterface(mockCtrl)
	listRelationsOptions := client.ClientListRelationsOptions{}
	mockListRelationsRequest.EXPECT().Options(listRelationsOptions).Return(mockListRelationsExecute)

	mockBody := mock_client.NewMockSdkClientListRelationsRequestInterface(mockCtrl)

	relations := []string{"viewer", "editor"}
	contextualTuples := []client.ClientContextualTupleKey{
		{User: "user:foo", Relation: "admin", Object: "doc:doc1"},
	}
	body := client.ClientListRelationsRequest{
		User:             "user:foo",
		Relations:        []string{"viewer", "editor"},
		Object:           "doc:doc1",
		ContextualTuples: contextualTuples,
		Context:          queryContext,
	}
	mockBody.EXPECT().Body(body).Return(mockListRelationsRequest)
	gomock.InOrder(
		mockFgaClient.EXPECT().ListRelations(context.Background()).Return(mockBody),
	)

	var clientConfig fga.ClientConfig

	output, err := listRelations(clientConfig, mockFgaClient, "user:foo", "doc:doc1", relations,
		contextualTuples, queryContext)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*output, expectedListRelationsResponse) {
		t.Errorf("Expect output %v actual %v", expectedListRelationsResponse, *output)
	}
}
