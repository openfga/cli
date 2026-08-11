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
	"io"
	"io/fs"
	"os"
)

// ErrNotRegularFile is returned when a referenced path is not a regular file
// (e.g. a device file or FIFO), which makes a subsequent read unsafe.
var ErrNotRegularFile = errors.New("not a regular file; refusing to read")

// checkRegularInfo returns an error unless info describes a regular file.
//
// Reading a non-regular file is unsafe in different ways depending on the file
// type: a FIFO (named pipe) with no writer blocks the read indefinitely,
// hanging the process, while an infinite device such as /dev/zero returns data
// endlessly, so the read grows its buffer without bound until the process is
// killed by the OOM killer.
func checkRegularInfo(name string, info fs.FileInfo) error {
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%q (mode %s): %w", name, info.Mode(), ErrNotRegularFile)
	}

	return nil
}

// checkRegularPath returns an error if the file at name is not a regular file.
//
// os.Stat performs a metadata-only lookup that does not open the file for I/O,
// so it neither blocks on a FIFO nor reads from a device, allowing the file to
// be rejected before either failure mode is triggered.
func checkRegularPath(name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("failed to stat file %q: %w", name, err)
	}

	return checkRegularInfo(name, info)
}

// readOpened reads an already-opened file after confirming, from the file
// descriptor itself, that it is a regular file. Checking the descriptor rather
// than the path means the file whose contents are returned is the same file that
// was validated.
//
// The buffer is pre-sized from the descriptor's reported size so the read does
// not repeatedly grow and copy it, matching what os.ReadFile does.
func readOpened(name string, file *os.File) ([]byte, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %q: %w", name, err)
	}

	if err := checkRegularInfo(name, info); err != nil {
		return nil, err
	}

	// Pre-size from the descriptor's reported size and grow only if the file
	// turned out to be larger, which is what os.ReadFile does. io.ReadAll would
	// instead start small and repeatedly double, allocating roughly twice the
	// file size in the process.
	data := make([]byte, 0, info.Size()+1)

	for {
		if len(data) >= cap(data) {
			data = append(data, 0)[:len(data)]
		}

		read, err := file.Read(data[len(data):cap(data)])
		data = data[:len(data)+read]

		if err != nil {
			if errors.Is(err, io.EOF) {
				return data, nil
			}

			return nil, fmt.Errorf("failed to read file %q: %w", name, err)
		}
	}
}

// ReadContained reads ref, resolved relative to basePath, and returns its
// contents. The read is contained to basePath: references escaping it via "..",
// an absolute path, or a symlink pointing outside the tree are rejected.
//
// The file is opened and read through an os.Root handle rather than resolved to
// a path and read separately. That matters because a lexical path and a
// root-relative lookup do not always agree: given a symlink "link" -> "sub/dir",
// os.Root resolves "link/../target" to "sub/target" whereas filepath.Join
// collapses it to "target". Validating one path and then reading the other would
// leave the containment check bypassable, so the same handle is used throughout.
//
// The target must also be a regular file. A metadata-only Stat runs first to
// reject non-regular files without opening anything for I/O. The Stat and the
// open are not atomic, so the open is nonblocking — a FIFO swapped into place
// between the two cannot block it — and the opened descriptor is re-checked
// before its contents are read.
func ReadContained(basePath, ref string) ([]byte, error) {
	root, err := os.OpenRoot(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open base directory %q: %w", basePath, err)
	}

	defer root.Close()

	// Metadata-only lookup through the root: rejects escapes and non-regular
	// files without opening anything for I/O.
	info, err := root.Stat(ref)
	if err != nil {
		return nil, fmt.Errorf("%q is not accessible within %q: %w", ref, basePath, err)
	}

	if err := checkRegularInfo(ref, info); err != nil {
		return nil, err
	}

	file, err := root.OpenFile(ref, os.O_RDONLY|openNonblock, 0)
	if err != nil {
		return nil, fmt.Errorf("%q is not accessible within %q: %w", ref, basePath, err)
	}

	defer file.Close()

	return readOpened(ref, file)
}

// ReadExternal reads name without containing it to any directory, for callers
// that have explicitly opted in to references outside the base directory. The
// target must still be a regular file, checked before the open and again on
// the opened descriptor; the open is nonblocking so a FIFO swapped into place
// after the check cannot block it.
func ReadExternal(name string) ([]byte, error) {
	if err := checkRegularPath(name); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(name, os.O_RDONLY|openNonblock, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", name, err)
	}

	defer file.Close()

	return readOpened(name, file)
}
