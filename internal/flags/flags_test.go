package flags

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetFlagRequired_NonPersistent_Success(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("foo", "", "foo flag")

	err := SetFlagRequired(cmd, "foo", "TestLocation", false)

	assert.NoError(t, err)
}

func TestSetFlagRequired_Persistent_Success(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.PersistentFlags().String("bar", "", "bar flag")

	err := SetFlagRequired(cmd, "bar", "TestLocation", true)

	assert.NoError(t, err)
}

func TestSetFlagRequired_NonPersistent_FlagNotFound(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}

	err := SetFlagRequired(cmd, "missing", "TestLocation", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error setting flag as required - (flag: missing, file: TestLocation):")
}

func TestSetFlagRequired_Persistent_FlagNotFound(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}

	err := SetFlagRequired(cmd, "missing", "TestLocation", true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error setting flag as required - (flag: missing, file: TestLocation):")
}

func TestSetFlagRequired_ErrorMessageFormat(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	flagName := "nonexistent"
	location := "SomeFunction"

	err := SetFlagRequired(cmd, flagName, location, false)

	require.Error(t, err)

	expectedPrefix := fmt.Sprintf("error setting flag as required - (flag: %s, file: %s):", flagName, location)
	assert.Contains(t, err.Error(), expectedPrefix)
}

func TestSetFlagsRequired_AllSuccess(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("foo", "", "foo flag")
	cmd.Flags().String("bar", "", "bar flag")
	cmd.Flags().String("baz", "", "baz flag")

	flags := []string{"foo", "bar", "baz"}
	err := SetFlagsRequired(cmd, flags, "TestLocation", false)

	assert.NoError(t, err)
}

func TestSetFlagsRequired_PersistentAllSuccess(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.PersistentFlags().String("foo", "", "foo flag")
	cmd.PersistentFlags().String("bar", "", "bar flag")

	flags := []string{"foo", "bar"}
	err := SetFlagsRequired(cmd, flags, "TestLocation", true)

	assert.NoError(t, err)
}

func TestSetFlagsRequired_SomeSuccess_SomeFail(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("foo", "", "foo flag")
	// "missing" flag is not defined

	flags := []string{"foo", "missing"}
	err := SetFlagsRequired(cmd, flags, "TestLocation", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error setting flag as required - (flag: missing, file: TestLocation):")
	// The error should not contain "foo" since that one succeeded
	assert.NotContains(t, err.Error(), "error setting flag as required - (flag: foo, file: TestLocation):")
}

func TestSetFlagsRequired_AllFail(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}

	flags := []string{"missing1", "missing2"}
	err := SetFlagsRequired(cmd, flags, "TestLocation", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error setting flag as required - (flag: missing1, file: TestLocation):")
	assert.Contains(t, err.Error(), "error setting flag as required - (flag: missing2, file: TestLocation):")
}

func TestSetFlagsRequired_EmptySlice(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}

	err := SetFlagsRequired(cmd, []string{}, "TestLocation", false)

	assert.NoError(t, err)
}

func TestSetFlagsRequired_NilSlice(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}

	err := SetFlagsRequired(cmd, nil, "TestLocation", false)

	assert.NoError(t, err)
}

func TestSetFlagsRequired_ErrorJoining(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}

	flags := []string{"missing1", "missing2", "missing3"}
	err := SetFlagsRequired(cmd, flags, "TestLocation", false)

	require.Error(t, err)

	for _, flag := range flags {
		expectedError := fmt.Sprintf("error setting flag as required - (flag: %s, file: TestLocation):", flag)
		assert.Contains(t, err.Error(), expectedError)
	}
}

func TestSetFlagsRequired_MixedPersistentAndNonPersistent(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("regular", "", "regular flag")
	cmd.PersistentFlags().String("persistent", "", "persistent flag")

	flags := []string{"regular", "persistent"}
	err := SetFlagsRequired(cmd, flags, "TestLocation", false)

	// The regular flag should succeed, but persistent flag should fail
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error setting flag as required - (flag: persistent, file: TestLocation):")
	assert.NotContains(t, err.Error(), "error setting flag as required - (flag: regular, file: TestLocation):")
}

func TestSetFlagRequired_NilCommand(t *testing.T) {
	t.Parallel()

	err := SetFlagRequired(nil, "foo", "TestLocation", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input: command cannot be nil")
}

func TestSetFlagRequired_EmptyFlag(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	err := SetFlagRequired(cmd, "", "TestLocation", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input: flag name cannot be empty")
}

func TestSetFlagRequired_WhitespaceFlag(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	err := SetFlagRequired(cmd, "   ", "TestLocation", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input: flag name cannot be empty")
}

func TestSetFlagsRequired_NilCommand(t *testing.T) {
	t.Parallel()

	err := SetFlagsRequired(nil, []string{"foo"}, "TestLocation", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input: command cannot be nil")
}

func TestSetFlagsRequired_EmptyFlagInSlice(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	flags := []string{"valid", "", "also-valid"}
	err := SetFlagsRequired(cmd, flags, "TestLocation", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input: flag name cannot be empty")
}
