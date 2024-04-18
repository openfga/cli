package storetest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/oklog/ulid/v2"
	pb "github.com/openfga/api/proto/openfga/v1"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/comparison"
)

type FGAContext struct {
	user             string
	relation         string
	object           string
	conditionName    string
	conditionContext *structpb.Struct
	objectType       string

	modelID string
	storeID string

	fgaServer *server.Server
	fgaClient *client.OpenFgaClient

	contextualTuples []client.ClientContextualTupleKey

	isLocalTest bool
}

func (f *FGAContext) writeLocalModel(authModel *authorizationmodel.AuthzModel, fgaServer *server.Server) error {
	f.fgaServer = fgaServer

	model := authModel.GetProtoModel()
	f.storeID = ulid.Make().String()
	authModelWriteReq := &pb.WriteAuthorizationModelRequest{
		StoreId:         f.storeID,
		TypeDefinitions: model.GetTypeDefinitions(),
		SchemaVersion:   model.GetSchemaVersion(),
		Conditions:      model.GetConditions(),
	}

	writtenModel, err := fgaServer.WriteAuthorizationModel(context.Background(), authModelWriteReq)
	if err != nil {
		return err //nolint:wrapcheck
	}

	f.modelID = writtenModel.GetAuthorizationModelId()

	return nil
}

func (f *FGAContext) tableToContext(table *godog.Table) (*structpb.Struct, error) {
	con := map[string]interface{}{}

	for _, row := range table.Rows {
		key := row.Cells[0].Value
		val := row.Cells[1].Value

		if strings.HasPrefix(val, "{") || strings.HasPrefix(val, "[") {
			var unmarshalledVal interface{}

			err := json.Unmarshal([]byte(val), &unmarshalledVal)
			if err != nil {
				return nil, err //nolint:wrapcheck
			}

			con[key] = unmarshalledVal
		} else {
			con[key] = val
		}
	}

	conditionContext, err := structpb.NewStruct(con)
	if err != nil {
		return nil, errors.New("failed to construct condition context") //nolint:goerr113
	}

	return conditionContext, nil
}

func (f *FGAContext) writeTuple() error {
	if f.isLocalTest {
		tuple := &pb.TupleKey{
			User:     f.user,
			Relation: f.relation,
			Object:   f.object,
		}

		if f.conditionName != "" {
			tuple.Condition = &pb.RelationshipCondition{
				Name:    f.conditionName,
				Context: f.conditionContext,
			}
		}

		writeRequest := &pb.WriteRequest{
			StoreId: f.storeID,
			Writes:  &pb.WriteRequestWrites{TupleKeys: []*pb.TupleKey{tuple}},
		}

		_, err := f.fgaServer.Write(context.Background(), writeRequest)
		if err != nil {
			return err //nolint:wrapcheck
		}
	} else {
		tuple := client.ClientContextualTupleKey{
			User:     f.user,
			Relation: f.relation,
			Object:   f.object,
		}

		if f.conditionName != "" {
			context := f.conditionContext.AsMap()
			tuple.Condition = &openfga.RelationshipCondition{
				Name:    f.conditionName,
				Context: &context,
			}
		}

		f.contextualTuples = append(f.contextualTuples, tuple)
	}

	// Clear the data after writing
	f.user = ""
	f.relation = ""
	f.object = ""
	f.conditionContext = nil
	f.conditionName = ""
	f.objectType = ""

	return nil
}

func (f *FGAContext) checkLocal() (bool, error) {
	req := &pb.CheckRequest{
		StoreId:              f.storeID,
		AuthorizationModelId: f.modelID,
		TupleKey: &pb.CheckRequestTupleKey{
			User:     f.user,
			Relation: f.relation,
			Object:   f.object,
		},
	}

	if f.conditionContext != nil {
		req.Context = f.conditionContext
	}

	res, err := f.fgaServer.Check(context.Background(), req)
	if err != nil {
		return false, err //nolint:wrapcheck
	}

	return res.GetAllowed(), nil
}

func (f *FGAContext) checkRemote() (bool, error) {
	req := client.ClientCheckRequest{
		User:     f.user,
		Relation: f.relation,
		Object:   f.object,

		ContextualTuples: f.contextualTuples,
	}

	if f.conditionContext != nil {
		context := f.conditionContext.AsMap()
		req.Context = &context
	}

	res, err := f.fgaClient.Check(context.Background()).Body(req).Execute()
	if err != nil {
		return false, err //nolint:wrapcheck
	}

	return res.GetAllowed(), nil
}

