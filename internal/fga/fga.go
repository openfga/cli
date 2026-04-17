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
	"errors"
	"fmt"
	"strings"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"

	"github.com/openfga/cli/internal/build"
)

const (
	MaxSdkRetry    = 15
	MinSdkWaitInMs = 500
)

var (
	userAgent = "openfga-cli/" + build.Version

	ErrInvalidHeaderFormat = errors.New("expected format \"Header-Name: value\"")
)

type ClientConfig struct {
	ApiUrl               string   `json:"api_url,omitempty"` //nolint:revive,stylecheck
	StoreID              string   `json:"store_id,omitempty"`
	AuthorizationModelID string   `json:"authorization_model_id,omitempty"`
	APIToken             string   `json:"api_token,omitempty"` //nolint:gosec
	APITokenIssuer       string   `json:"api_token_issuer,omitempty"`
	APIAudience          string   `json:"api_audience,omitempty"`
	APIScopes            []string `json:"api_scopes,omitempty"`
	ClientID             string   `json:"client_id,omitempty"`
	ClientSecret         string   `json:"client_secret,omitempty"` //nolint:gosec
	CustomHeaders        []string `json:"custom_headers,omitempty"`
	Debug                bool     `json:"debug,omitempty"`
}

func (c ClientConfig) GetFgaClient() (*client.OpenFgaClient, error) {
	clientConfig, err := c.getClientConfig()
	if err != nil {
		return nil, err
	}

	fgaClient, err := client.NewSdkClient(clientConfig)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return fgaClient, nil
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

func (c ClientConfig) getClientConfig() (*client.ClientConfiguration, error) {
	customHeaders, err := c.getCustomHeaders()
	if err != nil {
		return nil, fmt.Errorf("invalid custom headers configuration: %w", err)
	}

	return &client.ClientConfiguration{
		ApiUrl:               c.ApiUrl,
		StoreId:              c.StoreID,
		AuthorizationModelId: c.AuthorizationModelID,
		Credentials:          c.getCredentials(),
		UserAgent:            userAgent,
		RetryParams: &openfga.RetryParams{
			MaxRetry:    MaxSdkRetry,
			MinWaitInMs: MinSdkWaitInMs,
		},
		Debug:          c.Debug,
		DefaultHeaders: customHeaders,
	}, nil
}

func (c ClientConfig) getCustomHeaders() (map[string]string, error) {
	headers := make(map[string]string, len(c.CustomHeaders))

	for _, header := range c.CustomHeaders {
		name, value, _ := strings.Cut(header, ":")

		name, value = strings.TrimSpace(name), strings.TrimSpace(value)
		if name == "" {
			return nil, fmt.Errorf("invalid custom header %q: %w", header, ErrInvalidHeaderFormat)
		}

		headers[name] = value
	}

	return headers, nil
}
