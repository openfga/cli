package model

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	mockclient "github.com/openfga/cli/internal/mocks"
)

var errMockList = errors.New("mock error")

const model1JSON = `{"id":"01GXSA8YR785C4FYS3C0RTG7B1","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}]}` //nolint:all

func TestListModelsEmpty(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)

	var models []openfga.AuthorizationModel

	response := openfga.ReadAuthorizationModelsResponse{
		AuthorizationModels: models,
		ContinuationToken:   openfga.PtrString(""),
	}
	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)
	options := client.ClientReadAuthorizationModelsOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)
	mockFgaClient.EXPECT().ReadAuthorizationModels(context.Background()).Return(mockRequest)

	output, err := listModels(mockFgaClient, 5)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := "{\"authorization_models\":[]}"

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, output)
	}
}

func TestListModelsFail(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)

	mockExecute.EXPECT().Execute().Return(nil, errMockList)

	mockRequest := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)
	options := client.ClientReadAuthorizationModelsOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)
	mockFgaClient.EXPECT().ReadAuthorizationModels(context.Background()).Return(mockRequest)

	_, err := listModels(mockFgaClient, 5)
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestListModelsSinglePage(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)

	var model1 openfga.AuthorizationModel

	if err := json.Unmarshal([]byte(model1JSON), &model1); err != nil {
		t.Fatalf("%v", err)
	}

	models := []openfga.AuthorizationModel{
		model1,
	}

	response := openfga.ReadAuthorizationModelsResponse{
		AuthorizationModels: models,
		ContinuationToken:   openfga.PtrString(""),
	}
	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)
	options := client.ClientReadAuthorizationModelsOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)
	mockFgaClient.EXPECT().ReadAuthorizationModels(context.Background()).Return(mockRequest)

	output, err := listModels(mockFgaClient, 5)
	if err != nil {
		t.Error(err)
	}
	expectedOutput := `{"authorization_models":[{"id":"01GXSA8YR785C4FYS3C0RTG7B1","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}]}]}` //nolint:all

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, output)
	}
}

func TestListModelsMultiPage(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute1 := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)

	var model1 openfga.AuthorizationModel

	if err := json.Unmarshal([]byte(model1JSON), &model1); err != nil {
		t.Fatalf("%v", err)
	}

	models1 := []openfga.AuthorizationModel{
		model1,
	}
	continuationToken1 := openfga.PtrString("abcdef")
	response1 := openfga.ReadAuthorizationModelsResponse{
		AuthorizationModels: models1,
		ContinuationToken:   continuationToken1,
	}

	mockRequest1 := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)
	options1 := client.ClientReadAuthorizationModelsOptions{
		ContinuationToken: openfga.PtrString(""),
	}

	mockExecute2 := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)

	var model2 openfga.AuthorizationModel

	model2JSON := `{"id":"01GXSA8YR785C4FYS3C0RTG7B2","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}]}` //nolint:all
	if err := json.Unmarshal([]byte(model2JSON), &model2); err != nil {
		t.Fatalf("%v", err)
	}

	models2 := []openfga.AuthorizationModel{
		model2,
	}
	emptyToken := ""
	response2 := openfga.ReadAuthorizationModelsResponse{
		AuthorizationModels: models2,
		ContinuationToken:   &emptyToken,
	}

	gomock.InOrder(
		mockExecute1.EXPECT().Execute().Return(&response1, nil),
		mockExecute2.EXPECT().Execute().Return(&response2, nil),
	)

	mockRequest2 := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)
	options2 := client.ClientReadAuthorizationModelsOptions{
		ContinuationToken: continuationToken1,
	}
	gomock.InOrder(
		mockRequest1.EXPECT().Options(options1).Return(mockExecute1),
		mockRequest2.EXPECT().Options(options2).Return(mockExecute2),
	)
	gomock.InOrder(
		mockFgaClient.EXPECT().ReadAuthorizationModels(context.Background()).Return(mockRequest1),
		mockFgaClient.EXPECT().ReadAuthorizationModels(context.Background()).Return(mockRequest2),
	)

	output, err := listModels(mockFgaClient, 2)
	if err != nil {
		t.Error(err)
	}
	expectedOutput := `{"authorization_models":[{"id":"01GXSA8YR785C4FYS3C0RTG7B1","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}]},{"id":"01GXSA8YR785C4FYS3C0RTG7B2","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}]}]}` //nolint:all

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, output)
	}
}

// Data has multi page but we are exceeding the maximum allowable page.
func TestListModelsMultiPageMaxPage(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute1 := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)

	var model1 openfga.AuthorizationModel

	if err := json.Unmarshal([]byte(model1JSON), &model1); err != nil {
		t.Fatalf("%v", err)
	}

	models1 := []openfga.AuthorizationModel{
		model1,
	}
	continuationToken1 := openfga.PtrString("abcdef")
	response1 := openfga.ReadAuthorizationModelsResponse{
		AuthorizationModels: models1,
		ContinuationToken:   continuationToken1,
	}

	mockRequest1 := mockclient.NewMockSdkClientReadAuthorizationModelsRequestInterface(mockCtrl)
	options1 := client.ClientReadAuthorizationModelsOptions{
		ContinuationToken: openfga.PtrString(""),
	}

	mockExecute1.EXPECT().Execute().Return(&response1, nil)

	mockRequest1.EXPECT().Options(options1).Return(mockExecute1)
	mockFgaClient.EXPECT().ReadAuthorizationModels(context.Background()).Return(mockRequest1)

	output, err := listModels(mockFgaClient, 0)
	if err != nil {
		t.Error(err)
	}
	expectedOutput := `{"authorization_models":[{"id":"01GXSA8YR785C4FYS3C0RTG7B1","schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}]}]}` //nolint:all

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, output)
	}
}
