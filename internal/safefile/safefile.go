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

// Package safefile contains helpers for safely reading files referenced by
// user-supplied input (e.g. store YAML files).
package safefile

import (
	"errors"
	"fmt"
	"os"
)

// ErrNotRegularFile is returned when a referenced path is not a regular file
// (e.g. a device file or FIFO), which makes a subsequent read unsafe.
var ErrNotRegularFile = errors.New("not a regular file; refusing to read")

// CheckRegular returns an error if the file at name is not a regular file.
//
// Reading a non-regular file with os.ReadFile is unsafe in different ways
// depending on the file type: a FIFO (named pipe) with no writer blocks the
// read indefinitely, hanging the process, while an infinite device such as
// /dev/zero returns data endlessly, so os.ReadFile grows its buffer without
// bound until the process is killed by the OOM killer. os.Stat performs a
// metadata-only stat that does not open the file for I/O, so it neither blocks
// nor reads any data, allowing us to reject the file before either failure
// mode is triggered.
func CheckRegular(name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("failed to stat file %q: %w", name, err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("%q (mode %s): %w", name, info.Mode(), ErrNotRegularFile)
	}

	return nil
}
