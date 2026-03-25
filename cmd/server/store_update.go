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

var storeUpdateCmd = &cobra.Command{
	Use:   "update <server-id> <store-id>",
	Short: "Update a store entry within a server connection",
	Long: `Update the alias, model ID, or auth overrides for a known store entry.

Only the flags you explicitly pass are updated; all other fields are unchanged.`,
	Args: cobra.ExactArgs(2),
	Example: `  # Change alias
  fga server store update 01SERVERID 01STOREID --alias new-alias

  # Set per-store client credentials
  fga server store update 01SERVERID 01STOREID \
    --client-id new-cid --client-secret new-cs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]
		storeID := args[1]

		cs, err := serve.NewConfigStore()
		if err != nil {
			return err
		}

		cfg, err := cs.Read()
		if err != nil {
			return err
		}

		sIdx := -1
		for i := range cfg.Servers {
			if cfg.Servers[i].ID == serverID {
				sIdx = i
				break
			}
		}
		if sIdx == -1 {
			return fmt.Errorf("server %q not found", serverID) //nolint:err113
		}

		stIdx := -1
		for i := range cfg.Servers[sIdx].Stores {
			if cfg.Servers[sIdx].Stores[i].StoreID == storeID {
				stIdx = i
				break
			}
		}
		if stIdx == -1 {
			return fmt.Errorf("store %q not found in server %q", storeID, serverID) //nolint:err113
		}

		st := &cfg.Servers[sIdx].Stores[stIdx]

		if cmd.Flags().Changed("alias") {
			st.Alias, _ = cmd.Flags().GetString("alias")
		}
		if cmd.Flags().Changed("model-id") {
			st.ModelID, _ = cmd.Flags().GetString("model-id")
		}

		// Update auth override if any auth flag changed.
		authFlags := []string{"auth-method", "api-token", "client-id", "client-secret", "api-token-issuer", "api-audience"}
		for _, f := range authFlags {
			if cmd.Flags().Changed(f) {
				if st.Auth == nil {
					st.Auth = &serve.Auth{}
				}
				break
			}
		}
		if st.Auth != nil {
			if cmd.Flags().Changed("auth-method") {
				v, _ := cmd.Flags().GetString("auth-method")
				st.Auth.Method = serve.AuthMethod(v)
			}
			if cmd.Flags().Changed("api-token") {
				st.Auth.APIToken, _ = cmd.Flags().GetString("api-token")
			}
			if cmd.Flags().Changed("client-id") {
				st.Auth.ClientID, _ = cmd.Flags().GetString("client-id")
			}
			if cmd.Flags().Changed("client-secret") {
				st.Auth.ClientSecret, _ = cmd.Flags().GetString("client-secret")
			}
			if cmd.Flags().Changed("api-token-issuer") {
				st.Auth.TokenIssuer, _ = cmd.Flags().GetString("api-token-issuer")
			}
			if cmd.Flags().Changed("api-audience") {
				st.Auth.Audience, _ = cmd.Flags().GetString("api-audience")
			}
		}

		if err := cs.Write(cfg); err != nil {
			return err
		}

		return output.Display(st.Public())
	},
}

func init() {
	storeUpdateCmd.Flags().String("alias", "", "New alias")
	storeUpdateCmd.Flags().String("model-id", "", "New default authorization model ID")
	storeUpdateCmd.Flags().String("auth-method", "", `Auth override method: "none", "api_token", or "client_credentials"`)
	storeUpdateCmd.Flags().String("api-token", "", "API token override")
	storeUpdateCmd.Flags().String("client-id", "", "OAuth2 client ID override")
	storeUpdateCmd.Flags().String("client-secret", "", "OAuth2 client secret override")
	storeUpdateCmd.Flags().String("api-token-issuer", "", "Token issuer URL override")
	storeUpdateCmd.Flags().String("api-audience", "", "OAuth2 audience override")
}
