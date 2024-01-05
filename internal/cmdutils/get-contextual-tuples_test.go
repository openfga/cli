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

package cmdutils_test

import (
	"testing"

	"github.com/openfga/go-sdk/client"

	"github.com/openfga/cli/internal/cmdutils"
)

type tupleTestPassing struct {
	raw    []string
	parsed []client.ClientTupleKey
}

type tupleTestFailing struct {
	raw []string
}

func TestGetContextualTuplesWithNoError(t *testing.T) {
	t.Parallel()

	tests := []tupleTestPassing{{
		raw: []string{"user:anne can_view document:2"},
		parsed: []client.ClientTupleKey{
			{User: "user:anne", Relation: "can_view", Object: "document:2"},
		},
	}, {
		raw: []string{
			"group:product#member owner document:roadmap",
			"user:beth can_delete folder:marketing",
			"user:carl can_share repo:openfga/openfga",
		},
		parsed: []client.ClientTupleKey{
			{User: "group:product#member", Relation: "owner", Object: "document:roadmap"},
			{User: "user:beth", Relation: "can_delete", Object: "folder:marketing"},
			{User: "user:carl", Relation: "can_share", Object: "repo:openfga/openfga"},
		},
	}, {
		// Note that this is an invalid tuple, but the server will let us know that.
		// We can validate against it in a future iteration
		raw: []string{"anne can_view document-2"},
		parsed: []client.ClientTupleKey{
			{User: "anne", Relation: "can_view", Object: "document-2"},
		},
	}}

	for index := 0; index < len(tests); index++ {
		test := tests[index]
		t.Run("TestGetContextualTuplesWithNoError"+string(rune(index)), func(t *testing.T) {
			t.Parallel()
			tuples, err := cmdutils.ParseContextualTuplesInner(test.raw)
			if err != nil {
				t.Error(err)
			}

			if len(tuples) != len(test.parsed) {
				t.Errorf("Expected parsed tuples to have length %v actual %v", len(test.parsed), len(tuples))
			}

			for index := 0; index < len(tuples); index++ {
				if tuples[index].User != test.parsed[index].User ||
					tuples[index].Relation != test.parsed[index].Relation ||
					tuples[index].Object != test.parsed[index].Object {
					t.Errorf("Expected parsed tuples to match %v actual %v", test.parsed, tuples)
				}
			}
		})
	}
}

func TestGetContextualTuplesWithError(t *testing.T) {
	t.Parallel()

	tests := []tupleTestFailing{{
		raw: []string{"user:anne can_view"},
	}, {
		raw: []string{"can_view document:2"},
	}, {
		raw: []string{"can_view"},
	}, {
		raw: []string{"this is not a valid tuple"},
	}, {
		raw: []string{
			"group:product#member owner document:roadmap",
			"user:beth can_delete folder:marketing",
			"user:carl can_share repo:openfga/openfga",
			"user:dan can_share",
		},
	}}

	for index := 0; index < len(tests); index++ {
		test := tests[index]
		t.Run("TestGetContextualTuplesWithNoError"+string(rune(index)), func(t *testing.T) {
			t.Parallel()
			_, err := cmdutils.ParseContextualTuplesInner(test.raw)

			if err == nil {
				t.Error("Expect error but there is none")
			}
		})
	}
}
