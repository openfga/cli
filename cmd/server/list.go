/*
Copyright © 2023 OpenFGA

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/serve"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List OpenFGA server connections",
	Long:    `List all server connections stored in ~/.config/fga/servers.yaml.`,
	Example: `  fga server list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cs, err := serve.NewConfigStore()
		if err != nil {
			return err
		}

		cfg, err := cs.Read()
		if err != nil {
			return err
		}

		if len(cfg.Servers) == 0 {
			fmt.Fprintln(os.Stderr, "No servers found. Use 'fga server add' to create one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tAPI URL\tAUTH TYPE\tSTORES\tSTORE CRUD\tSTORE LIST")
		for _, s := range cfg.Servers {
			storeCRUD := "yes"
			if !s.Capabilities.StoreCRUD {
				storeCRUD = "no"
			}
			storeList := "yes"
			if !s.Capabilities.StoreList {
				storeList = "no"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
				s.ID, s.Name, s.APIURL, s.Auth.Method, len(s.Stores), storeCRUD, storeList)
		}
		return w.Flush()
	},
}
