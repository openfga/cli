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

// Package storetest contains cli specific store interfaces and functionality
package storetest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openfga/cli/internal/clierrors"
	"github.com/openfga/cli/internal/safefile"
	"github.com/openfga/cli/internal/tuplefile"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/authorizationmodel"
)

// Static error variables for validation.
var (
	ErrUserAndUsersConflict     = errors.New("cannot contain both 'user' and 'users'")
	ErrUserRequired             = errors.New("must specify 'user' or 'users'")
	ErrObjectAndObjectsConflict = errors.New("cannot contain both 'object' and 'objects'")
	ErrObjectRequired           = errors.New("must specify 'object' or 'objects'")

	errFailedProcessingTupleFiles = errors.New("failed to process one or more tuple files")
)

type ModelTestCheck struct {
	User       string          `json:"user,omitempty"    yaml:"user,omitempty"`
	Users      []string        `json:"users,omitempty"   yaml:"users,omitempty"`
	Object     string          `json:"object,omitempty"  yaml:"object,omitempty"`
	Objects    []string        `json:"objects,omitempty" yaml:"objects,omitempty"`
	Context    *map[string]any `json:"context"           yaml:"context,omitempty"`
	Assertions map[string]bool `json:"assertions"        yaml:"assertions"`
}

type ModelTestListObjects struct {
	User       string              `json:"user"       yaml:"user"`
	Type       string              `json:"type"       yaml:"type"`
	Context    *map[string]any     `json:"context"    yaml:"context"`
	Assertions map[string][]string `json:"assertions" yaml:"assertions"`
}

type ModelTestListUsers struct {
	Object     string                                 `json:"object"      yaml:"object"`
	UserFilter []openfga.UserTypeFilter               `json:"user_filter" yaml:"user_filter"` //nolint:tagliatelle
	Context    *map[string]any                        `json:"context"     yaml:"context,omitempty"`
	Assertions map[string]ModelTestListUsersAssertion `json:"assertions"  yaml:"assertions"`
}

type ModelTestListUsersAssertion struct {
	Users []string `json:"users" yaml:"users"`
}

type ModelTest struct {
	Name        string                            `json:"name"         yaml:"name"`
	Description string                            `json:"description"  yaml:"description,omitempty"`
	Tuples      []client.ClientContextualTupleKey `json:"tuples"       yaml:"tuples,omitempty"`
	TupleFile   string                            `json:"tuple_file"   yaml:"tuple_file,omitempty"` //nolint:tagliatelle
	Check       []ModelTestCheck                  `json:"check"        yaml:"check"`
	ListObjects []ModelTestListObjects            `json:"list_objects" yaml:"list_objects,omitempty"` //nolint:tagliatelle
	ListUsers   []ModelTestListUsers              `json:"list_users"   yaml:"list_users,omitempty"`   //nolint:tagliatelle
}

type StoreData struct {
	Name       string                            `json:"name"        yaml:"name"`
	Model      string                            `json:"model"       yaml:"model"`
	ModelFile  string                            `json:"model_file"  yaml:"model_file,omitempty"` //nolint:tagliatelle
	Tuples     []client.ClientContextualTupleKey `json:"tuples"      yaml:"tuples"`
	TupleFile  string                            `json:"tuple_file"  yaml:"tuple_file,omitempty"`  //nolint:tagliatelle
	TupleFiles []string                          `json:"tuple_files" yaml:"tuple_files,omitempty"` //nolint:tagliatelle
	Tests      []ModelTest                       `json:"tests"       yaml:"tests"`
}

// resolveFile resolves a file referenced from within a store YAML against
// basePath and returns a path that is safe to read.
//
// By default (allowExternal == false) the reference is contained to basePath
// using os.Root: any component that escapes the base directory via "..",
// an absolute path, or a symlink pointing outside the tree is rejected
// atomically and in a symlink-safe manner (see go.dev/blog/osroot). This
// prevents a malicious store YAML from reading arbitrary local files.
//
// When allowExternal == true the caller has explicitly opted in (via
// --allow-external-files) to references that live outside basePath, so
// containment is skipped and only the resolved path is returned.
//
// In both modes the target must be a regular file. Reading a non-regular file
// with os.ReadFile is unsafe: a FIFO with no writer blocks the read forever,
// while an infinite device such as /dev/zero returns endless data and grows the
// read buffer until the process is OOM-killed. os.Root.Stat and os.Stat perform
// metadata-only lookups that neither block nor read data, so we reject
// non-regular files before either failure mode is triggered.
func resolveFile(basePath, ref string, allowExternal bool) (string, error) {
	if allowExternal {
		joined := filepath.Join(basePath, ref)
		if err := safefile.CheckRegular(joined); err != nil {
			return "", fmt.Errorf("file reference %q: %w", ref, err)
		}

		return joined, nil
	}

	root, err := os.OpenRoot(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to open base directory %q: %w", basePath, err)
	}
	defer root.Close()

	// Stat through the root: symlink-safe, rejects any escape from basePath,
	// and (being metadata-only) neither blocks on FIFOs nor reads from devices.
	info, err := root.Stat(ref)
	if err != nil {
		return "", fmt.Errorf("file reference %q is not accessible within %q: %w", ref, basePath, err)
	}

	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("file reference %q (mode %s): %w", ref, info.Mode(), safefile.ErrNotRegularFile)
	}

	return filepath.Join(basePath, ref), nil
}

