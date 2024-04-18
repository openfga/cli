package storetest

import (
	"context"
	"errors"

	"github.com/cucumber/godog"
	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/authorizationmodel"
)

type ctxKey string

const (
	ctxKeyAuthModel   ctxKey = "authModel"
	ctxKeyFgaClient   ctxKey = "fgaClient"
	ctxKeyFgaServer   ctxKey = "fgaServer"
	ctxKeyIsLocalTest ctxKey = "isLocalTest"
)

func RunCucumberTests(
	path string,
	fgaClient *client.OpenFgaClient,
	storeData *StoreData,
	format authorizationmodel.ModelFormat,
	reporter string,
) (int, error) {
	fgaServer, authModel, stopServerFn, err := getLocalServerModelAndTuples(storeData, format)
	if err != nil {
		return 1, err
	}

	isLocalTest := authModel != nil

	defer stopServerFn()

	ctx := context.WithValue(context.Background(), ctxKeyFgaClient, fgaClient)
	ctx = context.WithValue(ctx, ctxKeyAuthModel, authModel)
	ctx = context.WithValue(ctx, ctxKeyFgaServer, fgaServer)
	ctx = context.WithValue(ctx, ctxKeyIsLocalTest, isLocalTest)

	if !isLocalTest {
		cfg := fgaClient.GetConfig()

		err = cfg.ValidateConfig()
		if err != nil {
			return 1, err //nolint:wrapcheck
		}

		if cfg.StoreId == "" {
			return 1, errors.New("store ID must be provided when running tests remotely") //nolint:goerr113
		}
	}

	opts := &godog.Options{
		Format:         reporter,
		DefaultContext: ctx,
	}

	if path != "" {
		opts.Paths = []string{path}
	} else {
		features, err := storeData.ToFeatures()
		if err != nil {
			return 1, err
		}

		opts.FeatureContents = features
	}

	status := godog.TestSuite{
		Name:                "test",
		ScenarioInitializer: InitializeScenario,
		Options:             opts,
	}.Run()

	return status, nil
}
