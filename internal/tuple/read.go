package tuple

import (
	"context"
	"fmt"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
)

const DefaultReadPageSize int32 = 50

func Read(
	fgaClient client.SdkClient,
	body *client.ClientReadRequest,
	maxPages int,
	consistency *openfga.ConsistencyPreference,
) (
	*openfga.ReadResponse, error,
) {
	tuples := make([]openfga.Tuple, 0)
	continuationToken := ""
	pageIndex := 0
	options := client.ClientReadOptions{
		PageSize: openfga.PtrInt32(DefaultReadPageSize),
	}

	if consistency != nil && *consistency != openfga.CONSISTENCYPREFERENCE_UNSPECIFIED {
		options.Consistency = consistency
	}

	for {
		options.ContinuationToken = &continuationToken

		response, err := fgaClient.Read(context.Background()).Body(*body).Options(options).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to read tuples due to %w", err)
		}

		tuples = append(tuples, response.Tuples...)
		pageIndex++

		if response.ContinuationToken == "" ||
			(maxPages != 0 && pageIndex >= maxPages) {
			break
		}

		continuationToken = response.ContinuationToken
	}

	return &openfga.ReadResponse{Tuples: tuples}, nil
}

func TupleKeyToTupleKeyWithoutCondition(tk client.ClientTupleKey) client.ClientTupleKeyWithoutCondition {
	return client.ClientTupleKeyWithoutCondition{
		Object:   tk.GetObject(),
		Relation: tk.GetRelation(),
		User:     tk.GetUser(),
	}
}

// TupleKeysToTupleKeysWithoutCondition converts ClientTupleKeys to a slice of
// ClientTupleKeyWithoutCondition, stripping out condition-related fields.
func TupleKeysToTupleKeysWithoutCondition(tks ...client.ClientTupleKey) []client.ClientTupleKeyWithoutCondition {
	converted := make([]client.ClientTupleKeyWithoutCondition, 0, len(tks))
	for _, tk := range tks {
		converted = append(converted, TupleKeyToTupleKeyWithoutCondition(tk))
	}
	return converted
}
