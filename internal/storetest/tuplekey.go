package storetest

import (
	pb "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/go-sdk/client"
)

func convertClientTupleKeysToProtoTupleKeys(tuples []client.ClientWriteRequestTupleKey) []*pb.WriteRequestTupleKey {
	pbTuples := []*pb.WriteRequestTupleKey{}

	for index := 0; index < len(tuples); index++ {
		tuple := tuples[index]
		tpl := pb.WriteRequestTupleKey{
			User:     tuple.User,
			Relation: tuple.Relation,
			Object:   tuple.Object,
		}
		pbTuples = append(pbTuples, &tpl)
	}

	return pbTuples
}
