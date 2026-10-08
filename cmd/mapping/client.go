package mapping

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	openfga "github.com/openfga/go-sdk"
	sdkclient "github.com/openfga/go-sdk/client"
	"github.com/openfga/mapper/apply"
	"github.com/openfga/mapper/language"

	internaltuple "github.com/openfga/cli/internal/tuple"
)

const maxWriteChunkSize = 40

// sdkWriteReadClient is the subset of sdkclient.SdkClient used by fgaTupleClient.
type sdkWriteReadClient interface {
	Read(ctx context.Context) sdkclient.SdkClientReadRequestInterface
	Write(ctx context.Context) sdkclient.SdkClientWriteRequestInterface
}

// fgaTupleClient adapts the OpenFGA SDK client to the apply.TupleClient interface.
type fgaTupleClient struct {
	inner sdkWriteReadClient
}

func (c *fgaTupleClient) ReadTuples(ctx context.Context, filter language.TupleFilter) ([]language.Tuple, error) {
	body := &sdkclient.ClientReadRequest{}
	if filter.User != "" {
		body.User = &filter.User
	}
	if filter.Relation != "" {
		body.Relation = &filter.Relation
	}
	if filter.Object != "" {
		body.Object = &filter.Object
	}

	consistency := openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY
	resp, err := internaltuple.Read(ctx, c.inner, body, 0, internaltuple.DefaultReadPageSize, &consistency)
	if err != nil {
		return nil, err
	}

	tuples := make([]language.Tuple, 0, len(resp.Tuples))
	for _, t := range resp.Tuples {
		key := t.GetKey()
		tuple := language.Tuple{
			User:     key.GetUser(),
			Relation: key.GetRelation(),
			Object:   key.GetObject(),
		}
		if cond, ok := key.GetConditionOk(); ok {
			tuple.Condition = cond.GetName()
			tuple.Context = cond.GetContext()
		}
		tuples = append(tuples, tuple)
	}

	return tuples, nil
}

func (c *fgaTupleClient) WriteTuples(ctx context.Context, tuples []language.Tuple) error {
	writes, deletes := partitionSyncTuples(tuples)
	if len(writes) == 0 && len(deletes) == 0 {
		return nil
	}

	writeOpts := sdkclient.ClientWriteOptions{
		Conflict: sdkclient.ClientWriteConflictOptions{
			OnDuplicateWrites: sdkclient.CLIENT_WRITE_REQUEST_ON_DUPLICATE_WRITES_IGNORE,
			OnMissingDeletes:  sdkclient.CLIENT_WRITE_REQUEST_ON_MISSING_DELETES_IGNORE,
		},
		Transaction: &sdkclient.TransactionOptions{
			Disable:     true,
			MaxPerChunk: maxWriteChunkSize,
		},
	}

	// Deletes must be issued before writes. The OpenFGA server rejects a write to a
	// URO that already holds a tuple with a different condition (TupleConditionConflictError),
	// so a condition change (delete old + write new on same URO) requires the delete to
	// land first. The SDK with Transaction.Disable sends writes first then deletes, so we
	// split them into two calls.
	if len(deletes) > 0 {
		if err := c.executeDeletes(ctx, writeOpts, deletes); err != nil {
			return err
		}
	}

	if len(writes) > 0 {
		if err := c.executeWrites(ctx, writeOpts, writes); err != nil {
			return err
		}
	}

	return nil
}

