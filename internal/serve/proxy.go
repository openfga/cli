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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// proxyClient is a dedicated HTTP client for upstream requests with a
// reasonable timeout to prevent goroutine leaks from slow/hung upstreams.
var proxyClient = &http.Client{Timeout: 30 * time.Second}

// handleProxy transparently forwards requests to the upstream OpenFGA server.
// URL pattern: /servers/{id}/proxy/{rest...}
//
// The handler:
//   - Looks up the server by ID
//   - Extracts the OpenFGA store ID from the request path (e.g. /stores/{storeId}/...)
//     and resolves the effective auth (server auth merged with any store auth override)
//   - Constructs the upstream URL by concatenating server.APIURL + rest path
//   - Injects the appropriate Authorization header
//   - Forwards the response (status, headers, body) back to the caller
func handleProxy(cs *ConfigStore, tc *tokenCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverID := r.PathValue("id")
		rest := r.PathValue("rest")

		srv, err := cs.FindServerByID(serverID)
		if err != nil {
			jsonError(w, fmt.Sprintf("config error: %v", err), http.StatusInternalServerError)
			return
		}
		if srv == nil {
			jsonError(w, "server not found", http.StatusNotFound)
			return
		}

		// Extract the OpenFGA store ID from the path (e.g. "stores/{storeId}/...").
		storeID := extractStoreID(rest)

		// Resolve effective auth: server base + any store-level override.
		auth := srv.ResolvedAuth(storeID)

		// Cache key is "serverID" for server-level auth, "serverID:storeID" when
		// a store has credential overrides.
		cacheKey := serverID
		if storeID != "" {
			for _, st := range srv.Stores {
				if st.StoreID == storeID && st.Auth != nil {
					cacheKey = serverID + ":" + storeID
					break
				}
			}
		}

		// Build the upstream URL.
		upstream := strings.TrimRight(srv.APIURL, "/")
		if rest != "" {
			upstream += "/" + strings.TrimLeft(rest, "/")
		}
		if r.URL.RawQuery != "" {
			upstream += "?" + r.URL.RawQuery
		}

		parsedURL, err := url.Parse(upstream)
		if err != nil {
			jsonError(w, "invalid upstream URL", http.StatusInternalServerError)
			return
		}

		proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, parsedURL.String(), r.Body)
		if err != nil {
			jsonError(w, fmt.Sprintf("building proxy request: %v", err), http.StatusInternalServerError)
			return
		}

		// Forward content-type and accept; do NOT forward caller's Authorization
		// (we inject our own below).
		for _, name := range []string{"Content-Type", "Accept"} {
			if v := r.Header.Get(name); v != "" {
				proxyReq.Header.Set(name, v)
			}
		}
		proxyReq.Header.Set("User-Agent", "openfga-playground-proxy/1.0")

		// Inject auth header resolved from the server (+ optional store override).
		authHeader, err := getAuthHeader(r.Context(), auth, cacheKey, tc)
		if err != nil {
			jsonError(w, fmt.Sprintf("auth: %v", err), http.StatusBadGateway)
			return
		}
		if authHeader != "" {
			proxyReq.Header.Set("Authorization", authHeader)
		}

		resp, err := proxyClient.Do(proxyReq)
		if err != nil {
			jsonError(w, fmt.Sprintf("upstream error: %v", err), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Forward response headers, excluding hop-by-hop headers.
		hopByHop := map[string]bool{
			"Connection": true, "Keep-Alive": true, "Transfer-Encoding": true,
			"Te": true, "Trailer": true, "Upgrade": true,
		}
		for name, values := range resp.Header {
			if hopByHop[name] {
				continue
			}
			for _, v := range values {
				w.Header().Add(name, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

// extractStoreID parses an OpenFGA API path and returns the store ID if present.
// OpenFGA paths that include a store ID follow the pattern "stores/{storeId}/...".
func extractStoreID(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	parts := strings.SplitN(trimmed, "/", 3)
	if len(parts) >= 2 && parts[0] == "stores" && parts[1] != "" {
		return parts[1]
	}
	return ""
}
