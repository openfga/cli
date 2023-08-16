package authorizationmodel

import (
	"context"

	"github.com/openfga/go-sdk/client"
)

func checkStringArraysEqual(array1 []string, array2 []string) bool {
	if len(array1) != len(array2) {
		return false
	}

	for index, value := range array1 {
		if value != array2[index] {
			return false
		}
	}

	return true
}

type ModelTestCheckSingleResult struct {
	Request client.ClientCheckRequest `json:"request"`
	// Response client.ClientCheckResponse `json:"response"`
	Expected   bool  `json:"expected"`
	Got        *bool `json:"got"`
	Error      error `json:"error"`
	TestResult bool  `json:"test_result"`
}

func (result ModelTestCheckSingleResult) IsPassing() bool {
	return result.Error == nil && *result.Got == result.Expected
}

type ModelTestListObjectsSingleResult struct {
	Request client.ClientListObjectsRequest `json:"request"`
	// Response client.ClientListObjectsResponse `json:"response"`
	Expected   []string  `json:"expected"`
	Got        *[]string `json:"got"`
	Error      error     `json:"error"`
	TestResult bool      `json:"test_result"`
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
	Tuples      []client.ClientTupleKey `json:"tuples"`
	Check       []ModelTestCheck        `json:"check"`
	ListObjects []ModelTestListObjects  `json:"list_objects" yaml:"list-objects"` //nolint:tagliatelle
}

func RunSingleCheckTest(
	fgaClient *client.OpenFgaClient,
	checkRequest client.ClientCheckRequest,
	expectation bool,
	modelID *string,
) ModelTestCheckSingleResult {
	response, err := fgaClient.Check(context.Background()).
		Body(checkRequest).
		Options(client.ClientCheckOptions{AuthorizationModelId: modelID}).
		Execute()

	result := ModelTestCheckSingleResult{
		Request:  checkRequest,
		Expected: expectation,
		Error:    err,
	}

	if response != nil {
		// result.Response = *response
		result.Got = response.Allowed
		result.TestResult = result.IsPassing()
	}

	return result
}

func RunCheckTest(
	fgaClient *client.OpenFgaClient,
	checkTest ModelTestCheck,
	tuples []client.ClientTupleKey,
	modelID *string,
) []ModelTestCheckSingleResult {
	results := []ModelTestCheckSingleResult{}

	for relation, expectation := range checkTest.Assertions {
		result := RunSingleCheckTest(fgaClient,
			client.ClientCheckRequest{
				User:             checkTest.User,
				Relation:         relation,
				Object:           checkTest.Object,
				ContextualTuples: &tuples,
			},
			expectation,
			modelID,
		)
		results = append(results, result)
	}

	return results
}

func RunSingleListObjectsTest(
	fgaClient *client.OpenFgaClient,
	listObjectsRequest client.ClientListObjectsRequest,
	expectation []string,
	modelID *string,
) ModelTestListObjectsSingleResult {
	response, err := fgaClient.ListObjects(context.Background()).
		Body(listObjectsRequest).
		Options(client.ClientListObjectsOptions{AuthorizationModelId: modelID}).
		Execute()

	result := ModelTestListObjectsSingleResult{
		Request:  listObjectsRequest,
		Expected: expectation,
		Error:    err,
	}

	if response != nil {
		// result.Response = *response
		result.Got = response.Objects
		result.TestResult = result.IsPassing()
	}

	return result
}

func RunListObjectsTest(
	fgaClient *client.OpenFgaClient,
	listObjectsTest ModelTestListObjects,
	tuples []client.ClientTupleKey,
	modelID *string,
) []ModelTestListObjectsSingleResult {
	results := []ModelTestListObjectsSingleResult{}

	for relation, expectation := range listObjectsTest.Assertions {
		result := RunSingleListObjectsTest(fgaClient,
			client.ClientListObjectsRequest{
				User:             listObjectsTest.User,
				Type:             listObjectsTest.Type,
				Relation:         relation,
				ContextualTuples: &tuples,
			},
			expectation,
			modelID,
		)
		results = append(results, result)
	}

	return results
}

func RunTest(fgaClient *client.OpenFgaClient, test ModelTest, modelID *string) TestResult {
	checkResults := []ModelTestCheckSingleResult{}

	for index := 0; index < len(test.Check); index++ {
		results := RunCheckTest(fgaClient, test.Check[index], test.Tuples, modelID)
		checkResults = append(checkResults, results...)
	}

	listObjectResults := []ModelTestListObjectsSingleResult{}

	for index := 0; index < len(test.ListObjects); index++ {
		results := RunListObjectsTest(fgaClient, test.ListObjects[index], test.Tuples, modelID)
		listObjectResults = append(listObjectResults, results...)
	}

	return TestResult{
		Name:               test.Name,
		Description:        test.Description,
		CheckResults:       checkResults,
		ListObjectsResults: listObjectResults,
	}
}

func RunTests(fgaClient *client.OpenFgaClient, tests []ModelTest, modelID *string) []TestResult {
	results := []TestResult{}

	for index := 0; index < len(tests); index++ {
		result := RunTest(fgaClient, tests[index], modelID)
		results = append(results, result)
	}

	return results
}
