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

// Package server implements the `fga server` sub-commands.
package server

import (
	"github.com/spf13/cobra"
)

// ServerCmd is the parent command for all `fga server` subcommands.
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage OpenFGA server connections",
	Long: `Manage OpenFGA server connections stored in ~/.config/fga/servers.yaml.

Each server entry describes how to connect to an OpenFGA instance: the API URL,
authentication credentials, capability flags, and an optional list of known stores.

The fga serve proxy uses these server entries to route and authenticate API calls.`,
}

// ServerStoreCmd is the parent for `fga server store` subcommands.
var ServerStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Manage stores within a server connection",
	Long: `Manage the list of known stores associated with a server connection.

Each store entry can have an alias for easy reference and optional auth overrides
that are merged with the server's base authentication.`,
}

func mustGetString(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return v
}

func init() {
	ServerCmd.AddCommand(addCmd)
	ServerCmd.AddCommand(listCmd)
	ServerCmd.AddCommand(removeCmd)
	ServerCmd.AddCommand(updateCmd)

	ServerStoreCmd.AddCommand(storeAddCmd)
	ServerStoreCmd.AddCommand(storeListCmd)
	ServerStoreCmd.AddCommand(storeRemoveCmd)
	ServerStoreCmd.AddCommand(storeUpdateCmd)

	ServerCmd.AddCommand(ServerStoreCmd)
}
