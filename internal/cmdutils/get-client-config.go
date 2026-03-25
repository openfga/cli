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
	"github.com/openfga/cli/internal/serve"
)

// GetClientConfig resolves an fga.ClientConfig for the given command.
//
// If the --server flag is set, the named server is loaded from
// $XDG_CONFIG_HOME/fga/servers.yaml and used as the base configuration.
// If --store-id is also set, any auth overrides from that store entry are merged.
// Any other flags that are explicitly set override the server/store values.
//
// If --server is not set, behaviour is unchanged (flags and env vars only).
func GetClientConfig(cmd *cobra.Command) fga.ClientConfig {
	if serverID, _ := cmd.Flags().GetString("server"); serverID != "" {
		cfg, err := loadServerConfig(cmd, serverID)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: could not load server %q: %v\n", serverID, err)
			// Fall through to flag-based config.
		} else {
			return cfg
		}
	}

	return buildConfigFromFlags(cmd)
}

// loadServerConfig builds a ClientConfig starting from the named server,
// merges any store-level auth overrides (if --store-id is set), then applies
// any explicitly-set command flags on top.
func loadServerConfig(cmd *cobra.Command, serverID string) (fga.ClientConfig, error) {
	cs, err := serve.NewConfigStore()
	if err != nil {
		return fga.ClientConfig{}, err
	}

	srv, err := cs.FindServerByID(serverID)
	if err != nil {
		return fga.ClientConfig{}, err
	}
	if srv == nil {
		return fga.ClientConfig{}, fmt.Errorf("server %q not found", serverID) //nolint:err113
	}

	// Resolve auth: start with server base, apply store override if --store-id is set.
	storeID, _ := cmd.Flags().GetString("store-id")
	auth := srv.ResolvedAuth(storeID)

	cfg := fga.ClientConfig{
		ApiUrl:         srv.APIURL,
		StoreID:        storeID,
		APIToken:       auth.APIToken,
		APITokenIssuer: auth.TokenIssuer,
		APIAudience:    auth.Audience,
		ClientID:       auth.ClientID,
		ClientSecret:   auth.ClientSecret,
	}

	// Command-line flags override server/store values when explicitly provided.
	if cmd.Flags().Changed("api-url") {
		cfg.ApiUrl, _ = cmd.Flags().GetString("api-url")
	}
	if cmd.Flags().Changed("store-id") {
		cfg.StoreID, _ = cmd.Flags().GetString("store-id")
	}
	if cmd.Flags().Changed("model-id") {
		cfg.AuthorizationModelID, _ = cmd.Flags().GetString("model-id")
	}
	if cmd.Flags().Changed("api-token") {
		cfg.APIToken, _ = cmd.Flags().GetString("api-token")
	}
	if cmd.Flags().Changed("api-token-issuer") {
		cfg.APITokenIssuer, _ = cmd.Flags().GetString("api-token-issuer")
	}
	if cmd.Flags().Changed("api-audience") {
		cfg.APIAudience, _ = cmd.Flags().GetString("api-audience")
	}
	if cmd.Flags().Changed("client-id") {
		cfg.ClientID, _ = cmd.Flags().GetString("client-id")
	}
	if cmd.Flags().Changed("client-secret") {
		cfg.ClientSecret, _ = cmd.Flags().GetString("client-secret")
	}
	if cmd.Flags().Changed("api-scopes") {
		cfg.APIScopes, _ = cmd.Flags().GetStringArray("api-scopes")
	}
	if cmd.Flags().Changed("debug") {
		cfg.Debug, _ = cmd.Flags().GetBool("debug")
	}

	return cfg, nil
}

// buildConfigFromFlags builds a ClientConfig purely from command flags (original behaviour).
func buildConfigFromFlags(cmd *cobra.Command) fga.ClientConfig {
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
	customHeaders, _ := cmd.Flags().GetStringArray("custom-headers")
	debug, _ := cmd.Flags().GetBool("debug")

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
		CustomHeaders:        customHeaders,
		Debug:                debug,
	}
}
