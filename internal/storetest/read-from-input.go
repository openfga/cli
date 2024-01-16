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

package storetest

import (
	"fmt"
	"os"

	"github.com/openfga/cli/internal/authorizationmodel"

	"gopkg.in/yaml.v3"
)

// ReadFromFile is used to read and parse the Store file.
func ReadFromFile(fileName string, basePath string) (authorizationmodel.ModelFormat, *StoreData, error) {
	format := authorizationmodel.ModelFormatDefault

	var storeData StoreData

	testFile, err := os.Open(fileName)
	if err != nil {
		return format, nil, fmt.Errorf("failed to read file %s due to %w", fileName, err)
	}

	decoder := yaml.NewDecoder(testFile)
	decoder.KnownFields(true)
	err = decoder.Decode(&storeData)

	if err != nil {
		return format, nil, fmt.Errorf("failed to unmarshal file %s due to %w", fileName, err)
	}

	format, err = storeData.LoadModel(basePath)
	if err != nil {
		return format, nil, err
	}

	err = storeData.LoadTuples(basePath)
	if err != nil {
		return format, nil, err
	}

	return format, &storeData, nil
}
