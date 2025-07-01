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
	"path"

	"github.com/openfga/cli/internal/clierrors"
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

	errMissingTuple               = errors.New("either tuple_file or tuple_files or tuples must be provided")
	errFailedProcessingTupleFiles = errors.New("failed to process one or more tuple files")
)

type ModelTestCheck struct {
	User       string                  `json:"user,omitempty"    yaml:"user,omitempty"`
	Users      []string                `json:"users,omitempty"   yaml:"users,omitempty"`
	Object     string                  `json:"object,omitempty"  yaml:"object,omitempty"`
	Objects    []string                `json:"objects,omitempty" yaml:"objects,omitempty"`
	Context    *map[string]interface{} `json:"context"           yaml:"context,omitempty"`
	Assertions map[string]bool         `json:"assertions"        yaml:"assertions"`
}

type ModelTestListObjects struct {
	User       string                  `json:"user"       yaml:"user"`
	Type       string                  `json:"type"       yaml:"type"`
	Context    *map[string]interface{} `json:"context"    yaml:"context"`
	Assertions map[string][]string     `json:"assertions" yaml:"assertions"`
}

type ModelTestListUsers struct {
	Object     string                                 `json:"object"      yaml:"object"`
	UserFilter []openfga.UserTypeFilter               `json:"user_filter" yaml:"user_filter"` //nolint:tagliatelle
	Context    *map[string]interface{}                `json:"context"     yaml:"context,omitempty"`
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

func (storeData *StoreData) LoadModel(basePath string) (authorizationmodel.ModelFormat, error) {
	format := authorizationmodel.ModelFormatDefault
	if storeData.Model != "" {
		return format, nil
	}

	if storeData.ModelFile == "" {
		return format, nil
	}

	var inputModel string

	storeName := storeData.Name
	if err := authorizationmodel.ReadFromFile(
		path.Join(basePath, storeData.ModelFile),
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

func (storeData *StoreData) LoadTuples(basePath string) error {
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

	if storeData.TupleFile == "" && len(storeData.TupleFiles) == 0 && len(allTuples) == 0 {
		errs = errors.Join(
			errs,
			errMissingTuple,
		)
	}

	errs = errors.Join(
		errs,
		storeData.loadAndAddTuplesFromFile(basePath, storeData.TupleFile, addTuples),
	)

	errs = errors.Join(
		errs,
		storeData.loadAndAddTuplesFromFiles(basePath, storeData.TupleFiles, addTuples),
	)

	if len(allTuples) > 0 {
		storeData.Tuples = allTuples
	}

	errs = errors.Join(
		errs,
		storeData.loadTestTuples(basePath),
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
	add func([]client.ClientContextualTupleKey),
) error {
	if file == "" {
		return nil
	}

	tuples, err := tuplefile.ReadTupleFile(path.Join(basePath, file))
	if err != nil {
		return fmt.Errorf("failed to process global tuple %s file due to %w", file, err)
	}

	add(tuples)

	return nil
}

func (storeData *StoreData) loadAndAddTuplesFromFiles(
	basePath string,
	files []string,
	add func([]client.ClientContextualTupleKey),
) error {
	var errs error

	for _, file := range files {
		tuples, err := tuplefile.ReadTupleFile(path.Join(basePath, file))
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

func (storeData *StoreData) loadTestTuples(basePath string) error {
	var errs error

	for testIndex, testCase := range storeData.Tests {
		if testCase.TupleFile == "" {
			continue
		}

		tuples, err := tuplefile.ReadTupleFile(path.Join(basePath, testCase.TupleFile))
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
