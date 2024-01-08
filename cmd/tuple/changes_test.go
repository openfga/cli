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

var errMockReadChanges = errors.New("mock error")

func TestReadChangesError(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	var response openfga.ReadChangesResponse

	mockExecute.EXPECT().Execute().Return(&response, errMockReadChanges)

	mockRequest := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)
	options := client.ClientReadChangesOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	body := client.ClientReadChangesRequest{
		Type: "document",
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().ReadChanges(context.Background()).Return(mockBody)

	_, err := readChanges(mockFgaClient, 5, "document", "")
	if err == nil {
		t.Error("Expect error but there is none")
	}
}

func TestReadChangesEmpty(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	changes := []openfga.TupleChange{}
	response := openfga.ReadChangesResponse{
		Changes:           changes,
		ContinuationToken: openfga.PtrString(""),
	}

	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)
	options := client.ClientReadChangesOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	body := client.ClientReadChangesRequest{
		Type: "document",
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().ReadChanges(context.Background()).Return(mockBody)

	output, err := readChanges(mockFgaClient, 5, "document", "")
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"changes":[],"continuation_token":""}`

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expect output %v actual %v", expectedOutput, output)
	}
}

func TestReadChangesSinglePage(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	tupleKey := openfga.TupleKey{
		User:     "user:user1",
		Relation: "reader",
		Object:   "document:doc1",
	}

	operation := openfga.WRITE
	changesTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	changes := []openfga.TupleChange{
		{
			TupleKey:  tupleKey,
			Operation: operation,
			Timestamp: changesTime,
		},
	}
	response := openfga.ReadChangesResponse{
		Changes:           changes,
		ContinuationToken: openfga.PtrString(""),
	}

	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)
	options := client.ClientReadChangesOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	body := client.ClientReadChangesRequest{
		Type: "document",
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().ReadChanges(context.Background()).Return(mockBody)

	output, err := readChanges(mockFgaClient, 5, "document", "")
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"changes":[{"operation":"TUPLE_OPERATION_WRITE","timestamp":"2009-11-10T23:00:00Z","tuple_key":{"object":"document:doc1","relation":"reader","user":"user:user1"}}],"continuation_token":""}` //nolint:lll

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expect output %v actual %v", expectedOutput, output)
	}
}

func TestReadChangesMultiPages(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	const continuationToken = "01GXSA8YR785C4FYS3C0RTG7B2" //nolint:gosec

	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)
	mockExecute := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	tupleKey1 := openfga.TupleKey{
		User:     "user:user1",
		Relation: "reader",
		Object:   "document:doc1",
	}

	operation1 := openfga.WRITE
	changesTime1 := time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC)

	changes1 := []openfga.TupleChange{
		{
			TupleKey:  tupleKey1,
			Operation: operation1,
			Timestamp: changesTime1,
		},
	}
	response1 := openfga.ReadChangesResponse{
		Changes:           changes1,
		ContinuationToken: openfga.PtrString(continuationToken),
	}

	operation2 := openfga.DELETE
	changesTime2 := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	changes2 := []openfga.TupleChange{
		{
			TupleKey:  tupleKey1,
			Operation: operation2,
			Timestamp: changesTime2,
		},
	}

	response2 := openfga.ReadChangesResponse{
		Changes:           changes2,
		ContinuationToken: openfga.PtrString(continuationToken),
	}
	gomock.InOrder(
		mockExecute.EXPECT().Execute().Return(&response1, nil),
		mockExecute.EXPECT().Execute().Return(&response2, nil),
	)

	mockRequest1 := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)
	options1 := client.ClientReadChangesOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest2 := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)
	options2 := client.ClientReadChangesOptions{
		ContinuationToken: openfga.PtrString(continuationToken),
	}

	gomock.InOrder(
		mockRequest1.EXPECT().Options(options1).Return(mockExecute),
		mockRequest2.EXPECT().Options(options2).Return(mockExecute),
	)

	mockBody1 := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)
	mockBody2 := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	body := client.ClientReadChangesRequest{
		Type: "document",
	}

	gomock.InOrder(
		mockBody1.EXPECT().Body(body).Return(mockRequest1),
		mockBody2.EXPECT().Body(body).Return(mockRequest2),
	)

	gomock.InOrder(
		mockFgaClient.EXPECT().ReadChanges(context.Background()).Return(mockBody1),
		mockFgaClient.EXPECT().ReadChanges(context.Background()).Return(mockBody2),
	)

	output, err := readChanges(mockFgaClient, 5, "document", "")
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"changes":[{"operation":"TUPLE_OPERATION_WRITE","timestamp":"2009-11-10T22:00:00Z","tuple_key":{"object":"document:doc1","relation":"reader","user":"user:user1"}},{"operation":"TUPLE_OPERATION_DELETE","timestamp":"2009-11-10T23:00:00Z","tuple_key":{"object":"document:doc1","relation":"reader","user":"user:user1"}}],"continuation_token":"01GXSA8YR785C4FYS3C0RTG7B2"}` //nolint:lll

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expect output %v actual %v", expectedOutput, output)
	}
}

func TestReadChangesMultiPagesLimit(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFgaClient := mock_client.NewMockSdkClient(mockCtrl)

	mockExecute := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	tupleKey := openfga.TupleKey{
		User:     "user:user1",
		Relation: "reader",
		Object:   "document:doc1",
	}

	operation := openfga.WRITE
	changesTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	changes := []openfga.TupleChange{
		{
			TupleKey:  tupleKey,
			Operation: operation,
			Timestamp: changesTime,
		},
	}
	response := openfga.ReadChangesResponse{
		Changes:           changes,
		ContinuationToken: openfga.PtrString("01GXSA8YR785C4FYS3C0RTG7B2"),
	}

	mockExecute.EXPECT().Execute().Return(&response, nil)

	mockRequest := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)
	options := client.ClientReadChangesOptions{
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest.EXPECT().Options(options).Return(mockExecute)

	mockBody := mock_client.NewMockSdkClientReadChangesRequestInterface(mockCtrl)

	body := client.ClientReadChangesRequest{
		Type: "document",
	}
	mockBody.EXPECT().Body(body).Return(mockRequest)

	mockFgaClient.EXPECT().ReadChanges(context.Background()).Return(mockBody)

	output, err := readChanges(mockFgaClient, 1, "document", "")
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"changes":[{"operation":"TUPLE_OPERATION_WRITE","timestamp":"2009-11-10T23:00:00Z","tuple_key":{"object":"document:doc1","relation":"reader","user":"user:user1"}}],"continuation_token":"01GXSA8YR785C4FYS3C0RTG7B2"}` //nolint:lll

	outputTxt, err := json.Marshal(output)
	if err != nil {
		t.Error(err)
	}

	if string(outputTxt) != expectedOutput {
		t.Errorf("Expect output %v actual %v", expectedOutput, output)
	}
}
