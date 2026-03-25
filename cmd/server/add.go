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

package server

import (
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/serve"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an OpenFGA server connection",
	Long: `Add a new server connection to ~/.config/fga/servers.yaml.

For API token authentication, pass --auth-method api_token --api-token <token>.
For OAuth2 client credentials, pass --auth-method client_credentials with the
appropriate --client-id, --client-secret, --api-token-issuer flags.`,
	Example: `  # No auth (default, for local OpenFGA)
  fga server add --name "Local" --api-url http://localhost:8080

  # API token
  fga server add --name "Staging" --api-url https://api.staging.example.com \
    --auth-method api_token --api-token sk_abc123

  # OAuth2 client credentials (e.g. Auth0 FGA)
  fga server add --name "Production" --api-url https://api.us1.fga.dev \
    --auth-method client_credentials \
    --client-id my-client \
    --client-secret my-secret \
    --api-token-issuer https://auth.fga.dev \
    --api-audience https://api.us1.fga.dev \
    --no-store-crud`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		apiURL, _ := cmd.Flags().GetString("api-url")
		if name == "" || apiURL == "" {
			return fmt.Errorf("--name and --api-url are required") //nolint:err113
		}

		authMethod := serve.AuthMethod(mustGetString(cmd, "auth-method"))
		if authMethod == "" {
			authMethod = serve.AuthMethodNone
		}
		storeCRUD, _ := cmd.Flags().GetBool("store-crud")
		storeList, _ := cmd.Flags().GetBool("store-list")

		srv := serve.Server{
			ID:     ulid.Make().String(),
			Name:   name,
			APIURL: apiURL,
			Auth: serve.Auth{
				Method:       authMethod,
				APIToken:     mustGetString(cmd, "api-token"),
				ClientID:     mustGetString(cmd, "client-id"),
				ClientSecret: mustGetString(cmd, "client-secret"),
				TokenIssuer:  mustGetString(cmd, "api-token-issuer"),
				Audience:     mustGetString(cmd, "api-audience"),
			},
			Capabilities: serve.Capabilities{
				StoreCRUD: storeCRUD,
				StoreList: storeList,
			},
		}

		cs, err := serve.NewConfigStore()
		if err != nil {
			return err
		}

		cfg, err := cs.Read()
		if err != nil {
			return err
		}

		cfg.Servers = append(cfg.Servers, srv)
		if err := cs.Write(cfg); err != nil {
			return err
		}

		return output.Display(srv.Public())
	},
}

func init() {
	addCmd.Flags().String("name", "", "Server name (required)")
	addCmd.Flags().String("api-url", "", "OpenFGA API URL (required)")
	addCmd.Flags().String("auth-method", "none", `Auth method: "none", "api_token", or "client_credentials"`)
	addCmd.Flags().String("api-token", "", "API token (for auth-type api-token)")
	addCmd.Flags().String("client-id", "", "OAuth2 client ID")
	addCmd.Flags().String("client-secret", "", "OAuth2 client secret")
	addCmd.Flags().String("api-token-issuer", "", "OAuth2 token issuer URL")
	addCmd.Flags().String("api-audience", "", "OAuth2 audience")
	addCmd.Flags().Bool("store-crud", true, "Enable store CRUD operations (CreateStore, GetStore, DeleteStore)")
	addCmd.Flags().Bool("store-list", true, "Enable listing stores (ListStores)")

	_ = addCmd.MarkFlagRequired("name")
	_ = addCmd.MarkFlagRequired("api-url")
}
