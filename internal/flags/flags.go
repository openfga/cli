// Package flags provides utility functions for working with cobra command flags.
// It simplifies the process of marking flags as required and handling related errors.
package flags

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// ErrFlagRequired is returned when a flag cannot be marked as required.
var ErrFlagRequired = errors.New("error setting flag as required")

// SetFlagRequired marks a single flag as required for a cobra command.
// If isPersistent is true, it marks the persistent flag as required.
// Returns an error if the flag cannot be marked as required.
func SetFlagRequired(cmd *cobra.Command, flag string, location string, isPersistent bool) error {
	if isPersistent {
		if err := cmd.MarkPersistentFlagRequired(flag); err != nil {
			return fmt.Errorf("%w - (flag: %s, file: %s): %w", ErrFlagRequired, flag, location, err)
		}
	} else {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			return fmt.Errorf("%w - (flag: %s, file: %s): %w", ErrFlagRequired, flag, location, err)
		}
	}

	return nil
}

// SetFlagsRequired marks multiple flags as required for a cobra command.
// If isPersistent is true, it marks the persistent flags as required.
// Returns a joined error if any flag cannot be marked as required.
func SetFlagsRequired(cmd *cobra.Command, flags []string, location string, isPersistent bool) error {
	flagErrors := make([]error, len(flags))

	for i, flag := range flags {
		if err := SetFlagRequired(cmd, flag, location, isPersistent); err != nil {
			flagErrors[i] = err
		}
	}

	return errors.Join(flagErrors...)
}
