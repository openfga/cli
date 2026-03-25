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

package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/oklog/ulid/v2"
)

// ServerInput is the JSON body accepted by the create and update server endpoints.
type ServerInput struct {
	Name         string        `json:"name"`
	APIURL       string        `json:"apiUrl"`
	Auth         *AuthInput    `json:"auth,omitempty"`
	Capabilities *Capabilities `json:"capabilities,omitempty"`
}

// AuthInput is the JSON body for auth configuration.
type AuthInput struct {
	Method       AuthMethod `json:"method"`
	APIToken     string     `json:"apiToken"`     //nolint:gosec
	ClientID     string     `json:"clientId"`
	ClientSecret string     `json:"clientSecret"` //nolint:gosec
	TokenIssuer  string     `json:"apiTokenIssuer"`
	Audience     string     `json:"apiAudience"`
}

func (ai *AuthInput) toAuth() Auth {
	if ai == nil {
		return Auth{Method: AuthMethodNone}
	}
	return Auth{
		Method:       ai.Method,
		APIToken:     ai.APIToken,
		ClientID:     ai.ClientID,
		ClientSecret: ai.ClientSecret,
		TokenIssuer:  ai.TokenIssuer,
		Audience:     ai.Audience,
	}
}

// StoreInput is the JSON body accepted by the create and update store endpoints.
type StoreInput struct {
	StoreID string     `json:"storeId"`
	Alias   string     `json:"alias"`
	ModelID string     `json:"modelId"`
	Auth    *AuthInput `json:"auth,omitempty"`
}

// ---- Server handlers ----

