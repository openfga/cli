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
	globalTuples []client.ClientTupleKey,
	model *authorizationmodel.AuthzModel,
	basePath string,
) (TestResult, error) {
	format, err := test.LoadModel(basePath)
	if err != nil {
		return TestResult{}, err
	}

	var authModel *authorizationmodel.AuthzModel

	if test.Model != "" {
		m := authorizationmodel.AuthzModel{}
		if err := m.ReadModelFromString(test.Model, format); err != nil {
			return TestResult{}, err //nolint:wrapcheck
		}

		authModel = &m
	} else if model != nil {
		authModel = model
	}

	testTuples := append(append([]client.ClientTupleKey{}, globalTuples...), test.Tuples...)

	if authModel == nil {
		return RunRemoteTest(fgaClient, test, testTuples), nil
	}

	return RunLocalTest(fgaServer, test, testTuples, authModel)
}

func RunTests(
	fgaClient *client.OpenFgaClient,
	storeData StoreData,
	basePath string,
) ([]TestResult, error) {
	results := []TestResult{}

	fgaServer, authModel, err := getLocalServerAndModel(storeData, basePath)
	if err != nil {
		return results, err
	}

	for index := 0; index < len(storeData.Tests); index++ {
		result, err := RunTest(
			fgaClient,
			fgaServer,
			storeData.Tests[index],
			storeData.Tuples,
			authModel,
			basePath,
		)
		if err != nil {
			return results, err
		}

		results = append(results, result)
	}

	return results, nil
}
