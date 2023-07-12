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
	"strings"

	"github.com/openfga/cli/internal/clierrors"
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
						"they must be of the format\"user relation object\" ")
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