func (f *FGAContext) check(expected bool) error {
	var actual bool

	var err error

	if f.isLocalTest {
		actual, err = f.checkLocal()
	} else {
		actual, err = f.checkRemote()
	}

	if err != nil {
		return err
	}

	if actual != expected {
		errorString := fmt.Sprintf(
			"ⅹ Check(user=%s,relation=%s,object=%s",
			f.user,
			f.relation,
			f.object,
		)

		if f.conditionContext != nil {
			errorString += fmt.Sprintf(", context:%v", f.conditionContext)
		}

		errorString += fmt.Sprintf(
			"): expected=%t, got=%t",
			expected,
			actual,
		)

		return errors.New(errorString) //nolint:goerr113
	}

	return nil
}

func (f *FGAContext) listObjectsLocal() ([]string, error) {
	req := &pb.ListObjectsRequest{
		StoreId:              f.storeID,
		AuthorizationModelId: f.modelID,
		User:                 f.user,
		Type:                 f.objectType,
		Relation:             f.relation,
	}

	if f.conditionContext != nil {
		req.Context = f.conditionContext
	}

	r, err := f.fgaServer.ListObjects(context.Background(), req)
	if err != nil {
		return []string{}, err //nolint:wrapcheck
	}

	return r.GetObjects(), nil
}

func (f *FGAContext) listObjectsRemote() ([]string, error) {
	req := client.ClientListObjectsRequest{
		User:             f.user,
		Type:             f.objectType,
		Relation:         f.relation,
		ContextualTuples: f.contextualTuples,
	}

	if f.conditionContext != nil {
		context := f.conditionContext.AsMap()
		req.Context = &context
	}

	res, err := f.fgaClient.ListObjects(context.Background()).Body(req).Execute()
	if err != nil {
		return []string{}, err //nolint:wrapcheck
	}

	return res.GetObjects(), nil
}

func (f *FGAContext) listObjects(expected []string) error {
	var actual []string

	var err error

	if f.isLocalTest {
		actual, err = f.listObjectsLocal()
	} else {
		actual, err = f.listObjectsRemote()
	}

	if err != nil {
		return err
	}

	if !comparison.CheckStringArraysEqual(actual, expected) {
		errorString := fmt.Sprintf(
			"ⅹ ListObjects(user=%s,relation=%s,type=%s",
			f.user,
			f.relation,
			f.objectType,
		)

		if f.conditionContext != nil {
			errorString += fmt.Sprintf(", context:%v", f.conditionContext)
		}

		errorString += fmt.Sprintf(
			"): expected=%s, got=%s",
			expected,
			actual,
		)

		return errors.New(errorString) //nolint:goerr113
	}

	return nil
}

func (f *FGAContext) thereIsATuple(user, relation, object string) error {
	f.user = user
	f.object = object
	f.relation = relation

	return f.writeTuple()
}

func (f *FGAContext) thereIsATupleWithCondition(
	user, relation, object, condition string,
	table *godog.Table,
) error {
	f.user = user
	f.object = object
	f.relation = relation
	f.conditionName = condition

	conditionContext, err := f.tableToContext(table)
	if err != nil {
		return err
	}

	f.conditionContext = conditionContext

	return f.writeTuple()
}

func (f *FGAContext) thereIsATupleWithConditionNoValue(user, relation, object, condition string) error {
	f.user = user
	f.object = object
	f.relation = relation
	f.conditionName = condition

	return f.writeTuple()
}

func (f *FGAContext) checkHasRelation(user, relation string) error {
	if strings.Contains(user, ":") {
		f.user = user
	}

	f.relation = relation

	return f.check(true)
}

func (f *FGAContext) checkHasRelations(user string, table *godog.Table) error {
	if strings.Contains(user, ":") {
		f.user = user
	}

	var errors []error

	for _, row := range table.Rows {
		f.relation = row.Cells[0].Value

		err := f.check(true)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%v", errors) //nolint:goerr113
	}

	return nil
}

func (f *FGAContext) checkDoesNotHaveRelation(user, relation string) error {
	if strings.Contains(user, ":") {
		f.user = user
	}

	f.relation = relation

	return f.check(false)
}

