package tuplefile

import (
	"fmt"
	"os"
	"path"

	openfga "github.com/openfga/go-sdk"
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
		err = fmt.Errorf("unsupported file format %q", path.Ext(fileName)) //nolint:goerr113
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse input tuples: %w", err)
	}

	return tuples, nil
}

func ClientTupleKeyToTupleKeyWithoutCondition(clientTupleKey []client.ClientTupleKey) []openfga.TupleKeyWithoutCondition {
	var tuples []openfga.TupleKeyWithoutCondition
	for _, tuple := range clientTupleKey {
		convertedTuple := openfga.TupleKeyWithoutCondition{
			User:     tuple.User,
			Relation: tuple.Relation,
			Object:   tuple.Object,
		}
		tuples = append(tuples, convertedTuple)
	}
	return tuples
}
