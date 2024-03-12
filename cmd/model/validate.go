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
	"context"
	"time"

	"github.com/oklog/ulid/v2"
	pb "github.com/openfga/api/proto/openfga/v1"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/openfga/pkg/typesystem"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/output"
)

type validationResult struct {
	ID        string     `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	IsValid   bool       `json:"is_valid"`
	Error     *string    `json:"error,omitempty"`
}

func validate(inputModel authorizationmodel.AuthzModel) validationResult {
	model := &pb.AuthorizationModel{}
	output := validationResult{
		IsValid: true,
	}

	modelJSONString, err := inputModel.GetAsJSONString()
	if err != nil {
		output.IsValid = false
		errorString := "unable to parse json input"
		output.Error = &errorString

		return output
	}

	err = protojson.Unmarshal([]byte(*modelJSONString), model)
	if err != nil {
		output.IsValid = false
		errorString := "unable to parse json input"
		output.Error = &errorString

		return output
	}

	if model.GetId() != "" {
		output.ID = model.GetId()

		modelID, err := ulid.Parse(output.ID)
		if err != nil {
			output.IsValid = false
			errorString := "unable to parse id: invalid ulid format"
			output.Error = &errorString

			return output
		}

		createdAt := time.Unix(int64(modelID.Time()/1_000), 0).UTC() //nolint:gomnd
		output.CreatedAt = &createdAt
	}

	if _, err = typesystem.NewAndValidate(context.Background(), model); err != nil {
		errString := err.Error()
		output.IsValid = false
		output.Error = &errString
	}

	return output
}

// validateCmd represents the validate command.
var validateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Validate Authorization Model",
	Long:    "Validates that an authorization model is valid.",
	Example: `fga model validate --file model.json`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputModel string
		if err := authorizationmodel.ReadFromInputFileOrArg(
			cmd,
			args,
			"file",
			false,
			&inputModel,
			openfga.PtrString(""),
			&validateInputFormat); err != nil {
			return err //nolint:wrapcheck
		}

		authModel := authorizationmodel.AuthzModel{}
		var err error

		if validateInputFormat == authorizationmodel.ModelFormatJSON {
			err = authModel.ReadFromJSONString(inputModel)
		} else {
			err = authModel.ReadFromDSLString(inputModel)
		}

		if err != nil {
			return err //nolint:wrapcheck
		}

		response := validate(authModel)

		return output.Display(response) //nolint:wrapcheck
	},
}

var validateInputFormat = authorizationmodel.ModelFormatDefault

func init() {
	validateCmd.Flags().String("file", "", "File Name. The file should have the model in the JSON or DSL format")
	validateCmd.Flags().Var(&validateInputFormat, "format", `Authorization model input format. Can be "fga" or "json"`)
}
