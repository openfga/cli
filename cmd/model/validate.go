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
	"github.com/openfga/cli/lib/output"
	"github.com/openfga/openfga/pkg/typesystem"
	"github.com/spf13/cobra"
	pb "go.buf.build/openfga/go/openfga/api/openfga/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type validationResult struct {
	ID        string     `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	IsValid   bool       `json:"is_valid"`
	Error     *string    `json:"error,omitempty"`
}

func validate(inputModel string) validationResult {
	model := &pb.AuthorizationModel{}
	output := validationResult{
		IsValid: true,
	}

	err := protojson.Unmarshal([]byte(inputModel), model)
	if err != nil {
		output.IsValid = false
		errorString := "unable to parse json input"
		output.Error = &errorString

		return output
	}

	if model.Id != "" {
		output.ID = model.Id

		modelID, err := ulid.Parse(model.Id)
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
	Use:   "validate",
	Short: "Validates an authorization model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		response := validate(args[0])

		return output.Display(response) //nolint:wrapcheck
	},
}

func init() {
}
