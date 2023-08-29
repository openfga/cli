package authorizationmodel

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/oklog/ulid/v2"
	pb "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/memory"
)

func checkStringArraysEqual(array1 []string, array2 []string) bool {
	if len(array1) != len(array2) {
		return false
	}

	sort.Strings(array1)
	sort.Strings(array2)

	for index, value := range array1 {
		if value != array2[index] {
			return false
		}
	}

	return true
}

type ModelTestCheckSingleResult struct {
	Request    client.ClientCheckRequest `json:"request"`
	Expected   bool                      `json:"expected"`
	Got        *bool                     `json:"got"`
	Error      error                     `json:"error"`
	TestResult bool                      `json:"test_result"`
}

func (result ModelTestCheckSingleResult) IsPassing() bool {
	return result.Error == nil && *result.Got == result.Expected
}

type ModelTestListObjectsSingleResult struct {
	Request    client.ClientListObjectsRequest `json:"request"`
	Expected   []string                        `json:"expected"`
	Got        *[]string                       `json:"got"`
	Error      error                           `json:"error"`
	TestResult bool                            `json:"test_result"`
}

func (result ModelTestListObjectsSingleResult) IsPassing() bool {
	return result.Error == nil && checkStringArraysEqual(*result.Got, result.Expected)
}

type TestResult struct {
	Name               string                             `json:"name"`
	Description        string                             `json:"description"`
	CheckResults       []ModelTestCheckSingleResult       `json:"check_results"`
	ListObjectsResults []ModelTestListObjectsSingleResult `json:"list_objects_results"`
}

//nolint:cyclop
func (result TestResult) FriendlyDisplay() string {
	totalCheckCount := len(result.CheckResults)
	failedCheckCount := 0
	totalListObjectsCount := len(result.ListObjectsResults)
	failedListObjectsCount := 0
	checkResultsOutput := ""
	listObjectsResultsOutput := ""

	if totalCheckCount > 0 {
		for index := 0; index < totalCheckCount; index++ {
			checkResult := result.CheckResults[index]

			if result.CheckResults[index].IsPassing() {
				checkResultsOutput = fmt.Sprintf(
					"%s\n✓ Check(user=%s,relation=%s,object=%s)",
					checkResultsOutput,
					checkResult.Request.User,
					checkResult.Request.Relation,
					checkResult.Request.Object,
				)
			} else {
				failedCheckCount++

				got := "N/A"
				if checkResult.Got != nil {
					got = fmt.Sprintf("%t", *checkResult.Got)
				}

				checkResultsOutput = fmt.Sprintf(
					"%s\nⅹ Check(user=%s,relation=%s,object=%s): expected=%t, got=%s, error=%v",
					checkResultsOutput,
					checkResult.Request.User,
					checkResult.Request.Relation,
					checkResult.Request.Object,
					checkResult.Expected,
					got,
					checkResult.Error,
				)
			}
		}
	}

	if totalListObjectsCount > 0 {
		for index := 0; index < totalListObjectsCount; index++ {
			listObjectsResult := result.ListObjectsResults[index]

			if result.ListObjectsResults[index].IsPassing() {
				listObjectsResultsOutput = fmt.Sprintf(
					"%s\n✓ ListObjects(user=%s,relation=%s,type=%s)",
					listObjectsResultsOutput,
					listObjectsResult.Request.User,
					listObjectsResult.Request.Relation,
					listObjectsResult.Request.Type,
				)
			} else {
				failedListObjectsCount++

				got := "N/A"
				if listObjectsResult.Got != nil {
					got = fmt.Sprintf("%s", *listObjectsResult.Got)
				}

				listObjectsResultsOutput = fmt.Sprintf(
					"%s\nⅹ ListObjects(user=%s,relation=%s,type=%s): expected=%s, got=%s, error=%v",
					listObjectsResultsOutput,
					listObjectsResult.Request.User,
					listObjectsResult.Request.Relation,
					listObjectsResult.Request.Type,
					listObjectsResult.Expected,
					got,
					listObjectsResult.Error,
				)
			}
		}
	}

	testStatus := "PASSING"
	if failedCheckCount+failedListObjectsCount != 0 {
		testStatus = "FAILING"
	}

	output := fmt.Sprintf(
		"(%s) %s: Checks (%d/%d passing) | ListObjects (%d/%d passing)",
		testStatus,
		result.Name,
		totalCheckCount-failedCheckCount,
		totalCheckCount,
		totalListObjectsCount-failedListObjectsCount,
		totalListObjectsCount,
	)

	if failedCheckCount > 0 {
		output = fmt.Sprintf("%s%s", output, checkResultsOutput)
	}

	if failedListObjectsCount > 0 {
		output = fmt.Sprintf("%s%s", output, listObjectsResultsOutput)
	}

	return output
}

