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

func TestModelTestCheckStructTagsOmitEmpty(t *testing.T) {
	t.Parallel()

	// Test that struct can handle single user/object
	checkSingle := ModelTestCheck{
		User:       "user:anne",
		Object:     "folder:product-2021",
		Assertions: map[string]bool{"can_view": true},
	}
	assert.NotEmpty(t, checkSingle.User)
	assert.Empty(t, checkSingle.Users)
	assert.NotEmpty(t, checkSingle.Object)
	assert.Empty(t, checkSingle.Objects)

	// Test that struct can handle multiple users
	checkUsers := ModelTestCheck{
		Users:      []string{"user:anne", "user:bob"},
		Object:     "folder:product-2021",
		Assertions: map[string]bool{"can_view": true},
	}
	assert.Empty(t, checkUsers.User)
	assert.NotEmpty(t, checkUsers.Users)
	assert.Len(t, checkUsers.Users, 2)

	// Test that struct can handle multiple objects
	checkObjects := ModelTestCheck{
		User:       "user:anne",
		Objects:    []string{"folder:product-2021", "folder:product-2022"},
		Assertions: map[string]bool{"can_view": true},
	}
	assert.NotEmpty(t, checkObjects.User)
	assert.Empty(t, checkObjects.Object)
	assert.NotEmpty(t, checkObjects.Objects)
	assert.Len(t, checkObjects.Objects, 2)
}
