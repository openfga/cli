package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// chdir changes into dir for the duration of the test, restoring the previous
// working directory afterwards.
func chdir(t *testing.T, dir string) {
	t.Helper()

	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir to %q: %v", dir, err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})
}

// userDirs returns the config and home directories initConfig resolves.
func userDirs(t *testing.T) (string, string) {
	t.Helper()

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("no user config directory: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no user home directory: %v", err)
	}

	return configDir, homeDir
}

// TestConfigInWorkingDirectoryIsNotRead is the regression test for config files
// in the working directory: such a file must not contribute any value, because
// the working directory is not necessarily under the user's control.
//
// Every name viper would resolve for SetConfigName(".fga") is covered, not just
// the YAML ones: viper searches every extension in viper.SupportedExts, so
// .fga.json and friends were loaded from the working directory too.
func TestConfigInWorkingDirectoryIsNotRead(t *testing.T) { //nolint:paralleltest
	// Not parallel: mutates the process working directory.
	const attackerURL = "https://attacker.example"

	configFiles := map[string]string{
		".fga.yaml":       "api-url: " + attackerURL + "\n",
		".fga.yml":        "api-url: " + attackerURL + "\n",
		".fga.json":       `{"api-url":"` + attackerURL + `"}`,
		".fga.toml":       `api-url = "` + attackerURL + `"` + "\n",
		".fga.ini":        "api-url = " + attackerURL + "\n",
		".fga.env":        "api-url=" + attackerURL + "\n",
		".fga.properties": "api-url = " + attackerURL + "\n",
		".fga":            "api-url: " + attackerURL + "\n",
	}

	// Subtests are not parallel either: each one chdirs into its own directory.
	for name, contents := range configFiles { //nolint:paralleltest
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()

			path := filepath.Join(dir, name)
			if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
				t.Fatalf("failed to write config %q: %v", path, err)
			}

			chdir(t, dir)

			viperInstance := viper.New()

			// Registered by the same code initConfig uses, so a search path added
			// there is covered here without restating the list.
			configDir, homeDir := userDirs(t)
			applyConfigSearch(viperInstance, configDir, homeDir)

			// A real config may exist in the developer's home directory, so assert
			// on where the value came from rather than on the error.
			_ = viperInstance.ReadInConfig()

			if used := viperInstance.ConfigFileUsed(); used != "" {
				if usedDir, err := filepath.EvalSymlinks(filepath.Dir(used)); err == nil {
					if wantDir, err := filepath.EvalSymlinks(dir); err == nil && usedDir == wantDir {
						t.Fatalf("config in the working directory was read: %q", used)
					}
				}
			}

			if got := viperInstance.GetString("api-url"); got == attackerURL {
				t.Fatalf("api-url was taken from the working-directory config %q", name)
			}
		})
	}
}
