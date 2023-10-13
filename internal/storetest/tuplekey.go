package storetest

import (
	"fmt"

	pb "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/go-sdk/client"
	"google.golang.org/protobuf/types/known/structpb"
)

func convertClientTupleKeysToProtoTupleKeys(
	tuples []client.ClientContextualTupleKey,
) ([]*pb.WriteRequestTupleKey, error) {
	pbTuples := []*pb.WriteRequestTupleKey{}

	for index := 0; index < len(tuples); index++ {
		tuple := tuples[index]
		tpl := pb.WriteRequestTupleKey{
			User:     tuple.User,
			Relation: tuple.Relation,
			Object:   tuple.Object,
		}

		if tuple.Condition != nil {
			conditionContext, err := structpb.NewStruct(*tuple.Condition.Context)
			if err != nil {
				return nil, fmt.Errorf("failed to construct a proto struct: %w", err)
			}

			tpl.Condition = &pb.RelationshipCondition{
				ConditionName: tuple.Condition.ConditionName,
				Context:       conditionContext,
			}
		}

		pbTuples = append(pbTuples, &tpl)
	}

	return pbTuples, nil
}
