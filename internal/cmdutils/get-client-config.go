/*
Copyright © 2023 OpenFGA

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmdutils

import (
	"github.com/openfga/cli/internal/fga"
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
