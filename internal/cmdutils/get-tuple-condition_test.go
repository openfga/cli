package cmdutils

import (
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTupleConditionCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("condition-name", "", "")
	cmd.Flags().String("condition-context", "", "")
	cmd.Flags().String("condition-expression", "", "")
	cmd.Flags().String("condition-parameters", "", "")
	return cmd
}

func TestParseTupleConditionExpression(t *testing.T) {
	t.Parallel()

	cmd := newTupleConditionCommand()
	require.NoError(t, cmd.Flags().Set("condition-expression", "channel_name == '#product-announcements'"))
	require.NoError(t, cmd.Flags().Set("condition-parameters", `{"channel_name":"string"}`))

	condition, err := ParseTupleCondition(cmd)
	require.NoError(t, err)
	assert.Equal(t, &openfga.RelationshipCondition{
		Name: "$expression",
		Context: &map[string]any{
			"expression": "channel_name == '#product-announcements'",
			"parameters": map[string]any{"channel_name": "string"},
		},
	}, condition)
}

func TestParseTupleConditionExpressionFlagValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name: "parameters must be valid JSON",
			flags: map[string]string{
				"condition-expression": "true",
				"condition-parameters": "not-json",
			},
			expectedError: "error parsing condition parameters",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cmd := newTupleConditionCommand()
			for name, value := range test.flags {
				require.NoError(t, cmd.Flags().Set(name, value))
			}

			_, err := ParseTupleCondition(cmd)
			require.ErrorContains(t, err, test.expectedError)
		})
	}
}
