package authorizationmodel

import (
	"context"
	"fmt"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/clierrors"
	"github.com/openfga/cli/internal/fga"
)

// ReadFromStore reads the model from the store with the given the AuthorizationModelID,
// or the latest model if no ID was provided.
func ReadFromStore(clientConfig fga.ClientConfig, fgaClient client.SdkClient) (*openfga.ReadAuthorizationModelResponse, error) { //nolint:lll
	authorizationModelID := clientConfig.AuthorizationModelID

	var (
		err   error
		model *openfga.ReadAuthorizationModelResponse
	)

	if authorizationModelID != "" {
		options := client.ClientReadAuthorizationModelOptions{
			AuthorizationModelId: openfga.PtrString(authorizationModelID),
		}
		model, err = fgaClient.ReadAuthorizationModel(context.Background()).Options(options).Execute()
	} else {
		options := client.ClientReadLatestAuthorizationModelOptions{}
		model, err = fgaClient.ReadLatestAuthorizationModel(context.Background()).Options(options).Execute()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get model %v due to %w", clientConfig.AuthorizationModelID, err)
	}

	if model.AuthorizationModel == nil {
		// If there is no model, try to get the store
		if _, err := fgaClient.GetStore(context.Background()).Execute(); err != nil {
			return nil, fmt.Errorf("failed to get model %v due to %w", clientConfig.AuthorizationModelID, err)
		}

		return nil, fmt.Errorf("%w", clierrors.ErrAuthorizationModelNotFound)
	}

	return model, nil
}
