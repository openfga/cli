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
	"github.com/openfga/cli/internal/clierrors"
	"github.com/openfga/cli/internal/output"
)

type validationResult struct {
	ID        string     `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	IsValid   bool       `json:"is_valid"`
	Error     *string    `json:"error,omitempty"`
	SizeKB    *float64   `json:"size_kb,omitempty"`
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

	err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(*modelJSONString), model)
	if err != nil {
		output.IsValid = false
		errorString := "unable to parse json input"
		output.Error = &errorString

		return output
	}

	sizeKB := authorizationmodel.ProtoModelSizeInKB(model)
	output.SizeKB = &sizeKB

	if model.GetId() != "" {
		output.ID = model.GetId()

		modelID, err := ulid.Parse(output.ID)
		if err != nil {
			output.IsValid = false
			errorString := "unable to parse id: invalid ulid format"
			output.Error = &errorString

			return output
		}

		createdAt := ulid.Time(modelID.Time()).UTC()
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
	Use:   "validate",
	Short: "Validate Authorization Model",
	Long:  "Validates that an authorization model is valid. When the input parses successfully, the JSON response includes size_kb, the protobuf-serialized size of the model in KB.",
	Example: `fga model validate --file model.json`,
	Annotations: map[string]string{
		"docs:response": `{"id":"01GPGWB8R33HWXS3KK6YG4ETGH","created_at":"2023-01-11T16:59:22Z","is_valid":true,"size_kb":0.05}

Invalid model:
{"id":"01GPGTVEH5NYTQ19RYFQKE0Q4Z","created_at":"2023-01-11T16:33:15Z","is_valid":false,"error":"invalid schema version","size_kb":0.05}`,
	},
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

		err := authModel.ReadModelFromString(inputModel, validateInputFormat)
		if err != nil {
			return err //nolint:wrapcheck
		}

		response := validate(authModel)

		err = output.Display(response)
		if err != nil {
			return err //nolint:wrapcheck
		}

		// Return an error if validation failed to ensure non-zero exit code
		if !response.IsValid {
			return clierrors.ValidationError("validate", *response.Error)
		}

		return nil
	},
}

var validateInputFormat = authorizationmodel.ModelFormatDefault

func init() {
	validateCmd.Flags().String("file", "", "File Name. The file should have the model in the JSON or DSL format or be an fga.mod file") //nolint:lll
	validateCmd.Flags().Var(&validateInputFormat, "format", `Authorization model input format. Can be "fga", "json", or "modular"`)     //nolint:lll
}
