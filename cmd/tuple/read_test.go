package tuple

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"go.uber.org/mock/gomock"

	mock_client "github.com/openfga/cli/internal/mocks"
	"github.com/openfga/cli/internal/tuple"
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
		PageSize:          openfga.PtrInt32(tuple.DefaultReadPageSize),
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

	mockFgaClient.EXPECT().Read(t.Context()).Return(mockBody)

	_, err := read(
		t.Context(),
		mockFgaClient,
		"user:user1",
		"reader",
		"document:doc1",
		5,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
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
		PageSize:          openfga.PtrInt32(tuple.DefaultReadPageSize),
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

	mockFgaClient.EXPECT().Read(t.Context()).Return(mockBody)

	output, err := read(
		t.Context(),
		mockFgaClient,
		"user:user1",
		"reader",
		"document:doc1",
		5,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
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
		PageSize:          openfga.PtrInt32(tuple.DefaultReadPageSize),
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

	mockFgaClient.EXPECT().Read(t.Context()).Return(mockBody)

	output, err := read(
		t.Context(),
		mockFgaClient,
		"user:user1",
		"reader",
		"document:doc1",
		5,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","tuples":[{"key":{"object":"document:doc1","relation":"reader","user":"user:user1"},"timestamp":"2009-11-10T23:00:00Z"}]}`
	simpleOutput := `[{"object":"document:doc1","relation":"reader","user":"user:user1"}]`

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
		PageSize:          openfga.PtrInt32(tuple.DefaultReadPageSize),
		ContinuationToken: openfga.PtrString(""),
	}
	mockRequest2 := mock_client.NewMockSdkClientReadRequestInterface(mockCtrl)
	options2 := client.ClientReadOptions{
		PageSize:          openfga.PtrInt32(tuple.DefaultReadPageSize),
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
		mockFgaClient.EXPECT().Read(t.Context()).Return(mockBody1),
		mockFgaClient.EXPECT().Read(t.Context()).Return(mockBody2),
	)

	output, err := read(
		t.Context(),
		mockFgaClient,
		"user:user1",
		"reader",
		"document:doc1",
		5,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","tuples":[{"key":{"object":"document:doc1","relation":"reader","user":"user:user1"},"timestamp":"2009-11-10T22:00:00Z"},{"key":{"object":"document:doc2","relation":"reader","user":"user:user1"},"timestamp":"2009-11-10T23:00:00Z"}]}`
	simpleOutput := `[{"object":"document:doc1","relation":"reader","user":"user:user1"},{"object":"document:doc2","relation":"reader","user":"user:user1"}]`

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
		PageSize:          openfga.PtrInt32(tuple.DefaultReadPageSize),
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

	mockFgaClient.EXPECT().Read(t.Context()).Return(mockBody)

	output, err := read(
		t.Context(),
		mockFgaClient,
		"user:user1",
		"reader",
		"document:doc1",
		1,
		openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
	)
	if err != nil {
		t.Error(err)
	}

	expectedOutput := `{"continuation_token":"","tuples":[{"key":{"object":"document:doc1","relation":"reader","user":"user:user1"},"timestamp":"2009-11-10T23:00:00Z"}]}`
	simpleOutput := `[{"object":"document:doc1","relation":"reader","user":"user:user1"}]`

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

func TestReadResponseCSVDTOParser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		readRes  readResponse
		expected []readResponseCSVDTO
	}{
		{
			readRes: readResponse{
				simple: []openfga.TupleKey{
					{
						User:     "user:anne",
						Relation: "reader",
						Object:   "document:secret.doc",
						Condition: &openfga.RelationshipCondition{
							Name:    "inOfficeIP",
							Context: toPointer(map[string]interface{}{"ip_addr": "10.0.0.1"}),
						},
					},
					{
						User:      "user:john",
						Relation:  "writer",
						Object:    "document:abc.doc",
						Condition: &openfga.RelationshipCondition{},
					},
				},
			},
			expected: []readResponseCSVDTO{
				{
					UserType:         "user",
					UserID:           "anne",
					Relation:         "reader",
					ObjectType:       "document",
					ObjectID:         "secret.doc",
					ConditionName:    "inOfficeIP",
					ConditionContext: "{\"ip_addr\":\"10.0.0.1\"}",
				},
				{
					UserType:   "user",
					UserID:     "john",
					Relation:   "writer",
					ObjectType: "document",
					ObjectID:   "abc.doc",
				},
			},
		},
	}

	for _, testCase := range testCases {
		outcome, _ := testCase.readRes.toCsvDTO()
		assert.Equal(t, testCase.expected, outcome)
	}
}

func toPointer[T any](p T) *T {
	return &p
}
