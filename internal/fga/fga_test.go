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

package fga

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfga/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCustomHeaders(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		headers  []string
		expected map[string]string
		err      error
	}{
		{
			name:     "no headers",
			headers:  []string{},
			expected: map[string]string{},
		},
		{
			name:    "single valid header",
			headers: []string{"X-Custom: value1"},
			expected: map[string]string{
				"X-Custom": "value1",
			},
		},
		{
			name:    "multiple valid headers",
			headers: []string{"X-Custom: value1", "X-Request-ID: abc123"},
			expected: map[string]string{
				"X-Custom":     "value1",
				"X-Request-ID": "abc123",
			},
		},
		{
			name:    "colon in value is preserved",
			headers: []string{"X-Custom: host:port"},
			expected: map[string]string{
				"X-Custom": "host:port",
			},
		},
		{
			name:    "whitespace is trimmed",
			headers: []string{"  X-Custom  :  value1  "},
			expected: map[string]string{
				"X-Custom": "value1",
			},
		},
		{
			name:    "empty value is valid",
			headers: []string{"X-Custom: "},
			expected: map[string]string{
				"X-Custom": "",
			},
		},
		{
			name:    "missing colon returns error",
			headers: []string{"nocolon"},
			err:     ErrInvalidHeaderFormat,
		},
		{
			name:    "empty string returns error",
			headers: []string{""},
			err:     ErrInvalidHeaderFormat,
		},
		{
			name:    "empty header name returns error",
			headers: []string{": value"},
			err:     ErrEmptyHeaderName,
		},
		{
			name:    "valid header before invalid stops at first error",
			headers: []string{"X-Good: ok", "bad-header"},
			err:     ErrInvalidHeaderFormat,
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cfg := ClientConfig{CustomHeaders: test.headers}
			result, err := cfg.getCustomHeaders()

			if test.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestCustomHeadersSentInRequest(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name            string
		customHeaders   []string
		expectedHeaders map[string]string
	}{
		{
			name:            "single header is sent",
			customHeaders:   []string{"X-Custom-Header: value1"},
			expectedHeaders: map[string]string{"X-Custom-Header": "value1"},
		},
		{
			name:          "multiple headers are sent",
			customHeaders: []string{"X-Custom-Header: value1", "X-Request-ID: abc123"},
			expectedHeaders: map[string]string{
				"X-Custom-Header": "value1",
				"X-Request-ID":    "abc123",
			},
		},
		{
			name:            "no custom headers",
			customHeaders:   []string{},
			expectedHeaders: map[string]string{},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			headersCh := make(chan http.Header, 1)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				headersCh <- r.Header.Clone()
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"stores": []}`))
			}))
			defer server.Close()

			cfg := ClientConfig{
				ApiUrl:        server.URL,
				StoreID:       "01H0H015178Y2V4CX10C2KGHF4",
				CustomHeaders: test.customHeaders,
			}

			fgaClient, err := cfg.GetFgaClient()
			require.NoError(t, err)

			_, err = fgaClient.ListStores(context.Background()).
				Options(client.ClientListStoresOptions{}).
				Execute()
			require.NoError(t, err)

			capturedHeaders := <-headersCh
			for name, value := range test.expectedHeaders {
				assert.Equal(t, value, capturedHeaders.Get(name),
					"expected header %s to have value %q", name, value)
			}
		})
	}
}
