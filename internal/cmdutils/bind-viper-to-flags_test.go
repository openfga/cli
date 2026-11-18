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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindViperToFlags(t *testing.T) {
	t.Parallel()

	const flagName = "header"

	testcases := []struct {
		name     string
		value    any
		expected []string
	}{
		{
			name: "slice value produces one flag value per element",
			value: []any{
				"X-Custom-Header: value1",
				"X-Request-ID: abc123",
			},
			expected: []string{"X-Custom-Header: value1", "X-Request-ID: abc123"},
		},
		{
			name:     "single element slice",
			value:    []any{"X-Custom-Header: value1"},
			expected: []string{"X-Custom-Header: value1"},
		},
		{
			name:     "empty slice leaves flag untouched",
			value:    []any{},
			expected: []string{},
		},
		{
			name:     "typed string slice produces one flag value per element",
			value:    []string{"X-Custom-Header: value1", "X-Request-ID: abc123"},
			expected: []string{"X-Custom-Header: value1", "X-Request-ID: abc123"},
		},
		{
			name:     "typed int slice produces one flag value per element",
			value:    []int{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "scalar string produces single flag value",
			value:    "https://api.fga.example",
			expected: []string{"https://api.fga.example"},
		},
		{
			name:     "space separated scalar string (env var style) splits into multiple flag values",
			value:    "X-Custom-Header:value1 X-Request-ID:abc123",
			expected: []string{"X-Custom-Header:value1", "X-Request-ID:abc123"},
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

			cmd := &cobra.Command{Use: "root"}
			cmd.Flags().StringArray(flagName, nil, "")

			viperInstance := viper.New()
			viperInstance.Set(flagName, test.value)

			BindViperToFlags(cmd, viperInstance)

			got, err := cmd.Flags().GetStringArray(flagName)
			require.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

// newDebugCmd builds a command with a bound debug flag and a viper instance
// that behaves like the real one: the "debug" key is present (as a config file
// would set it) and FGA_* environment variables are read automatically.
func newDebugCmd(t *testing.T, configDebug bool) (*cobra.Command, *viper.Viper) {
	t.Helper()

	cmd := &cobra.Command{Use: "root"}
	cmd.Flags().Bool("debug", false, "")

	viperInstance := viper.New()
	viperInstance.SetEnvPrefix("FGA")
	viperInstance.AutomaticEnv()

	if configDebug {
		// Simulate a config file having set debug: true.
		viperInstance.Set("debug", true)
	}

	return cmd, viperInstance
}

func debugValue(t *testing.T, cmd *cobra.Command) bool {
	t.Helper()

	val, err := cmd.Flags().GetBool("debug")
	require.NoError(t, err)

	return val
}

func TestDebugFlagIgnoredFromConfigFile(t *testing.T) {
	t.Setenv("FGA_DEBUG", "")

	cmd, viperInstance := newDebugCmd(t, true)
	BindViperToFlags(cmd, viperInstance)

	assert.False(t, debugValue(t, cmd), "config file debug:true must not enable debug")
}

func TestDebugFlagFromEmptyEnvIgnoresConfigFile(t *testing.T) {
	// An empty env var must not re-open the config-file binding path.
	t.Setenv("FGA_DEBUG", "")

	cmd, viperInstance := newDebugCmd(t, true)
	BindViperToFlags(cmd, viperInstance)

	assert.False(t, debugValue(t, cmd), "empty FGA_DEBUG must not enable debug from config file")
}

func TestDebugFlagFalseEnvIgnoresConfigFile(t *testing.T) {
	t.Setenv("FGA_DEBUG", "false")

	cmd, viperInstance := newDebugCmd(t, true)
	BindViperToFlags(cmd, viperInstance)

	assert.False(t, debugValue(t, cmd), "FGA_DEBUG=false must keep debug off despite config file")
}

func TestDebugFlagEnabledByEnv(t *testing.T) {
	t.Setenv("FGA_DEBUG", "true")

	cmd, viperInstance := newDebugCmd(t, false)
	BindViperToFlags(cmd, viperInstance)

	assert.True(t, debugValue(t, cmd), "FGA_DEBUG=true must enable debug")
}

func TestDebugFlagExplicitFlagWins(t *testing.T) {
	t.Setenv("FGA_DEBUG", "")

	cmd, viperInstance := newDebugCmd(t, true)
	require.NoError(t, cmd.Flags().Set("debug", "true"))

	BindViperToFlags(cmd, viperInstance)

	assert.True(t, debugValue(t, cmd), "an explicit --debug flag must be preserved")
}
