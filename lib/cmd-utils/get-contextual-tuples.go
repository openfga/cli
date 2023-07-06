package cmdutils

import (
	"strings"

	"github.com/openfga/cli/lib/clierrors"
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

func ParseContextualTuplesInner(contextualTuplesArray []string) ([]client.ClientTupleKey, error) {
	contextualTuples := []client.ClientTupleKey{}

	if len(contextualTuplesArray) > 0 {
		for index := 0; index < len(contextualTuplesArray); index++ {
			tuple := strings.Split(contextualTuplesArray[index], " ")
			if len(tuple) != 3 { //nolint:gomnd
				return contextualTuples,
					clierrors.ValidationError("ParseContextualTuplesInner", "Failed to parse contextual tuples, "+ //nolint:wrapcheck
						"they must be of the output\"user relation object\" ")
			}

			contextualTuples = append(contextualTuples, client.ClientTupleKey{
				User:     tuple[0],
				Relation: tuple[1],
				Object:   tuple[2],
			})
		}
	}

	return contextualTuples, nil
}

func ParseContextualTuples(cmd *cobra.Command) ([]client.ClientTupleKey, error) {
	contextualTuplesArray, _ := cmd.Flags().GetStringArray("contextual-tuple")

	return ParseContextualTuplesInner(contextualTuplesArray)
}
