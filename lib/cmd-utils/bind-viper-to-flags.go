package cmdutils

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// BindViperToFlags recursively binds viper configs to cobra commands and subcommands.
func BindViperToFlags(cmd *cobra.Command, viperInstance *viper.Viper) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		configName := flag.Name

		if !flag.Changed && viperInstance.IsSet(configName) {
			value := viperInstance.Get(configName)
			err := cmd.Flags().Set(flag.Name, fmt.Sprintf("%v", value))
			cobra.CheckErr(err)
		}
	})

	for _, subcmd := range cmd.Commands() {
		BindViperToFlags(subcmd, viperInstance)
	}
}
