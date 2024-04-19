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
	"github.com/openfga/openfga/pkg/storage"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/comparison"
)

type FGAStepsContext struct {
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
	dataStore storage.OpenFGADatastore

	contextualTuples []client.ClientContextualTupleKey

	isLocalTest bool
}

func (f *FGAStepsContext) writeLocalModel(authModel *authorizationmodel.AuthzModel) error {
	model := authModel.GetProtoModel()
	f.storeID = ulid.Make().String()
	authModelWriteReq := &pb.WriteAuthorizationModelRequest{
		StoreId:         f.storeID,
		TypeDefinitions: model.GetTypeDefinitions(),
		SchemaVersion:   model.GetSchemaVersion(),
		Conditions:      model.GetConditions(),
	}

	writtenModel, err := f.fgaServer.WriteAuthorizationModel(context.Background(), authModelWriteReq)
	if err != nil {
		return err //nolint:wrapcheck
	}

	f.modelID = writtenModel.GetAuthorizationModelId()

	return nil
}

func (f *FGAStepsContext) tableToContext(table *godog.Table) (*structpb.Struct, error) {
	con := map[string]interface{}{}

	for _, row := range table.Rows {
		key := row.Cells[0].Value
		val := row.Cells[1].Value

		// Determine if we need to unmarshal from a JSON object or array
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
		return nil, fmt.Errorf("failed to create condition context due to %w", err)
	}

	return conditionContext, nil
}

func (f *FGAStepsContext) writeTuple() error {
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
			return fmt.Errorf("failed to write tuple due to %w", err)
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

func (f *FGAStepsContext) checkLocal() (bool, error) {
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

func (f *FGAStepsContext) checkRemote() (bool, error) {
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

func (f *FGAStepsContext) check(expected bool) error {
	var actual bool

	var err error

	if f.isLocalTest {
		actual, err = f.checkLocal()
	} else {
		actual, err = f.checkRemote()
	}

	if actual != expected || err != nil {
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

		if err != nil {
			errorString += fmt.Sprintf(", error=%v", err)
		}

		return errors.New(errorString) //nolint:goerr113
	}

	return nil
}

func (f *FGAStepsContext) listObjectsLocal() ([]string, error) {
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

func (f *FGAStepsContext) listObjectsRemote() ([]string, error) {
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

func (f *FGAStepsContext) listObjects(expected []string) error {
	var actual []string

	var err error

	if f.isLocalTest {
		actual, err = f.listObjectsLocal()
	} else {
		actual, err = f.listObjectsRemote()
	}

	if !comparison.CheckStringArraysEqual(actual, expected) || err != nil {
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

		if err != nil {
			errorString += fmt.Sprintf(", error=%v", err)
		}

		return errors.New(errorString) //nolint:goerr113
	}

	return nil
}

func (f *FGAStepsContext) thereIsATuple(user, relation, object string) error {
	f.user = user
	f.object = object
	f.relation = relation

	return f.writeTuple()
}

func (f *FGAStepsContext) thereIsATupleWithCondition(
	user, relation, object, condition string,
	table *godog.Table,
) error {
	f.user = user
	f.object = object
	f.relation = relation
	f.conditionName = condition

	conditionContext, err := f.tableToContext(table)
	if err != nil {
		return fmt.Errorf("failed to convert context for condition %s due to %w", condition, err)
	}

	f.conditionContext = conditionContext

	return f.writeTuple()
}

func (f *FGAStepsContext) thereIsATupleWithConditionNoValue(user, relation, object, condition string) error {
	f.user = user
	f.object = object
	f.relation = relation
	f.conditionName = condition

	return f.writeTuple()
}

func (f *FGAStepsContext) checkHasRelation(user, relation string) error {
	if strings.Contains(user, ":") {
		f.user = user
	}

	f.relation = relation

	return f.check(true)
}

func (f *FGAStepsContext) checkHasRelations(user string, table *godog.Table) error {
	if strings.Contains(user, ":") {
		f.user = user
	}

	var errs []error

	for _, row := range table.Rows {
		f.relation = row.Cells[0].Value

		err := f.check(true)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (f *FGAStepsContext) checkDoesNotHaveRelation(user, relation string) error {
	if strings.Contains(user, ":") {
		f.user = user
	}

	f.relation = relation

	return f.check(false)
}

func (f *FGAStepsContext) checkDoesNotHaveRelations(user string, table *godog.Table) error {
	if strings.Contains(user, ":") {
		f.user = user
	}

	var errs []error

	for _, row := range table.Rows {
		f.relation = row.Cells[0].Value

		err := f.check(false)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (f *FGAStepsContext) addContext(user, object string, table *godog.Table) error {
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

func (f *FGAStepsContext) userAndObject(user, object string) {
	if strings.Contains(user, ":") {
		f.user = user
	}

	f.object = object
}

func (f *FGAStepsContext) conditionCheckNoUser(table *godog.Table) error {
	return f.addContext("", "", table)
}

func (f *FGAStepsContext) addType(user, objectType string) {
	if user != "" {
		f.user = user
	}

	f.objectType = objectType
}

func (f *FGAStepsContext) listObjectsStep(user, relation string, table *godog.Table) error {
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

func (f *FGAStepsContext) listNoObjects(user, relation string) error {
	if user != "" {
		f.user = user
	}

	f.relation = relation

	return f.listObjects([]string{})
}
