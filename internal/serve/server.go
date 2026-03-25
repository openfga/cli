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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	DefaultPort = 8880
	DefaultHost = "localhost"

	// MaxRequestBodyBytes is the maximum request body size (10 MB).
	MaxRequestBodyBytes int64 = 10 << 20

	// SessionTokenHeader is the HTTP header used to authenticate requests
	// when session-token mode is enabled.
	SessionTokenHeader = "X-Serve-Token"
)

// ProxyServer is the fga serve HTTP proxy server. It is API-only — the
// OpenFGA Playground is a standalone web application served separately.
type ProxyServer struct {
	host  string
	port  int
	token string // session token; empty = no token auth
	cs    *ConfigStore
	tc    *tokenCache
}

// NewServer creates a ProxyServer using the default config path.
// If port <= 0 the default port (8880) is used.
// If host is empty, "localhost" is used.
// If token is empty, session-token authentication is disabled.
func NewServer(host string, port int, token string) (*ProxyServer, error) {
	cs, err := NewConfigStore()
	if err != nil {
		return nil, fmt.Errorf("config store: %w", err)
	}
	return NewServerWithStore(host, port, token, cs), nil
}

// NewServerWithStore creates a ProxyServer using the provided ConfigStore.
// Exported for use in tests.
func NewServerWithStore(host string, port int, token string, cs *ConfigStore) *ProxyServer {
	if port <= 0 {
		port = DefaultPort
	}
	if host == "" {
		host = DefaultHost
	}
	return &ProxyServer{host: host, port: port, token: token, cs: cs, tc: newTokenCache()}
}

// GenerateToken creates a cryptographically random 32-byte hex-encoded token.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating session token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ListenAndServe starts the HTTP server on host:port and blocks.
func (s *ProxyServer) ListenAndServe() error {
	handler := s.buildHandler()

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	if s.token != "" {
		log.Printf("fga serve: listening on http://%s (session token required)", addr)
		log.Printf("fga serve: token: %s", s.token)
	} else {
		log.Printf("fga serve: listening on http://%s", addr)
		log.Printf("fga serve: WARNING: session token disabled — any local process can access the API")
	}
	log.Printf("fga serve: config at %s", s.cs.path)
	log.Printf("fga serve: this server is API-only; run the OpenFGA Playground separately and point it here")
	return http.ListenAndServe(addr, handler) //nolint:gosec
}

// Handler returns the HTTP handler (useful for testing without starting a listener).
func (s *ProxyServer) Handler() http.Handler {
	return s.buildHandler()
}

func (s *ProxyServer) buildHandler() http.Handler {
	mux := s.buildMux()
	var handler http.Handler = mux

	// Middleware stack (outermost runs first):
	// 1. maxBody — limit request body size
	// 2. CORS — handle preflight and set response headers
	// 3. origin validation — reject non-localhost browser origins
	// 4. session token — authenticate requests
	if s.token != "" {
		handler = sessionTokenMiddleware(s.token, handler)
	}
	handler = originValidationMiddleware(handler)
	handler = corsMiddleware(s.token != "", handler)
	handler = maxBodyMiddleware(handler)

	return handler
}

func (s *ProxyServer) buildMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Server management
	mux.Handle("/servers", handleServers(s.cs))
	mux.Handle("/servers/{id}", handleServerByID(s.cs, s.tc))

	// Store management within a server
	mux.Handle("/servers/{id}/stores", handleStores(s.cs))
	mux.Handle("/servers/{id}/stores/{storeId}", handleStoreByID(s.cs))

	// Transparent proxy — matches any method and any sub-path
	mux.Handle("/servers/{id}/proxy/{rest...}", handleProxy(s.cs, s.tc))

	// Health / readiness probe (always accessible, even without token)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, map[string]string{"status": "ok"})
	})

	// Root — explain that this is an API-only server.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		respondJSON(w, map[string]string{
			"status":  "ok",
			"service": "fga serve",
			"message": "This is the fga serve API. Run the OpenFGA Playground separately and configure it to use this URL as its backend.",
		})
	})

	return mux
}

// maxBodyMiddleware limits request bodies to MaxRequestBodyBytes to prevent
// memory exhaustion from oversized payloads.
func maxBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodyBytes)
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware attaches CORS response headers to all requests from browser
// origins, and handles preflight OPTIONS requests.
func corsMiddleware(tokenEnabled bool, next http.Handler) http.Handler {
	allowHeaders := "Content-Type, Authorization"
	if tokenEnabled {
		allowHeaders += ", " + SessionTokenHeader
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			w.Header().Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// originValidationMiddleware rejects browser requests whose Origin header
// refers to a non-localhost origin, preventing cross-site request forgery.
//
// Requests without an Origin header (CLI tools, curl, server-to-server) are
// always allowed.
func originValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && !isLocalOrigin(origin) {
			jsonError(w, "forbidden: origin not allowed", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// sessionTokenMiddleware requires a valid session token on all requests
// except /healthz and the root service-info endpoint.
func sessionTokenMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/healthz" || path == "/" {
			next.ServeHTTP(w, r)
			return
		}

		// Accept token from header or query param.
		provided := r.Header.Get(SessionTokenHeader)
		if provided == "" {
			provided = r.URL.Query().Get("token")
		}

		if provided != token {
			jsonError(w, "unauthorized: invalid or missing session token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// isLocalOrigin returns true if the origin refers to localhost / 127.0.0.1 / ::1.
func isLocalOrigin(origin string) bool {
	// Strip scheme (e.g. "http://")
	host := origin
	if idx := strings.Index(origin, "://"); idx != -1 {
		host = origin[idx+3:]
	}
	// Strip port (e.g. ":5173")
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	switch host {
	case "localhost", "127.0.0.1", "[::1]":
		return true
	}
	return false
}
