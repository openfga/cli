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
	"errors"
	"fmt"
	"maps"
	"strings"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/mapper"

	"github.com/openfga/cli/internal/authorizationmodel"
)

var (
	errTypeNotInModel        = errors.New("type not found in authorization model")
	errTypeHasNoRelations    = errors.New("type has no relations defined")
	errRelationNotOnType     = errors.New("relation does not exist on type")
	errUserTypeNotAssignable = errors.New("user type is not a valid assignee")
	errConditionNotInModel   = errors.New("condition not found in authorization model")
	errContextKeyNotParam    = errors.New("context key is not a parameter of condition")
)

// ruleCheckResult holds the outcome of validating one rule against a model.
type ruleCheckResult struct {
	Name string
	Err  error
}

// modelIndex is a pre-built lookup structure for an authorization model.
type modelIndex struct {
	types      map[string]*openfga.TypeDefinition
	conditions map[string]openfga.Condition
}

func buildModelIndex(model *authorizationmodel.AuthzModel) *modelIndex {
	typeDefs := model.GetTypeDefinitions()
	idx := &modelIndex{
		types:      make(map[string]*openfga.TypeDefinition, len(typeDefs)),
		conditions: make(map[string]openfga.Condition),
	}

	for i := range typeDefs {
		typeDef := &typeDefs[i]
		idx.types[typeDef.Type] = typeDef
	}

	if conds := model.GetConditions(); conds != nil {
		maps.Copy(idx.conditions, *conds)
	}

	return idx
}

// loadAuthzModel reads and parses an authorization model from path.
// Supports DSL (.fga), JSON (.json), and modular (.fga.mod) formats.
func loadAuthzModel(path string) (*authorizationmodel.AuthzModel, error) {
	var (
		input     string
		format    = authorizationmodel.ModelFormatDefault
		storeName string
	)

	if err := authorizationmodel.ReadFromFile(path, &input, &format, &storeName); err != nil {
		return nil, fmt.Errorf("reading model %s: %w", path, err)
	}

	var model authorizationmodel.AuthzModel
	if err := model.ReadModelFromString(input, format); err != nil {
		return nil, fmt.Errorf("parsing model %s: %w", path, err)
	}

	return &model, nil
}

// validateRulesAgainstModel checks that every rule's tuple templates are consistent
// with the authorization model. Fields that are fully interpolated (type cannot be
// determined statically) are skipped. Results are returned per-rule so callers can
// display per-rule status.
func validateRulesAgainstModel(rules []mapper.RuleSummary, model *authorizationmodel.AuthzModel) []ruleCheckResult {
	idx := buildModelIndex(model)
	results := make([]ruleCheckResult, len(rules))

	for ruleIdx, rule := range rules {
		var errs []error

		for _, tmpl := range rule.Tuples {
			if err := validateTupleTemplate(idx, rule.Name, tmpl); err != nil {
				errs = append(errs, err)
			}

			if err := validateTupleCondition(idx, rule.Name, tmpl); err != nil {
				errs = append(errs, err)
			}
		}

		for _, tmpl := range rule.IteratorTuples {
			if err := validateTupleTemplate(idx, rule.Name, tmpl); err != nil {
				errs = append(errs, err)
			}

			if err := validateTupleCondition(idx, rule.Name, tmpl); err != nil {
				errs = append(errs, err)
			}
		}

		for _, filterTmpl := range rule.TupleFilters {
			if err := validateTupleFilterTemplate(idx, rule.Name, filterTmpl); err != nil {
				errs = append(errs, err)
			}
		}

		results[ruleIdx] = ruleCheckResult{Name: rule.Name, Err: errors.Join(errs...)}
	}

	return results
}

func validateTupleTemplate(idx *modelIndex, ruleName string, tmpl mapper.TupleTemplate) error {
	objectType, _, objectOK := extractTypeFromField(tmpl.Object)
	if !objectOK {
		return nil
	}

	typeDef, exists := idx.types[objectType]
	if !exists {
		return fmt.Errorf("rule %q: type %q: %w", ruleName, objectType, errTypeNotInModel)
	}

	relation := tmpl.Relation
	if strings.Contains(relation, "{{") {
		return nil
	}

	relations := typeDef.GetRelations()
	if len(relations) == 0 {
		return fmt.Errorf("rule %q: type %q: %w", ruleName, objectType, errTypeHasNoRelations)
	}

	if _, relExists := relations[relation]; !relExists {
		return fmt.Errorf("rule %q: relation %q on type %q: %w", ruleName, relation, objectType, errRelationNotOnType)
	}

	userType, userRel, userOK := extractTypeFromField(tmpl.User)
	if !userOK {
		return nil
	}

	if !isValidAssignee(typeDef, relation, userType, userRel) {
		userDesc := userType
		if userRel != "" && userRel != "*" {
			userDesc = userType + "#" + userRel
		}

		return fmt.Errorf("rule %q: user type %q for relation %q on type %q: %w",
			ruleName, userDesc, relation, objectType, errUserTypeNotAssignable)
	}

	return nil
}

