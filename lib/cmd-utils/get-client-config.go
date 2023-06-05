package cmd_utils

import (
	"github.com/openfga/fga-cli/lib/fga"
	"github.com/spf13/cobra"
)

func GetClientConfig(cmd *cobra.Command, args []string) fga.FgaClientConfig {
	serverUrl, _ := cmd.Flags().GetString("server-url")
	storeId, _ := cmd.Flags().GetString("store-id")
	authorizationModelId, _ := cmd.Flags().GetString("authorization-model-id")
	apiToken, _ := cmd.Flags().GetString("api-token")
	clientCredentialsApiTokenIssuer, _ := cmd.Flags().GetString("api-token-issuer")
	clientCredentialsApiAudience, _ := cmd.Flags().GetString("api-audience")
	clientCredentialsClientId, _ := cmd.Flags().GetString("client-id")
	clientCredentialsClientSecret, _ := cmd.Flags().GetString("client-secret")

	return fga.FgaClientConfig{
		ServerUrl:                       serverUrl,
		StoreId:                         storeId,
		AuthorizationModelId:            authorizationModelId,
		ApiToken:                        apiToken,
		ClientCredentialsApiTokenIssuer: clientCredentialsApiTokenIssuer,
		ClientCredentialsApiAudience:    clientCredentialsApiAudience,
		ClientCredentialsClientId:       clientCredentialsClientId,
		ClientCredentialsClientSecret:   clientCredentialsClientSecret,
	}
}
