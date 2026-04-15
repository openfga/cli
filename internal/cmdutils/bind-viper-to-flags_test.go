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

package cmdutils_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openfga/cli/internal/cmdutils"
)

func TestBindViperToFlags_StringArrayFromYAML(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		config   map[string]any
		expected []string
	}{
		{
			name: "yaml list binds as multiple values",
			config: map[string]any{
				"custom-headers": []any{
					"X-Custom-Header: value1",
					"X-Request-ID: abc123",
				},
			},
			expected: []string{"X-Custom-Header: value1", "X-Request-ID: abc123"},
		},
		{
			name: "single element list",
			config: map[string]any{
				"custom-headers": []any{
					"X-Custom-Header: value1",
				},
			},
			expected: []string{"X-Custom-Header: value1"},
		},
		{
			name:     "no config leaves default",
			config:   map[string]any{},
			expected: []string{},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().StringArray("custom-headers", []string{}, "test flag")

			v := viper.New()
			for k, val := range test.config {
				v.Set(k, val)
			}

			cmdutils.BindViperToFlags(cmd, v)

			result, err := cmd.Flags().GetStringArray("custom-headers")
			require.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestBindViperToFlags_ScalarFlagUnchanged(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("api-url", "http://localhost:8080", "test flag")

	v := viper.New()
	v.Set("api-url", "https://api.fga.example")

	cmdutils.BindViperToFlags(cmd, v)

	result, err := cmd.Flags().GetString("api-url")
	require.NoError(t, err)
	assert.Equal(t, "https://api.fga.example", result)
}

func TestBindViperToFlags_CLIFlagTakesPrecedence(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringArray("custom-headers", []string{}, "test flag")

	require.NoError(t, cmd.Flags().Set("custom-headers", "X-CLI: from-flag"))

	v := viper.New()
	v.Set("custom-headers", []any{"X-Config: from-yaml"})

	cmdutils.BindViperToFlags(cmd, v)

	result, err := cmd.Flags().GetStringArray("custom-headers")
	require.NoError(t, err)
	assert.Equal(t, []string{"X-CLI: from-flag"}, result)
}
