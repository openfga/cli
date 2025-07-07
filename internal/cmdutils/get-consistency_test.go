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

package cmdutils_test

import (
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openfga/cli/internal/cmdutils"
)

func TestGetConsistency(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name      string
		stringVal string
		expected  *openfga.ConsistencyPreference
		err       string
	}{
		{
			name:      "handles parsing correct value",
			stringVal: "HIGHER_CONSISTENCY",
			expected:  openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr(),
		},
		{
			name:      "handles parsing value from lowercase",
			stringVal: "higher_consistency",
			expected:  openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY.Ptr(),
		},
		{
			name:      "handles no value",
			stringVal: "",
			expected:  openfga.CONSISTENCYPREFERENCE_UNSPECIFIED.Ptr(),
		},
		{
			name:      "throws for unknown values",
			stringVal: "invalid",
			err:       "invalid value 'invalid' for consistency",
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			consistency, err := cmdutils.ParseConsistency(test.stringVal)
			if err == nil {
				assert.Equal(t, test.expected, consistency)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.err)
			}
		})
	}
}
