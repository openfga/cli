package storetest

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/comparison"
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
	Got        []string                        `json:"got"`
	Error      error                           `json:"error"`
	TestResult bool                            `json:"test_result"`
}

func (result ModelTestListObjectsSingleResult) IsPassing() bool {
	return result.Error == nil && result.Got != nil && comparison.CheckStringArraysEqual(result.Got, result.Expected)
}

type TestResult struct {
	Name               string                             `json:"name"`
	Description        string                             `json:"description"`
	CheckResults       []ModelTestCheckSingleResult       `json:"check_results"`
	ListObjectsResults []ModelTestListObjectsSingleResult `json:"list_objects_results"`
}

// IsPassing - indicates whether a Test has succeeded completely or has any failing parts.
func (result TestResult) IsPassing() bool {
	for index := 0; index < len(result.CheckResults); index++ {
		if !result.CheckResults[index].IsPassing() {
			return false
		}
	}

	for index := 0; index < len(result.ListObjectsResults); index++ {
		if !result.ListObjectsResults[index].IsPassing() {
			return false
		}
	}

	return true
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

			if checkResult.IsPassing() {
				checkResultsOutput = fmt.Sprintf(
					"%s\n✓ Check(user=%s,relation=%s,object=%s, context=%v)",
					checkResultsOutput,
					checkResult.Request.User,
					checkResult.Request.Relation,
					checkResult.Request.Object,
					checkResult.Request.Context,
				)
			} else {
				failedCheckCount++

				got := "N/A"
				if checkResult.Got != nil {
					got = strconv.FormatBool(*checkResult.Got)
				}

				checkResultsOutput = fmt.Sprintf(
					"%s\nⅹ Check(user=%s,relation=%s,object=%s, context=%v): expected=%t, got=%s, error=%v",
					checkResultsOutput,
					checkResult.Request.User,
					checkResult.Request.Relation,
					checkResult.Request.Object,
					checkResult.Request.Context,
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

			if listObjectsResult.IsPassing() {
				listObjectsResultsOutput = fmt.Sprintf(
					"%s\n✓ ListObjects(user=%s,relation=%s,type=%s, context=%v)",
					listObjectsResultsOutput,
					listObjectsResult.Request.User,
					listObjectsResult.Request.Relation,
					listObjectsResult.Request.Type,
					listObjectsResult.Request.Context,
				)
			} else {
				failedListObjectsCount++

				got := "N/A"
				if listObjectsResult.Got != nil {
					got = fmt.Sprintf("%s", listObjectsResult.Got)
				}

				listObjectsResultsOutput = fmt.Sprintf(
					"%s\nⅹ ListObjects(user=%s,relation=%s,type=%s, context=%v): expected=%s, got=%s, error=%v",
					listObjectsResultsOutput,
					listObjectsResult.Request.User,
					listObjectsResult.Request.Relation,
					listObjectsResult.Request.Type,
					listObjectsResult.Request.Context,
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

type TestResults struct {
	Results []TestResult `json:"results"`
}

// IsPassing - indicates whether a Test Suite has succeeded completely or has any failing tests.
func (test TestResults) IsPassing() bool {
	for index := 0; index < len(test.Results); index++ {
		if !test.Results[index].IsPassing() {
			return false
		}
	}

	return true
}

func (test TestResults) FriendlyDisplay() string {
	friendlyResults := []string{}

	for index := 0; index < len(test.Results); index++ {
		friendlyResults = append(friendlyResults, test.Results[index].FriendlyDisplay())
	}

	return strings.Join(friendlyResults, "\n---\n")
}
