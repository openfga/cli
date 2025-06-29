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

       ext := path.Ext(fileName)
       switch ext {
       case ".json":
               err = yaml.Unmarshal(data, &tuples)
               if err != nil {
                       err = parseTuplesFromJSONL(data, &tuples, "json")
               }
               if err == nil && len(tuples) == 0 {
                       err = clierrors.EmptyTuplesFileError("json")
               }
       case ".jsonl":
               err = parseTuplesFromJSONL(data, &tuples, "jsonl")
               if err == nil && len(tuples) == 0 {
                       err = clierrors.EmptyTuplesFileError("jsonl")
               }
       case ".yaml", ".yml":
               err = yaml.Unmarshal(data, &tuples)
               if err == nil && len(tuples) == 0 {
                       err = clierrors.EmptyTuplesFileError(strings.TrimPrefix(ext, "."))
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

func parseTuplesFromJSONL(data []byte, tuples *[]client.ClientTupleKey, format string) error {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		var tuple client.ClientTupleKey
		if err := yaml.Unmarshal([]byte(trimmed), &tuple); err != nil {
			return fmt.Errorf("failed to parse tuple on line %d: %w", index+1, err)
		}

		*tuples = append(*tuples, tuple)
	}

       if len(*tuples) == 0 {
               return clierrors.EmptyTuplesFileError(format)
       }

	return nil
}
