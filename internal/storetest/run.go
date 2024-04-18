package storetest

import (
	"context"

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

	defer stopServerFn()

	ctx := context.WithValue(context.Background(), ctxKeyFgaClient, fgaClient)
	ctx = context.WithValue(ctx, ctxKeyAuthModel, authModel)
	ctx = context.WithValue(ctx, ctxKeyFgaServer, fgaServer)
	ctx = context.WithValue(ctx, ctxKeyIsLocalTest, authModel != nil)

	opts := &godog.Options{
		Format:         reporter,
		DefaultContext: ctx,
	}

	if path != "" {
		opts.Paths = []string{path}
	} else {
		opts.FeatureContents = storeData.ToFeatures()
	}

	status := godog.TestSuite{
		Name:                "test",
		ScenarioInitializer: InitializeScenario,
		Options:             opts,
	}.Run()

	return status, nil
}
