package storetest

import (
	"context"
	"github.com/oklog/ulid/v2"
	pb "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/memory"
	"math"
)

const writeMaxChunkSize = 40

func initLocalStore(
	fgaServer *server.Server,
	model *pb.AuthorizationModel,
	test ModelTest,
	testTuples []client.ClientTupleKey,
) (*string, *string, error) {
	var modelID *string

	storeID := ulid.Make().String()
	tuples := convertClientTupleKeysToProtoTupleKeys(testTuples)

	var authModelWriteReq *pb.WriteAuthorizationModelRequest

	if test.Model != "" {
		authModel := authorizationmodel.AuthzModel{}
		if err := authModel.ReadFromDSLString(test.Model); err != nil {
			return nil, nil, err //nolint:wrapcheck
		}

		protoModel := authModel.GetProtoModel()
		authModelWriteReq = &pb.WriteAuthorizationModelRequest{
			StoreId:         storeID,
			TypeDefinitions: protoModel.GetTypeDefinitions(),
			SchemaVersion:   protoModel.GetSchemaVersion(),
		}
	} else if model != nil {
		authModelWriteReq = &pb.WriteAuthorizationModelRequest{
			StoreId:         storeID,
			TypeDefinitions: model.GetTypeDefinitions(),
			SchemaVersion:   model.GetSchemaVersion(),
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

			_, err := fgaServer.Write(context.Background(), &pb.WriteRequest{
				StoreId: storeID,
				Writes:  &pb.TupleKeys{TupleKeys: writeChunk},
			})
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

	testLocalityCounts := storeData.GetTestLocalityCount()
	if testLocalityCounts.Local == 0 {
		return fgaServer, authModel, nil
	}

	// If we have at least one local test, initialize the local server
	datastore := memory.New()

	format, err := storeData.LoadModel(basePath)
	if err != nil {
		return nil, nil, err
	}

	fgaServer, err = server.NewServerWithOpts(server.WithDatastore(datastore))
	if err != nil || storeData.Model == "" {
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
