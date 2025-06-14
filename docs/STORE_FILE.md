# OpenFGA Store File Format

The store file is a YAML configuration file (`.fga.yaml`) that defines a complete OpenFGA store setup, including the authorization model, relationship tuples, and test cases. This file format enables easy management, testing, and deployment of OpenFGA configurations.

## File Structure

The store file uses YAML syntax and supports the following top-level properties:

```yaml
name: "Store Name"                    # Required: Name of the store
model_file: "./model.fga"             # Path to authorization model file
model: |                              # OR inline model definition
  model
    schema 1.1
  type user
  # ... more model definitions

tuple_file: "./tuples.yaml"           # Path to tuples file  
tuples:                               # OR inline tuples
  - user: user:anne
    relation: viewer
    object: document:1

tests:                                # Test definitions
  - name: "test-name"
    description: "Test description"   # Optional
    tuple_file: "./test-tuples.yaml"  # Test-specific tuples file
    tuples:                           # OR inline test tuples
      - user: user:bob
        relation: editor
        object: document:2
    check:                            # Authorization checks
      - user: user:anne
        object: document:1
        context:                      # Optional context for ABAC
          timestamp: "2023-05-03T21:25:23+00:00"
        assertions:
          viewer: true
          editor: false
    list_objects:                     # List objects tests
      - user: user:anne
        type: document
        context:                      # Optional context
          timestamp: "2023-05-03T21:25:23+00:00"
        assertions:
          viewer:
            - document:1
            - document:2
    list_users:                       # List users tests
      - object: document:1
        user_filter:
          - type: user
        context:                      # Optional context
          timestamp: "2023-05-03T21:25:23+00:00"
        assertions:
          viewer:
            users:
              - user:anne
              - user:bob
```

## Core Components

### 1. Store Metadata

- **`name`** (required): The display name for the store
- This name is used when creating a new store via import

### 2. Authorization Model

You can specify the authorization model in two ways:

#### Option A: External File Reference
```yaml
model_file: "./path/to/model.fga"
```

#### Option B: Inline Model Definition
```yaml
model: |
  model
    schema 1.1
  
  type user
  
  type document
    relations
      define viewer: [user]
      define editor: [user] and viewer
```

The model defines the authorization schema including:
- Types (user, document, folder, etc.)
- Relations (viewer, editor, owner, etc.)  
- Authorization rules and conditions

### 3. Relationship Tuples

Tuples define the actual relationships between users and objects. You can specify them in two ways:

#### Option A: External File Reference
```yaml
tuple_file: "./path/to/tuples.yaml"
```

#### Option B: Inline Tuple Definition
```yaml
tuples:
  - user: user:anne
    relation: viewer  
    object: document:1
  - user: user:bob
    relation: editor
    object: document:1
    condition:                        # Optional: for conditional relationships
      name: valid_ip
      context:
        ip_address: "192.168.1.100"
```

**Supported tuple file formats:**
- YAML (`.yaml`, `.yml`)
- JSON (`.json`)
- CSV (`.csv`)

### 4. Tests

The `tests` array contains test cases to validate your authorization model and tuples.

#### Test Structure
Each test can include:
- **`name`** (required): Test identifier
- **`description`** (optional): Human-readable test description
- **`tuple_file`** or **`tuples`**: Test-specific relationship tuples (appended to global tuples)
- **`check`**: Authorization check tests
- **`list_objects`**: List objects API tests  
- **`list_users`**: List users API tests

#### Check Tests
Validate whether a user has specific relations to an object:

```yaml
check:
  - user: user:anne
    object: document:1
    context:                          # Optional: for ABAC conditions
      current_time: "2023-05-03T21:25:23+00:00"
      user_ip: "192.168.0.1"
    assertions:
      viewer: true                    # Expected result
      editor: false
      owner: false
```

#### List Objects Tests
Validate which objects a user can access:

```yaml
list_objects:
  - user: user:anne
    type: document                    # Object type to query
    context:                          # Optional context
      current_time: "2023-05-03T21:25:23+00:00"
    assertions:
      viewer:                         # Objects user can view
        - document:1
        - document:2  
      editor:                         # Objects user can edit
        - document:1
```

#### List Users Tests
Validate which users have access to an object:

```yaml
list_users:
  - object: document:1
    user_filter:                      # Filter by user types
      - type: user
      - type: team
    context:                          # Optional context
      current_time: "2023-05-03T21:25:23+00:00"
    assertions:
      viewer:
        users:
          - user:anne
          - user:bob
```

