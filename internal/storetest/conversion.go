package storetest

import (
	"fmt"
	"strings"

	pb "github.com/openfga/api/proto/openfga/v1"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"google.golang.org/protobuf/types/known/structpb"
)

func convertClientTupleKeysToProtoTupleKeys(
	tuples []client.ClientContextualTupleKey,
) ([]*pb.TupleKey, error) {
	pbTuples := []*pb.TupleKey{}

	for _, tuple := range tuples {
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

func convertStoreObjectToObject(object string) (openfga.FgaObject, *pb.Object) {
	splitObject := strings.Split(object, ":")

	return openfga.FgaObject{
			Type: splitObject[0],
			Id:   splitObject[1],
		}, &pb.Object{
			Type: splitObject[0],
			Id:   splitObject[1],
		}
}

func convertPbUsersToStrings(users []*pb.User) []string {
	simpleUsers := []string{}

	for _, user := range users {
		switch typedUser := user.GetUser().(type) {
		case *pb.User_Object:
			simpleUsers = append(simpleUsers, typedUser.Object.GetType()+":"+typedUser.Object.GetId())
		case *pb.User_Userset:
			simpleUsers = append(
				simpleUsers,
				typedUser.Userset.GetType()+":"+typedUser.Userset.GetId()+"#"+typedUser.Userset.GetRelation(),
			)
		case *pb.User_Wildcard:
			simpleUsers = append(simpleUsers, typedUser.Wildcard.GetType()+":*")
		}
	}

	return simpleUsers
}

func convertPbObjectOrUsersetToStrings(users []*pb.ObjectOrUserset) []string {
	simpleUsers := []string{}

	for _, user := range users {
		switch typedUser := user.GetUser().(type) {
		case *pb.ObjectOrUserset_Object:
			simpleUsers = append(simpleUsers, typedUser.Object.GetType()+":"+typedUser.Object.GetId())
		case *pb.ObjectOrUserset_Userset:
			simpleUsers = append(
				simpleUsers,
				typedUser.Userset.GetType()+":"+typedUser.Userset.GetId()+"#"+typedUser.Userset.GetRelation(),
			)
		}
	}

	return simpleUsers
}

func convertOpenfgaUsers(users []openfga.User) []string {
	simpleUsers := []string{}

	for _, user := range users {
		switch {
		case user.Object != nil:
			simpleUsers = append(simpleUsers, user.Object.Type+":"+user.Object.Id)
		case user.Userset != nil:
			simpleUsers = append(simpleUsers, user.Userset.Type+":"+user.Userset.Id+"#"+user.Userset.Relation)
		case user.Wildcard != nil:
			simpleUsers = append(simpleUsers, user.Wildcard.Type+":*")
		}
	}

	return simpleUsers
}

func convertOpenfgaObjectOrUserset(users []openfga.ObjectOrUserset) []string {
	simpleUsers := []string{}

	for _, user := range users {
		switch {
		case user.Object != nil:
			simpleUsers = append(simpleUsers, user.Object.Type+":"+user.Object.Id)
		case user.Userset != nil:
			simpleUsers = append(simpleUsers, user.Userset.Type+":"+user.Userset.Id+"#"+user.Userset.Relation)
		}
	}

	return simpleUsers
}