func (c *fgaTupleClient) executeDeletes(
	ctx context.Context,
	opts sdkclient.ClientWriteOptions,
	deletes []sdkclient.ClientTupleKeyWithoutCondition,
) error {
	resp, err := c.inner.Write(ctx).Options(opts).Body(sdkclient.ClientWriteRequest{
		Deletes: deletes,
	}).Execute()
	if err != nil {
		return fmt.Errorf("deleting tuples: %w", err)
	}

	var errs []error
	seen := make(map[string]struct{})
	for _, d := range resp.Deletes {
		if d.Status == sdkclient.FAILURE {
			var e error
			if d.Error != nil {
				e = d.Error
			} else {
				e = fmt.Errorf("delete failed: %s %s %s", d.TupleKey.User, d.TupleKey.Relation, d.TupleKey.Object)
			}
			if _, ok := seen[e.Error()]; !ok {
				seen[e.Error()] = struct{}{}
				errs = append(errs, e)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("deleting tuples: %w", errors.Join(errs...))
	}

	return nil
}

func (c *fgaTupleClient) executeWrites(
	ctx context.Context,
	opts sdkclient.ClientWriteOptions,
	writes []sdkclient.ClientTupleKey,
) error {
	resp, err := c.inner.Write(ctx).Options(opts).Body(sdkclient.ClientWriteRequest{
		Writes: writes,
	}).Execute()
	if err != nil {
		return fmt.Errorf("writing tuples: %w", err)
	}

	var errs []error
	seen := make(map[string]struct{})
	for _, w := range resp.Writes {
		if w.Status == sdkclient.FAILURE {
			var e error
			if w.Error != nil {
				e = w.Error
			} else {
				e = fmt.Errorf("write failed: %s %s %s", w.TupleKey.User, w.TupleKey.Relation, w.TupleKey.Object)
			}
			if _, ok := seen[e.Error()]; !ok {
				seen[e.Error()] = struct{}{}
				errs = append(errs, e)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("writing tuples: %w", errors.Join(errs...))
	}

	return nil
}

// countingClient wraps a TupleClient and tracks how many tuples were written/deleted.
// Reset writes and dels between records. Correctness depends on apply setting t.Action
// on every tuple it passes to WriteTuples; an unset Action is treated as a write.
type countingClient struct {
	inner  apply.TupleClient
	writes int
	dels   int
}

func (c *countingClient) ReadTuples(ctx context.Context, filter language.TupleFilter) ([]language.Tuple, error) {
	return c.inner.ReadTuples(ctx, filter)
}

func (c *countingClient) WriteTuples(ctx context.Context, tuples []language.Tuple) error {
	for _, t := range tuples {
		if t.Action == language.ActionDelete {
			c.dels++
		} else {
			c.writes++
		}
	}
	return c.inner.WriteTuples(ctx, tuples)
}

// printOnlyClient reads from inner but prints planned writes/deletes to out instead of applying them.
// Used for --dry-run.
type printOnlyClient struct {
	inner apply.TupleClient
	out   io.Writer
}

func (c *printOnlyClient) ReadTuples(ctx context.Context, filter language.TupleFilter) ([]language.Tuple, error) {
	return c.inner.ReadTuples(ctx, filter)
}

func (c *printOnlyClient) WriteTuples(_ context.Context, tuples []language.Tuple) error {
	enc := json.NewEncoder(c.out)
	for _, t := range tuples {
		if err := enc.Encode(toOpTuple(t)); err != nil {
			return fmt.Errorf("encoding dry-run output: %w", err)
		}
	}
	return nil
}

func partitionSyncTuples(tuples []language.Tuple) ([]sdkclient.ClientTupleKey, []sdkclient.ClientTupleKeyWithoutCondition) {
	var writes []sdkclient.ClientTupleKey
	var deletes []sdkclient.ClientTupleKeyWithoutCondition

	for _, t := range tuples {
		if t.Action == language.ActionDelete {
			deletes = append(deletes, sdkclient.ClientTupleKeyWithoutCondition{
				User:     t.User,
				Relation: t.Relation,
				Object:   t.Object,
			})

			continue
		}

		key := sdkclient.ClientTupleKey{
			User:     t.User,
			Relation: t.Relation,
			Object:   t.Object,
		}
		if t.Condition != "" {
			rc := &openfga.RelationshipCondition{Name: t.Condition}
			if len(t.Context) > 0 {
				rc.Context = &t.Context
			}
			key.Condition = rc
		}
		writes = append(writes, key)
	}

	return writes, deletes
}
