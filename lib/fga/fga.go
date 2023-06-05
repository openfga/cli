package fga

import (
	openfga "github.com/openfga/go-sdk"
	. "github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
	"net/url"
)

type FgaClientConfig struct {
	ServerUrl                       string `json:"server_url,omitempty"`
	StoreId                         string `json:"store_id,omitempty"`
	AuthorizationModelId            string `json:"authorization_model_id,omitempty"`
	ApiToken                        string `json:"api_token,omitempty"`
	ClientCredentialsApiTokenIssuer string `json:"api_token_issuer,omitempty"`
	ClientCredentialsApiAudience    string `json:"api_audience,omitempty"`
	ClientCredentialsClientId       string `json:"client_id,omitempty"`
	ClientCredentialsClientSecret   string `json:"client_secret,omitempty"`
}

func (c FgaClientConfig) getCredentials() *credentials.Credentials {
	if c.ApiToken != "" {
		return &credentials.Credentials{
			Method: credentials.CredentialsMethodApiToken,
			Config: &credentials.Config{
				ApiToken: c.ApiToken,
			},
		}
	}

	if c.ClientCredentialsClientId != "" {
		return &credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       c.ClientCredentialsClientId,
				ClientCredentialsClientSecret:   c.ClientCredentialsClientSecret,
				ClientCredentialsApiAudience:    c.ClientCredentialsApiAudience,
				ClientCredentialsApiTokenIssuer: c.ClientCredentialsApiTokenIssuer,
			},
		}
	}

	return &credentials.Credentials{
		Method: credentials.CredentialsMethodNone,
	}
}

func (c FgaClientConfig) getApiUriParts() (*url.URL, error) {
	return url.Parse(c.ServerUrl)
}

func (c FgaClientConfig) getClientConfig() (*ClientConfiguration, error) {
	apiUriParts, err := c.getApiUriParts()
	if err != nil {
		return nil, err
	}

	return &ClientConfiguration{
		ApiScheme:            apiUriParts.Scheme,
		ApiHost:              apiUriParts.Host,
		StoreId:              c.StoreId,
		AuthorizationModelId: openfga.PtrString(c.AuthorizationModelId),
		Credentials:          c.getCredentials(),
	}, nil
}

func (c FgaClientConfig) GetFgaClient() (*OpenFgaClient, error) {
	config, err := c.getClientConfig()
	if err != nil {
		return nil, err
	}

	fgaClient, err := NewSdkClient(config)
	if err != nil {
		return nil, err
	}

	return fgaClient, nil
}
