package authorizationmodel

import (
	"context"
	"fmt"
	"sort"

	"github.com/openfga/go-sdk/client"
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
	Tuples      []client.ClientTupleKey `json:"tuples"`
	Check       []ModelTestCheck        `json:"check"`
	ListObjects []ModelTestListObjects  `json:"list_objects" yaml:"list-objects"` //nolint:tagliatelle
}

func RunSingleCheckTest(
	fgaClient *client.OpenFgaClient,
	checkRequest client.ClientCheckRequest,
	expectation bool,
) ModelTestCheckSingleResult {
	response, err := fgaClient.Check(context.Background()).
		Body(checkRequest).
		Execute()

	result := ModelTestCheckSingleResult{
		Request:  checkRequest,
		Expected: expectation,
		Error:    err,
	}

	if response != nil {
		result.Got = response.Allowed
		result.TestResult = result.IsPassing()
	}

	return result
}

func RunCheckTest(
	fgaClient *client.OpenFgaClient,
	checkTest ModelTestCheck,
	tuples []client.ClientTupleKey,
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
		)
		results = append(results, result)
	}

	return results
}

func RunSingleListObjectsTest(
	fgaClient *client.OpenFgaClient,
	listObjectsRequest client.ClientListObjectsRequest,
	expectation []string,
) ModelTestListObjectsSingleResult {
	response, err := fgaClient.ListObjects(context.Background()).
		Body(listObjectsRequest).
		Execute()

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

func RunListObjectsTest(
	fgaClient *client.OpenFgaClient,
	listObjectsTest ModelTestListObjects,
	tuples []client.ClientTupleKey,
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
		)
		results = append(results, result)
	}

	return results
}

func RunTest(fgaClient *client.OpenFgaClient, test ModelTest) TestResult {
	checkResults := []ModelTestCheckSingleResult{}

	for index := 0; index < len(test.Check); index++ {
		results := RunCheckTest(fgaClient, test.Check[index], test.Tuples)
		checkResults = append(checkResults, results...)
	}

	listObjectResults := []ModelTestListObjectsSingleResult{}

	for index := 0; index < len(test.ListObjects); index++ {
		results := RunListObjectsTest(fgaClient, test.ListObjects[index], test.Tuples)
		listObjectResults = append(listObjectResults, results...)
	}

	return TestResult{
		Name:               test.Name,
		Description:        test.Description,
		CheckResults:       checkResults,
		ListObjectsResults: listObjectResults,
	}
}

func RunTests(fgaClient *client.OpenFgaClient, tests []ModelTest) []TestResult {
	results := []TestResult{}

	for index := 0; index < len(tests); index++ {
		result := RunTest(fgaClient, tests[index])
		results = append(results, result)
	}

	return results
}
