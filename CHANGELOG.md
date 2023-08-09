# Changelog

## v0.1.0-beta4

### [0.1.0-beta4](https://github.com/openfga/go-sdk/releases/tag/v0.1.0-beta3...v0.1.0-beta4) (2023-08-09)

Fixed:
- Accept model ID for tuple import & write (c53da0589302fda17146c84bb29917ac4b72de8d)
- Empty model ID now considered not set (fe804e6cd936089b6919814d781effe995504627)

## v0.1.0-beta3

### [0.1.0-beta3](https://github.com/openfga/go-sdk/releases/tag/v0.1.0-beta2...v0.1.0-beta3) (2023-08-01)

Breaking:
- In model commands, default input/output is now the FGA DSL format (8dfc6976a16828e249da5a1bdb506e6627c0ced0)
- Response for Store creation has been updated (8dfc6976a16828e249da5a1bdb506e6627c0ced0)
- `fga model validate` has been temporarily removed (fb25b9634f707f1ea2a00f1c6051272e707a6e3c)
- `fga model list` now shows just model id and created date by default (fb25b9634f707f1ea2a00f1c6051272e707a6e3c)

Changed:
- Add model transform command (8dfc6976a16828e249da5a1bdb506e6627c0ced0)
- Allow initializing with a model during store creation (8dfc6976a16828e249da5a1bdb506e6627c0ced0)
- In all model commands, accept an FGA DSL output/input (8dfc6976a16828e249da5a1bdb506e6627c0ced0)

Fixed:
- Updating build information on release (55a3b61828c82fa756d24ac151c5937892d39fd9)
- Default max tuples per write in import is now `1` so that none of the writes are in transaction mode (024825213340973ba22f17acd4bddcf3d6baefd9)


## v0.1.0-beta2

### [0.1.0-beta2](https://github.com/openfga/go-sdk/releases/tag/v0.1.0-beta1...v0.1.0-beta2) (2023-07-12)
- Fix brew build
- Add command completions
- Add package builds for Linux
- Change brew and archive file names to fga

## v0.1.0-beta1

### [0.1.0-beta1](https://github.com/openfga/go-sdk/releases/tag/v0.1.0-beta...v0.1.0-beta1) (2023-07-12)
- Fix Release Pipeline

## v0.1.0-beta

### [0.1.0-beta](https://github.com/openfga/go-sdk/releases/tag/v0.1.0-beta) (2023-07-11)

Initial OpenFGA CLI release
- Support for [OpenFGA](https://github.com/openfga/openfga) API
  - Create, read, list and delete stores
  - Create, read, list and validate authorization models
  - Write, delete, read and import tuples
  - Read tuple changes
  - Run authorization checks
  - List objects a user has access to
  - List relations a user has on an object
  - Use Expand to understand why access was granted
