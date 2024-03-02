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
func (result TestResult) FriendlyFailuresDisplay() string {
	totalCheckCount := len(result.CheckResults)
	failedCheckCount := 0
	totalListObjectsCount := len(result.ListObjectsResults)
	failedListObjectsCount := 0
	checkResultsOutput := ""
	listObjectsResultsOutput := ""

	if totalCheckCount > 0 {
		for index := 0; index < totalCheckCount; index++ {
			checkResult := result.CheckResults[index]

			if !checkResult.IsPassing() {
				failedCheckCount++

				got := "N/A"
				if checkResult.Got != nil {
					got = strconv.FormatBool(*checkResult.Got)
				}

				checkResultsOutput += fmt.Sprintf(
					"\nⅹ Check(user=%s,relation=%s,object=%s",
					checkResult.Request.User,
					checkResult.Request.Relation,
					checkResult.Request.Object)

				if checkResult.Request.Context != nil {
					checkResultsOutput += fmt.Sprintf(", context:%v", checkResult.Request.Context)
				}

				checkResultsOutput += fmt.Sprintf("): expected=%t, got=%s", checkResult.Expected, got)

				if checkResult.Error != nil {
					checkResultsOutput += fmt.Sprintf(", error=%v", checkResult.Error)
				}
			}
		}
	}

	if totalListObjectsCount > 0 {
		for index := 0; index < totalListObjectsCount; index++ {
			listObjectsResult := result.ListObjectsResults[index]

			if !listObjectsResult.IsPassing() {
				failedListObjectsCount++

				got := "N/A"
				if listObjectsResult.Got != nil {
					got = fmt.Sprintf("%s", listObjectsResult.Got)
				}

				listObjectsResultsOutput += fmt.Sprintf(
					"\nⅹ ListObjects(user=%s,relation=%s,type=%s",
					listObjectsResult.Request.User,
					listObjectsResult.Request.Relation,
					listObjectsResult.Request.Type)

				if listObjectsResult.Request.Context != nil {
					listObjectsResultsOutput += fmt.Sprintf(", context:%v", listObjectsResult.Request.Context)
				}

				listObjectsResultsOutput += fmt.Sprintf("): expected=%s, got=%s", listObjectsResult.Expected, got)

				if listObjectsResult.Error != nil {
					listObjectsResultsOutput += fmt.Sprintf(", error=%v", listObjectsResult.Error)
				}
			}
		}
	}

	if failedCheckCount+failedListObjectsCount != 0 {
		testStatus := "FAILING"
		output := fmt.Sprintf("(%s) %s: ", testStatus, result.Name)

		if totalCheckCount > 0 {
			output += fmt.Sprintf("Checks (%d/%d passing)", totalCheckCount-failedCheckCount, totalCheckCount)
		}

		if totalCheckCount > 0 && totalListObjectsCount > 0 {
			output += " | "
		}

		if totalListObjectsCount > 0 {
			output += fmt.Sprintf("ListObjects (%d/%d passing)", totalListObjectsCount-failedListObjectsCount, totalListObjectsCount)
		}

		if failedCheckCount > 0 {
			output = fmt.Sprintf("%s%s", output, checkResultsOutput)
		}

		if failedListObjectsCount > 0 {
			output = fmt.Sprintf("%s%s", output, listObjectsResultsOutput)
		}

		return output
	}

	return ""
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
		if !test.Results[index].IsPassing() {
			friendlyResults = append(friendlyResults, test.Results[index].FriendlyFailuresDisplay())
		}
	}

	failuresText := strings.Join(friendlyResults, "\n---\n")

	totalTestCount := len(test.Results)
	failedTestCount := 0
	totalCheckCount := 0
	failedCheckCount := 0
	totalListObjectsCount := 0
	failedListObjectsCount := 0

	for _, testResult := range test.Results {
		if !testResult.IsPassing() {
			failedTestCount++
		}

		totalCheckCount += len(testResult.CheckResults)

		for _, checkResult := range testResult.CheckResults {
			if !checkResult.IsPassing() {
				failedCheckCount++
			}
		}

		totalListObjectsCount += len(testResult.ListObjectsResults)

		for _, listObjectsResult := range testResult.ListObjectsResults {
			if !listObjectsResult.IsPassing() {
				failedListObjectsCount++
			}
		}
	}

	summary := failuresText

	if totalTestCount > 0 {
		if failedTestCount > 0 {
			summary += "\n---\n"
		}

		summary += fmt.Sprintf("# Test Summary #\nTests %d/%d passing", totalTestCount-failedTestCount, totalTestCount)

		if totalCheckCount > 0 {
			summary += fmt.Sprintf("\nChecks %d/%d passing", totalCheckCount-failedCheckCount, totalCheckCount)
		}

		if totalListObjectsCount > 0 {
			summary += fmt.Sprintf("\nListObjects %d/%d passing", totalListObjectsCount-failedListObjectsCount, totalListObjectsCount)
		}
	}

	return summary
}
