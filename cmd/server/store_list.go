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

var storeListCmd = &cobra.Command{
	Use:     "list <server-id>",
	Short:   "List stores within a server connection",
	Long:    `List all known stores associated with a server connection.`,
	Args:    cobra.ExactArgs(1),
	Example: `  fga server store list 01ABCDEFGHJKMNPQRSTVWXYZ`,
	RunE: func(_ *cobra.Command, args []string) error {
		serverID := args[0]

		cs, err := serve.NewConfigStore()
		if err != nil {
			return err
		}

		srv, err := cs.FindServerByID(serverID)
		if err != nil {
			return err
		}
		if srv == nil {
			return fmt.Errorf("server %q not found", serverID) //nolint:err113
		}

		if len(srv.Stores) == 0 {
			fmt.Fprintln(os.Stderr, "No stores found. Use 'fga server store add' to add one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "STORE ID\tALIAS\tMODEL ID\tAUTH OVERRIDE")
		for _, st := range srv.Stores {
			hasAuth := "no"
			if st.Auth != nil {
				hasAuth = "yes (" + string(st.Auth.Method) + ")"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", st.StoreID, st.Alias, st.ModelID, hasAuth)
		}
		return w.Flush()
	},
}
