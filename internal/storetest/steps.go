package storetest

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/cucumber/godog"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/memory"

	"github.com/openfga/cli/internal/authorizationmodel"
)

// Extracted from ctx.Step calls to allow reuse in the formatter to determine what are check or listobject steps.
var (
	// Check regex.
	HasRelationRegex               = regexp.MustCompile(`^(?:(?:he|she|they)|(\S+:\S+)) (?:has|have) the (\S+) permission$`)              //nolint:lll
	HasMultiRelationsRegex         = regexp.MustCompile(`^(?:(?:he|she|they)|(\S+:\S+)) (?:has|have) the following permissions$`)         //nolint:lll
	DoesNotHaveRelationRegex       = regexp.MustCompile(`^(?:(?:he|she|they)|(\S+:\S+)) (?:does|do) not have the (\S+) permission$`)      //nolint:lll
	DoesNotHaveMultiRelationsRegex = regexp.MustCompile(`^(?:(?:he|she|they)|(\S+:\S+)) (?:does|do) not have the following permissions$`) //nolint:lll

	// ListObjects regex.
	HasPermissionsRegex         = regexp.MustCompile(`^(?:(?:he|she|they)|(\S+:\S+)) (?:has|have) the (\S+) permission for$`) //nolint:lll
	DoesNotHavePermissionsRegex = regexp.MustCompile(`^(?:(?:he|she|they)|(\S+:\S+)) (?:has|have) no (\S+) permission$`)
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	fga := FGAStepsContext{}

	// Handle spinning up the server and write the model
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		authModel, ok := ctx.Value(ctxKeyAuthModel).(*authorizationmodel.AuthzModel)
		if !ok {
			return nil, errors.New("couldn't determine auth model to be written") //nolint:goerr113
		}

		fga.isLocalTest = authModel != nil

		if fga.isLocalTest {
			fga.dataStore = memory.New()

			fgaServer, err := server.NewServerWithOpts(
				server.WithDatastore(fga.dataStore),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to start fga server due to %w", err)
			}

			fga.fgaServer = fgaServer

			err = fga.writeLocalModel(authModel)
			if err != nil {
				return nil, err
			}
		} else {
			fgaClient, ok := ctx.Value(ctxKeyFgaClient).(*client.OpenFgaClient)
			if !ok {
				return nil, errors.New("couldn't get fga client") //nolint:goerr113
			}

			fga.fgaClient = fgaClient
		}

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		fga.fgaServer.Close()
		fga.dataStore.Close()

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
