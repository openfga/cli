package storetest

import (
	"context"

	pb "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"
)

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
	tuples []client.ClientTupleKey,
	model *authorizationmodel.AuthzModel,
) (TestResult, error) {
	checkResults := []ModelTestCheckSingleResult{}
	listObjectResults := []ModelTestListObjectsSingleResult{}

	storeID, modelID, err := initLocalStore(fgaServer, model.GetProtoModel(), tuples)
	if err != nil {
		return TestResult{}, err
	}

	testOptions := ModelTestOptions{
		StoreID: storeID,
		ModelID: modelID,
	}

	for index := 0; index < len(test.Check); index++ {
		results := RunLocalCheckTest(fgaServer, test.Check[index], tuples, testOptions)
		checkResults = append(checkResults, results...)
	}

	for index := 0; index < len(test.ListObjects); index++ {
		results := RunLocalListObjectsTest(fgaServer, test.ListObjects[index], tuples, testOptions)
		listObjectResults = append(listObjectResults, results...)
	}

	return TestResult{
		Name:               test.Name,
		Description:        test.Description,
		CheckResults:       checkResults,
		ListObjectsResults: listObjectResults,
	}, nil
}