func (f *FGAContext) checkDoesNotHaveRelations(user string, table *godog.Table) error {
	if strings.Contains(user, ":") {
		f.user = user
	}

	var errors []error

	for _, row := range table.Rows {
		f.relation = row.Cells[0].Value

		err := f.check(false)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%v", errors) //nolint:goerr113
	}

	return nil
}

func (f *FGAContext) addContext(user, object string, table *godog.Table) error {
	if user != "" {
		f.user = user
	}

	if object != "" {
		f.object = object
	}

	if table != nil {
		conditionContext, err := f.tableToContext(table)
		if err != nil {
			return err
		}

		f.conditionContext = conditionContext
	}

	return nil
}

func (f *FGAContext) userAndObject(user, object string) {
	if strings.Contains(user, ":") {
		f.user = user
	}

	f.object = object
}

func (f *FGAContext) conditionCheckNoUser(table *godog.Table) error {
	return f.addContext("", "", table)
}

func (f *FGAContext) addType(user, objectType string) {
	if user != "" {
		f.user = user
	}

	f.objectType = objectType
}

func (f *FGAContext) listObjectsStep(user, relation string, table *godog.Table) error {
	if user != "" {
		f.user = user
	}

	f.relation = relation

	objects := []string{}
	for _, row := range table.Rows {
		objects = append(objects, row.Cells[0].Value)
	}

	return f.listObjects(objects)
}

func (f *FGAContext) listNoObjects(user, relation string) error {
	if user != "" {
		f.user = user
	}

	f.relation = relation

	return f.listObjects([]string{})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	fga := FGAContext{}

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		isLocalTest, ok := ctx.Value(ctxKeyIsLocalTest).(bool)
		if !ok {
			return nil, errors.New("couldnt check isLocalTest") //nolint:goerr113
		}

		fga.isLocalTest = isLocalTest

		if fga.isLocalTest { //nolint:nestif
			authModel, ok := ctx.Value(ctxKeyAuthModel).(*authorizationmodel.AuthzModel) //nolint:varnamelen
			if !ok {
				return nil, errors.New("no auth model") //nolint:goerr113
			}

			fgaServer, ok := ctx.Value(ctxKeyFgaServer).(*server.Server)
			if !ok {
				return nil, errors.New("no fga server") //nolint:goerr113
			}

			err := fga.writeLocalModel(authModel, fgaServer)
			if err != nil {
				return nil, err
			}
		} else {
			fgaClient, ok := ctx.Value(ctxKeyFgaClient).(*client.OpenFgaClient)
			if !ok {
				return nil, errors.New("couldnt get fgaClient") //nolint:goerr113
			}

			fga.fgaClient = fgaClient
		}

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		fga.user = ""
		fga.relation = ""
		fga.object = ""
		fga.conditionContext = nil
		fga.conditionName = ""
		fga.contextualTuples = []client.ClientContextualTupleKey{}
		fga.objectType = ""

		return ctx, nil
	})

	// Adding tuples using given
	ctx.Step(`^(\S+:\S+) is an? (\S+) of (\S+:\S+)$`, fga.thereIsATuple)
	ctx.Step(`^(\S+:\S+) is an? (\S+) of (\S+:\S+) with (\S+) being$`, fga.thereIsATupleWithCondition)
	ctx.Step(`^(\S+:\S+) is an? (\S+) of (\S+:\S+) with (\S+)$`, fga.thereIsATupleWithConditionNoValue)

	// Conditions
	ctx.When(`(\S+:\S+) (?:accesses|access) (\S+:\S+) with$`, fga.addContext)
	ctx.When(`context is$`, fga.conditionCheckNoUser)
	ctx.When(`^(?:(?:he|she|they)|(\S+:\S+)) (?:accesses|access) (\S+:\S+)`, fga.userAndObject)

	// Check related
	ctx.Then(HasRelationRegex, fga.checkHasRelation)
	ctx.Then(HasMultiRelationsRegex, fga.checkHasRelations)
	ctx.Then(DoesNotHaveRelationRegex, fga.checkDoesNotHaveRelation)
	ctx.Then(DoesNotHaveMultiRelationsRegex, fga.checkDoesNotHaveRelations)

	// List object related
	ctx.When(`^(?:(?:he|she|they)|(\S+:\S+)) searches for (\w+)$`, fga.addType)
	ctx.Then(HasPermissionsRegex, fga.listObjectsStep)
	ctx.Then(DoesNotHavePermissionsRegex, fga.listNoObjects)
}
