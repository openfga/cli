package storetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestStoreDataValidate(t *testing.T) {
	t.Parallel()

	validSingle := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Object: "doc:1", Assertions: map[string]bool{"read": true},
	}}}}}
	assert.NoError(t, validSingle.Validate())

	validUsers := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		Users: []string{"user:1", "user:2"}, Object: "doc:1", Assertions: map[string]bool{"read": true},
	}}}}}
	assert.NoError(t, validUsers.Validate())

	validObjects := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Objects: []string{"doc:1", "doc:2"}, Assertions: map[string]bool{"read": true},
	}}}}}
	assert.NoError(t, validObjects.Validate())

	invalidBoth := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Users: []string{"user:2"}, Object: "doc:1", Assertions: map[string]bool{"read": true},
	}}}}}
	require.Error(t, invalidBoth.Validate())

	invalidObjectBoth := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Object: "doc:1", Objects: []string{"doc:2"}, Assertions: map[string]bool{"read": true},
	}}}}}
	require.Error(t, invalidObjectBoth.Validate())

	invalidNone := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Assertions: map[string]bool{"read": true},
	}}}}}
	require.Error(t, invalidNone.Validate())
}

func TestModelTestCheckYAMLOmitEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    ModelTestCheck
		expected string
	}{
		{
			name: "only user field populated",
			input: ModelTestCheck{
				User:       "user:anne",
				Object:     "folder:product-2021",
				Assertions: map[string]bool{"can_view": true},
			},
			expected: `user: user:anne
object: folder:product-2021
assertions:
    can_view: true
`,
		},
		{
			name: "only users field populated",
			input: ModelTestCheck{
				Users:      []string{"user:anne", "user:bob"},
				Object:     "folder:product-2021",
				Assertions: map[string]bool{"can_view": true},
			},
			expected: `users:
    - user:anne
    - user:bob
object: folder:product-2021
assertions:
    can_view: true
`,
		},
		{
			name: "only object field populated",
			input: ModelTestCheck{
				User:       "user:anne",
				Object:     "folder:product-2021",
				Assertions: map[string]bool{"can_view": true},
			},
			expected: `user: user:anne
object: folder:product-2021
assertions:
    can_view: true
`,
		},
		{
			name: "only objects field populated",
			input: ModelTestCheck{
				User:       "user:anne",
				Objects:    []string{"folder:product-2021", "folder:product-2022"},
				Assertions: map[string]bool{"can_view": true},
			},
			expected: `user: user:anne
objects:
    - folder:product-2021
    - folder:product-2022
assertions:
    can_view: true
`,
		},
		{
			name: "users and objects fields populated",
			input: ModelTestCheck{
				Users:      []string{"user:anne", "user:bob"},
				Objects:    []string{"folder:product-2021", "folder:product-2022"},
				Assertions: map[string]bool{"can_view": true},
			},
			expected: `users:
    - user:anne
    - user:bob
objects:
    - folder:product-2021
    - folder:product-2022
assertions:
    can_view: true
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			yamlBytes, err := yaml.Marshal(tt.input)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, string(yamlBytes))

			// Also verify that empty fields are not present in the output
			yamlStr := string(yamlBytes)
			if tt.input.User == "" {
				assert.NotContains(t, yamlStr, "user: ")
			}
			if len(tt.input.Users) == 0 {
				assert.NotContains(t, yamlStr, "users:")
			}
			if tt.input.Object == "" {
				assert.NotContains(t, yamlStr, "object: ")
			}
			if len(tt.input.Objects) == 0 {
				assert.NotContains(t, yamlStr, "objects:")
			}
		})
	}
}
