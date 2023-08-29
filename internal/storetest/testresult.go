package storetest

import (
	"fmt"

	"github.com/openfga/cli/internal/comparison"
	"github.com/openfga/go-sdk/client"
)

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
	return result.Error == nil && result.Got != nil && comparison.CheckStringArraysEqual(*result.Got, result.Expected)
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
