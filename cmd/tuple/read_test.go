package tuple

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	mock_client "github.com/openfga/cli/internal/mocks"
)

var errMockRead = errors.New("mock error")

func TestReadError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	var response openfga.ReadResponse

	mockExecute.EXPECT().Execute().Return(&response, errMockRead)

	mockRequest := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)
	options := client.ClientReadOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	body := client.ClientReadRequest{
		User:     openfga.PtrString("user:user1"),
		Relation: openfga.PtrString("reader"),
		Object:   openfga.PtrString("document:doc1"),
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().Read(context.Background()).Return(mockBody)

	_, err := read(mockFgaClient, "user:user1", "reader", "document:doc1", 5)
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestReadEmpty(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	var tuples []openfga.Tuple
	response := openfga.ReadResponse{
		Tuples:            tuples,
		ContinuationToken: "",
	}

	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)
	options := client.ClientReadOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	body := client.ClientReadRequest{
		User:     openfga.PtrString("user:user1"),
		Relation: openfga.PtrString("reader"),
		Object:   openfga.PtrString("document:doc1"),
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().Read(context.Background()).Return(mockBody)

	output, err := read(mockFgaClient, "user:user1", "reader", "document:doc1", 5)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := "{\"continuation_token\":\"\",\"tuples\":[]}"
	simpleOutput := "[]"

	outputTxt, err := json.Marshal(output.complete)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, string(outputTxt))
	}

	simpleTxt, err := json.Marshal(output.simple)
	if err != nil {
		t.Error(err)
	}

	if string(simpleTxt) != simpleOutput {
		t.Errorf("Expected output %v actual %v", simpleOutput, string(simpleTxt))
	}
}

func TestReadSinglePage(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	key1 := openfga.TupleKey{
		User:     "user:user1",
		Relation: "reader",
		Object:   "document:doc1",
	}

	changesTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	tuples := []openfga.Tuple{
		{
			Key:       key1,
			Timestamp: changesTime,
		},
	}
	response := openfga.ReadResponse{
		Tuples:            tuples,
		ContinuationToken: "",
	}

	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)
	options := client.ClientReadOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	body := client.ClientReadRequest{
		User:     openfga.PtrString("user:user1"),
		Relation: openfga.PtrString("reader"),
		Object:   openfga.PtrString("document:doc1"),
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().Read(context.Background()).Return(mockBody)

	output, err := read(mockFgaClient, "user:user1", "reader", "document:doc1", 5)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","tuples":[{"key":{"object":"document:doc1","relation":"reader","user":"user:user1"},"timestamp":"2009-11-10T23:00:00Z"}]}` //nolint:lll
	simpleOutput := `[{"object":"document:doc1","relation":"reader","user":"user:user1"}]`                                                                                 //nolint:lll

	outputTxt, err := json.Marshal(output.complete)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, string(outputTxt))
	}

	simpleTxt, err := json.Marshal(output.simple)
	if err != nil {
		t.Error(err)
	}

	if string(simpleTxt) != simpleOutput {
		t.Errorf("Expected output %v actual %v", simpleOutput, string(simpleTxt))
	}
}

