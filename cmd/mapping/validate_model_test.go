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

package mapping

import (
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/stretchr/testify/assert"
)

func typeDefWithRefs(refs []openfga.RelationReference) *openfga.TypeDefinition {
	return &openfga.TypeDefinition{
		Type: "obj",
		Metadata: &openfga.Metadata{
			Relations: &map[string]openfga.RelationMetadata{
				"rel": {DirectlyRelatedUserTypes: &refs},
			},
		},
	}
}

func TestIsValidAssigneeCondition(t *testing.T) {
	t.Parallel()

	condName := "non_expired"

	t.Run("direct type ref with required condition", func(t *testing.T) {
		t.Parallel()

		typeDef := typeDefWithRefs([]openfga.RelationReference{
			{Type: "user", Condition: &condName},
		})

		assert.True(t, isValidAssignee(typeDef, "rel", "user", "", condName), "matching condition should be valid")
		assert.False(t, isValidAssignee(typeDef, "rel", "user", "", ""), "missing condition should be invalid")
		assert.False(t, isValidAssignee(typeDef, "rel", "user", "", "other"), "wrong condition should be invalid")
	})

	t.Run("direct type ref with no condition required", func(t *testing.T) {
		t.Parallel()

		typeDef := typeDefWithRefs([]openfga.RelationReference{
			{Type: "user"},
		})

		assert.True(t, isValidAssignee(typeDef, "rel", "user", "", ""), "no condition required accepts empty condition")
		assert.True(t, isValidAssignee(typeDef, "rel", "user", "", condName), "no condition required accepts any condition")
	})

	t.Run("wildcard ref with required condition", func(t *testing.T) {
		t.Parallel()

		wildcard := map[string]any{}
		typeDef := typeDefWithRefs([]openfga.RelationReference{
			{Type: "user", Wildcard: &wildcard, Condition: &condName},
		})

		assert.True(t, isValidAssignee(typeDef, "rel", "user", "*", condName), "matching condition should be valid")
		assert.False(t, isValidAssignee(typeDef, "rel", "user", "*", ""), "missing condition should be invalid")
	})

	t.Run("relation ref with required condition", func(t *testing.T) {
		t.Parallel()

		typeDef := typeDefWithRefs([]openfga.RelationReference{
			{Type: "group", Relation: openfga.PtrString("member"), Condition: &condName},
		})

		assert.True(t, isValidAssignee(typeDef, "rel", "group", "member", condName), "matching condition should be valid")
		assert.False(t, isValidAssignee(typeDef, "rel", "group", "member", ""), "missing condition should be invalid")
	})
}
