package storetest

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openfga/cli/internal/authorizationmodel"
)

// buildModelWithNTypes returns a valid FGA model DSL string with n type definitions.
func buildModelWithNTypes(n int) string {
	var builder strings.Builder

	builder.WriteString("model\n  schema 1.1\n\ntype user\n")

	for i := 1; i < n; i++ {
		fmt.Fprintf(&builder, "\ntype resource%d\n  relations\n    define owner: [user]\n", i)
	}

	return builder.String()
}

func TestGetLocalServerModelAndTuples_MaxTypesLimit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		numTypes    int
		maxTypes    int
		expectError bool
	}{
		{
			name:        "model within default limit succeeds",
			numTypes:    6,
			maxTypes:    100,
			expectError: false,
		},
		{
			name:        "model exceeding custom limit fails",
			numTypes:    6,
			maxTypes:    5,
			expectError: true,
		},
		{
			name:        "model within custom limit succeeds",
			numTypes:    6,
			maxTypes:    10,
			expectError: false,
		},
		{
			name:        "model at exact custom limit succeeds",
			numTypes:    5,
			maxTypes:    5,
			expectError: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			model := buildModelWithNTypes(testCase.numTypes)
			storeData := &StoreData{Model: model}
			config := LocalServerConfig{MaxTypesPerAuthorizationModel: testCase.maxTypes}

			fgaServer, authModel, stopFn, err := getLocalServerModelAndTuples(
				storeData, authorizationmodel.ModelFormatDefault, config,
			)
			require.NoError(t, err)

			defer stopFn()

			assert.NotNil(t, fgaServer)
			assert.NotNil(t, authModel)

			// Try writing the model to the embedded server — this is where the limit is enforced
			_, _, err = initLocalStore(context.Background(), fgaServer, authModel.GetProtoModel(), nil)

			if testCase.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "exceeds the allowed limit")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
