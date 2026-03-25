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

// Package serve implements the fga serve proxy server and its configuration.
package serve

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	configVersion  = 1
	configDirName  = "fga"
	configFileName = "servers.yaml"
)

// AuthMethod identifies how the proxy authenticates to the upstream OpenFGA server.
// Values match the CredentialsMethod enum in @openfga/sdk.
type AuthMethod string

const (
	// AuthMethodNone sends no authentication headers.
	AuthMethodNone AuthMethod = "none"
	// AuthMethodAPIToken sends a static bearer token.
	AuthMethodAPIToken AuthMethod = "api_token"
	// AuthMethodClientCredentials performs the OAuth2 client credentials flow.
	AuthMethodClientCredentials AuthMethod = "client_credentials"
)

// Auth holds authentication configuration for a server or store.
// Secret fields are stored on disk and never included in API responses.
type Auth struct {
	Method       AuthMethod `yaml:"method,omitempty"           json:"method,omitempty"`
	APIToken     string     `yaml:"api_token,omitempty"        json:"-"` //nolint:gosec
	ClientID     string     `yaml:"client_id,omitempty"        json:"-"`
	ClientSecret string     `yaml:"client_secret,omitempty"    json:"-"` //nolint:gosec
	TokenIssuer  string     `yaml:"api_token_issuer,omitempty" json:"apiTokenIssuer,omitempty"`
	Audience     string     `yaml:"api_audience,omitempty"     json:"apiAudience,omitempty"`
}

// PublicAuth is an Auth safe for API responses — all secret fields omitted.
type PublicAuth struct {
	Method      AuthMethod `json:"method,omitempty"`
	TokenIssuer string     `json:"apiTokenIssuer,omitempty"`
	Audience    string     `json:"apiAudience,omitempty"`
}

func (a Auth) Public() PublicAuth {
	return PublicAuth{
		Method:      a.Method,
		TokenIssuer: a.TokenIssuer,
		Audience:    a.Audience,
	}
}

// mergeAuth returns a copy of base with any non-zero fields from override applied.
// This allows a store entry to override specific credential fields (e.g. client_id,
// client_secret) while inheriting the rest from the server auth.
func mergeAuth(base, override Auth) Auth {
	result := base
	if override.Method != "" {
		result.Method = override.Method
	}
	if override.APIToken != "" {
		result.APIToken = override.APIToken
	}
	if override.ClientID != "" {
		result.ClientID = override.ClientID
	}
	if override.ClientSecret != "" {
		result.ClientSecret = override.ClientSecret
	}
	if override.TokenIssuer != "" {
		result.TokenIssuer = override.TokenIssuer
	}
	if override.Audience != "" {
		result.Audience = override.Audience
	}
	return result
}

// Capabilities describes what the upstream server supports.
// Defaults to all-true (open-source OpenFGA).
type Capabilities struct {
	StoreCRUD bool `yaml:"store_crud"  json:"storeCrud"`
	StoreList bool `yaml:"store_list"  json:"storeList"`
}

// DefaultCapabilities returns capabilities for an open-source OpenFGA instance.
func DefaultCapabilities() Capabilities {
	return Capabilities{StoreCRUD: true, StoreList: true}
}

// StoreEntry represents a known store within a server.
type StoreEntry struct {
	StoreID string `yaml:"store_id"           json:"storeId"`
	Alias   string `yaml:"alias,omitempty"    json:"alias,omitempty"`
	ModelID string `yaml:"model_id,omitempty" json:"modelId,omitempty"`
	// Auth overrides specific fields of the server's auth for this store.
	// Stored on disk, never included in API responses.
	Auth *Auth `yaml:"auth,omitempty" json:"-"`
}

// PublicStoreEntry is a StoreEntry safe for API responses.
type PublicStoreEntry struct {
	StoreID string `json:"storeId"`
	Alias   string `json:"alias,omitempty"`
	ModelID string `json:"modelId,omitempty"`
}

