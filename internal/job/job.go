package job

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/openfga/cli/internal/storage"
	"github.com/openfga/go-sdk/client"
	"math"
	"sync/atomic"
	"time"
)

func CreateJob(conn *sql.Conn, storeID string, tuples []client.ClientTupleKey) (string, error) {
	bulkJobID := ulid.Make().String()
	err := storage.InsertTuples(conn, bulkJobID, storeID, tuples)
	if err != nil {
		return "", err
	}
	return bulkJobID, nil
}

func ImportTuples(conn *sql.Conn, bulkJobID string,
	fgaClient client.SdkClient,
	requestRate int,
	maxRequests int,
	rampIntervalInSeconds int64,
) error {
	notInsertedTuplesCount, insertedTuplesCount, err := storage.GetTotalAndRemainingTuples(conn, bulkJobID)
	totalTuplesCount := insertedTuplesCount + notInsertedTuplesCount
	if err != nil {
		return err
	}
	completedTuples := atomic.Int64{}
	completedTuples.Store(insertedTuplesCount)

	rampStartTime := time.Now().Unix()
	currentRequestRate := requestRate
	for completedTuples.Load() < totalTuplesCount {
		startTime := time.Now()
		if rampStartTime > time.Now().Unix()+rampIntervalInSeconds {
			currentRequestRate = GetRequestRate(requestRate, maxRequests)
			rampStartTime = time.Now().Unix()
		}
		remainingTuples, e := storage.GetRemainingTuples(conn, bulkJobID, currentRequestRate)
		if e != nil {
			return e
		}
		for _, tuple := range remainingTuples {
			_, e = fgaClient.
				WriteTuples(context.Background()).
				Body(client.ClientWriteTuplesBody{tuple.Tuple}).
				Options(client.ClientWriteOptions{}).
				Execute()
			if e != nil {
				errStr := e.Error()
				storage.UpdateStatus(conn, tuple.Rowid, storage.NOT_INSERTED, errStr)
			} else {
				storage.UpdateStatus(conn, tuple.Rowid, storage.INSERTED, "")
			}
			completedTuples.Add(1)
		}
		elapsed := time.Since(startTime)

		fmt.Printf("Completed %d/%d. Requests Per Second - %d\n", completedTuples.Load(), totalTuplesCount, currentRequestRate)
		time.Sleep(time.Second - elapsed)
	}
	return nil
}

func GetRequestRate(currentRequestRate int, maxRequests int) int {
	increasedRequestRate := currentRequestRate + int(math.Ceil(float64(currentRequestRate)*0.3))
	if increasedRequestRate > maxRequests {
		return maxRequests
	}
	return increasedRequestRate
}
