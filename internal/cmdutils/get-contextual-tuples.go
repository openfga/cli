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
	"github.com/openfga/go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/clierrors"
)

func ParseContextualTuplesInner(contextualTuplesArray []string) ([]client.ClientContextualTupleKey, error) {
	contextualTuples := []client.ClientContextualTupleKey{}

	if len(contextualTuplesArray) > 0 {
		for index := 0; index < len(contextualTuplesArray); index++ {
			tuple := strings.Split(contextualTuplesArray[index], " ")
			if len(tuple) != 3 && len(tuple) != 4 {
				return contextualTuples,
					clierrors.ValidationError( //nolint:wrapcheck
						"ParseContextualTuplesInner",
						"Failed to parse contextual tuples, "+
							"they must be of the format \"user relation object\" or \"user relation object condition\", "+
							"where condition is a JSON string of the form { name, context }")
			}

			var condition *openfga.RelationshipCondition

			if len(tuple) == 4 { //nolint:gomnd
				cond, err := ParseTupleConditionString(tuple[3])
				if err != nil {
					return nil, fmt.Errorf("failed to parse condition due to %w", err)
				}

				condition = cond
			}

			contextualTuples = append(contextualTuples, client.ClientContextualTupleKey{
				User:      tuple[0],
				Relation:  tuple[1],
				Object:    tuple[2],
				Condition: condition,
			})
		}
	}

	return contextualTuples, nil
}

func ParseContextualTuples(cmd *cobra.Command) ([]client.ClientContextualTupleKey, error) {
	contextualTuplesArray, _ := cmd.Flags().GetStringArray("contextual-tuple")

	return ParseContextualTuplesInner(contextualTuplesArray)
}