func (storeData *StoreData) LoadModel(
	basePath string,
	allowExternalFiles bool,
) (authorizationmodel.ModelFormat, error) {
	format := authorizationmodel.ModelFormatDefault
	if storeData.Model != "" {
		return format, nil
	}

	if storeData.ModelFile == "" {
		return format, nil
	}

	modelPath, err := resolveFile(basePath, storeData.ModelFile, allowExternalFiles)
	if err != nil {
		return format, err
	}

	var inputModel string

	storeName := storeData.Name
	if err := authorizationmodel.ReadFromFile(
		modelPath,
		&inputModel,
		&format,
		&storeName); err != nil {
		return format, err //nolint:wrapcheck
	}

	if inputModel != "" {
		storeData.Model = inputModel
	}

	return format, nil
}

func (storeData *StoreData) LoadTuples(basePath string, allowExternalFiles bool) error {
	var (
		errs      error
		allTuples []client.ClientContextualTupleKey
	)

	addTuples := func(tuples []client.ClientContextualTupleKey) {
		allTuples = append(allTuples, tuples...)
	}

	if storeData.Tuples != nil {
		addTuples(storeData.Tuples)
	}

	if storeData.TupleFile != "" {
		errs = errors.Join(
			errs,
			storeData.loadAndAddTuplesFromFile(basePath, storeData.TupleFile, allowExternalFiles, addTuples),
		)
	}

	if len(storeData.TupleFiles) > 0 {
		errs = errors.Join(
			errs,
			storeData.loadAndAddTuplesFromFiles(basePath, storeData.TupleFiles, allowExternalFiles, addTuples),
		)
	}

	if len(allTuples) > 0 {
		storeData.Tuples = allTuples
	}

	errs = errors.Join(
		errs,
		storeData.loadTestTuples(basePath, allowExternalFiles),
	)
	if errs != nil {
		return errors.Join(
			errFailedProcessingTupleFiles,
			errs,
		)
	}

	return nil
}

//nolint:cyclop
func (storeData *StoreData) Validate() error {
	var errs error

	for _, test := range storeData.Tests {
		for index, check := range test.Check {
			if check.User != "" && len(check.Users) > 0 {
				err := fmt.Errorf("test %s check %d: %w", test.Name, index, ErrUserAndUsersConflict)
				errs = errors.Join(errs, err)
			} else if check.User == "" && len(check.Users) == 0 {
				err := fmt.Errorf("test %s check %d: %w", test.Name, index, ErrUserRequired)
				errs = errors.Join(errs, err)
			}

			if check.Object != "" && len(check.Objects) > 0 {
				err := fmt.Errorf("test %s check %d: %w", test.Name, index, ErrObjectAndObjectsConflict)
				errs = errors.Join(errs, err)
			} else if check.Object == "" && len(check.Objects) == 0 {
				err := fmt.Errorf("test %s check %d: %w", test.Name, index, ErrObjectRequired)
				errs = errors.Join(errs, err)
			}
		}
	}

	if errs != nil {
		return clierrors.ValidationError("StoreFormat", errs.Error()) //nolint:wrapcheck
	}

	return nil
}

func (storeData *StoreData) loadAndAddTuplesFromFile(
	basePath string,
	file string,
	allowExternalFiles bool,
	add func([]client.ClientContextualTupleKey),
) error {
	if file == "" {
		return nil
	}

	resolved, err := resolveFile(basePath, file, allowExternalFiles)
	if err != nil {
		return fmt.Errorf("failed to process global tuple %s file due to %w", file, err)
	}

	tuples, err := tuplefile.ReadTupleFile(resolved)
	if err != nil {
		return fmt.Errorf("failed to process global tuple %s file due to %w", file, err)
	}

	add(tuples)

	return nil
}

func (storeData *StoreData) loadAndAddTuplesFromFiles(
	basePath string,
	files []string,
	allowExternalFiles bool,
	add func([]client.ClientContextualTupleKey),
) error {
	var errs error

	for _, file := range files {
		resolved, err := resolveFile(basePath, file, allowExternalFiles)
		if err != nil {
			errs = errors.Join(
				errs,
				fmt.Errorf("failed to process tuple file %s due to %w", file, err),
			)

			continue
		}

		tuples, err := tuplefile.ReadTupleFile(resolved)
		if err != nil {
			errs = errors.Join(
				errs,
				fmt.Errorf("failed to process tuple file %s due to %w", file, err),
			)

			continue
		}

		add(tuples)
	}

	return errs
}

func (storeData *StoreData) loadTestTuples(basePath string, allowExternalFiles bool) error {
	var errs error

	for testIndex, testCase := range storeData.Tests {
		if testCase.TupleFile == "" {
			continue
		}

		resolved, err := resolveFile(basePath, testCase.TupleFile, allowExternalFiles)
		if err != nil {
			errs = errors.Join(
				errs,
				fmt.Errorf(
					"failed to process tuple file %s for test %s due to %w",
					testCase.TupleFile,
					testCase.Name,
					err,
				),
			)

			continue
		}

		tuples, err := tuplefile.ReadTupleFile(resolved)
		if err != nil {
			errs = errors.Join(
				errs,
				fmt.Errorf(
					"failed to process tuple file %s for test %s due to %w",
					testCase.TupleFile,
					testCase.Name,
					err,
				),
			)

			continue
		}

		storeData.Tests[testIndex].Tuples = append(storeData.Tests[testIndex].Tuples, tuples...)
	}

	return errs
}
