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

var removeCmd = &cobra.Command{
	Use:     "remove <id>",
	Short:   "Remove a server connection",
	Long:    `Remove a server connection from ~/.config/fga/servers.yaml by its ID.`,
	Args:    cobra.ExactArgs(1),
	Example: `  fga server remove 01ABCDEFGHJKMNPQRSTVWXYZ`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")

		cs, err := serve.NewConfigStore()
		if err != nil {
			return err
		}

		srv, err := cs.FindServerByID(id)
		if err != nil {
			return err
		}
		if srv == nil {
			return fmt.Errorf("server %q not found", id) //nolint:err113
		}

		if !force {
			confirmed, err := confirmation.AskForConfirmation(
				fmt.Sprintf("Delete server %q (%s)?", srv.Name, srv.ID))
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

		cfg.Servers = append(cfg.Servers[:idx], cfg.Servers[idx+1:]...)
		if err := cs.Write(cfg); err != nil {
			return err
		}

		fmt.Printf("Server %q removed.\n", id)
		return nil
	},
}

func init() {
	removeCmd.Flags().Bool("force", false, "Skip confirmation prompt")
}
