package tuplefile

import (
	"fmt"
	"os"
	"path"

	"github.com/openfga/go-sdk/client"
	"gopkg.in/yaml.v3"
)

func ReadTupleFile(fileName string) ([]client.ClientTupleKey, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", fileName, err)
	}

	var tuples []client.ClientTupleKey

	switch path.Ext(fileName) {
	case ".json", ".yaml", ".yml":
		err = yaml.Unmarshal(data, &tuples)
	case ".csv":
		err = parseTuplesFromCSV(data, &tuples)
	default:
		err = fmt.Errorf("unsupported file format %q", path.Ext(fileName)) //nolint:err113
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse input tuples: %w", err)
	}

	return tuples, nil
}

func TupleKeysToTupleKeysWithoutCondition(tks []client.ClientTupleKey) []client.ClientTupleKeyWithoutCondition {
	tkswc := make([]client.ClientTupleKeyWithoutCondition, 0, len(tks))
	for _, tk := range tks {
		tkswc = append(tkswc, client.ClientTupleKeyWithoutCondition{
			User:     tk.User,
			Relation: tk.Relation,
			Object:   tk.Object,
		})
	}
	return tkswc
}
