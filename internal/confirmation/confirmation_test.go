package confirmation

import (
	"bufio"
	"strings"
	"testing"
)

func TestConfirmation(t *testing.T) {
	t.Parallel()

	type test struct {
		_name  string
		input  string
		result bool
	}

	tests := []test{
		{
			_name:  "default",
			input:  "\n",
			result: false,
		},
		{
			_name:  "yes",
			input:  "  Yes\n",
			result: true,
		},

		{
			_name:  "Y",
			input:  "y\n",
			result: true,
		},
		{
			_name:  "NO",
			input:  "NO\n",
			result: false,
		},
		{
			_name:  "n",
			input:  "n\n",
			result: false,
		},
		{
			_name:  "other answer then no",
			input:  "other\nn\n",
			result: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test._name, func(t *testing.T) {
			t.Parallel()

			result, err := askForConfirmation(bufio.NewReader(strings.NewReader(test.input)), "test")
			if err != nil {
				t.Error(err)
			}

			if result != test.result {
				t.Errorf("Expect result %v actual %v", test.result, result)
			}
		})
	}
}
