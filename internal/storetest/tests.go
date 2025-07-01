package storetest

import (
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"

	"github.com/openfga/cli/internal/authorizationmodel"
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
	globalTuples []client.ClientContextualTupleKey,
	model *authorizationmodel.AuthzModel,
) (TestResult, error) {
	testTuples := append(append([]client.ClientContextualTupleKey{}, globalTuples...), test.Tuples...)

	if model == nil {
		return RunRemoteTest(fgaClient, test, testTuples), nil
	}

	return RunLocalTest(fgaServer, test, testTuples, model)
}

func RunTests(
	fgaClient *client.OpenFgaClient,
	storeData *StoreData,
	format authorizationmodel.ModelFormat,
) (TestResults, error) {
	testResults := TestResults{}

	if err := storeData.Validate(); err != nil {
		return testResults, err
	}

	fgaServer, authModel, stopServerFn, err := getLocalServerModelAndTuples(storeData, format)
	if err != nil {
		return testResults, err
	}

	defer stopServerFn()

	for _, test := range storeData.Tests {
		result, err := RunTest(
			fgaClient,
			fgaServer,
			test,
			storeData.Tuples,
			authModel,
		)
		if err != nil {
			return testResults, err
		}

		testResults.Results = append(testResults.Results, result)
	}

	return testResults, nil
}