## Context Support (ABAC)

The store file supports Attribute-Based Access Control (ABAC) through contextual information:

```yaml
# In tuples - for conditional relationships
tuples:
  - user: user:anne
    relation: viewer
    object: document:1
    condition:
      name: non_expired_grant
      context:
        grant_timestamp: "2023-05-03T21:25:20+00:00"
        grant_duration: "10m"

# In tests - for contextual evaluations
tests:
  - name: "time-based-access"
    check:
      - user: user:anne
        object: document:1
        context:
          current_timestamp: "2023-05-03T21:25:23+00:00"
        assertions:
          viewer: true
```

## File Composition

The store file supports flexible composition:

### Global + Test-Specific Data
- **Global tuples**: Applied to all tests
- **Test-specific tuples**: Appended to global tuples for individual tests
- Both `tuple_file` and `tuples` can be used together

### Mixed Inline and File References
```yaml
name: "Mixed Example"
model_file: "./model.fga"            # Model from file
tuples:                              # Inline global tuples
  - user: user:admin
    relation: owner
    object: system:main
tests:
  - name: "test-1"
    tuple_file: "./test1-tuples.yaml" # Additional tuples from file
    check:
      - user: user:admin
        object: system:main
        assertions:
          owner: true
```

## CLI Commands Using Store Files

### Store Import
Import a complete store configuration:
```bash
fga store import --file store.fga.yaml
```

### Model Testing  
Run tests against an authorization model:
```bash
fga model test --tests store.fga.yaml
```

### Store Export
Export store configuration to file:
```bash
fga store export --store-id 01H0H015178Y2V4CX10C2KGHF4 > exported-store.fga.yaml
```

## Examples

### Basic Store File
```yaml
name: "Document Management"
model_file: "./authorization-model.fga"
tuple_file: "./relationships.yaml"
tests:
  - name: "basic-permissions"
    check:
      - user: user:alice
        object: document:readme
        assertions:
          viewer: true
          editor: false
```

### Advanced Store with ABAC
```yaml
name: "Time-Based Access"
model: |
  model
    schema 1.1
  
  type user
  type document
    relations
      define viewer: [user with non_expired_grant]

  condition non_expired_grant(current_time: timestamp, grant_time: timestamp, duration: duration) {
    current_time < grant_time + duration
  }

tuples:
  - user: user:bob
    relation: viewer
    object: document:secret
    condition:
      name: non_expired_grant
      context:
        grant_time: "2023-05-03T21:25:20+00:00"
        duration: "1h"

tests:
  - name: "time-expiry-test"
    check:
      - user: user:bob
        object: document:secret
        context:
          current_time: "2023-05-03T21:30:00+00:00"  # Within 1 hour
        assertions:
          viewer: true
      - user: user:bob
        object: document:secret  
        context:
          current_time: "2023-05-03T22:30:00+00:00"  # After 1 hour
        assertions:
          viewer: false
```

### Multi-Test Store File
```yaml
name: "Comprehensive Testing"
model_file: "./model.fga"
tuple_file: "./base-tuples.yaml"

tests:
  - name: "admin-permissions"
    tuples:
      - user: user:admin
        relation: owner
        object: system:config
    check:
      - user: user:admin
        object: system:config
        assertions:
          owner: true
          viewer: true
    list_objects:
      - user: user:admin
        type: system
        assertions:
          owner:
            - system:config

  - name: "user-permissions" 
    tuple_file: "./user-test-tuples.yaml"
    check:
      - user: user:john
        object: document:public
        assertions:
          viewer: true
          editor: false
    list_users:
      - object: document:public
        user_filter:
          - type: user
        assertions:
          viewer:
            users:
              - user:john
              - user:jane
```

## Best Practices

1. **Use descriptive names**: Make store and test names clear and meaningful
2. **Organize with external files**: For complex models, use separate `.fga` files for models and `.yaml` files for tuples
3. **Comprehensive testing**: Include check, list_objects, and list_users tests to validate all API behaviors
4. **Context testing**: When using ABAC, test both positive and negative cases with different context values
5. **Modular tuples**: Use both global and test-specific tuples to avoid repetition
6. **Version control**: Store files work well with Git for tracking authorization changes over time

## File Extensions

- **Store files**: `.fga.yaml` (recommended) or `.yaml`
- **Model files**: `.fga` (recommended) or `.mod`
- **Tuple files**: `.yaml`, `.json`, or `.csv`

The `.fga.yaml` extension is the conventional naming pattern that makes store files easily identifiable and helps with tooling integration.