type ModelTestCheck struct {
	User       string          `json:"user"`
	Object     string          `json:"object"`
	Assertions map[string]bool `json:"assertions"`
}

type ModelTestListObjects struct {
	User       string              `json:"user"`
	Type       string              `json:"type"`
	Assertions map[string][]string `json:"assertions"`
}

type ModelTest struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Model       string                  `json:"model"`
	Tuples      []client.ClientTupleKey `json:"tuples"`
	Check       []ModelTestCheck        `json:"check"`
	ListObjects []ModelTestListObjects  `json:"list_objects" yaml:"list-objects"` //nolint:tagliatelle
}

type ModelTestOptions struct {
	StoreID *string
	ModelID *string
	Remote  bool
}

func RunSingleLocalCheckTest(
	fgaServer *server.Server,
	checkRequest *pb.CheckRequest,
	tuples []client.ClientTupleKey,
	expectation bool,
) ModelTestCheckSingleResult {
	res, err := fgaServer.Check(context.Background(), checkRequest)

	result := ModelTestCheckSingleResult{
		Request: client.ClientCheckRequest{
			User:             checkRequest.GetTupleKey().GetUser(),
			Relation:         checkRequest.GetTupleKey().GetRelation(),
			Object:           checkRequest.GetTupleKey().GetObject(),
			ContextualTuples: &tuples,
		},
		Expected: expectation,
		Error:    err,
	}

	if err == nil && res != nil {
		result.Got = &res.Allowed
		result.TestResult = result.IsPassing()
	}

	return result
}

func RunLocalCheckTest(
	fgaServer *server.Server,
	checkTest ModelTestCheck,
	tuples []client.ClientTupleKey,
	options ModelTestOptions,
) []ModelTestCheckSingleResult {
	results := []ModelTestCheckSingleResult{}

	for relation, expectation := range checkTest.Assertions {
		result := RunSingleLocalCheckTest(
			fgaServer,
			&pb.CheckRequest{
				StoreId:              *options.StoreID,
				AuthorizationModelId: *options.ModelID,
				TupleKey: &pb.TupleKey{
					User:     checkTest.User,
					Relation: relation,
					Object:   checkTest.Object,
				},
			},
			tuples,
			expectation,
		)
		results = append(results, result)
	}

	return results
}

func RunSingleLocalListObjectsTest(
	fgaServer *server.Server,
	listObjectsRequest *pb.ListObjectsRequest,
	tuples []client.ClientTupleKey,
	expectation []string,
) ModelTestListObjectsSingleResult {
	response, err := fgaServer.ListObjects(context.Background(), listObjectsRequest)

	result := ModelTestListObjectsSingleResult{
		Request: client.ClientListObjectsRequest{
			User:             listObjectsRequest.GetUser(),
			Relation:         listObjectsRequest.GetRelation(),
			Type:             listObjectsRequest.GetType(),
			ContextualTuples: &tuples,
		},
		Expected: expectation,
		Error:    err,
	}

	if response != nil {
		result.Got = &response.Objects
		result.TestResult = result.IsPassing()
	}

	return result
}

func RunLocalListObjectsTest(
	fgaServer *server.Server,
	listObjectsTest ModelTestListObjects,
	tuples []client.ClientTupleKey,
	options ModelTestOptions,
) []ModelTestListObjectsSingleResult {
	results := []ModelTestListObjectsSingleResult{}

	for relation, expectation := range listObjectsTest.Assertions {
		result := RunSingleLocalListObjectsTest(fgaServer,
			&pb.ListObjectsRequest{
				StoreId:              *options.StoreID,
				AuthorizationModelId: *options.ModelID,
				User:                 listObjectsTest.User,
				Type:                 listObjectsTest.Type,
				Relation:             relation,
			},
			tuples,
			expectation,
		)
		results = append(results, result)
	}

	return results
}

