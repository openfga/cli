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

// Package serve implements the `fga serve` command.
package serve

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/serve"
)

// ServeCmd is the `fga serve` cobra command.
var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the fga serve API proxy for the OpenFGA Playground",
	Long: `Start a local HTTP API server that manages server connection profiles
and acts as an authenticated proxy to upstream OpenFGA servers. Used by the
OpenFGA Playground, which is a separate web application you run independently
and point at this server.

Profile configuration is stored in ~/.config/fga/servers.yaml.

By default, a random session token is generated and printed to stderr. All
requests (except /healthz and the service-info root) must include this token
via the X-Serve-Token header or ?token= query parameter. Use --token to set a
specific token, or --no-token to disable token authentication entirely.

Endpoints:
  GET    /                     — service info (no auth)
  GET    /servers              — list servers (secrets redacted)
  POST   /servers              — create a server
  PUT    /servers/:id          — update a server
  DELETE /servers/:id          — delete a server
  ANY    /servers/:id/proxy/*  — proxy to the server's OpenFGA instance
  GET    /healthz              — health check`,
	Example: `  # Start with auto-generated session token (recommended)
  fga serve

  # Start with a specific token
  fga serve --token my-secret-token

  # Start without token authentication (not recommended)
  fga serve --no-token

  # Bind to a specific host and port
  fga serve --host 127.0.0.1 --port 9000`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		port, _ := cmd.Flags().GetInt("port")
		host, _ := cmd.Flags().GetString("host")
		tokenFlag, _ := cmd.Flags().GetString("token")
		noToken, _ := cmd.Flags().GetBool("no-token")

		var token string
		switch {
		case noToken:
			token = ""
		case tokenFlag != "":
			token = tokenFlag
		default:
			t, err := serve.GenerateToken()
			if err != nil {
				return fmt.Errorf("generating session token: %w", err)
			}
			token = t
		}

		srv, err := serve.NewServer(host, port, token)
		if err != nil {
			return fmt.Errorf("initializing server: %w", err)
		}

		return srv.ListenAndServe()
	},
}

func init() {
	ServeCmd.Flags().Int("port", serve.DefaultPort, "Port to listen on")
	ServeCmd.Flags().String("host", serve.DefaultHost, "Host/IP to bind to")
	ServeCmd.Flags().String("token", "", "Session token (auto-generated if not set)")
	ServeCmd.Flags().Bool("no-token", false, "Disable session token authentication")
}
