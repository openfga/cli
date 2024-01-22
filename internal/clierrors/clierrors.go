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

// Package clierrors contains errors used throughout this package
package clierrors

import (
	"errors"
	"fmt"
)

var (
	ErrValidation                 = errors.New("validation error")
	ErrInvalidFormat              = errors.New("invalid format")
	ErrStoreNotFound              = errors.New("store not found")
	ErrAuthorizationModelNotFound = errors.New("authorization model not found")
	ErrModelInputMissing          = errors.New("model input not provided")
	ErrRequiredCsvHeaderMissing   = errors.New("csv header missing")
)

func ValidationError(op string, details string) error {
	return fmt.Errorf("%w - %s: %s", ErrValidation, op, details)
}

func MissingRequiredCsvHeaderError(headerName string) error {
	return fmt.Errorf("%w (\"%s\")", ErrRequiredCsvHeaderMissing, headerName)
}