func TestReadMultiPages(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	const continuationToken = "01GXSA8YR785C4FYS3C0RTG7B2" //nolint:gosec

	mockExecute1 := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	key1 := openfga.TupleKey{
		User:     "user:user1",
		Relation: "reader",
		Object:   "document:doc1",
	}
	changesTime1 := time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC)

	tuples1 := []openfga.Tuple{
		{
			Key:       key1,
			Timestamp: changesTime1,
		},
	}
	response1 := openfga.ReadResponse{
		Tuples:            tuples1,
		ContinuationToken: continuationToken,
	}

	mockExecute2 := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	key2 := openfga.TupleKey{
		User:     "user:user1",
		Relation: "reader",
		Object:   "document:doc2",
	}

	changesTime2 := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	tuples2 := []openfga.Tuple{
		{
			Key:       key2,
			Timestamp: changesTime2,
		},
	}
	response2 := openfga.ReadResponse{
		Tuples:            tuples2,
		ContinuationToken: "",
	}

	gomock.InOrder(
		mockExecute1.EXPECT().Execute().Return(&response1, nil),
		mockExecute2.EXPECT().Execute().Return(&response2, nil),
	)

	mockRequest1 := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)
	options1 := client.ClientReadOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest2 := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)
	options2 := client.ClientReadOptions{
		ContinuationToken: openfga.PtrString(continuationToken),
	}
	gomock.InOrder(
		mockRequest1.EXPECT().Options(options1).Return(mockExecute1),
		mockRequest2.EXPECT().Options(options2).Return(mockExecute2),
	)

	mockBody1 := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)
	mockBody2 := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	body := client.ClientReadRequest{
		User:     openfga.PtrString("user:user1"),
		Relation: openfga.PtrString("reader"),
		Object:   openfga.PtrString("document:doc1"),
	}
	gomock.InOrder(
		mockBody1.EXPECT().Body(body).Return(mockRequest1),
		mockBody2.EXPECT().Body(body).Return(mockRequest2),
	)

	gomock.InOrder(
		mockFgaClient.EXPECT().Read(context.Background()).Return(mockBody1),
		mockFgaClient.EXPECT().Read(context.Background()).Return(mockBody2),
	)

	output, err := read(mockFgaClient, "user:user1", "reader", "document:doc1", 5)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","tuples":[{"key":{"object":"document:doc1","relation":"reader","user":"user:user1"},"timestamp":"2009-11-10T22:00:00Z"},{"key":{"object":"document:doc2","relation":"reader","user":"user:user1"},"timestamp":"2009-11-10T23:00:00Z"}]}` //nolint:lll
	simpleOutput := `[{"object":"document:doc1","relation":"reader","user":"user:user1"},{"object":"document:doc2","relation":"reader","user":"user:user1"}]`                                                                                                                            //nolint:lll

	outputTxt, err := json.Marshal(output.complete)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, string(outputTxt))
	}

	simpleTxt, err := json.Marshal(output.simple)
	if err != nil {
		t.Error(err)
	}

	if string(simpleTxt) != simpleOutput {
		t.Errorf("Expected output %v actual %v", simpleOutput, string(simpleTxt))
	}
}

func TestReadMultiPagesMaxLimit(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	key1 := openfga.TupleKey{
		User:     "user:user1",
		Relation: "reader",
		Object:   "document:doc1",
	}

	changesTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	tuples := []openfga.Tuple{
		{
			Key:       key1,
			Timestamp: changesTime,
		},
	}
	response := openfga.ReadResponse{
		Tuples:            tuples,
		ContinuationToken: "ABCDEFG",
	}

	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)
	options := client.ClientReadOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)

	body := client.ClientReadRequest{
		User:     openfga.PtrString("user:user1"),
		Relation: openfga.PtrString("reader"),
		Object:   openfga.PtrString("document:doc1"),
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().Read(context.Background()).Return(mockBody)

	output, err := read(mockFgaClient, "user:user1", "reader", "document:doc1", 1)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","tuples":[{"key":{"object":"document:doc1","relation":"reader","user":"user:user1"},"timestamp":"2009-11-10T23:00:00Z"}]}` //nolint:lll
	simpleOutput := `[{"object":"document:doc1","relation":"reader","user":"user:user1"}]`                                                                                 //nolint:lll

	outputTxt, err := json.Marshal(output.complete)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expected output %v actual %v", expectedOutput, string(outputTxt))
	}

	simpleTxt, err := json.Marshal(output.simple)
	if err != nil {
		t.Error(err)
	}

	if string(simpleTxt) != simpleOutput {
		t.Errorf("Expected output %v actual %v", simpleOutput, string(simpleTxt))
	}
}
