package storetest

import (
	"context"
	"math"

	"github.com/oklog/ulid/v2"
	pb "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/memory"
)

const writeMaxChunkSize = 40

func initLocalStore(
	fgaServer *server.Server,
	model *pb.AuthorizationModel,
	testTuples []client.ClientContextualTupleKey,
) (*string, *string, error) {
	var modelID *string

	storeID := ulid.Make().String()

	tuples, err := convertClientTupleKeysToProtoTupleKeys(testTuples)
	if err != nil {
		return nil, nil, err
	}

	var authModelWriteReq *pb.WriteAuthorizationModelRequest

	if model != nil {
		authModelWriteReq = &pb.WriteAuthorizationModelRequest{
			StoreId:         storeID,
			TypeDefinitions: model.GetTypeDefinitions(),
			SchemaVersion:   model.GetSchemaVersion(),
			Conditions:      model.GetConditions(),
		}
	}

	if authModelWriteReq != nil {
		writtenModel, err := fgaServer.WriteAuthorizationModel(context.Background(), authModelWriteReq)
		if err != nil {
			return nil, nil, err //nolint:wrapcheck
		}

		modelID = &writtenModel.AuthorizationModelId
	}

	tuplesLength := len(tuples)
	if tuplesLength > 0 {
		for i := 0; i < tuplesLength; i += writeMaxChunkSize {
			end := int(math.Min(float64(i+writeMaxChunkSize), float64(tuplesLength)))
			writeChunk := (tuples)[i:end]

			writeRequest := &pb.WriteRequest{
				StoreId: storeID,
				Writes:  &pb.WriteRequestTupleKeys{TupleKeys: writeChunk},
			}

			_, err := fgaServer.Write(context.Background(), writeRequest)
			if err != nil {
				return nil, nil, err //nolint:wrapcheck
			}
		}
	}

	return &storeID, modelID, nil
}

func getLocalServerAndModel(
	storeData StoreData,
	basePath string,
) (*server.Server, *authorizationmodel.AuthzModel, error) {
	var fgaServer *server.Server

	var authModel *authorizationmodel.AuthzModel

	format, err := storeData.LoadModel(basePath)
	if err != nil {
		return nil, nil, err
	}

	if storeData.Model == "" {
		return fgaServer, authModel, nil
	}

	// If we have at least one local test, initialize the local server
	datastore := memory.New()

	fgaServer, err = server.NewServerWithOpts(server.WithDatastore(datastore))
	if err != nil {
		return nil, nil, err //nolint:wrapcheck
	}

	tempModel := authorizationmodel.AuthzModel{}
	if format == authorizationmodel.ModelFormatJSON {
		if err := tempModel.ReadFromJSONString(storeData.Model); err != nil {
			return nil, nil, err //nolint:wrapcheck
		}
	} else {
		if err := tempModel.ReadFromDSLString(storeData.Model); err != nil {
			return nil, nil, err //nolint:wrapcheck
		}
	}

	authModel = &tempModel

	return fgaServer, authModel, nil
}
