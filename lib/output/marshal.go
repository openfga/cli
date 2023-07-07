package output

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	json "github.com/neilotoole/jsoncolor"
	"github.com/nwidger/jsoncolor"
)

// EmptyStruct is used when we wish to return an empty object.
type EmptyStruct struct{}

func displayColorTerminal(data any) error {
	// create custom formatter
	f := jsoncolor.NewFormatter()

	dst, err := jsoncolor.MarshalWithFormatter(data, f)
	if err != nil {
		return fmt.Errorf("unable to display output with error %w", err)
	}

	fmt.Println(string(dst))

	return nil
}

func displayTerminal(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	err := enc.Encode(data)
	if err != nil {
		return fmt.Errorf("unable to encode output with error %w", err)
	}

	return nil
}

func outputNonTerminal(data any) error {
	result, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal json with error %w", err)
	}

	fmt.Println(string(result))

	return nil
}

// Display will decorate the output if possible.  Otherwise, will print out the standard JSON.
func Display(data any) error {
	if json.IsColorTerminal(os.Stdout) {
		return displayColorTerminal(data)
	} else if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return displayTerminal(data)
	}

	return outputNonTerminal(data)
}
