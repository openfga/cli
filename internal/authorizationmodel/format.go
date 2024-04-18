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

	"github.com/openfga/cli/internal/clierrors"
)

type ModelFormat string

const (
	ModelFormatDefault ModelFormat = "default"
	ModelFormatJSON    ModelFormat = "json"
	ModelFormatFGA     ModelFormat = "fga"
	ModelFormatModular ModelFormat = "modular"
)

func (format *ModelFormat) String() string {
	return string(*format)
}

func (format *ModelFormat) Set(v string) error {
	switch v {
	case "json", "fga", "modular":
		*format = ModelFormat(v)

		return nil
	default:
		return fmt.Errorf(`%w: must be one of "%v" or "%v"`, clierrors.ErrInvalidFormat, ModelFormatJSON, ModelFormatFGA)
	}
}

func (format *ModelFormat) Type() string {
	return "format"
}
