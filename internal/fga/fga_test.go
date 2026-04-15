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
	"errors"
	"testing"

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
				assert.True(t, errors.Is(err, test.err))
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}
