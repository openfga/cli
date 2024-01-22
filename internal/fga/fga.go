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

// Package fga handles configuration and setup of the OpenFGA SDK
package fga

import (
	"strings"

	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"

	"github.com/openfga/cli/internal/build"
)

var userAgent = "openfga-cli/" + build.Version

type ClientConfig struct {
	ApiUrl               string   `json:"api_url,omitempty"` //nolint:revive,stylecheck
	StoreID              string   `json:"store_id,omitempty"`
	AuthorizationModelID string   `json:"authorization_model_id,omitempty"`
	APIToken             string   `json:"api_token,omitempty"`
	APITokenIssuer       string   `json:"api_token_issuer,omitempty"`
	APIAudience          string   `json:"api_audience,omitempty"`
	APIScopes            []string `json:"api_scopes,omitempty"`
	ClientID             string   `json:"client_id,omitempty"`
	ClientSecret         string   `json:"client_secret,omitempty"`
}

func (c ClientConfig) getCredentials() *credentials.Credentials {
	if c.APIToken != "" {
		return &credentials.Credentials{
			Method: credentials.CredentialsMethodApiToken,
			Config: &credentials.Config{
				ApiToken: c.APIToken,
			},
		}
	}

	if c.ClientID != "" {
		return &credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       c.ClientID,
				ClientCredentialsClientSecret:   c.ClientSecret,
				ClientCredentialsApiAudience:    c.APIAudience,
				ClientCredentialsApiTokenIssuer: c.APITokenIssuer,
				ClientCredentialsScopes:         strings.Join(c.APIScopes, " "),
			},
		}
	}

	return &credentials.Credentials{
		Method: credentials.CredentialsMethodNone,
	}
}

func (c ClientConfig) getClientConfig() *client.ClientConfiguration {
	return &client.ClientConfiguration{
		ApiUrl:               c.ApiUrl,
		StoreId:              c.StoreID,
		AuthorizationModelId: c.AuthorizationModelID,
		Credentials:          c.getCredentials(),
		UserAgent:            userAgent,
	}
}

func (c ClientConfig) GetFgaClient() (*client.OpenFgaClient, error) {
	fgaClient, err := client.NewSdkClient(c.getClientConfig())
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return fgaClient, nil
}
