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
	"github.com/openfga/cli/internal/safefile"

	"gopkg.in/yaml.v3"
)

// ReadFromFile is used to read and parse the Store file.
//
// Files referenced from within the store YAML (model_file, tuple_file,
// tuple_files, and per-test tuple_file) are, by default, contained to the
// directory holding the store file and must be regular files. Set
// allowExternalFiles to true to permit references that resolve outside that
// directory (e.g. via "..") for trusted workflows.
func ReadFromFile(
	fileName string,
	basePath string,
	allowExternalFiles bool,
) (authorizationmodel.ModelFormat, *StoreData, error) {
	format := authorizationmodel.ModelFormatDefault

	var storeData StoreData

	absFileName := fileName

	// Only join with basePath if fileName is not absolute and basePath is provided
	if !filepath.IsAbs(fileName) && basePath != "" {
		absFileName = filepath.Join(basePath, fileName)
	}

	// Reject non-regular files before reading: a FIFO with no writer blocks the
	// read forever, and an endless device such as /dev/zero grows the read buffer
	// until the process is OOM-killed.
	if err := safefile.CheckRegular(absFileName); err != nil {
		return format, nil, fmt.Errorf("cannot read store file: %w", err)
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

	format, err = storeData.LoadModel(resolvedBasePath, allowExternalFiles)
	if err != nil {
		return format, nil, err
	}

	err = storeData.LoadTuples(resolvedBasePath, allowExternalFiles)
	if err != nil {
		return format, nil, err
	}

	if err = storeData.Validate(); err != nil {
		return format, nil, err
	}

	return format, &storeData, nil
}
