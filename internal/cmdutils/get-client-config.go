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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/fga"
)

func GetClientConfig(cmd *cobra.Command) fga.ClientConfig {
	apiURL, _ := cmd.Flags().GetString("api-url")
	if !cmd.Flags().Changed("api-url") && cmd.Flags().Changed("server-url") {
		apiURL, _ = cmd.Flags().GetString("server-url")
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Using 'server-url' value of '%s'. Note that 'server-url' has been deprecated in favour of 'api-url' and may be removed in subsequent releases.\n", //nolint:lll
			apiURL,
		)
	}

	storeID, _ := cmd.Flags().GetString("store-id")
	authorizationModelID, _ := cmd.Flags().GetString("model-id")
	apiToken, _ := cmd.Flags().GetString("api-token")
	clientCredentialsAPITokenIssuer, _ := cmd.Flags().GetString("api-token-issuer")
	clientCredentialsAPIAudience, _ := cmd.Flags().GetString("api-audience")
	clientCredentialsClientID, _ := cmd.Flags().GetString("client-id")
	clientCredentialsClientSecret, _ := cmd.Flags().GetString("client-secret")
	clientCredentialsScopes, _ := cmd.Flags().GetStringArray("api-scopes")

	return fga.ClientConfig{
		ApiUrl:               apiURL,
		StoreID:              storeID,
		AuthorizationModelID: authorizationModelID,
		APIToken:             apiToken,
		APITokenIssuer:       clientCredentialsAPITokenIssuer,
		APIAudience:          clientCredentialsAPIAudience,
		ClientID:             clientCredentialsClientID,
		ClientSecret:         clientCredentialsClientSecret,
		APIScopes:            clientCredentialsScopes,
	}
}
