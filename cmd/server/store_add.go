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

var storeAddCmd = &cobra.Command{
	Use:   "add <server-id>",
	Short: "Add a store to a server connection",
	Long: `Associate a known store with a server connection in ~/.config/fga/servers.yaml.

Optionally provide an alias and auth overrides. Auth fields left blank inherit
from the parent server's authentication configuration.`,
	Args: cobra.ExactArgs(1),
	Example: `  # Add a store with an alias
  fga server store add 01SERVERID --store-id 01STOREID --alias prod-env

  # Add a store with per-store client credentials
  fga server store add 01SERVERID --store-id 01STOREID --alias staging \
    --client-id cid --client-secret cs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

		storeID, _ := cmd.Flags().GetString("store-id")
		if storeID == "" {
			return fmt.Errorf("--store-id is required") //nolint:err113
		}

		entry := serve.StoreEntry{
			StoreID: storeID,
			Alias:   mustGetString(cmd, "alias"),
			ModelID: mustGetString(cmd, "model-id"),
		}

		// Build auth override only if any credential flag was set.
		hasAuth := false
		for _, f := range []string{"auth-method", "api-token", "client-id", "client-secret", "api-token-issuer", "api-audience"} {
			if cmd.Flags().Changed(f) {
				hasAuth = true
				break
			}
		}
		if hasAuth {
			a := serve.Auth{
				Method:       serve.AuthMethod(mustGetString(cmd, "auth-method")),
				APIToken:     mustGetString(cmd, "api-token"),
				ClientID:     mustGetString(cmd, "client-id"),
				ClientSecret: mustGetString(cmd, "client-secret"),
				TokenIssuer:  mustGetString(cmd, "api-token-issuer"),
				Audience:     mustGetString(cmd, "api-audience"),
			}
			entry.Auth = &a
		}

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
			if cfg.Servers[i].ID == serverID {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("server %q not found", serverID) //nolint:err113
		}

		cfg.Servers[idx].Stores = append(cfg.Servers[idx].Stores, entry)
		if err := cs.Write(cfg); err != nil {
			return err
		}

		return output.Display(entry.Public())
	},
}

func init() {
	storeAddCmd.Flags().String("store-id", "", "OpenFGA store ID (required)")
	storeAddCmd.Flags().String("alias", "", "Short alias for this store")
	storeAddCmd.Flags().String("model-id", "", "Default authorization model ID")
	storeAddCmd.Flags().String("auth-method", "", `Auth override method: "none", "api_token", or "client_credentials"`)
	storeAddCmd.Flags().String("api-token", "", "API token override")
	storeAddCmd.Flags().String("client-id", "", "OAuth2 client ID override")
	storeAddCmd.Flags().String("client-secret", "", "OAuth2 client secret override")
	storeAddCmd.Flags().String("api-token-issuer", "", "OAuth2 token issuer URL override")
	storeAddCmd.Flags().String("api-audience", "", "OAuth2 audience override")

	_ = storeAddCmd.MarkFlagRequired("store-id")
}
