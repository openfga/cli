package storetest

import (
	"fmt"

	pb "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/go-sdk/client"
	"google.golang.org/protobuf/types/known/structpb"
)

func convertClientTupleKeysToProtoTupleKeys(
	tuples []client.ClientContextualTupleKey,
) ([]*pb.TupleKey, error) {
	pbTuples := []*pb.TupleKey{}

	for index := 0; index < len(tuples); index++ {
		tuple := tuples[index]
		tpl := pb.TupleKey{
			User:     tuple.User,
			Relation: tuple.Relation,
			Object:   tuple.Object,
		}

		if tuple.Condition != nil {
			conditionContext, err := structpb.NewStruct(tuple.Condition.GetContext())
			if err != nil {
				return nil, fmt.Errorf("failed to construct a proto struct: %w", err)
			}

			tpl.Condition = &pb.RelationshipCondition{
				Name:    tuple.Condition.Name,
				Context: conditionContext,
			}
		}

		pbTuples = append(pbTuples, &tpl)
	}

	return pbTuples, nil
}
