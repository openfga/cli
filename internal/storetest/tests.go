package storetest

import (
	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"
)

type ModelTestOptions struct {
	StoreID *string
	ModelID *string
	Remote  bool
}

func RunTest(
	fgaClient *client.OpenFgaClient,
	fgaServer *server.Server,
	test ModelTest,
	globalTuples []client.ClientWriteRequestTupleKey,
	model *authorizationmodel.AuthzModel,
) (TestResult, error) {
	testTuples := append(append([]client.ClientWriteRequestTupleKey{}, globalTuples...), test.Tuples...)

	if model == nil {
		return RunRemoteTest(fgaClient, test, testTuples), nil
	}

	return RunLocalTest(fgaServer, test, testTuples, model)
}

func RunTests(
	fgaClient *client.OpenFgaClient,
	storeData StoreData,
	basePath string,
) (TestResults, error) {
	test := TestResults{}

	fgaServer, authModel, err := getLocalServerAndModel(storeData, basePath)
	if err != nil {
		return test, err
	}

	for index := 0; index < len(storeData.Tests); index++ {
		result, err := RunTest(
			fgaClient,
			fgaServer,
			storeData.Tests[index],
			storeData.Tuples,
			authModel,
		)
		if err != nil {
			return test, err
		}

		test.Results = append(test.Results, result)
	}

	return test, nil
}
