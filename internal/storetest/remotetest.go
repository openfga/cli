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

	users := GetEffectiveUsers(checkTest)

	for _, user := range users {
		for relation, expectation := range checkTest.Assertions {
			result := RunSingleRemoteCheckTest(
				fgaClient,
				client.ClientCheckRequest{
					User:             user,
					Relation:         relation,
					Object:           checkTest.Object,
					Context:          checkTest.Context,
					ContextualTuples: tuples,
				},
				expectation,
			)
			results = append(results, result)
		}
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

func RunSingleRemoteListUsersTest(
	fgaClient *client.OpenFgaClient,
	listUsersRequest client.ClientListUsersRequest,
	expectation ModelTestListUsersAssertion,
) ModelTestListUsersSingleResult {
	response, err := fgaClient.ListUsers(context.Background()).Body(listUsersRequest).Execute()

	result := ModelTestListUsersSingleResult{
		Request:  listUsersRequest,
		Expected: expectation,
		Error:    err,
	}

	if response != nil {
		result.Got = ModelTestListUsersAssertion{
			Users: convertOpenfgaUsers(response.GetUsers()),
		}
		result.TestResult = result.IsPassing()
	}

	return result
}

func RunRemoteListUsersTest(
	fgaClient *client.OpenFgaClient,
	listUsersTest ModelTestListUsers,
	tuples []client.ClientContextualTupleKey,
) []ModelTestListUsersSingleResult {
	results := []ModelTestListUsersSingleResult{}

	object, _ := convertStoreObjectToObject(listUsersTest.Object)
	for relation, expectation := range listUsersTest.Assertions {
		result := RunSingleRemoteListUsersTest(fgaClient,
			client.ClientListUsersRequest{
				Object:           object,
				Relation:         relation,
				UserFilters:      listUsersTest.UserFilter,
				Context:          listUsersTest.Context,
				ContextualTuples: tuples,
			},
			expectation,
		)

		results = append(results, result)
	}

	return results
}

func RunRemoteTest(
	fgaClient *client.OpenFgaClient,
	test ModelTest,
	testTuples []client.ClientContextualTupleKey,
) TestResult {
	checkResults := []ModelTestCheckSingleResult{}

	for index := range test.Check {
		results := RunRemoteCheckTest(fgaClient, test.Check[index], testTuples)
		checkResults = append(checkResults, results...)
	}

	listObjectResults := []ModelTestListObjectsSingleResult{}

	for index := range test.ListObjects {
		results := RunRemoteListObjectsTest(fgaClient, test.ListObjects[index], testTuples)
		listObjectResults = append(listObjectResults, results...)
	}

	listUserResults := []ModelTestListUsersSingleResult{}

	for index := range test.ListUsers {
		results := RunRemoteListUsersTest(fgaClient, test.ListUsers[index], testTuples)
		listUserResults = append(listUserResults, results...)
	}

	return TestResult{
		Name:               test.Name,
		Description:        test.Description,
		CheckResults:       checkResults,
		ListObjectsResults: listObjectResults,
		ListUsersResults:   listUserResults,
	}
}
