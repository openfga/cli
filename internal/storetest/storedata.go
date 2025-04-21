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

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/tuplefile"
)

type ModelTestCheck struct {
	User       string                  `json:"user"       yaml:"user"`
	Object     string                  `json:"object"     yaml:"object"`
	Context    *map[string]interface{} `json:"context"    yaml:"context,omitempty"`
	Assertions map[string]bool         `json:"assertions" yaml:"assertions"`
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
	Name      string                            `json:"name"       yaml:"name"`
	Model     string                            `json:"model"      yaml:"model"`
	ModelFile string                            `json:"model_file" yaml:"model_file,omitempty"` //nolint:tagliatelle
	Tuples    []client.ClientContextualTupleKey `json:"tuples"     yaml:"tuples"`
	TupleFile string                            `json:"tuple_file" yaml:"tuple_file,omitempty"` //nolint:tagliatelle
	Tests     []ModelTest                       `json:"tests"      yaml:"tests"`
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
	var errs error

	if storeData.TupleFile != "" {
		tuples, err := tuplefile.ReadTupleFile(path.Join(basePath, storeData.TupleFile))
		if err != nil { //nolint:gocritic
			errs = fmt.Errorf("failed to process global tuple %s file due to %w", storeData.TupleFile, err)
		} else if storeData.Tuples == nil {
			storeData.Tuples = tuples
		} else {
			storeData.Tuples = append(storeData.Tuples, tuples...)
		}
	}

	for index, test := range storeData.Tests {
		if test.TupleFile == "" {
			continue
		}

		tuples, err := tuplefile.ReadTupleFile(path.Join(basePath, test.TupleFile))
		if err != nil {
			errs = errors.Join(
				errs,
				fmt.Errorf("failed to process tuple file %s for test %s due to %w", test.TupleFile, test.Name, err),
			)
		} else {
			storeData.Tests[index].Tuples = tuples
		}
	}

	if errs != nil {
		return errors.Join(errors.New("failed to process one or more tuple files"), errs) //nolint:err113
	}

	return nil
}