func validateTupleFilterTemplate( //nolint:cyclop
	idx *modelIndex,
	ruleName string,
	filterTmpl mapper.TupleFilterTemplate,
) error {
	if filterTmpl.Object == "" {
		return nil
	}

	objectType, _, objectOK := extractTypeFromField(filterTmpl.Object)
	if !objectOK {
		return nil
	}

	typeDef, exists := idx.types[objectType]
	if !exists {
		return fmt.Errorf("rule %q tuple_filter: type %q: %w", ruleName, objectType, errTypeNotInModel)
	}

	// Empty relation is a wildcard; interpolated relation can't be validated statically.
	if filterTmpl.Relation == "" || strings.Contains(filterTmpl.Relation, "{{") {
		return nil
	}

	relations := typeDef.GetRelations()
	if len(relations) == 0 {
		return fmt.Errorf("rule %q tuple_filter: type %q: %w", ruleName, objectType, errTypeHasNoRelations)
	}

	if _, relExists := relations[filterTmpl.Relation]; !relExists {
		return fmt.Errorf("rule %q tuple_filter: relation %q on type %q: %w",
			ruleName, filterTmpl.Relation, objectType, errRelationNotOnType)
	}

	if filterTmpl.User == "" {
		return nil
	}

	userType, userRel, userOK := extractTypeFromField(filterTmpl.User)
	if !userOK {
		return nil
	}

	if !isValidAssignee(typeDef, filterTmpl.Relation, userType, userRel) {
		userDesc := userType
		if userRel != "" && userRel != "*" {
			userDesc = userType + "#" + userRel
		}

		return fmt.Errorf("rule %q tuple_filter: user type %q for relation %q on type %q: %w",
			ruleName, userDesc, filterTmpl.Relation, objectType, errUserTypeNotAssignable)
	}

	return nil
}

func validateTupleCondition(idx *modelIndex, ruleName string, tmpl mapper.TupleTemplate) error {
	if tmpl.Condition == "" {
		return nil
	}

	cond, exists := idx.conditions[tmpl.Condition]
	if !exists {
		return fmt.Errorf("rule %q: condition %q: %w", ruleName, tmpl.Condition, errConditionNotInModel)
	}

	if len(tmpl.Context) == 0 {
		return nil
	}

	params := cond.GetParameters()
	if params == nil {
		return nil
	}

	var errs []error

	for key := range tmpl.Context {
		if _, ok := params[key]; !ok {
			errs = append(errs, fmt.Errorf("rule %q: context key %q for condition %q: %w",
				ruleName, key, tmpl.Condition, errContextKeyNotParam))
		}
	}

	return errors.Join(errs...)
}

// extractTypeFromField extracts the type (and optional relation) from a tuple field.
// Returns ("", "", false) if the type cannot be determined statically.
// Examples: "user:alice" → ("user","",true), "group:eng#member" → ("group","member",true).
func extractTypeFromField(field string) (string, string, bool) {
	if strings.HasPrefix(field, "{{") {
		return "", "", false
	}

	typeName, rest, ok := strings.Cut(field, ":")
	if !ok || typeName == "" || strings.Contains(typeName, "{{") {
		return "", "", false
	}

	if rest == "*" {
		return typeName, "*", true
	}

	if hashIdx := strings.LastIndex(rest, "#"); hashIdx >= 0 {
		rel := rest[hashIdx+1:]
		if !strings.Contains(rel, "{{") && rel != "" {
			return typeName, rel, true
		}

		return "", "", false
	}

	return typeName, "", true
}

// isValidAssignee checks whether a user type (with optional relation) is allowed
// as an assignee for the given relation on the given type definition.
// Returns true when the model lacks metadata needed for assignee validation.
func isValidAssignee(typeDef *openfga.TypeDefinition, relation, userType, userRel string) bool { //nolint:cyclop
	meta := typeDef.GetMetadata()
	if meta.Relations == nil {
		return true
	}

	relMeta := *meta.Relations
	relationMeta, exists := relMeta[relation]

	if !exists {
		return true
	}

	refs := relationMeta.GetDirectlyRelatedUserTypes()

	for _, ref := range refs {
		if ref.Type != userType {
			continue
		}

		switch {
		case userRel == "*":
			if ref.Wildcard != nil {
				return true
			}
		case userRel != "":
			if ref.Relation != nil && *ref.Relation == userRel {
				return true
			}
		default:
			if ref.Relation == nil && ref.Wildcard == nil {
				return true
			}
		}
	}

	return false
}
