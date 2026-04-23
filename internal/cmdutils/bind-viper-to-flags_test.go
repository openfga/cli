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

package cmdutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViperValueToStrings(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		value    any
		expected []string
	}{
		{
			name:     "scalar string produces single-element slice",
			value:    "X-Custom-Header: value1",
			expected: []string{"X-Custom-Header: value1"},
		},
		{
			name:     "typed string slice produces one string per element",
			value:    []string{"X-Custom-Header: value1", "X-Request-ID: abc123"},
			expected: []string{"X-Custom-Header: value1", "X-Request-ID: abc123"},
		},
		{
			name:     "any slice produces one string per element",
			value:    []any{"X-Custom-Header: value1", "X-Request-ID: abc123"},
			expected: []string{"X-Custom-Header: value1", "X-Request-ID: abc123"},
		},
		{
			name:     "empty slice",
			value:    []any{},
			expected: []string{},
		},
		{
			name:     "JSON array string is parsed into multiple entries",
			value:    `["X-Foo: bar","X-Baz: qux"]`,
			expected: []string{"X-Foo: bar", "X-Baz: qux"},
		},
		{
			name:     "JSON array with single element",
			value:    `["X-Foo: bar"]`,
			expected: []string{"X-Foo: bar"},
		},
		{
			name:     "string starting with [ but not ending with ] is treated as scalar",
			value:    "[not-an-array",
			expected: []string{"[not-an-array"},
		},
		{
			name:     "string ending with ] but not starting with [ is treated as scalar",
			value:    "not-an-array]",
			expected: []string{"not-an-array]"},
		},
		{
			name:     "invalid JSON array is treated as scalar",
			value:    "[not valid json]",
			expected: []string{"[not valid json]"},
		},
		{
			name:     "boolean value is stringified",
			value:    true,
			expected: []string{"true"},
		},
		{
			name:     "integer value is stringified",
			value:    42,
			expected: []string{"42"},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := viperValueToStrings(test.value)
			assert.Equal(t, test.expected, result)
		})
	}
}
