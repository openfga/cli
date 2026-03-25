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

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/serve"
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a server connection",
	Long: `Update fields of an existing server connection in ~/.config/fga/servers.yaml.

Only the flags you explicitly pass are updated; all other fields are unchanged.`,
	Args: cobra.ExactArgs(1),
	Example: `  # Change the API URL
  fga server update 01ABCDEF --api-url https://new-api.example.com

  # Rotate an API token
  fga server update 01ABCDEF --api-token new_token

  # Disable store CRUD
  fga server update 01ABCDEF --no-store-crud`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		cs, err := serve.NewConfigStore()
		if err != nil {
			return err
		}

		cfg, err := cs.Read()
		if err != nil {
			return err
		}

		idx := -1
		for i := range cfg.Servers {
			if cfg.Servers[i].ID == id {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("server %q not found", id) //nolint:err113
		}

		s := &cfg.Servers[idx]

		if cmd.Flags().Changed("name") {
			s.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("api-url") {
			s.APIURL, _ = cmd.Flags().GetString("api-url")
		}
		if cmd.Flags().Changed("auth-method") {
			v, _ := cmd.Flags().GetString("auth-method")
			s.Auth.Method = serve.AuthMethod(v)
		}
		if cmd.Flags().Changed("api-token") {
			s.Auth.APIToken, _ = cmd.Flags().GetString("api-token")
		}
		if cmd.Flags().Changed("client-id") {
			s.Auth.ClientID, _ = cmd.Flags().GetString("client-id")
		}
		if cmd.Flags().Changed("client-secret") {
			s.Auth.ClientSecret, _ = cmd.Flags().GetString("client-secret")
		}
		if cmd.Flags().Changed("api-token-issuer") {
			s.Auth.TokenIssuer, _ = cmd.Flags().GetString("api-token-issuer")
		}
		if cmd.Flags().Changed("api-audience") {
			s.Auth.Audience, _ = cmd.Flags().GetString("api-audience")
		}
		if cmd.Flags().Changed("store-crud") {
			s.Capabilities.StoreCRUD, _ = cmd.Flags().GetBool("store-crud")
		}
		if cmd.Flags().Changed("store-list") {
			s.Capabilities.StoreList, _ = cmd.Flags().GetBool("store-list")
		}

		if err := cs.Write(cfg); err != nil {
			return err
		}

		return output.Display(s.Public())
	},
}

func init() {
	updateCmd.Flags().String("name", "", "New server name")
	updateCmd.Flags().String("api-url", "", "New API URL")
	updateCmd.Flags().String("auth-method", "", `Auth method: "none", "api_token", or "client_credentials"`)
	updateCmd.Flags().String("api-token", "", "New API token")
	updateCmd.Flags().String("client-id", "", "New OAuth2 client ID")
	updateCmd.Flags().String("client-secret", "", "New OAuth2 client secret")
	updateCmd.Flags().String("api-token-issuer", "", "New token issuer URL")
	updateCmd.Flags().String("api-audience", "", "New OAuth2 audience")
	updateCmd.Flags().Bool("store-crud", true, "Enable store CRUD operations (CreateStore, GetStore, DeleteStore)")
	updateCmd.Flags().Bool("store-list", true, "Enable listing stores (ListStores)")
}
