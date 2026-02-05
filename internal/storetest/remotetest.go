package storetest

import (
	"context"

	"github.com/openfga/go-sdk/client"
)

func RunSingleRemoteCheckTest(
	ctx context.Context,
	fgaClient *client.OpenFgaClient,
	checkRequest client.ClientCheckRequest,
	expectation bool,
) ModelTestCheckSingleResult {
	res, err := fgaClient.Check(ctx).Body(checkRequest).Execute()

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
	ctx context.Context,
	fgaClient *client.OpenFgaClient,
	checkTest ModelTestCheck,
	tuples []client.ClientContextualTupleKey,
) []ModelTestCheckSingleResult {
	users := GetEffectiveUsers(checkTest)
	objects := GetEffectiveObjects(checkTest)
	results := make([]ModelTestCheckSingleResult, 0, len(users)*len(objects)*len(checkTest.Assertions))

	for _, user := range users {
		for _, object := range objects {
			for relation, expectation := range checkTest.Assertions {
				result := RunSingleRemoteCheckTest(
					ctx,
					fgaClient,
					client.ClientCheckRequest{
						User:             user,
						Relation:         relation,
						Object:           object,
						Context:          checkTest.Context,
						ContextualTuples: tuples,
					},
					expectation,
				)
				results = append(results, result)
			}
		}
	}

	return results
}

func RunSingleRemoteListObjectsTest(
	ctx context.Context,
	fgaClient *client.OpenFgaClient,
	listObjectsRequest client.ClientListObjectsRequest,
	expectation []string,
) ModelTestListObjectsSingleResult {
	response, err := fgaClient.ListObjects(ctx).Body(listObjectsRequest).Execute()

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
	ctx context.Context,
	fgaClient *client.OpenFgaClient,
	listObjectsTest ModelTestListObjects,
	tuples []client.ClientContextualTupleKey,
) []ModelTestListObjectsSingleResult {
	results := make([]ModelTestListObjectsSingleResult, 0, len(listObjectsTest.Assertions))

	for relation, expectation := range listObjectsTest.Assertions {
		result := RunSingleRemoteListObjectsTest(ctx, fgaClient,
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
	ctx context.Context,
	fgaClient *client.OpenFgaClient,
	listUsersRequest client.ClientListUsersRequest,
	expectation ModelTestListUsersAssertion,
) ModelTestListUsersSingleResult {
	response, err := fgaClient.ListUsers(ctx).Body(listUsersRequest).Execute()

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
	ctx context.Context,
	fgaClient *client.OpenFgaClient,
	listUsersTest ModelTestListUsers,
	tuples []client.ClientContextualTupleKey,
) []ModelTestListUsersSingleResult {
	results := make([]ModelTestListUsersSingleResult, 0, len(listUsersTest.Assertions))

	object, _ := convertStoreObjectToObject(listUsersTest.Object)
	for relation, expectation := range listUsersTest.Assertions {
		result := RunSingleRemoteListUsersTest(ctx, fgaClient,
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
	ctx context.Context,
	fgaClient *client.OpenFgaClient,
	test ModelTest,
	testTuples []client.ClientContextualTupleKey,
) TestResult {
	checkResults := make([]ModelTestCheckSingleResult, 0, len(test.Check))

	for index := range test.Check {
		results := RunRemoteCheckTest(ctx, fgaClient, test.Check[index], testTuples)
		checkResults = append(checkResults, results...)
	}

	listObjectResults := make([]ModelTestListObjectsSingleResult, 0, len(test.ListObjects))

	for index := range test.ListObjects {
		results := RunRemoteListObjectsTest(ctx, fgaClient, test.ListObjects[index], testTuples)
		listObjectResults = append(listObjectResults, results...)
	}

	listUserResults := make([]ModelTestListUsersSingleResult, 0, len(test.ListUsers))

	for index := range test.ListUsers {
		results := RunRemoteListUsersTest(ctx, fgaClient, test.ListUsers[index], testTuples)
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