func RunLocalTest(
	fgaServer *server.Server,
	test ModelTest,
	model *AuthzModel,
) (TestResult, error) {
	checkResults := []ModelTestCheckSingleResult{}
	listObjectResults := []ModelTestListObjectsSingleResult{}

	storeID, modelID, err := initLocalStore(fgaServer, model.GetProtoModel(), test)
	if err != nil {
		return TestResult{}, err
	}

	testOptions := ModelTestOptions{
		StoreID: storeID,
		ModelID: modelID,
	}

	for index := 0; index < len(test.Check); index++ {
		results := RunLocalCheckTest(fgaServer, test.Check[index], test.Tuples, testOptions)
		checkResults = append(checkResults, results...)
	}

	for index := 0; index < len(test.ListObjects); index++ {
		results := RunLocalListObjectsTest(fgaServer, test.ListObjects[index], test.Tuples, testOptions)
		listObjectResults = append(listObjectResults, results...)
	}

	return TestResult{
		Name:               test.Name,
		Description:        test.Description,
		CheckResults:       checkResults,
		ListObjectsResults: listObjectResults,
	}, nil
}

func RunSingleRemoteCheckTest(
	fgaClient *client.OpenFgaClient,
	checkRequest client.ClientCheckRequest,
	expectation bool,
) ModelTestCheckSingleResult {
	res, err := fgaClient.Check(context.Background()).Body(checkRequest).Execute()

	result := ModelTestCheckSingleResult{
		Request:  checkRequest,
		Expected: expectation,
		Error:    err,
	}

	if err == nil && res != nil {
		result.Got = res.Allowed
		result.TestResult = result.IsPassing()
	}

	return result
}

func RunRemoteCheckTest(
	fgaClient *client.OpenFgaClient,
	checkTest ModelTestCheck,
	tuples []client.ClientTupleKey,
) []ModelTestCheckSingleResult {
	results := []ModelTestCheckSingleResult{}

	for relation, expectation := range checkTest.Assertions {
		result := RunSingleRemoteCheckTest(
			fgaClient,
			client.ClientCheckRequest{
				User:             checkTest.User,
				Relation:         relation,
				Object:           checkTest.Object,
				ContextualTuples: &tuples,
			},
			expectation,
		)
		results = append(results, result)
	}

	return results
}

func RunSingleRemoteListObjectsTest(
	fgaClient *client.OpenFgaClient,
	listObjectsRequest client.ClientListObjectsRequest,
	expectation []string,
) ModelTestListObjectsSingleResult {
	response, err := fgaClient.ListObjects(context.Background()).Body(listObjectsRequest).Execute()

	result := ModelTestListObjectsSingleResult{
		Request:  listObjectsRequest,
		Expected: expectation,
		Error:    err,
	}

	if response != nil {
		result.Got = response.Objects
		result.TestResult = result.IsPassing()
	}

	return result
}

func RunRemoteListObjectsTest(
	fgaClient *client.OpenFgaClient,
	listObjectsTest ModelTestListObjects,
	tuples []client.ClientTupleKey,
) []ModelTestListObjectsSingleResult {
	results := []ModelTestListObjectsSingleResult{}

	for relation, expectation := range listObjectsTest.Assertions {
		result := RunSingleRemoteListObjectsTest(fgaClient,
			client.ClientListObjectsRequest{
				User:             listObjectsTest.User,
				Type:             listObjectsTest.Type,
				Relation:         relation,
				ContextualTuples: &tuples,
			},
			expectation,
		)
		results = append(results, result)
	}

	return results
}

func RunRemoteTest(fgaClient *client.OpenFgaClient, test ModelTest) TestResult {
	checkResults := []ModelTestCheckSingleResult{}

	for index := 0; index < len(test.Check); index++ {
		results := RunRemoteCheckTest(fgaClient, test.Check[index], test.Tuples)
		checkResults = append(checkResults, results...)
	}

	listObjectResults := []ModelTestListObjectsSingleResult{}

	for index := 0; index < len(test.ListObjects); index++ {
		results := RunRemoteListObjectsTest(fgaClient, test.ListObjects[index], test.Tuples)
		listObjectResults = append(listObjectResults, results...)
	}

	return TestResult{
		Name:               test.Name,
		Description:        test.Description,
		CheckResults:       checkResults,
		ListObjectsResults: listObjectResults,
	}
}

