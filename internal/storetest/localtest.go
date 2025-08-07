package storetest

import (
	"context"

	pb "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/openfga/cli/internal/authorizationmodel"
)

func RunSingleLocalCheckTest(
	ctx context.Context,
	fgaServer *server.Server,
	checkRequest *pb.CheckRequest,
) (*pb.CheckResponse, error) {
	return fgaServer.Check(ctx, checkRequest) //nolint:wrapcheck
}

func RunLocalCheckTest(
	ctx context.Context,
	fgaServer *server.Server,
	checkTest ModelTestCheck,
	tuples []client.ClientContextualTupleKey,
	options ModelTestOptions,
) []ModelTestCheckSingleResult {
	results := []ModelTestCheckSingleResult{}
	users := GetEffectiveUsers(checkTest)

	objects := GetEffectiveObjects(checkTest)
	for _, user := range users {
		for _, object := range objects {
			for relation, expectation := range checkTest.Assertions {
				result := ModelTestCheckSingleResult{
					Request: client.ClientCheckRequest{
						User:             user,
						Relation:         relation,
						Object:           object,
						ContextualTuples: tuples,
						Context:          checkTest.Context,
					},
					Expected: expectation,
				}

				var (
					reqCtx *structpb.Struct
					err    error
				)

				if checkTest.Context != nil {
					reqCtx, err = structpb.NewStruct(*checkTest.Context)
				}

				if err != nil {
					result.Error = err
				} else {
					response, err := RunSingleLocalCheckTest(ctx, fgaServer,
						&pb.CheckRequest{
							StoreId:              *options.StoreID,
							AuthorizationModelId: *options.ModelID,
							TupleKey: &pb.CheckRequestTupleKey{
								User:     user,
								Relation: relation,
								Object:   object,
							},
							Context: reqCtx,
						},
					)
					if err != nil {
						result.Error = err
					}

					if response != nil {
						result.Got = &response.Allowed
						result.TestResult = result.IsPassing()
					}
				}

				results = append(results, result)
			}
		}
	}

	return results
}

func RunSingleLocalListObjectsTest(
	ctx context.Context,
	fgaServer *server.Server,
	listObjectsRequest *pb.ListObjectsRequest,
) (*pb.ListObjectsResponse, error) {
	return fgaServer.ListObjects(ctx, listObjectsRequest) //nolint:wrapcheck
}

func RunLocalListObjectsTest(
	ctx context.Context,
	fgaServer *server.Server,
	listObjectsTest ModelTestListObjects,
	tuples []client.ClientContextualTupleKey,
	options ModelTestOptions,
) []ModelTestListObjectsSingleResult {
	results := []ModelTestListObjectsSingleResult{}

	for relation, expectation := range listObjectsTest.Assertions {
		result := ModelTestListObjectsSingleResult{
			Request: client.ClientListObjectsRequest{
				User:             listObjectsTest.User,
				Relation:         relation,
				Type:             listObjectsTest.Type,
				ContextualTuples: tuples,
				Context:          listObjectsTest.Context,
			},
			Expected: expectation,
		}

		var (
			reqCtx *structpb.Struct
			err    error
		)

		if listObjectsTest.Context != nil {
			reqCtx, err = structpb.NewStruct(*listObjectsTest.Context)
		}

		if err != nil {
			result.Error = err
		} else {
			response, err := RunSingleLocalListObjectsTest(ctx, fgaServer,
				&pb.ListObjectsRequest{
					StoreId:              *options.StoreID,
					AuthorizationModelId: *options.ModelID,
					User:                 listObjectsTest.User,
					Type:                 listObjectsTest.Type,
					Relation:             relation,
					Context:              reqCtx,
				},
			)
			if err != nil {
				result.Error = err
			}

			if response != nil {
				result.Got = response.GetObjects()
				result.TestResult = result.IsPassing()
			}
		}

		results = append(results, result)
	}

	return results
}

func RunSingleLocalListUsersTest(
	ctx context.Context,
	fgaServer *server.Server,
	listUsersRequest *pb.ListUsersRequest,
) (*pb.ListUsersResponse, error) {
	return fgaServer.ListUsers(ctx, listUsersRequest) //nolint:wrapcheck
}

func RunLocalListUsersTest(
	ctx context.Context,
	fgaServer *server.Server,
	listUsersTest ModelTestListUsers,
	tuples []client.ClientContextualTupleKey,
	options ModelTestOptions,
) []ModelTestListUsersSingleResult {
	results := []ModelTestListUsersSingleResult{}

	object, pbObject := convertStoreObjectToObject(listUsersTest.Object)

	userFilter := &pb.UserTypeFilter{
		Type:     listUsersTest.UserFilter[0].GetType(),
		Relation: listUsersTest.UserFilter[0].GetRelation(),
	}

	for relation, expectation := range listUsersTest.Assertions {
		result := ModelTestListUsersSingleResult{
			Request: client.ClientListUsersRequest{
				Object:           object,
				Relation:         relation,
				UserFilters:      listUsersTest.UserFilter,
				ContextualTuples: tuples,
				Context:          listUsersTest.Context,
			},
			Expected: expectation,
		}

		var (
			reqCtx *structpb.Struct
			err    error
		)

		if listUsersTest.Context != nil {
			reqCtx, err = structpb.NewStruct(*listUsersTest.Context)
		}

		if err != nil {
			result.Error = err
		} else {
			response, err := RunSingleLocalListUsersTest(ctx, fgaServer,
				&pb.ListUsersRequest{
					StoreId:              *options.StoreID,
					AuthorizationModelId: *options.ModelID,
					Object:               pbObject,
					Relation:             relation,
					UserFilters:          []*pb.UserTypeFilter{userFilter},
					Context:              reqCtx,
				},
			)
			if err != nil {
				result.Error = err
			}

			if response != nil {
				result.Got = ModelTestListUsersAssertion{
					Users: convertPbUsersToStrings(response.GetUsers()),
				}
				result.TestResult = result.IsPassing()
			}
		}

		results = append(results, result)
	}

	return results
}

func RunLocalTest(
	ctx context.Context,
	fgaServer *server.Server,
	test ModelTest,
	tuples []client.ClientContextualTupleKey,
	model *authorizationmodel.AuthzModel,
) (TestResult, error) {
	checkResults := []ModelTestCheckSingleResult{}
	listObjectResults := []ModelTestListObjectsSingleResult{}
	listUsersResults := []ModelTestListUsersSingleResult{}

	storeID, modelID, err := initLocalStore(ctx, fgaServer, model.GetProtoModel(), tuples)
	if err != nil {
		return TestResult{}, err
	}

	testOptions := ModelTestOptions{
		StoreID: storeID,
		ModelID: modelID,
	}

	for index := range test.Check {
		results := RunLocalCheckTest(ctx, fgaServer, test.Check[index], tuples, testOptions)
		checkResults = append(checkResults, results...)
	}

	for index := range test.ListObjects {
		results := RunLocalListObjectsTest(ctx, fgaServer, test.ListObjects[index], tuples, testOptions)
		listObjectResults = append(listObjectResults, results...)
	}

	for index := range test.ListUsers {
		results := RunLocalListUsersTest(ctx, fgaServer, test.ListUsers[index], tuples, testOptions)
		listUsersResults = append(listUsersResults, results...)
	}

	return TestResult{
		Name:               test.Name,
		Description:        test.Description,
		CheckResults:       checkResults,
		ListObjectsResults: listObjectResults,
		ListUsersResults:   listUsersResults,
	}, nil
}