// handleServers dispatches GET /servers and POST /servers.
func handleServers(cs *ConfigStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listServers(cs, w)
		case http.MethodPost:
			createServer(cs, w, r)
		default:
			jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleServerByID dispatches PUT /servers/{id} and DELETE /servers/{id}.
func handleServerByID(cs *ConfigStore, tc *tokenCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		switch r.Method {
		case http.MethodPut:
			updateServer(cs, tc, w, r, id)
		case http.MethodDelete:
			deleteServer(cs, tc, w, id)
		default:
			jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func listServers(cs *ConfigStore, w http.ResponseWriter) {
	cfg, err := cs.Read()
	if err != nil {
		jsonError(w, "config unavailable", http.StatusInternalServerError)
		return
	}
	result := make([]PublicServer, len(cfg.Servers))
	for i, s := range cfg.Servers {
		result[i] = s.Public()
	}
	respondJSON(w, result)
}

func createServer(cs *ConfigStore, w http.ResponseWriter, r *http.Request) {
	var input ServerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.Name == "" || input.APIURL == "" {
		jsonError(w, "name and apiUrl are required", http.StatusBadRequest)
		return
	}
	if err := validateAPIURL(input.APIURL); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	caps := DefaultCapabilities()
	if input.Capabilities != nil {
		caps = *input.Capabilities
	}
	auth := input.Auth.toAuth()
	if auth.Method == "" {
		auth.Method = AuthMethodNone
	}

	srv := Server{
		ID:           ulid.Make().String(),
		Name:         input.Name,
		APIURL:       input.APIURL,
		Auth:         auth,
		Stores:       []StoreEntry{},
		Capabilities: caps,
	}

	cfg, err := cs.Read()
	if err != nil {
		jsonError(w, "config unavailable", http.StatusInternalServerError)
		return
	}
	cfg.Servers = append(cfg.Servers, srv)
	if err := cs.Write(cfg); err != nil {
		jsonError(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, srv.Public())
}

func updateServer(cs *ConfigStore, tc *tokenCache, w http.ResponseWriter, r *http.Request, id string) {
	var input ServerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	cfg, err := cs.Read()
	if err != nil {
		jsonError(w, "config unavailable", http.StatusInternalServerError)
		return
	}

	idx := findServerIdx(cfg, id)
	if idx == -1 {
		jsonError(w, "server not found", http.StatusNotFound)
		return
	}

	credentialsChanged := false
	if input.Name != "" {
		cfg.Servers[idx].Name = input.Name
	}
	if input.APIURL != "" {
		cfg.Servers[idx].APIURL = input.APIURL
	}
	if input.Auth != nil {
		a := input.Auth.toAuth()
		if a.Method != "" {
			cfg.Servers[idx].Auth.Method = a.Method
		}
		if a.APIToken != "" {
			cfg.Servers[idx].Auth.APIToken = a.APIToken
			credentialsChanged = true
		}
		if a.ClientID != "" {
			cfg.Servers[idx].Auth.ClientID = a.ClientID
			credentialsChanged = true
		}
		if a.ClientSecret != "" {
			cfg.Servers[idx].Auth.ClientSecret = a.ClientSecret
			credentialsChanged = true
		}
		if a.TokenIssuer != "" {
			cfg.Servers[idx].Auth.TokenIssuer = a.TokenIssuer
		}
		if a.Audience != "" {
			cfg.Servers[idx].Auth.Audience = a.Audience
		}
	}
	if input.Capabilities != nil {
		cfg.Servers[idx].Capabilities = *input.Capabilities
	}

	if err := cs.Write(cfg); err != nil {
		jsonError(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	if credentialsChanged {
		tc.invalidatePrefix(id)
	}

	respondJSON(w, cfg.Servers[idx].Public())
}

func deleteServer(cs *ConfigStore, tc *tokenCache, w http.ResponseWriter, id string) {
	cfg, err := cs.Read()
	if err != nil {
		jsonError(w, "config unavailable", http.StatusInternalServerError)
		return
	}

	idx := findServerIdx(cfg, id)
	if idx == -1 {
		jsonError(w, "server not found", http.StatusNotFound)
		return
	}

	cfg.Servers = append(cfg.Servers[:idx], cfg.Servers[idx+1:]...)
	if err := cs.Write(cfg); err != nil {
		jsonError(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	tc.invalidatePrefix(id)
	w.WriteHeader(http.StatusNoContent)
}

// ---- Store handlers ----

// handleStores dispatches GET /servers/{id}/stores and POST /servers/{id}/stores.
func handleStores(cs *ConfigStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		switch r.Method {
		case http.MethodGet:
			listStores(cs, w, id)
		case http.MethodPost:
			addStore(cs, w, r, id)
		default:
			jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleStoreByID dispatches PUT/DELETE /servers/{id}/stores/{storeId}.
func handleStoreByID(cs *ConfigStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverID := r.PathValue("id")
		storeID := r.PathValue("storeId")
		switch r.Method {
		case http.MethodPut:
			updateStore(cs, w, r, serverID, storeID)
		case http.MethodDelete:
			removeStore(cs, w, serverID, storeID)
		default:
			jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func listStores(cs *ConfigStore, w http.ResponseWriter, serverID string) {
	cfg, err := cs.Read()
	if err != nil {
		jsonError(w, "config unavailable", http.StatusInternalServerError)
		return
	}
	idx := findServerIdx(cfg, serverID)
	if idx == -1 {
		jsonError(w, "server not found", http.StatusNotFound)
		return
	}
	result := make([]PublicStoreEntry, len(cfg.Servers[idx].Stores))
	for i, st := range cfg.Servers[idx].Stores {
		result[i] = st.Public()
	}
	respondJSON(w, result)
}

func addStore(cs *ConfigStore, w http.ResponseWriter, r *http.Request, serverID string) {
	var input StoreInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.StoreID == "" {
		jsonError(w, "storeId is required", http.StatusBadRequest)
		return
	}

	cfg, err := cs.Read()
	if err != nil {
		jsonError(w, "config unavailable", http.StatusInternalServerError)
		return
	}
	idx := findServerIdx(cfg, serverID)
	if idx == -1 {
		jsonError(w, "server not found", http.StatusNotFound)
		return
	}

	entry := StoreEntry{
		StoreID: input.StoreID,
		Alias:   input.Alias,
		ModelID: input.ModelID,
	}
	if input.Auth != nil {
		a := input.Auth.toAuth()
		entry.Auth = &a
	}

	cfg.Servers[idx].Stores = append(cfg.Servers[idx].Stores, entry)
	if err := cs.Write(cfg); err != nil {
		jsonError(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, entry.Public())
}

func updateStore(cs *ConfigStore, w http.ResponseWriter, r *http.Request, serverID, storeID string) {
	var input StoreInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	cfg, err := cs.Read()
	if err != nil {
		jsonError(w, "config unavailable", http.StatusInternalServerError)
		return
	}
	sIdx := findServerIdx(cfg, serverID)
	if sIdx == -1 {
		jsonError(w, "server not found", http.StatusNotFound)
		return
	}

	stIdx := findStoreIdx(&cfg.Servers[sIdx], storeID)
	if stIdx == -1 {
		jsonError(w, "store not found", http.StatusNotFound)
		return
	}

	st := &cfg.Servers[sIdx].Stores[stIdx]
	if input.Alias != "" {
		st.Alias = input.Alias
	}
	if input.ModelID != "" {
		st.ModelID = input.ModelID
	}
	if input.Auth != nil {
		a := input.Auth.toAuth()
		st.Auth = &a
	}

	if err := cs.Write(cfg); err != nil {
		jsonError(w, "failed to save config", http.StatusInternalServerError)
		return
	}
	respondJSON(w, st.Public())
}

func removeStore(cs *ConfigStore, w http.ResponseWriter, serverID, storeID string) {
	cfg, err := cs.Read()
	if err != nil {
		jsonError(w, "config unavailable", http.StatusInternalServerError)
		return
	}
	sIdx := findServerIdx(cfg, serverID)
	if sIdx == -1 {
		jsonError(w, "server not found", http.StatusNotFound)
		return
	}

	stIdx := findStoreIdx(&cfg.Servers[sIdx], storeID)
	if stIdx == -1 {
		jsonError(w, "store not found", http.StatusNotFound)
		return
	}

	stores := cfg.Servers[sIdx].Stores
	cfg.Servers[sIdx].Stores = append(stores[:stIdx], stores[stIdx+1:]...)
	if err := cs.Write(cfg); err != nil {
		jsonError(w, "failed to save config", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----

// validateAPIURL checks that the URL has an http/https scheme and a host.
func validateAPIURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid apiUrl: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("apiUrl scheme must be http or https")
	}
	if u.Host == "" {
		return fmt.Errorf("apiUrl must include a host")
	}
	return nil
}

func findServerIdx(cfg *Config, id string) int {
	for i := range cfg.Servers {
		if cfg.Servers[i].ID == id {
			return i
		}
	}
	return -1
}

func findStoreIdx(srv *Server, storeID string) int {
	for i := range srv.Stores {
		if srv.Stores[i].StoreID == storeID {
			return i
		}
	}
	return -1
}

func respondJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
