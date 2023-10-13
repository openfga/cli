package storetest

import (
	"context"

	"github.com/openfga/go-sdk/client"
)

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
	tuples []client.ClientContextualTupleKey,
) []ModelTestCheckSingleResult {
	results := []ModelTestCheckSingleResult{}

	for relation, expectation := range checkTest.Assertions {
		result := RunSingleRemoteCheckTest(
			fgaClient,
			client.ClientCheckRequest{
				User:             checkTest.User,
				Relation:         relation,
				Object:           checkTest.Object,
				Context:          checkTest.Context,
				ContextualTuples: tuples,
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
		result.Got = response.GetObjects()
		result.TestResult = result.IsPassing()
	}

	return result
}

func RunRemoteListObjectsTest(
	fgaClient *client.OpenFgaClient,
	listObjectsTest ModelTestListObjects,
	tuples []client.ClientContextualTupleKey,
) []ModelTestListObjectsSingleResult {
	results := []ModelTestListObjectsSingleResult{}

	for relation, expectation := range listObjectsTest.Assertions {
		result := RunSingleRemoteListObjectsTest(fgaClient,
			client.ClientListObjectsRequest{
				User:             listObjectsTest.User,
				Type:             listObjectsTest.Type,
				Relation:         relation,
				Context:          listObjectsTest.Context,
				ContextualTuples: tuples,
			},
			expectation,
		)
		results = append(results, result)
	}

	return results
}

func RunRemoteTest(fgaClient *client.OpenFgaClient, test ModelTest, testTuples []client.ClientContextualTupleKey) TestResult {
	checkResults := []ModelTestCheckSingleResult{}

	for index := 0; index < len(test.Check); index++ {
		results := RunRemoteCheckTest(fgaClient, test.Check[index], testTuples)
		checkResults = append(checkResults, results...)
	}

	listObjectResults := []ModelTestListObjectsSingleResult{}

	for index := 0; index < len(test.ListObjects); index++ {
		results := RunRemoteListObjectsTest(fgaClient, test.ListObjects[index], testTuples)
		listObjectResults = append(listObjectResults, results...)
	}

	return TestResult{
		Name:               test.Name,
		Description:        test.Description,
		CheckResults:       checkResults,
		ListObjectsResults: listObjectResults,
	}
}
