package store

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	mockclient "github.com/openfga/cli/internal/mocks"
)

var errMockListStores = errors.New("mock error")

func TestListStoresError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)

	var response openfga.ListStoresResponse

	mockExecute.EXPECT().Execute().Return(&response, errMockListStores)

	mockRequest := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)
	options := client.ClientListStoresOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)
	mockFgaClient.EXPECT().ListStores(context.Background()).Return(mockRequest)

	_, err := listStores(mockFgaClient, 5)
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestListStoresEmpty(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)

	var stores []openfga.Store

	response := openfga.ListStoresResponse{
		Stores:            stores,
		ContinuationToken: "",
	}
	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)
	options := client.ClientListStoresOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)
	mockFgaClient.EXPECT().ListStores(context.Background()).Return(mockRequest)

	output, err := listStores(mockFgaClient, 5)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","stores":[]}`

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, string(outputTxt))
	}
}

func TestListStoresSinglePage(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)

	expectedTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	stores := []openfga.Store{
		{
			Id:        "12345",
			Name:      "foo",
			CreatedAt: expectedTime,
			UpdatedAt: expectedTime,
		},
	}

	response := openfga.ListStoresResponse{
		Stores:            stores,
		ContinuationToken: "",
	}
	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)
	options := client.ClientListStoresOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)
	mockFgaClient.EXPECT().ListStores(context.Background()).Return(mockRequest)

	output, err := listStores(mockFgaClient, 5)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","stores":[{"created_at":"2009-11-10T23:00:00Z","id":"12345","name":"foo","updated_at":"2009-11-10T23:00:00Z"}]}` //nolint:lll

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, string(outputTxt))
	}
}

func TestListStoresMultiPage(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	const continuationToken = "01GXSA8YR785C4FYS3C0RTG7B2" //nolint:gosec

	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute1 := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)

	expectedTime1 := time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC)

	stores1 := []openfga.Store{
		{
			Id:        "abcde",
			Name:      "moo",
			CreatedAt: expectedTime1,
			UpdatedAt: expectedTime1,
		},
	}

	response1 := openfga.ListStoresResponse{
		Stores:            stores1,
		ContinuationToken: continuationToken,
	}

	mockExecute2 := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)

	expectedTime2 := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	stores2 := []openfga.Store{
		{
			Id:        "12345",
			Name:      "foo",
			CreatedAt: expectedTime2,
			UpdatedAt: expectedTime2,
		},
	}

	response2 := openfga.ListStoresResponse{
		Stores:            stores2,
		ContinuationToken: "",
	}
	gomock.InOrder(
		mockExecute1.EXPECT().Execute().Return(&response1, nil),
		mockExecute2.EXPECT().Execute().Return(&response2, nil),
	)

	mockRequest1 := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)
	options1 := client.ClientListStoresOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest2 := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)
	options2 := client.ClientListStoresOptions{
		ContinuationToken: openfga.PtrString(continuationToken),
	}
	gomock.InOrder(
		mockRequest1.EXPECT().Options(options1).Return(mockExecute1),
		mockRequest2.EXPECT().Options(options2).Return(mockExecute2),
	)
	gomock.InOrder(
		mockFgaClient.EXPECT().ListStores(context.Background()).Return(mockRequest1),
		mockFgaClient.EXPECT().ListStores(context.Background()).Return(mockRequest2),
	)

	output, err := listStores(mockFgaClient, 5)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","stores":[{"created_at":"2009-11-10T22:00:00Z","id":"abcde","name":"moo","updated_at":"2009-11-10T22:00:00Z"},{"created_at":"2009-11-10T23:00:00Z","id":"12345","name":"foo","updated_at":"2009-11-10T23:00:00Z"}]}` //nolint:lll

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, string(outputTxt))
	}
}

func TestListStoresMultiPageMaxPage(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	const continuationToken = "01GXSA8YR785C4FYS3C0RTG7B2" //nolint:gosec

	mockFgaClient := mockclient.NewMockSdkClient(mockCtrl)

	mockExecute1 := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)

	expectedTime1 := time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC)

	stores1 := []openfga.Store{
		{
			Id:        "abcde",
			Name:      "moo",
			CreatedAt: expectedTime1,
			UpdatedAt: expectedTime1,
		},
	}

	response1 := openfga.ListStoresResponse{
		Stores:            stores1,
		ContinuationToken: continuationToken,
	}

	mockExecute1.EXPECT().Execute().Return(&response1, nil)

	mockRequest1 := mockclient.NewMockSdkClientListStoresRequestInterface(mockCtrl)
	options1 := client.ClientListStoresOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest1.EXPECT().Options(options1).Return(mockExecute1)
	mockFgaClient.EXPECT().ListStores(context.Background()).Return(mockRequest1)

	output, err := listStores(mockFgaClient, 1)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","stores":[{"created_at":"2009-11-10T22:00:00Z","id":"abcde","name":"moo","updated_at":"2009-11-10T22:00:00Z"}]}` //nolint:lll

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, string(outputTxt))
	}
}
