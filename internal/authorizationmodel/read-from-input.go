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

package authorizationmodel

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/clierrors"
)

func ReadFromFile(
	fileName string,
	input *string,
	format *ModelFormat,
	storeName *string,
) error {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file %s due to %w", fileName, err)
	}

	*input = string(file)

	// if the input format is set as the default, set it from the file extension (and default to fga)
	if *format == ModelFormatDefault {
		if strings.HasSuffix(fileName, "json") {
			*format = ModelFormatJSON
		} else {
			*format = ModelFormatFGA
		}
	}

	if *storeName == "" {
		*storeName = strings.TrimSuffix(path.Base(fileName), filepath.Ext(fileName))
	}

	return nil
}

func ReadFromInputFileOrArg(
	cmd *cobra.Command,
	args []string,
	fileNameArg string,
	isOptional bool,
	input *string,
	storeName *string,
	format *ModelFormat,
) error {
	fileName, err := cmd.Flags().GetString(fileNameArg)
	if err != nil {
		return fmt.Errorf("failed to parse file name due to %w", err)
	}

	switch {
	case fileName != "":
		return ReadFromFile(fileName, input, format, storeName)
	case len(args) > 0 && args[0] != "-":
		*input = args[0]
		// if the input format is set as the default, set it from the file extension (and default to fga)
		if *format == ModelFormatDefault {
			*format = ModelFormatFGA
		}
	case !isOptional:
		_ = cmd.Help() // print out the help message so users know what the command expects

		return fmt.Errorf("%w", clierrors.ErrModelInputMissing)
	}

	return nil
}

func ReadFromInputFile(
	fileName string,
	format *ModelFormat,
) (*string, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s due to %w", fileName, err)
	}

	model := string(file)

	// if the input format is set as the default, set it from the file extension (and default to fga)
	if *format == ModelFormatDefault {
		if strings.HasSuffix(fileName, "json") {
			*format = ModelFormatJSON
		} else {
			*format = ModelFormatFGA
		}
	}

	return &model, nil
}
