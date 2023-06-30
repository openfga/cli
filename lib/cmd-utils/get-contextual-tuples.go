package cmdutils

import (
	"fmt"
	"strings"

	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"
)

func ParseContextualTuples(cmd *cobra.Command) ([]client.ClientTupleKey, error) {
	contextualTuples := []client.ClientTupleKey{}
	contextualTuplesArray, _ := cmd.Flags().GetStringArray("contextual-tuple")

	if len(contextualTuplesArray) > 0 {
		for index := 0; index < len(contextualTuplesArray); index++ {
			tuple := strings.Split(contextualTuplesArray[index], " ")
			if len(tuple) != 3 {
				return contextualTuples,
					fmt.Errorf("Failed to parse contextual tuples, they must be of the format\"user_type:user_id relation object_type:object_id\"")
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
