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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// authClient is a dedicated HTTP client for OAuth2 token exchanges with a
// shorter timeout than the proxy client (token endpoints should respond fast).
var authClient = &http.Client{Timeout: 10 * time.Second}

// tokenCache caches OAuth2 access tokens.
// Cache keys are "serverID" for server-level auth and "serverID:storeID" for
// store-level auth overrides.
type tokenCache struct {
	mu    sync.Mutex
	store map[string]*cachedToken
}

type cachedToken struct {
	value     string
	expiresAt time.Time
}

func newTokenCache() *tokenCache {
	return &tokenCache{store: make(map[string]*cachedToken)}
}

// get returns the cached token for the key, or "" if absent/expired.
func (tc *tokenCache) get(key string) string {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	t, ok := tc.store[key]
	if !ok || time.Now().After(t.expiresAt) {
		return ""
	}
	return t.value
}

// set caches a token, scheduling expiry with a 30-second safety buffer.
func (tc *tokenCache) set(key, token string, expiresInSec int) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	margin := 30
	if expiresInSec <= margin {
		margin = 0
	}
	expiry := time.Now().Add(time.Duration(expiresInSec-margin) * time.Second)
	tc.store[key] = &cachedToken{value: token, expiresAt: expiry}
}

// invalidate removes the cached token for the exact key.
func (tc *tokenCache) invalidate(key string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	delete(tc.store, key)
}

// invalidatePrefix removes all cached tokens whose key starts with prefix.
// Used when a server is updated/deleted to clear all store-level entries too.
func (tc *tokenCache) invalidatePrefix(prefix string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	for k := range tc.store {
		if k == prefix || strings.HasPrefix(k, prefix+":") {
			delete(tc.store, k)
		}
	}
}

type tokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// exchangeClientCredentials performs the OAuth2 client credentials flow and
// returns the access token and its lifetime in seconds.
func exchangeClientCredentials(ctx context.Context, auth Auth) (string, int, error) {
	issuer := strings.TrimRight(auth.TokenIssuer, "/")
	if issuer == "" {
		return "", 0, fmt.Errorf("api_token_issuer is required for client-credentials auth")
	}
	tokenURL := issuer + "/oauth/token"

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {auth.ClientID},
		"client_secret": {auth.ClientSecret},
	}
	if auth.Audience != "" {
		form.Set("audience", auth.Audience)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, fmt.Errorf("building token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := authClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("token exchange failed (HTTP %d)", resp.StatusCode)
	}

	var tr tokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", 0, fmt.Errorf("decoding token response: %w", err)
	}
	if tr.ExpiresIn <= 0 {
		tr.ExpiresIn = 3600
	}
	return tr.AccessToken, tr.ExpiresIn, nil
}

// getAuthHeader resolves the Authorization header value for the given auth config.
// cacheKey is used to namespace token cache entries (e.g. "serverID" or "serverID:storeID").
func getAuthHeader(ctx context.Context, auth Auth, cacheKey string, cache *tokenCache) (string, error) {
	switch auth.Method {
	case AuthMethodAPIToken:
		if auth.APIToken == "" {
			return "", nil
		}
		return "Bearer " + auth.APIToken, nil

	case AuthMethodClientCredentials:
		if token := cache.get(cacheKey); token != "" {
			return "Bearer " + token, nil
		}
		token, expiresIn, err := exchangeClientCredentials(ctx, auth)
		if err != nil {
			return "", err
		}
		cache.set(cacheKey, token, expiresIn)
		return "Bearer " + token, nil

	default: // AuthMethodNone
		return "", nil
	}
}
