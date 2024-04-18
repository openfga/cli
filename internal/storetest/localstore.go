package storetest

import (
	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/memory"

	"github.com/openfga/cli/internal/authorizationmodel"
)

func getLocalServerModelAndTuples(
	storeData *StoreData,
	format authorizationmodel.ModelFormat,
) (*server.Server, *authorizationmodel.AuthzModel, func(), error) {
	var fgaServer *server.Server

	var authModel *authorizationmodel.AuthzModel

	stopServerFn := func() {}

	if storeData == nil || storeData.Model == "" {
		return fgaServer, authModel, stopServerFn, nil
	}

	// If we have at least one local test, initialize the local server
	datastore := memory.New()

	fgaServer, err := server.NewServerWithOpts(
		server.WithDatastore(datastore),
	)
	if err != nil {
		return nil, nil, stopServerFn, err //nolint:wrapcheck
	}

	tempModel := authorizationmodel.AuthzModel{}

	err = tempModel.ReadModelFromString(storeData.Model, format)
	if err != nil {
		return nil, nil, stopServerFn, err //nolint:wrapcheck
	}

	authModel = &tempModel

	stopServerFn = func() {
		datastore.Close()
		fgaServer.Close()
	}

	return fgaServer, authModel, stopServerFn, nil
}
