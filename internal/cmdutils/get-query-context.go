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

	"github.com/spf13/cobra"
)

func ParseQueryContextInner(contextString string) (*map[string]interface{}, error) {
	queryContext := map[string]interface{}{}

	if contextString == "" {
		return &queryContext, nil
	}

	data := []byte(contextString)
	if err := json.Unmarshal(data, &queryContext); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &queryContext, nil
}

func ParseQueryContext(cmd *cobra.Command, queryParamName string) (*map[string]interface{}, error) {
	contextString, _ := cmd.Flags().GetString(queryParamName)

	return ParseQueryContextInner(contextString)
}
