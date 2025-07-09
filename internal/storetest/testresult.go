package storetest

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/comparison"
)

const NoValueString = "N/A"

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

type ModelTestListUsersSingleResult struct {
	Request    client.ClientListUsersRequest `json:"request"`
	Expected   ModelTestListUsersAssertion   `json:"expected"`
	Got        ModelTestListUsersAssertion   `json:"got"`
	Error      error                         `json:"error"`
	TestResult bool                          `json:"test_result"`
}

func (result ModelTestListUsersSingleResult) IsPassing() bool {
	return result.Error == nil &&
		comparison.CheckStringArraysEqual(result.Got.Users, result.Expected.Users)
}

type TestResult struct {
	Name               string                             `json:"name"`
	Description        string                             `json:"description"`
	CheckResults       []ModelTestCheckSingleResult       `json:"check_results"`
	ListObjectsResults []ModelTestListObjectsSingleResult `json:"list_objects_results"`
	ListUsersResults   []ModelTestListUsersSingleResult   `json:"list_users_results"`
}

// IsPassing - indicates whether a Test has succeeded completely or has any failing parts.
func (result TestResult) IsPassing() bool {
	for _, test := range result.CheckResults {
		if !test.IsPassing() {
			return false
		}
	}

	for _, test := range result.ListObjectsResults {
		if !test.IsPassing() {
			return false
		}
	}

	for _, test := range result.ListUsersResults {
		if !test.IsPassing() {
			return false
		}
	}

	return true
}

func (result TestResult) FriendlyFailuresDisplay() string {
	totalCheckCount := len(result.CheckResults)
	failedCheckCount := 0
	totalListObjectsCount := len(result.ListObjectsResults)
	failedListObjectsCount := 0
	totalListUsersCount := len(result.ListUsersResults)
	failedListUsersCount := 0
	checkResultsOutput := ""
	listObjectsResultsOutput := ""
	listUsersResultsOutput := ""

	if totalCheckCount > 0 {
		failedCheckCount, checkResultsOutput = buildCheckTestResults(
			result, failedCheckCount, checkResultsOutput)
	}

	if totalListObjectsCount > 0 {
		failedListObjectsCount, listObjectsResultsOutput = buildListObjectsTestResults(
			result, failedListObjectsCount, listObjectsResultsOutput)
	}

	if totalListUsersCount > 0 {
		failedListUsersCount, listUsersResultsOutput = buildListUsersTestResults(
			result, failedListUsersCount, listUsersResultsOutput)
	}

	if failedCheckCount+failedListObjectsCount+failedListUsersCount != 0 {
		return buildTestResultOutput(
			result,
			totalCheckCount, failedCheckCount,
			totalListObjectsCount, failedListObjectsCount,
			totalListUsersCount, failedListUsersCount,
			checkResultsOutput, listObjectsResultsOutput, listUsersResultsOutput)
	}

	return ""
}

