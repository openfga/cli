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

	"github.com/openfga/cli/internal/confirmation"
	"github.com/openfga/cli/internal/serve"
)

var storeRemoveCmd = &cobra.Command{
	Use:     "remove <server-id> <store-id>",
	Short:   "Remove a store from a server connection",
	Long:    `Remove a known store entry from a server connection in ~/.config/fga/servers.yaml.`,
	Args:    cobra.ExactArgs(2),
	Example: `  fga server store remove 01SERVERID 01STOREID`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]
		storeID := args[1]
		force, _ := cmd.Flags().GetBool("force")

		cs, err := serve.NewConfigStore()
		if err != nil {
			return err
		}

		srv, err := cs.FindServerByID(serverID)
		if err != nil {
			return err
		}
		if srv == nil {
			return fmt.Errorf("server %q not found", serverID) //nolint:err113
		}

		found := false
		for _, st := range srv.Stores {
			if st.StoreID == storeID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("store %q not found in server %q", storeID, serverID) //nolint:err113
		}

		if !force {
			confirmed, err := confirmation.AskForConfirmation(
				fmt.Sprintf("Remove store %q from server %q?", storeID, serverID))
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
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

		stores := cfg.Servers[sIdx].Stores
		stIdx := -1
		for i := range stores {
			if stores[i].StoreID == storeID {
				stIdx = i
				break
			}
		}
		cfg.Servers[sIdx].Stores = append(stores[:stIdx], stores[stIdx+1:]...)

		if err := cs.Write(cfg); err != nil {
			return err
		}

		fmt.Printf("Store %q removed from server %q.\n", storeID, serverID)
		return nil
	},
}

func init() {
	storeRemoveCmd.Flags().Bool("force", false, "Skip confirmation prompt")
}
