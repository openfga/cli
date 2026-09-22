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
	"encoding/json"
	"fmt"

	openfga "github.com/openfga/go-sdk"
	"github.com/spf13/cobra"
)

func ParseTupleConditionString(conditionString string) (*openfga.RelationshipCondition, error) {
	condition := openfga.RelationshipCondition{}
	if conditionString == "" {
		return &condition, nil
	}

	data := []byte(conditionString)
	if err := json.Unmarshal(data, &condition); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &condition, nil
}

func ParseTupleCondition(cmd *cobra.Command) (*openfga.RelationshipCondition, error) {
	var condition *openfga.RelationshipCondition

	if cmd.Flags().Lookup("condition-expression") != nil {
		conditionExpression, err := cmd.Flags().GetString("condition-expression")
		if err != nil {
			return nil, fmt.Errorf("failed to parse condition expression due to %w", err)
		}

		conditionParameters, err := cmd.Flags().GetString("condition-parameters")
		if err != nil {
			return nil, fmt.Errorf("failed to parse condition parameters due to %w", err)
		}

		if cmd.Flags().Changed("condition-expression") {
			conditionContext := map[string]any{"expression": conditionExpression}
			if cmd.Flags().Changed("condition-parameters") && conditionParameters != "" {
				parameters, err := ParseQueryContextInner(conditionParameters)
				if err != nil {
					return nil, fmt.Errorf("error parsing condition parameters: %w", err)
				}
				conditionContext["parameters"] = *parameters
			}

			return &openfga.RelationshipCondition{
				Name:    "$expression",
				Context: &conditionContext,
			}, nil
		}
	}

	conditionName, err := cmd.Flags().GetString("condition-name")
	if err != nil {
		return nil, fmt.Errorf("failed to parse condition name due to %w", err)
	}

	if conditionName != "" {
		conditionContext, err := ParseQueryContext(cmd, "condition-context")
		if err != nil {
			return nil, fmt.Errorf("error parsing condition context: %w", err)
		}

		condition = &openfga.RelationshipCondition{
			Name:    conditionName,
			Context: conditionContext,
		}
	}

	return condition, err //nolint:wrapcheck
}
