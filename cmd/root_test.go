package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid config file", func(t *testing.T) {
		t.Parallel()

		configFile := filepath.Join(t.TempDir(), ".fga.yaml")
		require.NoError(t, os.WriteFile(configFile, []byte("api-url: https://api.fga.example\n"), 0o600))

		viperInstance := viper.New()
		viperInstance.SetConfigFile(configFile)

		var stderr bytes.Buffer
		readConfig(viperInstance, &stderr)

		assert.Equal(t, "Using config file: "+configFile+"\n", stderr.String())
		assert.Equal(t, "https://api.fga.example", viperInstance.GetString("api-url"))
	})

	t.Run("invalid config file", func(t *testing.T) {
		t.Parallel()

		configFile := filepath.Join(t.TempDir(), ".fga.yaml")
		// YAML doesn't allow tabs for indentation
		content := "api-url: https://api.fga.example\nstore-id: 01H0H015178Y2V4CX10C2KGHF4\n\tapi-token: token\n"
		require.NoError(t, os.WriteFile(configFile, []byte(content), 0o600))

		viperInstance := viper.New()
		viperInstance.SetConfigFile(configFile)

		var stderr bytes.Buffer
		readConfig(viperInstance, &stderr)

		assert.Contains(t, stderr.String(), "Warning: ignoring config file "+configFile+": ")
		assert.Empty(t, viperInstance.GetString("api-url"))
	})

	t.Run("no config file found", func(t *testing.T) {
		t.Parallel()

		viperInstance := viper.New()
		viperInstance.AddConfigPath(t.TempDir())
		viperInstance.SetConfigName(".fga")

		var stderr bytes.Buffer
		readConfig(viperInstance, &stderr)

		assert.Empty(t, stderr.String())
	})
}
