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

package cmdutils

import (
	"fmt"
	"strings"

	openfga "github.com/openfga/go-sdk"
	"github.com/spf13/cobra"
)

func ParseConsistency(consistency string) (*openfga.ConsistencyPreference, error) {
	if consistency == "" {
		return openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(), nil
	}

	val := openfga.ConsistencyPreference(consistency)
	if val.IsValid() {
		return &val, nil
	}

	val = openfga.ConsistencyPreference(strings.ToUpper(consistency))
	if val.IsValid() {
		return &val, nil
	}

	return nil, fmt.Errorf( //nolint:err113
		"invalid value '%s' for consistency. Valid values are HIGHER_CONSISTENCY and MINIMIZE_LATENCY",
		consistency,
	)
}

func ParseConsistencyFromCmd(cmd *cobra.Command) (*openfga.ConsistencyPreference, error) {
	consistency, err := cmd.Flags().GetString("consistency")
	if err != nil {
		return nil, fmt.Errorf("failed to parse consistency due to %w", err)
	}

	return ParseConsistency(consistency)
}
