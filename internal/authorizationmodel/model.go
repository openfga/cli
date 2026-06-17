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

package authorizationmodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"time"

	"github.com/oklog/ulid/v2"
	pb "github.com/openfga/api/proto/openfga/v1"
	openfga "github.com/openfga/go-sdk"
	language "github.com/openfga/language/pkg/go/transformer"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/openfga/cli/internal/slices"
)

func getCreatedAtFromModelID(id string) (*time.Time, error) {
	modelID, err := ulid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("error parsing model id %w", err)
	}

	createdAt := ulid.Time(modelID.Time()).UTC()

	return &createdAt, nil
}

type AuthzModelList struct {
	AuthorizationModels []AuthzModel `json:"authorization_models"`
}

type AuthzModel struct {
	ID              *string                       `json:"id,omitempty"`
	CreatedAt       *time.Time                    `json:"created_at,omitempty"`
	SizeKB          *float64                      `json:"size_kb,omitempty"`
	SchemaVersion   *string                       `json:"schema_version,omitempty"`
	TypeDefinitions *[]openfga.TypeDefinition     `json:"type_definitions,omitempty"`
	Conditions      map[string]*openfga.Condition `json:"conditions,omitempty"`
}

func (model *AuthzModel) GetID() string {
	if model == nil || model.ID == nil {
		var ret string

		return ret
	}

	return *model.ID
}

func (model *AuthzModel) GetSchemaVersion() string {
	if model == nil || model.SchemaVersion == nil {
		var ret string

		return ret
	}

	return *model.SchemaVersion
}

func (model *AuthzModel) GetTypeDefinitions() []openfga.TypeDefinition {
	if model == nil || model.TypeDefinitions == nil {
		var ret []openfga.TypeDefinition

		return ret
	}

	return *model.TypeDefinitions
}

func (model *AuthzModel) GetConditions() *map[string]openfga.Condition {
	conditions := make(map[string]openfga.Condition)

	if model == nil || model.Conditions == nil {
		return &conditions
	}

	for conditionName, condition := range model.Conditions {
		conditions[conditionName] = *condition
	}

	return &conditions
}

func (model *AuthzModel) GetProtoModel() *pb.AuthorizationModel {
	if model == nil {
		return nil
	}

	var pbModel pb.AuthorizationModel

	jsonModel, err := model.GetAsJSONString()
	if err != nil {
		return nil
	}

	if err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(*jsonModel), &pbModel); err != nil {
		return nil
	}

	return &pbModel
}

func (model *AuthzModel) GetSizeInKB() float64 {
	pbModel := model.GetProtoModel()
	if pbModel == nil {
		return 0
	}

	sizeKB := float64(proto.Size(pbModel)) / 1024.0 //nolint:mnd

	return math.Round(sizeKB*100) / 100 //nolint:mnd
}

func (model *AuthzModel) GetCreatedAt() *time.Time {
	if model == nil {
		return nil
	}

	if model.CreatedAt != nil {
		return model.CreatedAt
	}

	if model.ID != nil {
		createdAt, _ := getCreatedAtFromModelID(model.GetID())

		model.CreatedAt = createdAt

		return createdAt
	}

	return nil
}

func (model *AuthzModel) Set(authzModel openfga.AuthorizationModel) {
	model.ID = &authzModel.Id
	model.SchemaVersion = &authzModel.SchemaVersion
	model.TypeDefinitions = &authzModel.TypeDefinitions

	if model.ID != nil {
		model.setCreatedAt()
	}

	conditions := authzModel.GetConditions()
	if len(conditions) > 0 {
		model.Conditions = make(map[string]*openfga.Condition, len(conditions))

		for k, v := range conditions {
			condition := v
			model.Conditions[k] = &condition
		}
	}
}

func (model *AuthzModel) ReadFromJSONString(jsonString string) error {
	jsonAuthModel := &openfga.AuthorizationModel{}

	err := json.Unmarshal([]byte(jsonString), jsonAuthModel)
	if err != nil {
		return fmt.Errorf("failed to parse input as json due to %w", err)
	}

	model.Set(*jsonAuthModel)

	return nil
}

func (model *AuthzModel) ReadFromDSLString(dslString string) error {
	parsedAuthModel, err := language.TransformDSLToProto(dslString)
	if err != nil {
		return fmt.Errorf("failed to transform due to %w", err)
	}

	bytes, err := protojson.Marshal(parsedAuthModel)
	if err != nil {
		return fmt.Errorf("failed to transform due to %w", err)
	}

	jsonAuthModel := openfga.AuthorizationModel{}

	err = json.Unmarshal(bytes, &jsonAuthModel)
	if err != nil {
		return fmt.Errorf("failed to transform due to %w", err)
	}

	model.Set(jsonAuthModel)

	return nil
}

