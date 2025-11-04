package tuplefile

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/clierrors"
	"github.com/openfga/cli/internal/cmdutils"
)

func parseTuplesFromCSV(data []byte, tuples *[]client.ClientTupleKey) error {
	reader := csv.NewReader(bytes.NewReader(data))

	columns, err := readHeaders(reader)
	if err != nil {
		return err
	}

	for index := 0; true; index++ {
		tuple, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to read tuple from csv file: %w", err)
		}

		tupleUserKey := tuple[columns.UserType] + ":" + tuple[columns.UserID]
		if columns.UserRelation != -1 && tuple[columns.UserRelation] != "" {
			tupleUserKey += "#" + tuple[columns.UserRelation]
		}

		condition, err := parseConditionColumnsForRow(columns, tuple, index)
		if err != nil {
			return err
		}

		tupleKey := client.ClientTupleKey{
			User:      tupleUserKey,
			Relation:  tuple[columns.Relation],
			Object:    tuple[columns.ObjectType] + ":" + tuple[columns.ObjectID],
			Condition: condition,
		}

		*tuples = append(*tuples, tupleKey)
	}

	return nil
}

func parseConditionColumnsForRow(
	columns *csvColumns,
	tuple []string,
	index int,
) (*openfga.RelationshipCondition, error) {
	var condition *openfga.RelationshipCondition

	if columns.ConditionName != -1 && tuple[columns.ConditionName] != "" {
		conditionContext := &(map[string]any{})

		if columns.ConditionContext != -1 {
			var err error

			conditionContext, err = cmdutils.ParseQueryContextInner(tuple[columns.ConditionContext])
			if err != nil {
				return nil, fmt.Errorf("failed to read condition context on line %d: %w", index, err)
			}
		}

		condition = &openfga.RelationshipCondition{
			Name:    tuple[columns.ConditionName],
			Context: conditionContext,
		}
	}

	return condition, nil
}

type csvColumns struct {
	UserType         int
	UserID           int
	UserRelation     int
	Relation         int
	ObjectType       int
	ObjectID         int
	ConditionName    int
	ConditionContext int
}

func (columns *csvColumns) setHeaderIndex(headerName string, index int) error {
	switch headerName {
	case "user_type":
		columns.UserType = index
	case "user_id":
		columns.UserID = index
	case "user_relation":
		columns.UserRelation = index
	case "relation":
		columns.Relation = index
	case "object_type":
		columns.ObjectType = index
	case "object_id":
		columns.ObjectID = index
	case "condition_name":
		columns.ConditionName = index
	case "condition_context":
		columns.ConditionContext = index
	default:
		return fmt.Errorf("invalid header %q, valid headers are user_type,user_id,user_relation,relation,object_type,object_id,condition_name,condition_context", headerName) //nolint:err113,lll
	}

	return nil
}

func (columns *csvColumns) validate() error {
	if columns.UserType == -1 {
		return clierrors.MissingRequiredCsvHeaderError("user_type") //nolint:wrapcheck
	}

	if columns.UserID == -1 {
		return clierrors.MissingRequiredCsvHeaderError("user_id") //nolint:wrapcheck
	}

	if columns.Relation == -1 {
		return clierrors.MissingRequiredCsvHeaderError("relation") //nolint:wrapcheck
	}

	if columns.ObjectType == -1 {
		return clierrors.MissingRequiredCsvHeaderError("object_type") //nolint:wrapcheck
	}

	if columns.ObjectID == -1 {
		return clierrors.MissingRequiredCsvHeaderError("object_id") //nolint:wrapcheck
	}

	if columns.ConditionContext != -1 && columns.ConditionName == -1 {
		return errors.New( //nolint:err113
			"missing \"condition_name\" header which is required when \"condition_context\" is present",
		)
	}

	return nil
}

func readHeaders(reader *csv.Reader) (*csvColumns, error) {
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv headers: %w", err)
	}

	columns := &csvColumns{
		UserType:         -1,
		UserID:           -1,
		UserRelation:     -1,
		Relation:         -1,
		ObjectType:       -1,
		ObjectID:         -1,
		ConditionName:    -1,
		ConditionContext: -1,
	}
	for index, header := range headers {
		err = columns.setHeaderIndex(strings.TrimSpace(header), index)
		if err != nil {
			return nil, err
		}
	}

	return columns, columns.validate()
}
