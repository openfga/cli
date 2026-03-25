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

package serve_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openfga/cli/internal/serve"
)

func TestConfigStore_EmptyWhenMissing(t *testing.T) {
	cs := serve.NewConfigStoreAt(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	cfg, err := cs.Read()
	require.NoError(t, err)
	assert.Empty(t, cfg.Servers)
}

func TestConfigStore_RoundTrip(t *testing.T) {
	cs := serve.NewConfigStoreAt(filepath.Join(t.TempDir(), "config.yaml"))

	original := &serve.Config{
		Version: 1,
		Servers: []serve.Server{
			{
				ID:     "01TEST",
				Name:   "Local",
				APIURL: "http://localhost:8080",
				Auth: serve.Auth{
					Method: serve.AuthMethodAPIToken,
					APIToken: "secret-on-disk",
				},
				Capabilities: serve.DefaultCapabilities(),
			},
		},
	}

	require.NoError(t, cs.Write(original))

	loaded, err := cs.Read()
	require.NoError(t, err)
	require.Len(t, loaded.Servers, 1)
	assert.Equal(t, "01TEST", loaded.Servers[0].ID)
	assert.Equal(t, "Local", loaded.Servers[0].Name)
	// Secret is preserved in the on-disk config file.
	assert.Equal(t, "secret-on-disk", loaded.Servers[0].Auth.APIToken)
}

func TestConfigStore_FindServerByID(t *testing.T) {
	cs := serve.NewConfigStoreAt(filepath.Join(t.TempDir(), "config.yaml"))
	cfg := &serve.Config{
		Version: 1,
		Servers: []serve.Server{
			{ID: "aaa", Name: "A", APIURL: "http://a", Auth: serve.Auth{Method: serve.AuthMethodNone}, Capabilities: serve.DefaultCapabilities()},
			{ID: "bbb", Name: "B", APIURL: "http://b", Auth: serve.Auth{Method: serve.AuthMethodNone}, Capabilities: serve.DefaultCapabilities()},
		},
	}
	require.NoError(t, cs.Write(cfg))

	s, err := cs.FindServerByID("bbb")
	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, "B", s.Name)

	missing, err := cs.FindServerByID("zzz")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestServer_SecretsRedactedInPublic(t *testing.T) {
	s := serve.Server{
		ID:     "test-id",
		Name:   "Secure",
		APIURL: "http://localhost",
		Auth: serve.Auth{
			Method: serve.AuthMethodClientCredentials,
			APIToken:     "tok-do-not-leak",
			ClientID:     "cid-do-not-leak",
			ClientSecret: "cs-do-not-leak",
		},
		Capabilities: serve.DefaultCapabilities(),
	}

	pub := s.Public()
	assert.Equal(t, "test-id", pub.ID)
	assert.Equal(t, serve.AuthMethodClientCredentials, pub.Auth.Method)

	// JSON serialisation of the public view must not contain secret values.
	data, err := json.Marshal(pub)
	require.NoError(t, err)
	str := string(data)
	assert.NotContains(t, str, "tok-do-not-leak")
	assert.NotContains(t, str, "cid-do-not-leak")
	assert.NotContains(t, str, "cs-do-not-leak")
}

func TestServer_ResolvedAuth_NoStores(t *testing.T) {
	s := &serve.Server{
		Auth: serve.Auth{Method: serve.AuthMethodAPIToken, APIToken: "server-token"},
	}
	got := s.ResolvedAuth("")
	assert.Equal(t, "server-token", got.APIToken)
}

func TestServer_ResolvedAuth_StoreOverride(t *testing.T) {
	storeAuth := serve.Auth{ClientID: "store-cid", ClientSecret: "store-cs"}
	s := &serve.Server{
		Auth: serve.Auth{
			Method: serve.AuthMethodClientCredentials,
			TokenIssuer: "https://issuer.example.com",
			Audience:    "https://api.example.com",
		},
		Stores: []serve.StoreEntry{
			{StoreID: "01STORE", Auth: &storeAuth},
		},
	}
	got := s.ResolvedAuth("01STORE")
	// Inherits issuer and audience from server, overrides client credentials from store.
	assert.Equal(t, serve.AuthMethodClientCredentials, got.Method)
	assert.Equal(t, "https://issuer.example.com", got.TokenIssuer)
	assert.Equal(t, "store-cid", got.ClientID)
	assert.Equal(t, "store-cs", got.ClientSecret)
}

func TestServer_ResolvedAuth_UnknownStore(t *testing.T) {
	s := &serve.Server{
		Auth: serve.Auth{Method: serve.AuthMethodAPIToken, APIToken: "server-token"},
		Stores: []serve.StoreEntry{
			{StoreID: "01STORE"},
		},
	}
	got := s.ResolvedAuth("OTHER")
	assert.Equal(t, "server-token", got.APIToken)
}
