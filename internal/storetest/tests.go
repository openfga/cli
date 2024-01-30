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
	test := TestResults{}

	fgaServer, authModel, stopServerFn, err := getLocalServerModelAndTuples(storeData, format)
	if err != nil {
		return test, err
	}

	defer stopServerFn()

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
