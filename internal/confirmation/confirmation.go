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

package confirmation

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// askForConfirmation is the internal implementation that prompt user to confirm their choice.
// Default answer is no.
func askForConfirmation(ioReader *bufio.Reader, question string) (bool, error) {
	for {
		_, err := fmt.Fprintln(os.Stdout, question+"(y/N)")
		if err != nil {
			return false, fmt.Errorf("unable to ask for confirmation with error %w", err)
		}

		s, err := ioReader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("unable to read string with error %w", err)
		}

		trimmedString := strings.ToLower(strings.TrimSpace(s))
		switch trimmedString {
		case "":
			return false, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}
	}
}

// AskForConfirmation will prompt user to confirm their choice.
// Default answer is no.
func AskForConfirmation(question string) (bool, error) {
	return askForConfirmation(bufio.NewReader(os.Stdin), question)
}
