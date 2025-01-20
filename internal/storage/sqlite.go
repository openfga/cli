package storage

import (
	"database/sql"
	"fmt"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"strings"
	"time"
)

const (
	NOT_INSERTED = iota
	INSERTED
)

var CREATE_TABLE = `
	create table if not exists import_job 
	(
		bulk_job_id INTEGER, STORE_ID CHAR(26), inserted_at INT NOT NULL, 
		imported_at INT, subject VARCHAR(256), relation VARCHAR(256), 
		object VARCHAR(256), condition TEXT, status INT, reason TEXT
)`

var INSERT_TUPLES = `
 	INSERT INTO import_job (
	bulk_job_id, store_id, inserted_at, 
	imported_at, subject, object, relation, 
	condition, status, reason) values %s
`

var READ_TUPLES = `SELECT 
		ROWID, subject, object, relation, condition
		from import_job where bulk_job_id = ? and status = ? order by ROWID limit ?
`

var GET_TUPLES_COUNT = `
	SELECT COUNT(*) from import_job where bulk_job_id = ? and status = ?
`

func NewDatabase() (*sql.DB, error) {
	dsnURI := "file:cli.db?_journal_mode=WAL"
	db, err := sql.Open("sqlite", dsnURI)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(CREATE_TABLE)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InsertTuples(db *sql.DB, bulkJobID string, storeID string, tuples []client.ClientTupleKey) error {
	valueStrings := make([]string, 0, len(tuples))
	valueArgs := make([]interface{}, 0, len(tuples)*10)
	for _, tuple := range tuples {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, bulkJobID)
		valueArgs = append(valueArgs, storeID)
		valueArgs = append(valueArgs, time.Now().Unix())
		valueArgs = append(valueArgs, nil)
		valueArgs = append(valueArgs, tuple.User)
		valueArgs = append(valueArgs, tuple.Object)
		valueArgs = append(valueArgs, tuple.Relation)
		if tuple.Condition != nil {
			json, err := tuple.Condition.MarshalJSON()
			if err != nil {
				return err
			}
			valueArgs = append(valueArgs, json)
		} else {
			valueArgs = append(valueArgs, "")
		}
		valueArgs = append(valueArgs, NOT_INSERTED)
		valueArgs = append(valueArgs, "")
	}

	stmt := fmt.Sprintf(INSERT_TUPLES, strings.Join(valueStrings, ","))
	_, err := db.Exec(stmt, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func GetRemainingTuples(db *sql.DB, bulkJobID string, count int) ([]TuplesResult, error) {
	row, err := db.Query(READ_TUPLES, bulkJobID, NOT_INSERTED, count)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	var tuples []TuplesResult
	for row.Next() {
		result := client.ClientTupleKey{Condition: &openfga.RelationshipCondition{}}
		var rowid int64
		var condition string
		err = row.Scan(&rowid, &result.User, &result.Object, &result.Relation, &condition)
		if err != nil {
			return nil, err
		}
		if condition != "" {
			nullable := openfga.NullableRelationshipCondition{}
			err = nullable.UnmarshalJSON([]byte(condition))
			if err != nil {
				return nil, err
			}
			if nullable.IsSet() {
				result.Condition = nullable.Get()
			} else {
				result.Condition = nil
			}
		} else {
			result.Condition = nil
		}
		tuples = append(tuples, TuplesResult{Rowid: rowid, Tuple: result})
	}
	fmt.Printf("%+v", tuples)
	return tuples, nil
}

type TuplesResult struct {
	Tuple client.ClientTupleKey
	Rowid int64
}

func GetTotalAndRemainingTuples(db *sql.DB, bulkJobID string) (int64, int64, error) {
	var notInsertedCount, insertedCount int64
	row, err := db.Query(GET_TUPLES_COUNT, bulkJobID, NOT_INSERTED)
	if err != nil {
		return 0, 0, err
	}
	row.Next()
	row.Scan(&notInsertedCount)
	row, err = db.Query(GET_TUPLES_COUNT, bulkJobID, INSERTED)
	if err != nil {
		return 0, 0, err
	}
	row.Next()
	row.Scan(&insertedCount)
	return notInsertedCount, insertedCount, nil
}
