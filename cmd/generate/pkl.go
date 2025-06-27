package generate

import (
	"encoding/json"
	"fmt"
	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/generate"
	"github.com/openfga/cli/internal/output"
	openfga "github.com/openfga/go-sdk"
	"github.com/spf13/cobra"
)

var pklCmd = &cobra.Command{
	Use:   "pkl",
	Short: "Generate pkl test utilities",
	Long:  "Generate pkl test utilities based on the given model",
	Example: `fga generate pkl --file=model.json
fga generate --file=fga.mod
fga generate '{"type_definitions":[{"type":"user"},{"type":"document","relations":{"can_view":{"this":{}}},"metadata":{"relations":{"can_view":{"directly_related_user_types":[{"type":"user"}]}}}}],"schema_version":"1.1"}' --format=json
fga generate --file=fga.mod --out=testing
fga generate --file=fga.mod --out=testing --config='{"user": {"base_type_name": "Awesome"}}'`, //nolint:lll
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		_, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		var inputModel string
		if err := authorizationmodel.ReadFromInputFileOrArg(
			cmd,
			args,
			"file",
			false,
			&inputModel,
			openfga.PtrString(""),
			&writeInputFormat); err != nil {
			return err //nolint:wrapcheck
		}

		authModel := authorizationmodel.AuthzModel{}

		err = authModel.ReadModelFromString(inputModel, writeInputFormat)
		if err != nil {
			return err //nolint:wrapcheck
		}

		out, err := cmd.Flags().GetString("out")
		if err != nil {
			return fmt.Errorf("failed to parse output directory due to %w", err)
		}
		config, err := cmd.Flags().GetString("config")
		if err != nil {
			return fmt.Errorf("failed to parse config due to %w", err)
		}
		cn := make(map[string]generate.PklConventionConfig)
		err = json.Unmarshal([]byte(config), &cn)
		if err != nil {
			return fmt.Errorf("failed to parse config content due to %w", err)
		}

		g := &generate.PklGenerator{
			Model:      authModel.TypeDefinitions,
			Convention: &generate.PklConvention{Config: cn},
		}
		files, err := g.Generate()
		err = files.SaveAll(out)
		if err != nil {
			return fmt.Errorf("failed to save generated files due to %w", err)
		}
		return output.Display(fmt.Sprintf("generated files in directory %v successfully", out))
	},
}

var writeInputFormat = authorizationmodel.ModelFormatDefault

func init() {
	pklCmd.Flags().String("out", "testing", "Output of testing directory")
	pklCmd.Flags().String("config", "{}", "Generator configurations")
	pklCmd.Flags().String("file", "", "File Name. The file should have the model in the JSON or DSL format")
	pklCmd.Flags().Var(&writeInputFormat, "format", `Authorization model input format. Can be "fga", "json", or "modular"`) //nolint:lll
}
