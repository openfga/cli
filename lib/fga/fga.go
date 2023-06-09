package fga

import (
	"net/url"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
)

type ClientConfig struct {
	ServerURL            string `json:"server_url,omitempty"`
	StoreID              string `json:"store_id,omitempty"`
	AuthorizationModelID string `json:"authorization_model_id,omitempty"`
	APIToken             string `json:"api_token,omitempty"`
	APITokenIssuer       string `json:"api_token_issuer,omitempty"`
	APIAudience          string `json:"api_audience,omitempty"`
	ClientID             string `json:"client_id,omitempty"`
	ClientSecret         string `json:"client_secret,omitempty"`
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
			},
		}
	}

	return &credentials.Credentials{
		Method: credentials.CredentialsMethodNone,
	}
}

func (c ClientConfig) getClientConfig() (*client.ClientConfiguration, error) {
	apiURIParts, err := url.Parse(c.ServerURL)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &client.ClientConfiguration{
		ApiScheme:            apiURIParts.Scheme,
		ApiHost:              apiURIParts.Host,
		StoreId:              c.StoreID,
		AuthorizationModelId: openfga.PtrString(c.AuthorizationModelID),
		Credentials:          c.getCredentials(),
	}, nil
}

func (c ClientConfig) GetFgaClient() (*client.OpenFgaClient, error) {
	config, err := c.getClientConfig()
	if err != nil {
		return nil, err
	}

	fgaClient, err := client.NewSdkClient(config)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return fgaClient, nil
}