func buildCheckTestResults(
	result TestResult,
	failedCheckCount int,
	checkResultsOutput string,
) (int, string) {
	for _, checkResult := range result.CheckResults {
		if !checkResult.IsPassing() {
			failedCheckCount++

			got := NoValueString
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

	return failedCheckCount, checkResultsOutput
}

func buildListObjectsTestResults(
	result TestResult,
	failedListObjectsCount int,
	listObjectsResultsOutput string,
) (int, string) {
	for _, listObjectsResult := range result.ListObjectsResults {
		if !listObjectsResult.IsPassing() {
			failedListObjectsCount++

			got := NoValueString
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

	return failedListObjectsCount, listObjectsResultsOutput
}

func buildListUsersTestResults(
	result TestResult,
	failedListUsersCount int,
	listUsersResultsOutput string,
) (int, string) {
	for _, listUsersResult := range result.ListUsersResults {
		if !listUsersResult.IsPassing() {
			failedListUsersCount++

			got := NoValueString

			if listUsersResult.Got.Users != nil {
				got = fmt.Sprintf("%+v", listUsersResult.Got)
			}

			userFilter := listUsersResult.Request.UserFilters[0]

			listUsersResultsOutput += fmt.Sprintf(
				"\nⅹ ListUsers(object=%+v,relation=%s,user_filter=%+v",
				listUsersResult.Request.Object,
				listUsersResult.Request.Relation,
				userFilter)

			if listUsersResult.Request.Context != nil {
				listUsersResultsOutput += fmt.Sprintf(", context:%v", listUsersResult.Request.Context)
			}

			listUsersResultsOutput += fmt.Sprintf("): expected=%+v, got=%+v", listUsersResult.Expected, got)

			if listUsersResult.Error != nil {
				listUsersResultsOutput += fmt.Sprintf(", error=%v", listUsersResult.Error)
			}
		}
	}

	return failedListUsersCount, listUsersResultsOutput
}

func buildTestResultOutput(result TestResult, totalCheckCount int, failedCheckCount int, //nolint:cyclop
	totalListObjectsCount int, failedListObjectsCount int,
	totalListUsersCount int, failedListUsersCount int,
	checkResultsOutput string, listObjectsResultsOutput string, listUsersOutput string,
) string {
	testStatus := "FAILING"
	output := fmt.Sprintf("(%s) %s: ", testStatus, result.Name)

	if totalCheckCount > 0 {
		output += fmt.Sprintf("Checks (%d/%d passing)", totalCheckCount-failedCheckCount, totalCheckCount)
	}

	if totalCheckCount > 0 && totalListObjectsCount > 0 {
		output += " | "
	}

	if totalListObjectsCount > 0 {
		output += fmt.Sprintf("ListObjects (%d/%d passing)",
			totalListObjectsCount-failedListObjectsCount, totalListObjectsCount)
	}

	if totalListObjectsCount > 0 && totalListUsersCount > 0 {
		output += " | "
	}

	if totalListUsersCount > 0 {
		output += fmt.Sprintf("ListUsers(%d/%d passing)",
			totalListUsersCount-failedListUsersCount, totalListUsersCount)
	}

	if failedCheckCount > 0 {
		output = fmt.Sprintf("%s%s", output, checkResultsOutput)
	}

	if failedListObjectsCount > 0 {
		output = fmt.Sprintf("%s%s", output, listObjectsResultsOutput)
	}

	if failedListUsersCount > 0 {
		output = fmt.Sprintf("%s%s", output, listUsersOutput)
	}

	return output
}

type TestResults struct {
	Results []TestResult `json:"results"`
}

// IsPassing - indicates whether a Test Suite has succeeded completely or has any failing tests.
func (test TestResults) IsPassing() bool {
	for _, testResult := range test.Results {
		if !testResult.IsPassing() {
			return false
		}
	}

	return true
}

func (test TestResults) FriendlyDisplay() string { //nolint:cyclop
	friendlyResults := []string{}

	for _, testResult := range test.Results {
		if !testResult.IsPassing() {
			friendlyResults = append(friendlyResults, testResult.FriendlyFailuresDisplay())
		}
	}

	failuresText := strings.Join(friendlyResults, "\n---\n")

	totalTestCount := len(test.Results)
	failedTestCount := 0
	totalCheckCount := 0
	failedCheckCount := 0
	totalListObjectsCount := 0
	failedListObjectsCount := 0
	totalListUsersCount := 0
	failedListUsersCount := 0

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

		totalListUsersCount += len(testResult.ListUsersResults)

		for _, listUsersResult := range testResult.ListUsersResults {
			if !listUsersResult.IsPassing() {
				failedListUsersCount++
			}
		}
	}

	summary := failuresText

	if totalTestCount > 0 {
		summary = buildTestSummary(
			failedTestCount, summary, totalTestCount,
			totalCheckCount, failedCheckCount,
			totalListObjectsCount, failedListObjectsCount,
			totalListUsersCount, failedListUsersCount,
		)
	}

	return summary
}

func (test TestResults) FriendlyBody() string {
	fullDisplay := test.FriendlyDisplay()
	// Remove the "# Test Summary #\n" header if present
	if strings.HasPrefix(fullDisplay, "# Test Summary #\n") {
		return strings.TrimPrefix(fullDisplay, "# Test Summary #\n")
	}
	return fullDisplay
}

func buildTestSummary(failedTestCount int, summary string, totalTestCount int,
	totalCheckCount int, failedCheckCount int,
	totalListObjectsCount int, failedListObjectsCount int,
	totalListUsersCount int, failedListUsersCount int,
) string {
	if failedTestCount > 0 {
		summary += "\n---\n"
	}

	summary += fmt.Sprintf("# Test Summary #\nTests %d/%d passing",
		totalTestCount-failedTestCount, totalTestCount)

	if totalCheckCount > 0 {
		summary += fmt.Sprintf("\nChecks %d/%d passing",
			totalCheckCount-failedCheckCount, totalCheckCount)
	}

	if totalListObjectsCount > 0 {
		summary += fmt.Sprintf("\nListObjects %d/%d passing",
			totalListObjectsCount-failedListObjectsCount, totalListObjectsCount)
	}

	if totalListUsersCount > 0 {
		summary += fmt.Sprintf("\nListUsers %d/%d passing",
			totalListUsersCount-failedListUsersCount, totalListUsersCount)
	}

	return summary
}
