package output

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// Marshal encodes the object in JSON and format it accordingly.
func Marshal(v any, format string) ([]byte, error) {
	switch format {
	case string(StandardJSON):
		return json.Marshal(v) //nolint:wrapcheck
	case string(PrettyJSON):
		return json.MarshalIndent(v, "", "  ") //nolint:wrapcheck
	default:
		return []byte{}, ErrUnknownFormat
	}
}

func Display(cmd *cobra.Command, v any) error {
	format := cmd.Flags().Lookup("format").Value

	output, err := Marshal(v, format.String())
	if err != nil {
		return fmt.Errorf("unable to marshal %w ", err)
	}

	fmt.Print(string(output))

	return nil
}
