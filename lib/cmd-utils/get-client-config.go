package cmdutils

import (
	"github.com/openfga/fga-cli/lib/fga"
	"github.com/spf13/cobra"
)

func GetClientConfig(cmd *cobra.Command) fga.ClientConfig {
	serverURL, _ := cmd.Flags().GetString("server-url")
	storeID, _ := cmd.Flags().GetString("store-id")
	authorizationModelID, _ := cmd.Flags().GetString("model-id")
	apiToken, _ := cmd.Flags().GetString("api-token")
	clientCredentialsAPITokenIssuer, _ := cmd.Flags().GetString("api-token-issuer")
	clientCredentialsAPIAudience, _ := cmd.Flags().GetString("api-audience")
	clientCredentialsClientID, _ := cmd.Flags().GetString("client-id")
	clientCredentialsClientSecret, _ := cmd.Flags().GetString("client-secret")

	return fga.ClientConfig{
		ServerURL:            serverURL,
		StoreID:              storeID,
		AuthorizationModelID: authorizationModelID,
		APIToken:             apiToken,
		APITokenIssuer:       clientCredentialsAPITokenIssuer,
		APIAudience:          clientCredentialsAPIAudience,
		ClientID:             clientCredentialsClientID,
		ClientSecret:         clientCredentialsClientSecret,
	}
}