const writeMaxChunkSize = 40

func initLocalStore(
	fgaServer *server.Server,
	model *pb.AuthorizationModel,
	test ModelTest,
) (*string, *string, error) {
	var modelID *string

	storeID := ulid.Make().String()
	tuples := []*pb.TupleKey{}

	for index := 0; index < len(test.Tuples); index++ {
		tuple := test.Tuples[index]
		tpl := pb.TupleKey{
			User:     tuple.User,
			Relation: tuple.Relation,
			Object:   tuple.Object,
		}
		tuples = append(tuples, &tpl)
	}

	var authModelWriteReq *pb.WriteAuthorizationModelRequest

	if test.Model != "" {
		authModel := AuthzModel{}
		if err := authModel.ReadFromDSLString(test.Model); err != nil {
			return nil, nil, err
		}

		protoModel := authModel.GetProtoModel()
		authModelWriteReq = &pb.WriteAuthorizationModelRequest{
			StoreId:         storeID,
			TypeDefinitions: protoModel.GetTypeDefinitions(),
			SchemaVersion:   protoModel.GetSchemaVersion(),
		}
	} else if model != nil {
		authModelWriteReq = &pb.WriteAuthorizationModelRequest{
			StoreId:         storeID,
			TypeDefinitions: model.GetTypeDefinitions(),
			SchemaVersion:   model.GetSchemaVersion(),
		}
	}

	if authModelWriteReq != nil {
		writtenModel, err := fgaServer.WriteAuthorizationModel(context.Background(), authModelWriteReq)
		if err != nil {
			return nil, nil, err //nolint:wrapcheck
		}

		modelID = &writtenModel.AuthorizationModelId
	}

	tuplesLength := len(tuples)
	if tuplesLength > 0 {
		for i := 0; i < tuplesLength; i += writeMaxChunkSize {
			end := int(math.Min(float64(i+writeMaxChunkSize), float64(tuplesLength)))
			writeChunk := (tuples)[i:end]

			_, err := fgaServer.Write(context.Background(), &pb.WriteRequest{
				StoreId: storeID,
				Writes:  &pb.TupleKeys{TupleKeys: writeChunk},
			})
			if err != nil {
				return nil, nil, err //nolint:wrapcheck
			}
		}
	}

	return &storeID, modelID, nil
}

func readAuthzModel(modelFileName string, testInputFormat ModelFormat) (*AuthzModel, error) {
	if modelFileName == "" {
		return nil, nil //nolint:nilnil
	}

	inputModel, err := ReadFromInputFile(
		modelFileName,
		&testInputFormat)
	if err != nil {
		return nil, err
	}

	authModel := AuthzModel{}

	if testInputFormat == ModelFormatJSON {
		err = authModel.ReadFromJSONString(*inputModel)
	} else {
		err = authModel.ReadFromDSLString(*inputModel)
	}

	if err != nil {
		return nil, err
	}

	return &authModel, nil
}

func RunTests(
	fgaClient *client.OpenFgaClient,
	tests []ModelTest,
	modelFileName string,
	testInputFormat ModelFormat,
	remote bool,
) ([]TestResult, error) {
	results := []TestResult{}

	if !remote {
		model, err := readAuthzModel(modelFileName, testInputFormat)
		if err != nil {
			return results, err
		}

		ds := memory.New()

		fgaServer, err := server.NewServerWithOpts(server.WithDatastore(ds))
		if err != nil {
			return results, err //nolint:wrapcheck
		}

		for index := 0; index < len(tests); index++ {
			result, err := RunLocalTest(fgaServer, tests[index], model)
			if err != nil {
				return results, err
			}

			results = append(results, result)
		}
	} else {
		for index := 0; index < len(tests); index++ {
			result := RunRemoteTest(fgaClient, tests[index])

			results = append(results, result)
		}
	}

	return results, nil
}
