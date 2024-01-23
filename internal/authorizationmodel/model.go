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
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	pb "github.com/openfga/api/proto/openfga/v1"
	openfga "github.com/openfga/go-sdk"
	language "github.com/openfga/language/pkg/go/transformer"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/openfga/cli/internal/slices"
)

func getCreatedAtFromModelID(id string) (*time.Time, error) {
	modelID, err := ulid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("error parsing model id %w", err)
	}

	createdAt := time.Unix(int64(modelID.Time()/1_000), 0).UTC() //nolint:gomnd

	return &createdAt, nil
}

type AuthzModelList struct {
	AuthorizationModels []AuthzModel `json:"authorization_models"`
}

type AuthzModel struct {
	ID              *string                       `json:"id,omitempty"`
	CreatedAt       *time.Time                    `json:"created_at,omitempty"`
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

	if err = protojson.Unmarshal([]byte(*jsonModel), &pbModel); err != nil {
		return nil
	}

	return &pbModel
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
	case ModelFormatFGA:
	case ModelFormatDefault:
		if err := model.ReadFromDSLString(input); err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (model *AuthzModel) setCreatedAt() {
	if *model.ID != "" {
		modelID, err := ulid.Parse(*model.ID)
		if err == nil {
			createdAt := time.Unix(int64(modelID.Time()/1_000), 0).UTC() //nolint:gomnd
			model.CreatedAt = &createdAt
		}
	}
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

	return newModel
}

func (model *AuthzModel) DisplayAsDSL(fields []string) (*string, error) {
	modelPb := pb.AuthorizationModel{}

	if len(fields) < 1 {
		fields = append(fields, "model")
	}

	dslModel := ""

	if slices.Contains(fields, "id") {
		if model.ID != nil {
			dslModel += fmt.Sprintf("# Model ID: %v\n", *model.ID)
		} else {
			dslModel += fmt.Sprintf("# Model ID: %v\n", "N/A")
		}
	}

	if slices.Contains(fields, "created_at") {
		if model.CreatedAt != nil {
			dslModel += fmt.Sprintf("# Created At: %v\n", *model.CreatedAt)
		} else {
			dslModel += fmt.Sprintf("# Created At: %v\n", "N/A")
		}
	}

	if slices.Contains(fields, "model") {
		modelJSON, err := model.GetAsJSONString()
		if err != nil {
			return nil, err
		}

		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal([]byte(*modelJSON), &modelPb)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal model json string due to: %w", err)
		}

		transformedJSON, err := language.TransformJSONProtoToDSL(&modelPb)
		if err != nil {
			return nil, fmt.Errorf("error transforming from JSON due to: %w", err)
		}

		dslModel += fmt.Sprintf("%v\n", transformedJSON)
	}

	return &dslModel, nil
}
