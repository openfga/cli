package generate

import "github.com/spf13/cobra"

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code",
	Long:  "Generate code to facilitate testing, modeling and working with OpenFGA.",
}

func init() {
	GenerateCmd.AddCommand(pklCmd)
}
