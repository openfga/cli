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

package model

import (
	"fmt"

	openfga "github.com/openfga/go-sdk"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/output"
)

// transformCmd represents the transform command.
var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transforms an authorization model",
	Example: `fga model transform --file=model.json
fga model transform --file=model.fga
fga model transform '{ "schema_version": "1.1", "type_definitions":[{"type":"user"}] }' --input-format json`,

	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputModel string
		if err := authorizationmodel.ReadFromInputFileOrArg(
			cmd,
			args,
			"file",
			false,
			&inputModel,
			openfga.PtrString(""),
			&transformInputFormat); err != nil {
			return err //nolint:wrapcheck
		}

		authModel := authorizationmodel.AuthzModel{}
		if err := authModel.ReadModelFromString(inputModel, transformInputFormat); err != nil {
			return err //nolint:wrapcheck
		}

		if transformInputFormat == authorizationmodel.ModelFormatJSON {
			dslModel, err := authModel.DisplayAsDSL([]string{"model"})
			if err != nil {
				return fmt.Errorf("failed to transform model due to %w", err)
			}
			fmt.Printf("%v", *dslModel)

			return nil
		}

		if err := authModel.ReadFromDSLString(inputModel); err != nil {
			return err //nolint:wrapcheck
		}

		return output.Display(authModel.DisplayAsJSON([]string{"model"})) //nolint:wrapcheck
	},
}

var transformInputFormat = authorizationmodel.ModelFormatDefault

func init() {
	transformCmd.Flags().String("file", "", "File Name. The file should have the model in the JSON or DSL format")
	transformCmd.Flags().Var(&transformInputFormat, "input-format", `Authorization model input format. Can be "fga" or "json"`) //nolint:lll
}
