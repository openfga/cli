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
	"path"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/go-sdk/client"
)

type ModelTestCheck struct {
	User       string          `json:"user"       yaml:"user"`
	Object     string          `json:"object"     yaml:"object"`
	Assertions map[string]bool `json:"assertions" yaml:"assertions"`
}

type ModelTestListObjects struct {
	User       string              `json:"user"       yaml:"user"`
	Type       string              `json:"type"       yaml:"type"`
	Assertions map[string][]string `json:"assertions" yaml:"assertions"`
}

type ModelTest struct {
	Name        string                  `json:"name"         yaml:"name"`
	Description string                  `json:"description"  yaml:"description"`
	Model       string                  `json:"model"        yaml:"model"`
	ModelFile   string                  `json:"model_file"   yaml:"model-file"` //nolint:tagliatelle
	Tuples      []client.ClientTupleKey `json:"tuples"       yaml:"tuples"`
	Check       []ModelTestCheck        `json:"check"        yaml:"check"`
	ListObjects []ModelTestListObjects  `json:"list_objects" yaml:"list-objects"` //nolint:tagliatelle
}

func (test *ModelTest) LoadModel(basePath string) (authorizationmodel.ModelFormat, error) {
	format := authorizationmodel.ModelFormatDefault

	if test.Model != "" {
		return format, nil
	}

	if test.ModelFile == "" {
		return format, nil
	}

	var inputModel string

	storeName := ""
	if err := authorizationmodel.ReadFromFile(
		path.Join(basePath, test.ModelFile),
		&inputModel,
		&format,
		&storeName); err != nil {
		return format, err //nolint:wrapcheck
	}

	if inputModel != "" {
		test.Model = inputModel
	}

	return format, nil
}

type StoreData struct {
	Name      string                  `json:"name"       yaml:"name"`
	Model     string                  `json:"model"      yaml:"model"`
	ModelFile string                  `json:"model_file" yaml:"model-file"` //nolint:tagliatelle
	Tuples    []client.ClientTupleKey `json:"tuples"     yaml:"tuples"`
	Tests     []ModelTest             `json:"tests"      yaml:"tests"`
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

type TestLocalityCount struct {
	Remote int
	Local  int
}

func (storeData *StoreData) GetTestLocalityCount() TestLocalityCount {
	counts := TestLocalityCount{
		Remote: 0,
		Local:  0,
	}

	if storeData.Model != "" || storeData.ModelFile != "" {
		counts.Local = len(storeData.Tests)

		return counts
	}

	for index := 0; index < len(storeData.Tests); index++ {
		test := storeData.Tests[index]
		if test.Model != "" || test.ModelFile != "" {
			counts.Local++
		} else {
			counts.Remote++
		}
	}

	return counts
}