func (model *AuthzModel) ReadModelFromModFGA(modFile string) error {
	modFileContents, err := os.ReadFile(modFile)
	if err != nil {
		return fmt.Errorf("failed to read fga.mod file due to %w", err)
	}

	parsedModFile, err := language.TransformModFile(string(modFileContents))
	if err != nil {
		return fmt.Errorf("failed to transform fga.mod file due to %w", err)
	}

	moduleFiles := []language.ModuleFile{}

	var fileReadErrors []error

	directory := path.Dir(modFile)

	for _, fileName := range parsedModFile.Contents.Value {
		filePath := path.Join(directory, fileName.Value)

		fileContents, err := os.ReadFile(filePath)
		if err != nil {
			fileReadErrors = append(
				fileReadErrors,
				fmt.Errorf("failed to read module file %s due to %w", fileName.Value, err),
			)

			continue
		}

		moduleFiles = append(moduleFiles, language.ModuleFile{
			Name:     fileName.Value,
			Contents: string(fileContents),
		})
	}

	if len(fileReadErrors) != 0 {
		return errors.Join(fileReadErrors...)
	}

	parsedAuthModel, err := language.TransformModuleFilesToModel(moduleFiles, parsedModFile.Schema.Value)
	if err != nil {
		return fmt.Errorf("failed to transform module to model due to %w", err)
	}

	bytes, err := protojson.Marshal(parsedAuthModel)
	if err != nil {
		return fmt.Errorf("failed to transform due to %w", err)
	}

	jsonAuthModel := openfga.AuthorizationModel{}

	err = json.Unmarshal(bytes, &jsonAuthModel)
	if err != nil {
		return fmt.Errorf("failed to transform due to %w", err)
	}

	model.Set(jsonAuthModel)

	return nil
}

func (model *AuthzModel) ReadModelFromString(input string, format ModelFormat) error {
	if input == "" {
		return nil
	}

	switch format {
	case ModelFormatJSON:
		if err := model.ReadFromJSONString(input); err != nil {
			return err
		}

		return nil
	case ModelFormatFGA, ModelFormatDefault:
		if err := model.ReadFromDSLString(input); err != nil {
			return err
		}

		return nil
	case ModelFormatModular:
		if err := model.ReadModelFromModFGA(input); err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (model *AuthzModel) GetAsJSONString() (*string, error) {
	bytes, err := json.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal due to %w", err)
	}

	jsonString := string(bytes)

	return &jsonString, nil
}

func (model *AuthzModel) DisplayAsJSON(fields []string) AuthzModel {
	newModel := AuthzModel{}

	if len(fields) < 1 {
		fields = append(fields, "model")
	}

	if slices.Contains(fields, "id") {
		newModel.ID = model.ID
	}

	if slices.Contains(fields, "created_at") {
		newModel.CreatedAt = model.CreatedAt
	}

	if slices.Contains(fields, "model") {
		newModel.SchemaVersion = model.SchemaVersion
		newModel.TypeDefinitions = model.TypeDefinitions
		newModel.Conditions = model.Conditions
	}

	if slices.Contains(fields, "size") {
		size := model.GetSizeInKB()
		newModel.SizeKB = &size
	}

	return newModel
}

func (model *AuthzModel) DisplayAsDSL(fields []string) (*string, error) {
	modelPb := pb.AuthorizationModel{}

	if len(fields) < 1 {
		fields = append(fields, "model")
	}

	dslModel := model.buildDSLMetadata(fields)

	if slices.Contains(fields, "model") {
		modelJSON, err := model.GetAsJSONString()
		if err != nil {
			return nil, err
		}

		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal([]byte(*modelJSON), &modelPb)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal model json string due to: %w", err)
		}

		transformedJSON, err := language.TransformJSONProtoToDSL(&modelPb, language.WithIncludeSourceInformation(true))
		if err != nil {
			return nil, fmt.Errorf("error transforming from JSON due to: %w", err)
		}

		dslModel += fmt.Sprintf("%v\n", transformedJSON)
	}

	return &dslModel, nil
}

func (model *AuthzModel) buildDSLMetadata(fields []string) string {
	metadata := ""

	if slices.Contains(fields, "id") {
		if model.ID != nil {
			metadata += fmt.Sprintf("# Model ID: %v\n", *model.ID)
		} else {
			metadata += fmt.Sprintf("# Model ID: %v\n", "N/A")
		}
	}

	if slices.Contains(fields, "created_at") {
		if model.CreatedAt != nil {
			metadata += fmt.Sprintf("# Created At: %v\n", *model.CreatedAt)
		} else {
			metadata += fmt.Sprintf("# Created At: %v\n", "N/A")
		}
	}

	if slices.Contains(fields, "size") {
		metadata += fmt.Sprintf("# Size: %.2f KB\n", model.GetSizeInKB())
	}

	return metadata
}

func (model *AuthzModel) setCreatedAt() {
	if *model.ID != "" {
		modelID, err := ulid.Parse(*model.ID)
		if err == nil {
			createdAt := ulid.Time(modelID.Time()).UTC()
			model.CreatedAt = &createdAt
		}
	}
}
