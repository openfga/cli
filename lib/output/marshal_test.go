package output

import "testing"

type DUTObject struct {
	Name   string
	Field1 string
}

func TestMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		v        any
		format   string
		output   string
		hasError bool
	}{
		{
			name: "standard",
			v: DUTObject{
				Name:   "Foo",
				Field1: "ABC",
			},
			output:   `{"Name":"Foo","Field1":"ABC"}`,
			format:   string(StandardJSON),
			hasError: false,
		},
		{
			name: "pretty",
			v: DUTObject{
				Name:   "Foo",
				Field1: "ABC",
			},
			output: `{
  "Name": "Foo",
  "Field1": "ABC"
}`,
			format:   string(PrettyJSON),
			hasError: false,
		},
		{
			name: "other",
			v: DUTObject{
				Name:   "Foo",
				Field1: "ABC",
			},
			output:   `{"Name":"Foo","Field1":"ABC"}`,
			format:   string(StandardJSON),
			hasError: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			out, err := Marshal(test.v, test.format)
			if err != nil {
				if !test.hasError {
					t.Error(err)
				}
			} else {
				if string(out) != test.output {
					t.Errorf("Expect %v actual %v", test.output, string(out))
				}
			}
		})
	}
}
