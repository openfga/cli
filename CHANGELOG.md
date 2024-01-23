# Changelog

## v0.2.5

### [0.2.5](https://github.com/openfga/cli/compare/v0.2.4...v0.2.5) (2024-01-23)

Added:
- Add support for oauth2 credentials with scopes instead of audience (#232) - thanks @le-yams
- `fga tuple import` now supports any columns order and optional fields are no longer required (#230) - thanks @le-yams  
- Support for mixed operators in the model.
  `define viewer: ([user] but not blocked) or owner or viewer from parent` is now supported!
  See [openfga/language#107](https://github.com/openfga/language/pull/107#issue-1990426478) for more details on supported and unsupported functionality

Fixed:
- Fixed `fga model write` not writing models with conditions (#236)
- Re-added support for condition parameters as identifiers and relation names (e.g. `list` and `map`)

## v0.2.4

### [0.2.4](https://github.com/openfga/cli/compare/v0.2.3...v0.2.4) (2024-01-16)

Fixed:
- Fixed support for reading json models (#228)


## v0.2.3

### [0.2.3](https://github.com/openfga/cli/compare/v0.2.2...v0.2.3) (2024-01-11)

Changed:
- add support for using csv files to import tuples (#222)


## v0.2.2

### [0.2.2](https://github.com/openfga/cli/compare/v0.2.1...v0.2.2) (2024-01-08)

Changed:
- add `fga store import` to import store data (#215)
- allow specifying `tuple_file` to reference tuples in the FGA store format (#212)
- support continuation token in `fga tuple changes` method and expose it in the output (#218)
- add support for specifying condition/context in queries and in write (#219) 
- allow more comments in the model (#221)

Fixed:
- fixed issue writing models with conditions

## v0.2.1

### [0.2.1](https://github.com/openfga/cli/compare/v0.2.0...v0.2.1) (2023-12-14)

Changed:
- dependency updates
  - now using OpenFGA `v1.4.0` with Conditions enabled by default
  - now using Go SDK `v0.3.0` with support for targeting servers hosted under a custom path e.g. `https://api.fga.example:8080/fga-service1`

## v0.2.0

### [0.2.0](https://github.com/openfga/cli/compare/v0.1.2...v0.2.0) (2023-11-03)

Changed:
- add support for models with conditions

## v0.1.2

### [0.1.2](https://github.com/openfga/cli/compare/v0.1.1...v0.1.2) (2023-10-12)

Changed:
- tuple write/delete cmd accepts file; deprecate tuple import (openfga/cli#165) (@gabrielbussolo)
- read now returns all pages when `--max-pages` is set to `0` (openfga/cli#174)
- read accepts a `--simple-output` flag to output the tuples directly (openfga/cli#168) (@gabrielbussolo)
  now you can run the following to export then import tuples:

  ```shell
  fga tuple read --simple-output --max-pages 0 > tuples.json
  fga tuple import --file tuples.json
  ```

## v0.1.1

### [0.1.1](https://github.com/openfga/cli/compare/v0.1.0...v0.1.1) (2023-09-22)

Fixed:
- Running `fga model test` now correctly exists with error code one if any test fails (https://github.com/openfga/cli/pull/157)

## v0.1.0

### [0.1.0](https://github.com/openfga/cli/compare/v0.1.0-beta6...v0.1.0) (2023-09-18)

Changed:
- The config `server-url` has been changed to `api-url` to be consistent with the upcoming SDK changes
  `server-url` is still accepted but is considered deprecated and may be removed in future releases.
  If present, `api-url` will take precedence over `server-url`.

Fixed:
- The core language is now able to better handle extraneous spaces, and panics should be heavily reduced

## v0.1.0-beta6

### [0.1.0-beta6](https://github.com/openfga/cli/compare/v0.1.0-beta6...v0.1.0) (2023-09-06)

Changed:
- [`model test`](https://github.com/openfga/cli?tab=readme-ov-file#run-tests-on-an-authorization-model) now runs test in a built in FGA instance (https://github.com/openfga/cli/pull/142)

BREAKING: The structure of the test file has changed, it is now `list_objects` instead of `list-objects`

## v0.1.0-beta5

### [0.1.0-beta5](https://github.com/openfga/cli/compare/v0.1.0-beta4...v0.1.0-beta5) (2023-08-28)

Added:
- Add [`model test`](https://github.com/openfga/cli?tab=readme-ov-file#run-tests-on-an-authorization-model) (https://github.com/openfga/cli/pull/131)
- Re-add [`model validate`](https://github.com/openfga/cli?tab=readme-ov-file#validate-an-authorization-model) (https://github.com/openfga/cli/pull/117)

Fixes:
- Upgrade dependencies, fixes a few issues in parsing

## v0.1.0-beta4

### [0.1.0-beta4](https://github.com/openfga/cli/compare/v0.1.0-beta3...v0.1.0-beta4) (2023-08-09)

Fixed:
- Accept model ID for tuple import & write (c53da0589302fda17146c84bb29917ac4b72de8d)
- Empty model ID now considered not set (fe804e6cd936089b6919814d781effe995504627)

## v0.1.0-beta3

### [0.1.0-beta3](https://github.com/openfga/cli/compare/v0.1.0-beta2...v0.1.0-beta3) (2023-08-01)

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

### [0.1.0-beta2](https://github.com/openfga/cli/compare/v0.1.0-beta1...v0.1.0-beta2) (2023-07-12)
- Fix brew build
- Add command completions
- Add package builds for Linux
- Change brew and archive file names to fga

## v0.1.0-beta1

### [0.1.0-beta1](https://github.com/openfga/cli/compare/v0.1.0-beta...v0.1.0-beta1) (2023-07-12)
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
