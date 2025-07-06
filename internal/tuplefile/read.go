package tuplefile

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	case ".jsonl":
		err = parseTuplesFromJSONL(data, &tuples)
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

func parseTuplesFromJSONL(data []byte, tuples *[]client.ClientTupleKey) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var tuple client.ClientTupleKey
		if err := json.Unmarshal([]byte(line), &tuple); err != nil {
			return fmt.Errorf("failed to read tuple from jsonl file on line %d: %w", lineNum, err)
		}

		*tuples = append(*tuples, tuple)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read jsonl file: %w", err)
	}

	if len(*tuples) == 0 {
		return clierrors.EmptyTuplesFileError("jsonl")
	}

	return nil
}