func (s StoreEntry) Public() PublicStoreEntry {
	return PublicStoreEntry{
		StoreID: s.StoreID,
		Alias:   s.Alias,
		ModelID: s.ModelID,
	}
}

// Server represents a connection to an OpenFGA server.
// Secret fields are stored on disk and never included in API responses.
type Server struct {
	ID           string       `yaml:"id"`
	Name         string       `yaml:"name"`
	APIURL       string       `yaml:"api_url"`
	Auth         Auth         `yaml:"auth"`
	Stores       []StoreEntry `yaml:"stores,omitempty"`
	Capabilities Capabilities `yaml:"capabilities"`
}

// PublicServer is a Server safe for API responses — all secret fields omitted.
type PublicServer struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	APIURL       string             `json:"apiUrl"`
	Auth         PublicAuth         `json:"auth"`
	Stores       []PublicStoreEntry `json:"stores,omitempty"`
	Capabilities Capabilities       `json:"capabilities"`
}

// Public returns a secrets-redacted view of the server.
func (s Server) Public() PublicServer {
	stores := make([]PublicStoreEntry, len(s.Stores))
	for i, st := range s.Stores {
		stores[i] = st.Public()
	}
	return PublicServer{
		ID:           s.ID,
		Name:         s.Name,
		APIURL:       s.APIURL,
		Auth:         s.Auth.Public(),
		Stores:       stores,
		Capabilities: s.Capabilities,
	}
}

// ResolvedAuth returns the effective Auth for the given storeID.
// If the store has auth overrides, they are merged onto the server's base auth.
// If storeID is empty or no matching store is found, the server's base auth is returned.
func (s *Server) ResolvedAuth(storeID string) Auth {
	if storeID == "" {
		return s.Auth
	}
	for _, st := range s.Stores {
		if st.StoreID == storeID && st.Auth != nil {
			return mergeAuth(s.Auth, *st.Auth)
		}
	}
	return s.Auth
}

// Config is the top-level structure of $XDG_CONFIG_HOME/fga/servers.yaml.
type Config struct {
	Version int      `yaml:"version"`
	Servers []Server `yaml:"servers"`
}

// ConfigStore manages read/write access to the on-disk config file.
// It is safe for concurrent use.
type ConfigStore struct {
	path string
	mu   sync.RWMutex
}

// NewConfigStore creates a ConfigStore backed by the default config path
// ($XDG_CONFIG_HOME/fga/servers.yaml).
func NewConfigStore() (*ConfigStore, error) {
	path, err := DefaultConfigPath()
	if err != nil {
		return nil, err
	}
	return &ConfigStore{path: path}, nil
}

// NewConfigStoreAt creates a ConfigStore backed by the specified path.
func NewConfigStoreAt(path string) *ConfigStore {
	return &ConfigStore{path: path}
}

// DefaultConfigPath returns $XDG_CONFIG_HOME/fga/servers.yaml (defaults to ~/.config/fga/servers.yaml).
func DefaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}
	return filepath.Join(dir, configDirName, configFileName), nil
}

// Read loads and returns the current config. If the file does not exist,
// an empty config is returned (not an error).
func (cs *ConfigStore) Read() (*Config, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	data, err := os.ReadFile(cs.path)
	if os.IsNotExist(err) {
		return &Config{Version: configVersion, Servers: []Server{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Servers == nil {
		cfg.Servers = []Server{}
	}
	return &cfg, nil
}

// Write atomically persists the config to disk.
// The file is created with permissions 0600; the directory with 0700.
func (cs *ConfigStore) Write(cfg *Config) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(cs.path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write to a temp file then rename for atomicity.
	tmp := cs.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Rename(tmp, cs.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("installing config: %w", err)
	}
	return nil
}

// FindServerByID returns the server with the given ID, or nil if not found.
func (cs *ConfigStore) FindServerByID(id string) (*Server, error) {
	cfg, err := cs.Read()
	if err != nil {
		return nil, err
	}
	for i := range cfg.Servers {
		if cfg.Servers[i].ID == id {
			return &cfg.Servers[i], nil
		}
	}
	return nil, nil
}
