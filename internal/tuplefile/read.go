package tuplefile

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/openfga/go-sdk/client"
	"gopkg.in/yaml.v3"

	"github.com/openfga/cli/internal/clierrors"
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
		if err == nil && len(tuples) == 0 {
			err = clierrors.EmptyTuplesFileError(strings.TrimPrefix(path.Ext(fileName), "."))
		}
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
