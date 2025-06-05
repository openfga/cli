# Store File Format

A store file (typically named `*.fga.yaml`) describes everything needed to recreate an OpenFGA store.
It can be exported using `fga store export` and imported using `fga store import`.
The same format can be used with `fga model test` when providing a model and test suite in a single file.

## Top Level Fields

- `name` (string, optional): Name of the store.
- `model` (string, optional): Authorization model in [OpenFGA DSL](https://openfga.dev/docs/modeling/overview) or JSON.
- `model_file` (string, optional): Path to a file containing the authorization model. If both `model` and `model_file` are provided the inline `model` value is used.
- `tuples` (array, optional): List of tuples that will be written to the store.
- `tuple_file` (string, optional): Path to a file containing tuples in YAML, JSON or CSV format. Tuples from the file are appended to any tuples listed in `tuples`.
- `tests` (array, optional): List of tests to execute against the store. These are the same as the tests used by `fga model test`.

## Tests

Each entry under `tests` may contain the following fields:

- `name` (string): Name of the test.
- `description` (string, optional): Description of what is being tested.
- `tuples` (array, optional): Additional tuples used only for this test.
- `tuple_file` (string, optional): File containing tuples to be loaded for this test.
- `check` (array, optional): List of check requests with expected boolean results.
- `list_objects` (array, optional): List objects requests with expected object ids per relation.
- `list_users` (array, optional): List users requests with expected users per relation.

### Check

A check entry consists of:

- `user` (string): User to check.
- `object` (string): Object to check against.
- `assertions` (map[string]bool): Expected boolean result for each relation.

### List Objects

A list objects entry consists of:

- `user` (string): User to list objects for.
- `type` (string): Object type to list.
- `context` (map, optional): Contextual data for [conditions](https://openfga.dev/docs/concepts/conditions).
- `assertions` (map[string][]string): Expected object ids for each relation.

### List Users

A list users entry consists of:

- `object` (string): Object to list users for.
- `user_filter` (array): Filters describing which users to return.
- `context` (map, optional): Contextual data for conditions.
- `assertions` (map[string].users): Expected users for each relation. Each relation contains a `users` array.

## Example

```yaml
name: Test
model: |+
  model
    schema 1.1

  type user
  type group
    relations
      define member: [user]
      define moderator: [user]

tuples:
  - user: user:1
    relation: member
    object: group:admins
  - user: user:1
    relation: member
    object: group:employees

tests:
  - name: Example
    check:
      - user: user:1
        object: group:admins
        assertions:
          member: true
    list_objects:
      - user: user:1
        type: group
        assertions:
          member:
            - group:admins
            - group:employees
    list_users:
      - object: group:admins
        user_filter:
          - type: user
        assertions:
          member:
            users:
              - user:1
```

Additional examples can be found in the [`example/`](../example/) directory and in the [sample-stores repository](https://github.com/openfga/sample-stores/).
