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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/openfga/go-sdk/client"

	"gopkg.in/yaml.v3"

	"github.com/openfga/cli/internal/authorizationmodel"
)

type ModelTestCheck struct {
	User       string                  `json:"user"       yaml:"user"`
	Object     string                  `json:"object"     yaml:"object"`
	Context    *map[string]interface{} `json:"context"    yaml:"context"`
	Assertions map[string]bool         `json:"assertions" yaml:"assertions"`
}

type ModelTestListObjects struct {
	User       string                  `json:"user"       yaml:"user"`
	Type       string                  `json:"type"       yaml:"type"`
	Context    *map[string]interface{} `json:"context"    yaml:"context"`
	Assertions map[string][]string     `json:"assertions" yaml:"assertions"`
}

type ModelTest struct {
	Name        string                            `json:"name"         yaml:"name"`
	Description string                            `json:"description"  yaml:"description"`
	Tuples      []client.ClientContextualTupleKey `json:"tuples"       yaml:"tuples"`
	TupleFile   string                            `json:"tuple_file"   yaml:"tuple_file"` //nolint:tagliatelle
	Check       []ModelTestCheck                  `json:"check"        yaml:"check"`
	ListObjects []ModelTestListObjects            `json:"list_objects" yaml:"list_objects"` //nolint:tagliatelle
}

type StoreData struct {
	Name      string                            `json:"name"       yaml:"name"`
	Model     string                            `json:"model"      yaml:"model"`
	ModelFile string                            `json:"model_file" yaml:"model_file"` //nolint:tagliatelle
	Tuples    []client.ClientContextualTupleKey `json:"tuples"     yaml:"tuples"`
	TupleFile string                            `json:"tuple_file" yaml:"tuple_file"` //nolint:tagliatelle
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
		tuples, err := readTupleFile(path.Join(basePath, storeData.TupleFile))
		if err != nil {
			errs = fmt.Errorf("failed to process global tuple %s file due to %w", storeData.TupleFile, err)
		} else {
			storeData.Tuples = tuples
		}
	}

	for index := 0; index < len(storeData.Tests); index++ {
		test := storeData.Tests[index]
		if test.TupleFile == "" {
			continue
		}

		tuples, err := readTupleFile(path.Join(basePath, test.TupleFile))
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
		return errors.Join(errors.New("failed to process one or more tuple files"), errs) //nolint:goerr113
	}

	return nil
}

func readTupleFile(tuplePath string) ([]client.ClientContextualTupleKey, error) {
	var tuples []client.ClientContextualTupleKey

	tupleFile, err := os.Open(tuplePath)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	defer tupleFile.Close()

	switch path.Ext(tuplePath) {
	case ".json":
		contents, err := io.ReadAll(tupleFile)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		err = json.Unmarshal(contents, &tuples)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
	case ".yaml", ".yml":
		decoder := yaml.NewDecoder(tupleFile)
		decoder.KnownFields(true)

		err = decoder.Decode(&tuples)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
	default:
		return nil, fmt.Errorf("unsupported file format %s", path.Ext(tuplePath)) //nolint:goerr113
	}

	return tuples, nil
}
