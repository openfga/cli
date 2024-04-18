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
	"path"

	"github.com/cucumber/godog"
	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/tuplefile"
)

type ModelTestCheck struct {
	User       string                  `json:"user"       yaml:"user"`
	Object     string                  `json:"object"     yaml:"object"`
	Context    *map[string]interface{} `json:"context"    yaml:"context"`
	Assertions map[string]bool         `json:"assertions" yaml:"assertions"`
}

func (m *ModelTestCheck) toScenario() (string, error) {
	scenarios := ""

	for relation, value := range m.Assertions {
		scenario := fmt.Sprintf("\n\tScenario: check %s %s %s\n", m.User, m.Object, relation)

		scenario += fmt.Sprintf("\t\tWhen %s accesses %s\n", m.User, m.Object)
		if m.Context != nil {
			scenario += "\t\tWhen context is\n"

			for key, value := range *m.Context {
				row, err := contextToTableRow(key, value)
				if err != nil {
					return "", err
				}

				scenario += row
			}
		}

		if value {
			scenario += fmt.Sprintf("\t\tThen %s has the %s permission\n", m.User, relation)
		} else {
			scenario += fmt.Sprintf("\t\tThen %s does not have the %s permission\n", m.User, relation)
		}

		scenarios += scenario + "\n"
	}

	return scenarios, nil
}

type ModelTestListObjects struct {
	User       string                  `json:"user"       yaml:"user"`
	Type       string                  `json:"type"       yaml:"type"`
	Context    *map[string]interface{} `json:"context"    yaml:"context"`
	Assertions map[string][]string     `json:"assertions" yaml:"assertions"`
}

func (m *ModelTestListObjects) toScenario() (string, error) {
	scenarios := ""

	for relation, objects := range m.Assertions {
		scenario := fmt.Sprintf("Scenario: list objects %s %s %s\n", m.User, m.Type, objects)

		if m.Context != nil {
			scenario += "\t\tWhen context is\n"

			for key, value := range *m.Context {
				row, err := contextToTableRow(key, value)
				if err != nil {
					return "", err
				}

				scenario += row
			}
		}

		scenario += fmt.Sprintf("\t\tWhen %s searches for %s\n", m.User, m.Type)

		if len(objects) > 0 {
			scenario += fmt.Sprintf("\t\tThen %s has the %s permission for\n", m.User, relation)
			for _, object := range objects {
				scenario += fmt.Sprintf("\t\t\t| %s |\n", object)
			}
		} else {
			scenario += fmt.Sprintf("\t\tThen %s has no %s permission\n", m.User, relation)
		}

		scenarios += scenario + "\n"
	}

	return scenarios, nil
}

type ModelTest struct {
	Name        string                            `json:"name"         yaml:"name"`
	Description string                            `json:"description"  yaml:"description"`
	Tuples      []client.ClientContextualTupleKey `json:"tuples"       yaml:"tuples"`
	TupleFile   string                            `json:"tuple_file"   yaml:"tuple_file"` //nolint:tagliatelle
	Check       []ModelTestCheck                  `json:"check"        yaml:"check"`
	ListObjects []ModelTestListObjects            `json:"list_objects" yaml:"list_objects"` //nolint:tagliatelle
}

func (m *ModelTest) toFeature(globalTuples string) (godog.Feature, error) {
	featureName := m.Name

	if featureName == "" {
		featureName = "Test"
	}

	feature := godog.Feature{
		Name: featureName,
	}

	featureString := fmt.Sprintf("Feature: %s\n", featureName)

	localTuples := ""

	if len(m.Tuples) > 0 {
		tuples, err := tuplesToGherkin(m.Tuples)
		if err != nil {
			return feature, err
		}

		localTuples += tuples
	}

	if globalTuples != "" || localTuples != "" {
		featureString += "\tBackground:\n" + globalTuples + localTuples
	}

	checkScenarios, err := m.toCheckScenarios()
	if err != nil {
		return feature, err
	}

	featureString += checkScenarios

	listObjectsScenarios, err := m.toListObjectsScenarios()
	if err != nil {
		return feature, err
	}

	featureString += listObjectsScenarios

	feature.Contents = []byte(featureString)

	return feature, nil
}

func (m *ModelTest) toCheckScenarios() (string, error) {
	scenarios := ""

	for _, check := range m.Check {
		scenario, err := check.toScenario()
		if err != nil {
			return "", err
		}

		scenarios += "\n\t" + scenario
	}

	return scenarios, nil
}

func (m *ModelTest) toListObjectsScenarios() (string, error) {
	scenarios := ""

	for _, check := range m.ListObjects {
		scenario, err := check.toScenario()
		if err != nil {
			return "", err
		}

		scenarios += "\n\t" + scenario
	}

	return scenarios, nil
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
		tuples, err := tuplefile.ReadTupleFile(path.Join(basePath, storeData.TupleFile))
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
		return errors.Join(errors.New("failed to process one or more tuple files"), errs) //nolint:goerr113
	}

	return nil
}

func (storeData *StoreData) ToFeatures() ([]godog.Feature, error) {
	features := []godog.Feature{}
	globalTuples := ""

	if len(storeData.Tuples) > 0 {
		tuples, err := tuplesToGherkin(storeData.Tuples)
		if err != nil {
			return features, err
		}

		globalTuples += tuples
	}

	for _, test := range storeData.Tests {
		feature, err := test.toFeature(globalTuples)
		if err != nil {
			return features, err
		}

		features = append(features, feature)
	}

	return features, nil
}

func tuplesToGherkin(tuples []client.ClientContextualTupleKey) (string, error) {
	givenString := ""
	for _, tuple := range tuples {
		givenString += fmt.Sprintf("\t\tGiven %s is a %s of %s", tuple.User, tuple.Relation, tuple.Object)

		if tuple.Condition != nil {
			if tuple.Condition.Context != nil {
				givenString += fmt.Sprintf(" with %s being\n", tuple.Condition.Name)

				for key, value := range *tuple.Condition.Context {
					row, err := contextToTableRow(key, value)
					if err != nil {
						return "", err
					}

					givenString += row
				}
			} else {
				givenString += fmt.Sprintf(" with %s\n", tuple.Condition.Name)
			}
		}

		givenString += "\n"
	}

	return givenString, nil
}

func contextToTableRow(key string, value interface{}) (string, error) {
	switch value.(type) {
	case map[string]any, []any:
		value, err := json.Marshal(value)
		if err != nil {
			return "", err //nolint:wrapcheck
		}

		return fmt.Sprintf("\t\t\t| %s | %s |\n", key, string(value)), nil
	default:
		return fmt.Sprintf("\t\t\t| %s | %v |\n", key, value), nil
	}
}
