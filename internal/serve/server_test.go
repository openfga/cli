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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openfga/cli/internal/serve"
)

// newTestServer creates a ProxyServer backed by a temp config file with no
// session token (for backward-compatible CRUD/proxy tests).
func newTestServer(t *testing.T) (*serve.ProxyServer, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cs := serve.NewConfigStoreAt(path)
	srv := serve.NewServerWithStore("", 0, "", cs)
	return srv, path
}

// newTestServerWithToken creates a ProxyServer with session-token auth enabled.
func newTestServerWithToken(t *testing.T, token string) *serve.ProxyServer {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cs := serve.NewConfigStoreAt(path)
	return serve.NewServerWithStore("", 0, token, cs)
}

func doRequest(t *testing.T, handler http.Handler, method, path string, body any) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec.Result()
}

func decodeJSON(t *testing.T, r *http.Response, v any) {
	t.Helper()
	defer r.Body.Close()
	require.NoError(t, json.NewDecoder(r.Body).Decode(v))
}

// ---- health check ----

func TestHealthz(t *testing.T) {
	srv, _ := newTestServer(t)
	resp := doRequest(t, srv.Handler(), http.MethodGet, "/healthz", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ---- origin validation ----

func TestOriginValidation_LocalAllowed(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOriginValidation_RemoteForbidden(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://attacker.example.com")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestOriginValidation_NoOriginAllowed(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	// no Origin header
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOriginValidation_127001Allowed(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---- CORS preflight ----

func TestCORSPreflight(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodOptions, "/servers", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSPreflight_IncludesTokenHeader(t *testing.T) {
	srv := newTestServerWithToken(t, "test-token")
	req := httptest.NewRequest(http.MethodOptions, "/servers", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), serve.SessionTokenHeader)
}

// ---- session token ----

func TestSessionToken_Required(t *testing.T) {
	srv := newTestServerWithToken(t, "my-secret")
	h := srv.Handler()

	// Request without token → 401
	resp := doRequest(t, h, http.MethodGet, "/servers", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSessionToken_ValidHeader(t *testing.T) {
	srv := newTestServerWithToken(t, "my-secret")
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/servers", nil)
	req.Header.Set(serve.SessionTokenHeader, "my-secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSessionToken_ValidQueryParam(t *testing.T) {
	srv := newTestServerWithToken(t, "my-secret")
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/servers?token=my-secret", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSessionToken_WrongToken(t *testing.T) {
	srv := newTestServerWithToken(t, "my-secret")
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/servers", nil)
	req.Header.Set(serve.SessionTokenHeader, "wrong-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSessionToken_HealthzExempt(t *testing.T) {
	srv := newTestServerWithToken(t, "my-secret")
	h := srv.Handler()

	// /healthz should work without a token
	resp := doRequest(t, h, http.MethodGet, "/healthz", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSessionToken_Disabled(t *testing.T) {
	// Empty token = disabled, all requests pass through
	srv, _ := newTestServer(t) // no token
	resp := doRequest(t, srv.Handler(), http.MethodGet, "/servers", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGenerateToken_Unique(t *testing.T) {
	t1, err := serve.GenerateToken()
	require.NoError(t, err)
	t2, err := serve.GenerateToken()
	require.NoError(t, err)
	assert.NotEqual(t, t1, t2)
	assert.Len(t, t1, 64) // 32 bytes hex-encoded
}

// ---- APIURL validation ----

func TestCreateServer_InvalidAPIURL(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	tests := []struct {
		name   string
		apiUrl string
	}{
		{"no scheme", "localhost:8080"},
		{"ftp scheme", "ftp://example.com"},
		{"empty host", "http://"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, h, http.MethodPost, "/servers", map[string]any{
				"name":   "Test",
				"apiUrl": tc.apiUrl,
			})
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestCreateServer_ValidAPIURL(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	for _, u := range []string{"http://localhost:8080", "https://api.fga.dev"} {
		resp := doRequest(t, h, http.MethodPost, "/servers", map[string]any{
			"name":   "Test",
			"apiUrl": u,
		})
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}
}

// ---- max body size ----

func TestMaxBodySize(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	// Send a body larger than MaxRequestBodyBytes (10 MB)
	bigBody := strings.NewReader(strings.Repeat("x", 11<<20))
	req := httptest.NewRequest(http.MethodPost, "/servers", bigBody)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	// Should fail — either 400 (bad JSON) or 413 (body too large)
	assert.True(t, rec.Code >= 400)
}

// ---- server CRUD ----

func TestServerCRUD(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	// GET /servers — empty list
	resp := doRequest(t, h, http.MethodGet, "/servers", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var servers []serve.PublicServer
	decodeJSON(t, resp, &servers)
	assert.Empty(t, servers)

	// POST /servers — create
	input := map[string]any{
		"name":   "Local",
		"apiUrl": "http://localhost:8080",
	}
	resp = doRequest(t, h, http.MethodPost, "/servers", input)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created serve.PublicServer
	decodeJSON(t, resp, &created)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "Local", created.Name)
	assert.Equal(t, "http://localhost:8080", created.APIURL)
	assert.Equal(t, serve.AuthMethodNone, created.Auth.Method)
	assert.True(t, created.Capabilities.StoreCRUD)
	assert.True(t, created.Capabilities.StoreList)

	// GET /servers — now has one server
	resp = doRequest(t, h, http.MethodGet, "/servers", nil)
	decodeJSON(t, resp, &servers)
	assert.Len(t, servers, 1)

	// PUT /servers/{id} — update name
	resp = doRequest(t, h, http.MethodPut, "/servers/"+created.ID,
		map[string]any{"name": "Local Dev"})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updated serve.PublicServer
	decodeJSON(t, resp, &updated)
	assert.Equal(t, "Local Dev", updated.Name)

	// DELETE /servers/{id}
	resp = doRequest(t, h, http.MethodDelete, "/servers/"+created.ID, nil)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// GET /servers — empty again
	resp = doRequest(t, h, http.MethodGet, "/servers", nil)
	decodeJSON(t, resp, &servers)
	assert.Empty(t, servers)
}

func TestCreateServer_MissingFields(t *testing.T) {
	srv, _ := newTestServer(t)
	resp := doRequest(t, srv.Handler(), http.MethodPost, "/servers",
		map[string]any{"name": "no-url"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateServer_NotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	resp := doRequest(t, srv.Handler(), http.MethodPut, "/servers/nonexistent",
		map[string]any{"name": "x"})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteServer_NotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	resp := doRequest(t, srv.Handler(), http.MethodDelete, "/servers/nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ---- store CRUD ----

func TestStoreCRUD(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	// Create a server first
	resp := doRequest(t, h, http.MethodPost, "/servers",
		map[string]any{"name": "S", "apiUrl": "http://localhost:8080"})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var s serve.PublicServer
	decodeJSON(t, resp, &s)

	// GET /servers/{id}/stores — empty
	resp = doRequest(t, h, http.MethodGet, "/servers/"+s.ID+"/stores", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var stores []serve.PublicStoreEntry
	decodeJSON(t, resp, &stores)
	assert.Empty(t, stores)

	// POST /servers/{id}/stores — add a store
	resp = doRequest(t, h, http.MethodPost, "/servers/"+s.ID+"/stores",
		map[string]any{"storeId": "01STORE", "alias": "my-store"})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var st serve.PublicStoreEntry
	decodeJSON(t, resp, &st)
	assert.Equal(t, "01STORE", st.StoreID)
	assert.Equal(t, "my-store", st.Alias)

	// GET /servers/{id}/stores — now has one
	resp = doRequest(t, h, http.MethodGet, "/servers/"+s.ID+"/stores", nil)
	decodeJSON(t, resp, &stores)
	assert.Len(t, stores, 1)

	// PUT /servers/{id}/stores/{storeId} — update alias
	resp = doRequest(t, h, http.MethodPut, "/servers/"+s.ID+"/stores/01STORE",
		map[string]any{"alias": "renamed"})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var updatedStore serve.PublicStoreEntry
	decodeJSON(t, resp, &updatedStore)
	assert.Equal(t, "renamed", updatedStore.Alias)

	// DELETE /servers/{id}/stores/{storeId}
	resp = doRequest(t, h, http.MethodDelete, "/servers/"+s.ID+"/stores/01STORE", nil)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// GET /servers/{id}/stores — empty again
	resp = doRequest(t, h, http.MethodGet, "/servers/"+s.ID+"/stores", nil)
	decodeJSON(t, resp, &stores)
	assert.Empty(t, stores)
}

func TestAddStore_MissingStoreID(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	resp := doRequest(t, h, http.MethodPost, "/servers", map[string]any{"name": "S", "apiUrl": "http://localhost"})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var s serve.PublicServer
	decodeJSON(t, resp, &s)

	resp = doRequest(t, h, http.MethodPost, "/servers/"+s.ID+"/stores", map[string]any{"alias": "no-id"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestStores_ServerNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	resp := doRequest(t, srv.Handler(), http.MethodGet, "/servers/nope/stores", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ---- secrets redacted ----

func TestSecretsRedacted(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	// Create a server with an API token
	input := map[string]any{
		"name":   "Secure",
		"apiUrl": "http://localhost:8080",
		"auth": map[string]any{
			"method":   "api_token",
			"apiToken": "super-secret-token",
		},
	}
	resp := doRequest(t, h, http.MethodPost, "/servers", input)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// GET /servers must NOT include the token
	resp = doRequest(t, h, http.MethodGet, "/servers", nil)
	body, _ := io.ReadAll(resp.Body)
	assert.NotContains(t, string(body), "super-secret-token")
}

// ---- config file permissions ----

func TestConfigFilePermissions(t *testing.T) {
	srv, configPath := newTestServer(t)
	h := srv.Handler()

	doRequest(t, h, http.MethodPost, "/servers", map[string]any{
		"name":   "Test",
		"apiUrl": "http://localhost:8080",
	})

	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

// ---- proxy ----

func TestProxy_ForwardsRequest(t *testing.T) {
	// Start a fake upstream OpenFGA server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/stores", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stores":[]}`))
	}))
	defer upstream.Close()

	srv, _ := newTestServer(t)
	h := srv.Handler()

	// Create a server pointing at the test upstream
	resp := doRequest(t, h, http.MethodPost, "/servers", map[string]any{
		"name":   "Test",
		"apiUrl": upstream.URL,
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created serve.PublicServer
	decodeJSON(t, resp, &created)

	// Proxy a request
	resp = doRequest(t, h, http.MethodGet, "/servers/"+created.ID+"/proxy/stores", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "stores")
}

func TestProxy_ServerNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	resp := doRequest(t, srv.Handler(), http.MethodGet, "/servers/nope/proxy/stores", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProxy_InjectsAPITokenHeader(t *testing.T) {
	var receivedAuth string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer upstream.Close()

	srv, _ := newTestServer(t)
	h := srv.Handler()

	resp := doRequest(t, h, http.MethodPost, "/servers", map[string]any{
		"name":   "TokenServer",
		"apiUrl": upstream.URL,
		"auth": map[string]any{
			"method":   "api_token",
			"apiToken": "my-test-token",
		},
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var s serve.PublicServer
	decodeJSON(t, resp, &s)

	resp = doRequest(t, h, http.MethodGet, "/servers/"+s.ID+"/proxy/stores", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer my-test-token", receivedAuth)
}

func TestProxy_StoreAuthOverride(t *testing.T) {
	var receivedAuth string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer upstream.Close()

	srv, _ := newTestServer(t)
	h := srv.Handler()

	// Create server with server-level token
	resp := doRequest(t, h, http.MethodPost, "/servers", map[string]any{
		"name":   "S",
		"apiUrl": upstream.URL,
		"auth": map[string]any{
			"method":   "api_token",
			"apiToken": "server-token",
		},
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var s serve.PublicServer
	decodeJSON(t, resp, &s)

	// Add a store with its own token override
	resp = doRequest(t, h, http.MethodPost, "/servers/"+s.ID+"/stores", map[string]any{
		"storeId": "01STORE",
		"auth": map[string]any{
			"method":   "api_token",
			"apiToken": "store-token",
		},
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Proxy a request scoped to that store — should use store token
	resp = doRequest(t, h, http.MethodGet,
		"/servers/"+s.ID+"/proxy/stores/01STORE/authorization-models", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer store-token", receivedAuth)
}
