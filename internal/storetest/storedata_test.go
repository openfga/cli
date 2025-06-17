package storetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	invalidBoth := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		User: "user:1", Users: []string{"user:2"}, Object: "doc:1", Assertions: map[string]bool{"read": true},
	}}}}}
	require.Error(t, invalidBoth.Validate())

	invalidNone := StoreData{Tests: []ModelTest{{Name: "t1", Check: []ModelTestCheck{{
		Object: "doc:1", Assertions: map[string]bool{"read": true},
	}}}}}
	require.Error(t, invalidNone.Validate())
}
