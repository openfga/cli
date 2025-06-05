package tuple

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/openfga/go-sdk/client"
	"gopkg.in/yaml.v3"
)

const (
	csvFormat  = "csv"
	yamlFormat = "yaml"
	jsonFormat = "json"
)

// tupleCSVDTO represents a tuple in CSV format.
type tupleCSVDTO struct {
	UserType         string `csv:"user_type"`
	UserID           string `csv:"user_id"`
	UserRelation     string `csv:"user_relation,omitempty"`
	Relation         string `csv:"relation"`
	ObjectType       string `csv:"object_type"`
	ObjectID         string `csv:"object_id"`
	ConditionName    string `csv:"condition_name,omitempty"`
	ConditionContext string `csv:"condition_context,omitempty"`
}

func tuplesToCSVDTO(tuples []client.ClientTupleKey) ([]tupleCSVDTO, error) {
	result := make([]tupleCSVDTO, 0, len(tuples))

	for _, tupleKey := range tuples {
		userParts := strings.SplitN(tupleKey.User, ":", 2)
		if len(userParts) != 2 {
			continue
		}

		userType := userParts[0]
		userIDRel := userParts[1]
		userID := userIDRel
		userRel := ""

		if strings.Contains(userIDRel, "#") {
			parts := strings.SplitN(userIDRel, "#", 2)
			userID = parts[0]
			userRel = parts[1]
		}

		objParts := strings.SplitN(tupleKey.Object, ":", 2)
		if len(objParts) != 2 {
			continue
		}

		condName := ""
		condCtx := ""

		if tupleKey.Condition != nil {
			condName = tupleKey.Condition.Name

			if tupleKey.Condition.Context != nil {
				b, err := json.Marshal(tupleKey.Condition.Context)
				if err != nil {
					return nil, fmt.Errorf("failed to convert condition context to CSV: %w", err)
				}

				condCtx = string(b)
			}
		}

		result = append(result, tupleCSVDTO{
			UserType:         userType,
			UserID:           userID,
			UserRelation:     userRel,
			Relation:         tupleKey.Relation,
			ObjectType:       objParts[0],
			ObjectID:         objParts[1],
			ConditionName:    condName,
			ConditionContext: condCtx,
		})
	}

	return result, nil
}

func formatFromExtension(fileName string) string {
	switch strings.ToLower(path.Ext(fileName)) {
	case ".csv":
		return csvFormat
	case ".yaml", ".yml":
		return yamlFormat
	default:
		return jsonFormat
	}
}

func formatTuples(tuples []client.ClientTupleKey, format string) (string, error) {
	switch format {
	case csvFormat:
		dto, err := tuplesToCSVDTO(tuples)
		if err != nil {
			return "", err
		}

		b, err := gocsv.MarshalBytes(dto)

		return string(b), err
	case yamlFormat:
		b, err := yaml.Marshal(tuples)

		return string(b), err
	default:
		b, err := json.MarshalIndent(tuples, "", "  ")

		return string(b), err
	}
}
