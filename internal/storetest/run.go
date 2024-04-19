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
	ctxKeyAuthModel  ctxKey = "authModel"
	ctxKeyFgaClient  ctxKey = "fgaClient"
	ctxKeyFgaContext ctxKey = "fgaContext"
)

func RunTests(
	path string,
	fgaClient *client.OpenFgaClient,
	storeData *StoreData,
	format authorizationmodel.ModelFormat,
	reporter string,
) (int, error) {
	authModel, err := getAuthModel(storeData, format)
	if err != nil {
		return 1, err
	}

	isLocalTest := authModel != nil

	ctx := context.WithValue(context.Background(), ctxKeyFgaClient, fgaClient)
	ctx = context.WithValue(ctx, ctxKeyAuthModel, authModel)

	if !isLocalTest {
		// Validate the config for the fga client before running tests
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

func getAuthModel(storeData *StoreData, format authorizationmodel.ModelFormat) (*authorizationmodel.AuthzModel, error) {
	var authModel *authorizationmodel.AuthzModel

	if storeData == nil || storeData.Model == "" {
		return authModel, nil
	}

	tempModel := authorizationmodel.AuthzModel{}

	err := tempModel.ReadModelFromString(storeData.Model, format)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &tempModel, nil
}
