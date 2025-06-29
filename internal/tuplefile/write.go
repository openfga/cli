package tuplefile

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/openfga/go-sdk/client"
)

// CSVHeaderRow returns the header row used for CSV tuple files.
func CSVHeaderRow() string {
	return strings.Join([]string{
		"user_type",
		"user_id",
		"user_relation",
		"relation",
		"object_type",
		"object_id",
		"condition_name",
		"condition_context",
	}, ",")
}

// SerializeTuple returns the tuple in the provided format.
// format must be one of "json", "yaml", or "csv".
func SerializeTuple(format string, tuple client.ClientTupleKey) ([]byte, error) {
	switch format {
	case "yaml", "yml":
		return yaml.Marshal(tuple)
	case "csv":
		return serializeTupleCSV(tuple)
	default: // json
		return json.Marshal(tuple)
	}
}

func serializeTupleCSV(tuple client.ClientTupleKey) ([]byte, error) {
	userType, userID, userRelation := splitUser(tuple.User)
	objectType, objectID := splitObject(tuple.Object)

	var conditionName string
	var conditionContext string
	if tuple.Condition != nil {
		conditionName = tuple.Condition.Name
		if tuple.Condition.Context != nil {
			b, err := json.Marshal(tuple.Condition.Context)
			if err != nil {
				return nil, err
			}
			conditionContext = string(b)
		}
	}

	record := []string{
		userType,
		userID,
		userRelation,
		tuple.Relation,
		objectType,
		objectID,
		conditionName,
		conditionContext,
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(record); err != nil {
		return nil, err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func splitUser(user string) (string, string, string) {
	parts := strings.SplitN(user, ":", 2)
	if len(parts) != 2 {
		return user, "", ""
	}
	userType := parts[0]
	rest := parts[1]
	userID := rest
	userRel := ""
	if idx := strings.Index(rest, "#"); idx != -1 {
		userID = rest[:idx]
		userRel = rest[idx+1:]
	}
	return userType, userID, userRel
}

func splitObject(object string) (string, string) {
	parts := strings.SplitN(object, ":", 2)
	if len(parts) != 2 {
		return object, ""
	}
	return parts[0], parts[1]
}
