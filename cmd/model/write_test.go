package model

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/authorizationmodel"
	mockclient "github.com/openfga/cli/internal/mocks"
)

var errMockWrite = errors.New("mock error")

func TestWriteModelFail(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	modelJSONTxt := `{"schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}],"conditions":{}}` //nolint:lll
	body := &client.ClientWriteAuthorizationModelRequest{}

	err := json.Unmarshal([]byte(modelJSONTxt), &body)
	if err != nil {
		t.Fatal(err)
	}

	mockExecute := mockclient.NewMockSdkClientWriteAuthorizationModelRequestInterface(mockCtrl)
	mockExecute.EXPECT().Execute().Return(nil, errMockWrite)

	mockRequest := mockclient.NewMockSdkClientWriteAuthorizationModelRequestInterface(mockCtrl)
	mockRequest.EXPECT().Body(*body).Return(mockExecute)

	mockFgaClient.EXPECT().WriteAuthorizationModel(context.Background()).Return(mockRequest)

	model := authorizationmodel.AuthzModel{}
	err = model.ReadFromJSONString(modelJSONTxt)

	if err != nil {
		return
	}

	_, err = Write(mockFgaClient, model)
	if err == nil {
		t.Fatalf("Expect error but there is none")
	}
}

func TestWriteModel(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	modelJSONTxt := `{"schema_version":"1.1","type_definitions":[{"relations":{"viewer":{"this":{}}},"type":"github-repo"}],"conditions":{}}` //nolint:lll

	body := &client.ClientWriteAuthorizationModelRequest{}

	err := json.Unmarshal([]byte(modelJSONTxt), &body)
	if err != nil {
		t.Fatal(err)
	}

	mockExecute := mockclient.NewMockSdkClientWriteAuthorizationModelRequestInterface(mockCtrl)

	modelID := "01GXSB8YR785C4FYS3C0RTG7C2"
	response := client.ClientWriteAuthorizationModelResponse{
		AuthorizationModelId: modelID,
	}
	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mockclient.NewMockSdkClientWriteAuthorizationModelRequestInterface(mockCtrl)
	mockRequest.EXPECT().Body(*body).Return(mockExecute)

	mockFgaClient.EXPECT().WriteAuthorizationModel(context.Background()).Return(mockRequest)

	model := authorizationmodel.AuthzModel{}

	err = model.ReadFromJSONString(modelJSONTxt)
	if err != nil {
		return
	}

	output, err := Write(mockFgaClient, model)
	if err != nil {
		t.Fatal(err)
	}

	if *output != response {
		t.Fatalf("Expected output %v actual %v", response, *output)
	}
}
