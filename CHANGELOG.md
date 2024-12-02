# Changelog

### [0.6.2](https://github.com/openfga/cli/compare/v0.6.1...v0.6.2) (2024-12-02)

Fixed:
- Fixed issue where `fga store import` would error when importing a store with no tuples (#408) - thanks @ap0calypse8
- Fixed repetition in `fga query check` error output (#405) - thanks @Siddhant-K-code

### [0.6.1](https://github.com/openfga/cli/compare/v0.6.0...v0.6.1) (2024-09-09)

Fixed:
- Fixed issue where `fga store import`, `fga tuple write` and `fga tuple delete` could not be ran due to an issue with the `--max-tuples-per-write` and `--max-parallel-requests` options (#389)
- Fixed an issue where List Users failed test output did not include the returned response (#391)

### [0.6.0](https://github.com/openfga/cli/compare/v0.5.3...v0.6.0) (2024-09-08)

Added:
- Support usage of consistency parameter (#381)

### [0.5.3](https://github.com/openfga/cli/compare/v0.5.2...v0.5.3) (2024-08-15)

Fixed:
- Bump OpenFGA to v1.5.9 to fix an issue in the `check` API [CVE-2024-42473](https://github.com/openfga/openfga/security/advisories/GHSA-3f6g-m4hr-59h8)

### [0.5.2](https://github.com/openfga/cli/compare/v0.5.1...v0.5.2) (2024-08-08)

Fixed:
- Fixed issue where an error in getting the store in`fga store import` fails the import (#365)

### [0.5.1](https://github.com/openfga/cli/compare/v0.5.0...v0.5.1) (2024-06-25)

Fixed:
- Fixed issue where `fga store import` output could no longer be piped to `jq` (#355) - thanks @Siddhant-K-code

### [0.5.0](https://github.com/openfga/cli/compare/v0.4.1...v0.5.0) (2024-06-18)

Added:
- `fga store import` now shows a progress bar when writing tuples to show (#348) - thanks @Siddhant-K-code

Changed:
- `excluded_users` has been removed from the `fga query list-users` output and the testing for ListUsers (#347)

BREAKING CHANGE:

This version removes the `excluded_users` property from the ``fga query list-users` output and the ListUsers testing,
for more details see the [associated API change](https://github.com/openfga/api/pull/171).

### [0.4.1](https://github.com/openfga/cli/compare/v0.4.0...v0.4.1) (2024-06-05)

Added:
- Support asserting the `excluded_users` in `list_users` tests (#337)

Fixed:
- `fga store import` now outputs store and model details when writing to an existing store (#336)

### [0.4.0](https://github.com/openfga/cli/compare/v0.3.1...v0.4.0) (2024-05-07)

Added:
- Support querying the [list users API](https://openfga.dev/blog/list-users-announcement) with `fga query list-users` (#314)
- Support for running list users tests via `fga model test` (#315)

Changed:
- `fga store import` now uses the filename if no name is provided (#318) - thanks @NeerajNagure

### [0.3.1](https://github.com/openfga/cli/compare/v0.3.0...v0.3.1) (2024-04-29)

Added:

- `fga store import` now outputs the store and model details (#299) - thanks @NeerajNagure
- `fga store export` to support exporting a store (#306)
- Support specifying output format using `--output-format` for `fga model transform` (#308)

Changed:
- `fga tuple write` returns simpler error messages (#303) - thanks @shruti2522

### [0.3.0](https://github.com/openfga/cli/compare/v0.2.7...v0.3.0) (2024-03-28)

Added:
- Support for [modular models](https://github.com/openfga/rfcs/blob/main/20231212-modular-models.md) ([#262](https://github.com/openfga/cli/issues/262))

## v0.3.0-beta.1

### [0.3.0-beta.1](https://github.com/openfga/cli/compare/v0.2.6...v0.3.0-beta.1) (2024-03-13)

Added:
- Support for [modular models](https://github.com/openfga/rfcs/blob/main/20231212-modular-models.md) ([#262](https://github.com/openfga/cli/issues/262))

## v0.2.7

### [0.2.7](https://github.com/openfga/cli/compare/v0.2.6...v0.2.7) (2024-03-13)

Added:
- Support for exporting tuples as CSV (#250) - thanks @edwin-Marrima

Changed:
- Simplify the output of `model test` (#265)
- go > v1.21.8 is now required (#272)

Deprecated:
- The `--simple-output` flag in `tuple read` has been deprecated in favour of `--output-format=simple-json` (#250) - thanks @edwin-Marrima

## v0.2.6

### [0.2.6](https://github.com/openfga/cli/compare/v0.2.5...v0.2.6) (2024-02-27)

Fixed:
- allow transforming from JSON when `this` is not in first place (openfga/language#166)

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
