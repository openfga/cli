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

// Package _import contains commands to manage import job into OpenFGA.
package _import

import (
	"github.com/spf13/cobra"
)

// ImportCmd represents the store command.
var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Create a job to insert a large volume of tuples into OpenFGA",
	Long:  "import jobs are backed by database that can track the successful inserts of tuples and we can retry failed inserts or follow the status of the job.",
}

func init() {
	ImportCmd.AddCommand(createCmd)
	ImportCmd.AddCommand(statusCmd)
	ImportCmd.AddCommand(retryCmd)
	ImportCmd.AddCommand(listCmd)
}
