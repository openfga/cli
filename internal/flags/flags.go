// Package flags provides utility functions for working with cobra command flags.
// It simplifies the process of marking flags as required and handling related errors.
package flags

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// ErrFlagRequired is returned when a flag cannot be marked as required.
	ErrFlagRequired = errors.New("error setting flag as required")

	// ErrInvalidInput is returned when invalid input is provided.
	ErrInvalidInput = errors.New("invalid input")
)

// buildFlagRequiredError creates a consistent error message for flag requirement failures.
// It wraps the original error with context about which flag and location failed.
func buildFlagRequiredError(flag, location string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%w - (flag: %s, file: %s): %w", ErrFlagRequired, flag, location, err)
}

// SetFlagRequired marks a single flag as required for a cobra command.
//
// Parameters:
//   - cmd: The cobra command to modify
//   - flag: The name of the flag to mark as required
//   - location: A string identifying the calling location (for error context)
//   - isPersistent: If true, marks the persistent flag as required; otherwise marks the regular flag
//
// Returns an error if:
//   - cmd is nil
//   - flag is empty
//   - the flag cannot be marked as required (e.g., flag doesn't exist)
func SetFlagRequired(cmd *cobra.Command, flag string, location string, isPersistent bool) error {
	if cmd == nil {
		return fmt.Errorf("%w: command cannot be nil", ErrInvalidInput)
	}

	if strings.TrimSpace(flag) == "" {
		return fmt.Errorf("%w: flag name cannot be empty", ErrInvalidInput)
	}

	if isPersistent {
		if err := cmd.MarkPersistentFlagRequired(flag); err != nil {
			return buildFlagRequiredError(flag, location, err)
		}
	} else {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			return buildFlagRequiredError(flag, location, err)
		}
	}

	return nil
}

// SetFlagsRequired marks multiple flags as required for a cobra command.
//
// Parameters:
//   - cmd: The cobra command to modify
//   - flags: A slice of flag names to mark as required
//   - location: A string identifying the calling location (for error context)
//   - isPersistent: If true, marks the persistent flags as required; otherwise marks the regular flags
//
// Returns a joined error containing all individual flag requirement failures.
// If no flags are provided or all succeed, returns nil.
//
// Note: This function continues processing all flags even if some fail,
// allowing you to see all failures at once rather than stopping at the first error.
func SetFlagsRequired(cmd *cobra.Command, flags []string, location string, isPersistent bool) error {
	if cmd == nil {
		return fmt.Errorf("%w: command cannot be nil", ErrInvalidInput)
	}

	if len(flags) == 0 {
		return nil
	}

	// Pre-allocate slice with exact capacity needed
	flagErrors := make([]error, 0, len(flags))

	for _, flag := range flags {
		if err := SetFlagRequired(cmd, flag, location, isPersistent); err != nil {
			flagErrors = append(flagErrors, err)
		}
	}

	if len(flagErrors) > 0 {
		return errors.Join(flagErrors...)
	}

	return nil
}
