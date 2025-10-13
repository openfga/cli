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
	"path/filepath"

	"github.com/openfga/cli/internal/authorizationmodel"

	"gopkg.in/yaml.v3"
)

// ReadFromFile is used to read and parse the Store file.
func ReadFromFile(fileName string, basePath string) (authorizationmodel.ModelFormat, *StoreData, error) {
	format := authorizationmodel.ModelFormatDefault

	var storeData StoreData

	absFileName := fileName

	// Only join with basePath if fileName is not absolute and basePath is provided
	if !filepath.IsAbs(fileName) && basePath != "" {
		absFileName = filepath.Join(basePath, fileName)
	}

	testFile, err := os.Open(absFileName)
	if err != nil {
		return format, nil, fmt.Errorf(
			"failed to read file %q (resolved path: %q): %w",
			fileName, absFileName, err,
		)
	}
	defer testFile.Close()

	decoder := yaml.NewDecoder(testFile)
	decoder.KnownFields(true)

	err = decoder.Decode(&storeData)
	if err != nil {
		return format, nil, fmt.Errorf("failed to unmarshal file %s due to %w", fileName, err)
	}

	// Use the directory of the resolved file path for nested file references
	resolvedBasePath := filepath.Dir(absFileName)

	format, err = storeData.LoadModel(resolvedBasePath)
	if err != nil {
		return format, nil, err
	}

	err = storeData.LoadTuples(resolvedBasePath)
	if err != nil {
		return format, nil, err
	}

	if err = storeData.Validate(); err != nil {
		return format, nil, err
	}

	return format, &storeData, nil
}
