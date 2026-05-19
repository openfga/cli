package storetest

import (
	"context"

	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"

	"github.com/openfga/cli/internal/authorizationmodel"
)

type ModelTestOptions struct {
	StoreID *string
	ModelID *string
	Remote  bool
}

// LocalServerConfig holds configuration for the embedded OpenFGA server
// used during local model testing. Additional server options can be added
// here as needed (see https://github.com/openfga/cli/issues/564).
type LocalServerConfig struct {
	MaxTypesPerAuthorizationModel int
}

func RunTest(
	ctx context.Context,
	fgaClient *client.OpenFgaClient,
	fgaServer *server.Server,
	test ModelTest,
	globalTuples []client.ClientContextualTupleKey,
	model *authorizationmodel.AuthzModel,
) (TestResult, error) {
	testTuples := append(append([]client.ClientContextualTupleKey{}, globalTuples...), test.Tuples...)

	if model == nil {
		return RunRemoteTest(ctx, fgaClient, test, testTuples), nil
	}

	return RunLocalTest(ctx, fgaServer, test, testTuples, model)
}

func RunTests(
	ctx context.Context,
	fgaClient *client.OpenFgaClient,
	storeData *StoreData,
	format authorizationmodel.ModelFormat,
	serverConfig LocalServerConfig,
) (TestResults, error) {
	testResults := TestResults{}

	if err := storeData.Validate(); err != nil {
		return testResults, err
	}

	fgaServer, authModel, stopServerFn, err := getLocalServerModelAndTuples(storeData, format, serverConfig)
	if err != nil {
		return testResults, err
	}

	defer stopServerFn()

	for _, test := range storeData.Tests {
		result, err := RunTest(
			ctx,
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
