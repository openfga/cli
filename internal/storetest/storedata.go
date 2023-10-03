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
	User       string                 `json:"user"       yaml:"user"`
	Object     string                 `json:"object"     yaml:"object"`
	Context    map[string]interface{} `json:"context"  yaml:"context"`
	Assertions map[string]bool        `json:"assertions" yaml:"assertions"`
}

type ModelTestListObjects struct {
	User       string                 `json:"user"       yaml:"user"`
	Type       string                 `json:"type"       yaml:"type"`
	Context    map[string]interface{} `json:"context"  yaml:"context"`
	Assertions map[string][]string    `json:"assertions" yaml:"assertions"`
}

type ModelTest struct {
	Name        string                              `json:"name"         yaml:"name"`
	Description string                              `json:"description"  yaml:"description"`
	Tuples      []client.ClientWriteRequestTupleKey `json:"tuples"       yaml:"tuples"`
	Check       []ModelTestCheck                    `json:"check"        yaml:"check"`
	ListObjects []ModelTestListObjects              `json:"list_objects" yaml:"list_objects"` //nolint:tagliatelle
}

type StoreData struct {
	Name      string                              `json:"name"       yaml:"name"`
	Model     string                              `json:"model"      yaml:"model"`
	ModelFile string                              `json:"model_file" yaml:"model_file"` //nolint:tagliatelle
	Tuples    []client.ClientWriteRequestTupleKey `json:"tuples"     yaml:"tuples"`
	Tests     []ModelTest                         `json:"tests"      yaml:"tests"`
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
