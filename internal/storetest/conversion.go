package storetest

import (
	"fmt"
	"strings"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"google.golang.org/protobuf/types/known/structpb"
)

func convertClientTupleKeysToProtoTupleKeys(
	tuples []client.ClientContextualTupleKey,
) ([]*openfgav1.TupleKey, error) {
	pbTuples := []*openfgav1.TupleKey{}

	for _, tuple := range tuples {
		tpl := openfgav1.TupleKey{
			User:     tuple.User,
			Relation: tuple.Relation,
			Object:   tuple.Object,
		}

		if tuple.Condition != nil {
			conditionContext, err := structpb.NewStruct(tuple.Condition.GetContext())
			if err != nil {
				return nil, fmt.Errorf("failed to construct a proto struct: %w", err)
			}

			tpl.Condition = &openfgav1.RelationshipCondition{
				Name:    tuple.Condition.Name,
				Context: conditionContext,
			}
		}

		pbTuples = append(pbTuples, &tpl)
	}

	return pbTuples, nil
}

func convertStoreObjectToObject(object string) (openfga.FgaObject, *openfgav1.Object) {
	splitObject := strings.Split(object, ":")

	return openfga.FgaObject{
		Type: splitObject[0],
		Id:   splitObject[1],
	}, &openfgav1.Object{
		Type: splitObject[0],
		Id:   splitObject[1],
	}
}

func convertPbUsersToStrings(users []*openfgav1.User) []string {
	simpleUsers := []string{}

	for _, user := range users {
		switch typedUser := user.GetUser().(type) {
		case *openfgav1.User_Object:
			simpleUsers = append(simpleUsers, typedUser.Object.GetType()+":"+typedUser.Object.GetId())
		case *openfgav1.User_Userset:
			simpleUsers = append(
				simpleUsers,
				typedUser.Userset.GetType()+":"+typedUser.Userset.GetId()+"#"+typedUser.Userset.GetRelation(),
			)
		case *openfgav1.User_Wildcard:
			simpleUsers = append(simpleUsers, typedUser.Wildcard.GetType()+":*")
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
