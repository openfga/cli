# OpenFGA CLI

A cross-platform CLI to interact with an OpenFGA server

[![Go Reference](https://pkg.go.dev/badge/github.com/openfga/cli.svg)](https://pkg.go.dev/github.com/openfga/cli)
[![Release](https://img.shields.io/github/v/release/openfga/cli?sort=semver&color=green)](https://github.com/openfga/cli/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](./LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenfga%2Fcli.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenfga%2Fcli?ref=badge_shield)
[![X](https://img.shields.io/twitter/follow/openfga?color=%23179CF0&logo=x "@openfga on X")](https://x.com/openfga)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/openfga/cli/badge?label=OpenSSF%20Scorecard)](https://securityscorecards.dev/viewer/?uri=github.com/openfga/cli)
[![Join our community](https://img.shields.io/badge/slack-cncf_%23openfga-40abb8.svg?logo=slack)](https://openfga.dev/community)

## Table of Contents
- [About OpenFGA](#about)
- [Resources](#resources)
- [Installation](#installation)
- [Building from Source](#building-from-source)
- [Usage](#usage)
  - [Configuration](#configuration)
  - [Custom Headers](#custom-headers)
<!-- BEGIN_COMMANDS_TOC -->
  - [Commands](#commands)
    - [Stores](#stores)
      - [Create Store](#create-store)
      - [Get Store](#get-store)
      - [List Stores](#list-stores)
      - [Delete Store](#delete-store)
      - [Import Store Data](#import-store-data)
      - [Export Store Data](#export-store-data)
    - [Authorization Models](#authorization-models)
      - [Write Authorization Model](#write-authorization-model)
      - [Read a Single Authorization Model](#read-a-single-authorization-model)
      - [Read Authorization Models](#read-authorization-models)
      - [Validate Authorization Model](#validate-authorization-model)
      - [Transform an Authorization Model](#transform-an-authorization-model)
      - [Test an Authorization Model](#test-an-authorization-model)
    - [Relationship Tuples](#relationship-tuples)
      - [Create Relationship Tuples](#create-relationship-tuples)
      - [Delete Relationship Tuples](#delete-relationship-tuples)
      - [Read Relationship Tuples](#read-relationship-tuples)
      - [Read Relationship Tuple Changes (Watch)](#read-relationship-tuple-changes-watch)
    - [Relationship Queries](#relationship-queries)
      - [Check](#check)
      - [Expand](#expand)
      - [List Objects](#list-objects)
      - [List Relations](#list-relations)
      - [List Users](#list-users)
    - [Manage JSON-to-tuple mappings](#manage-jsontotuple-mappings)
      - [Scaffold a starter mapping file](#scaffold-a-starter-mapping-file)
      - [Validate a mapping file](#validate-a-mapping-file)
      - [Run the embedded tests in a mapping file](#run-the-embedded-tests-in-a-mapping-file)
      - [Evaluate a mapping against JSON input and emit tuple operations](#evaluate-a-mapping-against-json-input-and-emit-tuple-operations)
<!-- END_COMMANDS_TOC -->
- [Contributing](#contributing)
- [License](#license)


## About
[OpenFGA](https://openfga.dev) is an open source Fine-Grained Authorization solution inspired by [Google's Zanzibar paper](https://research.google/pubs/pub48190/). It was created by the FGA team at [Auth0/Okta](https://auth0.com) based on [Auth0 Fine-Grained Authorization (FGA)](https://fga.dev), available under [a permissive license (Apache-2)](https://github.com/openfga/rfcs/blob/main/LICENSE) and welcomes community contributions.

OpenFGA is designed to make it easy for application builders to model their permission layer, and to add and integrate fine-grained authorization into their applications. OpenFGA’s design is optimized for reliability and low latency at a high scale.

## Resources

- [OpenFGA Documentation](https://openfga.dev/docs)
- [OpenFGA API Documentation](https://openfga.dev/api/service)
- [Zanzibar Academy](https://zanzibar.academy)
- [Google's Zanzibar Paper (2019)](https://research.google/pubs/pub48190/)

## Installation

### Brew
```shell
brew install openfga/tap/fga
```

### Linux (deb, rpm and apk) packages
Download the .deb, .rpm or .apk packages from the [releases page](https://github.com/openfga/cli/releases).

Debian:
```shell
sudo apt install ./fga_<version>_linux_<arch>.deb
```

Fedora:
```shell
sudo dnf install ./fga_<version>_linux_<arch>.rpm
```

Alpine Linux:
```shell
sudo apk add --allow-untrusted ./fga_<version>_linux_<arch>.apk
```

### Windows

via [Scoop](https://scoop.sh/)
```shell
scoop install openfga
```

### Docker
```shell
docker pull openfga/cli; docker run -it openfga/cli
```

The Docker image is multi-platform and includes the system CA certificates needed for endpoints that use publicly trusted certificate authorities. Private or internal CAs still need to be provided by the user.

### Go

```shell
go install github.com/openfga/cli/cmd/fga@latest
```

### Manually
Download the pre-compiled binaries from the [releases page](https://github.com/openfga/cli/releases).

## Building from Source

Make sure you have Go 1.25 or later installed. See the [Go downloads](https://go.dev/dl/) page.

1. Clone the repo to a local directory, and navigate to that directory:

   ```bash
   git clone https://github.com/openfga/cli.git && cd cli
   ```

2. Then use the build command:

   ```bash
   go build -o ./dist/fga ./cmd/fga/main.go
   ```

   or if you have `make` installed, just run:

   ```bash
   make build
   ```

3. Run the OpenFGA CLI with:

   ```bash
   ./dist/fga
   ```

## Usage

### Configuration

For any command that interacts with an OpenFGA server, these configuration values can be passed (where applicable)

| Name                   | Flag                 | CLI                    | ~/.fga.yaml        |
|------------------------|----------------------|------------------------|--------------------|
| API Url                | `--api-url`          | `FGA_API_URL`          | `api-url`          |
| Shared Secret          | `--api-token`        | `FGA_API_TOKEN`        | `api-token`        |
| Client ID              | `--client-id`        | `FGA_CLIENT_ID`        | `client-id`        |
| Client Secret          | `--client-secret`    | `FGA_CLIENT_SECRET`    | `client-secret`    |
| Scopes                 | `--api-scopes`       | `FGA_API_SCOPES`       | `api-scopes`       |
| Token Issuer           | `--api-token-issuer` | `FGA_API_TOKEN_ISSUER` | `api-token-issuer` |
| Token Audience         | `--api-audience`     | `FGA_API_AUDIENCE`     | `api-audience`     |
| Store ID               | `--store-id`         | `FGA_STORE_ID`         | `store-id`         |
| Authorization Model ID | `--model-id`         | `FGA_MODEL_ID`         | `model-id`         |
| Custom Headers         | `--custom-headers`   | `FGA_CUSTOM_HEADERS`   | `custom-headers`   |

If you are authenticating with a shared secret, you should specify the API Token value. If you are authenticating using OAuth, you should specify the Client ID, Client Secret, API Audience and Token Issuer. For example:

```
# Note: This example is for Auth0 FGA
api-url: https://api.us1.fga.dev
client-id: 4Zb..UYjaHreLKOJuU8
client-secret: J3...2pBwiauD
api-audience: https://api.us1.fga.dev/
api-token-issuer: auth.fga.dev
store-id: 01H0H015178Y2V4CX10C2KGHF4
```

#### Custom Headers

You can add custom HTTP headers to all requests sent to the API using the `--custom-headers` flag. Headers are specified in `<name>: <value>` format, and the flag can be repeated to add multiple headers.

##### Flag
```shell
--custom-headers "Header-Name: header-value"
```

##### Example
```shell
fga store list --custom-headers "X-Custom-Header: value1" --custom-headers "X-Request-ID: abc123"
```

##### Configuration

Custom headers can also be configured via the CLI environment variable or the configuration file:

| Name           | Flag                 | CLI                    | ~/.fga.yaml        |
|----------------|----------------------|------------------------|---------------------|
| Custom Headers | `--custom-headers`   | `FGA_CUSTOM_HEADERS`   | `custom-headers`    |

Example `~/.fga.yaml`:
```yaml
api-url: https://api.fga.example
store-id: 01H0H015178Y2V4CX10C2KGHF4
custom-headers:
  - "X-Custom-Header: value1"
  - "X-Request-ID: abc123"
```

### Commands

<!-- BEGIN_COMMANDS -->
#### Stores

Create, Get, Delete and List OpenFGA Stores

##### Create Store

Create an OpenFGA store.

###### Command

```
fga store create [flags]
```

###### Parameters

* `--format`: Authorization model input format. Can be "fga", "json" or "modular
* `--model`: Authorization Model File Name
* `--name`: Store Name

###### Example

```bash
fga store create --name "FGA Demo Store"
fga store create --model Model.fga
export FGA_STORE_ID=$(fga store create --model model.fga | jq -r .store.id)
```

###### Response

```json
{
    "id": "01H0H015178Y2V4CX10C2KGHF4",
    "name": "FGA Demo Store",
    "created_at": "2023-05-19T16:10:07.637585677Z",
    "updated_at": "2023-05-19T16:10:07.637585677Z"
}

Response for fga store create --model Model.fga:
{
  "store": {
    "id":"01H6H9CNQRP2TVCFR7899XGNY8",
    "name":"Model",
    "created_at":"2023-07-29T16:58:28.984402Z",
    "updated_at":"2023-07-29T16:58:28.984402Z"
  },
  "model": {
    "authorization_model_id":"01H6H9CNQV36Y9WS1RJGRN8D06"
  }
}
```

##### Get Store

Get a particular store.

###### Command

```
fga store get [flags]
```

###### Parameters

* `--store-id`: Store ID

###### Example

```bash
fga store get --store-id=01H0H015178Y2V4CX10C2KGHF4
```

###### Response

```json
{
    "id": "01H0H015178Y2V4CX10C2KGHF4",
    "name": "FGA Demo Store",
    "created_at": "2023-05-19T16:10:07.637585677Z",
    "updated_at": "2023-05-19T16:10:07.637585677Z"
}
```

##### List Stores

Get a list of stores.

###### Command

```
fga store list [flags]
```

###### Parameters

* `--max-pages`: Max number of pages to get.
* `--name`: Filter stores by exact name. Substrings and regexes are not supported.

###### Example

```bash
fga store list
```

###### Response

```json
{
  "stores": [{
    "id": "..",
    "name": "..",
    "created_at": "",
    "updated_at": "",
    "deleted_at": ""
  }]
}
```

##### Delete Store

Mark a store as deleted.

###### Command

```
fga store delete [flags]
```

###### Parameters

* `--force`: Force delete without confirmation
* `--store-id`: Store ID

###### Example

```bash
fga store delete --store-id=01H0H015178Y2V4CX10C2KGHF4
```

###### Response

```json
{}
```

##### Import Store Data

Import a store: updating the name, model and appending the global tuples

###### Command

```
fga store import [flags]
```

###### Parameters

* `--allow-external-files`: Allow model_file, tuple_file and tuple_files references in the store file to resolve to paths outside the store file's directory. Only enable this for store files you trust.
* `--file`: File Name. The file should have the store
* `--max-parallel-requests`: Max number of requests to issue to the server in parallel.
* `--max-tuples-per-write`: Max tuples per write chunk.
* `--store-id`: Store ID

###### Example

```bash
fga store import --file=model.fga.yaml
```

###### Response

```json
{}
```

##### Export Store Data

Export a store to YAML. If --output-file is specified the response will be written to that file; otherwise it is written to stdout.

###### Command

```
fga store export [flags]
```

###### Parameters

* `--max-tuples`: max number of tuples to return in the output
* `--model-id`: Authorization Model ID
* `--output-file`: name of the file to export the store to
* `--store-id`: store ID

###### Example

```bash
fga store export --store-id=01H0H015178Y2V4CX10C2KGHF4
```

###### Response

```yaml
name: Test
model: |+
  model
    schema 1.1

  type user

tuples:
  - user: user:1
    relation: member
    object: group:admins
tests: []
```

#### Authorization Models

Write, read, list and validate authorization models.

##### Parameters

* `--store-id`: Store ID

##### Write Authorization Model

Writes a new authorization model.

###### Command

```
fga model write [flags]
```

###### Parameters

* `--file`: File Name. The file should have the model in the JSON or DSL format
* `--format`: Authorization model input format. Can be "fga", "json", or "modular"
* `--store-id`: Store ID

###### Example

```bash
fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file=model.json
fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file=fga.mod
fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 '{"type_definitions":[{"type":"user"},{"type":"document","relations":{"can_view":{"this":{}}},"metadata":{"relations":{"can_view":{"directly_related_user_types":[{"type":"user"}]}}}}],"schema_version":"1.1"}' --format=json
```

###### Response

```json
{
  "authorization_model_id":"01GXSA8YR785C4FYS3C0RTG7B1"
}
```

##### Read a Single Authorization Model

Read an authorization model. Omit --model-id to get the latest model.

###### Command

```
fga model get [flags]
```

###### Parameters

* `--field`: Fields to display, choices are: id, created_at, size and model
* `--format`: Authorization model output format. Can be "fga" or "json"
* `--model-id`: Authorization Model ID
* `--store-id`: Store ID

###### Example

```bash
fga model get --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1 --field size --field model --field id --field created_at
fga model get --store-id=01H0H015178Y2V4CX10C2KGHF4
```

###### Response

```fga
# Model ID: 01GXSA8YR785C4FYS3C0RTG7B1
# Created At: 2023-04-11 23:26:34.759 +0000 UTC
# Size: 20.05 KB
model
  schema 1.1

type user

type document
  relations
    define can_view: [user]
```

##### Read Authorization Models

List all authorization models for a store, in descending order by creation date.

###### Command

```
fga model list [flags]
```

###### Parameters

* `--field`: Fields to display, choices are: id, created_at and model
* `--max-pages`: Max number of pages to get.
* `--store-id`: Store ID

###### Example

```bash
fga model list --store-id=01H0H015178Y2V4CX10C2KGHF4
```

###### Response

```json
{
  "authorization_models": [
    {
      "id":"01H6H9XH1G5Q6DK6PFMGDZNH9S",
      "created_at":"2023-07-29T17:07:41Z"
    },
    {
      "id":"01H6H9PPR6C3P45R75X55ZFP46",
      "created_at":"2023-07-29T17:03:57Z"
    }
  ]
}
```

##### Validate Authorization Model

Validates that an authorization model is valid. When the input parses successfully, the JSON response includes size_kb, the protobuf-serialized size of the model in KB.

###### Command

```
fga model validate [flags]
```

###### Parameters

* `--file`: File Name. The file should have the model in the JSON or DSL format or be an fga.mod file
* `--format`: Authorization model input format. Can be "fga", "json", or "modular"

###### Example

```bash
fga model validate --file model.json
```

###### Response

```json
{"id":"01GPGWB8R33HWXS3KK6YG4ETGH","created_at":"2023-01-11T16:59:22Z","is_valid":true,"size_kb":0.05}

Invalid model:
{"id":"01GPGTVEH5NYTQ19RYFQKE0Q4Z","created_at":"2023-01-11T16:33:15Z","is_valid":false,"error":"invalid schema version","size_kb":0.05}
```

##### Transform an Authorization Model

Convert an authorization model between formats (.fga, .json, .mod).

###### Command

```
fga model transform [flags]
```

###### Parameters

* `--file`: File Name. The file should have the model in the JSON or DSL format or be an `fga.mod` format
* `--input-format`: Authorization model input format. Can be "fga", "json", or "modular"
* `--output-format`: Authorization model output format. Can be "fga" or "json"."

###### Example

```bash
fga model transform --file=model.json
fga model transform --file=model.fga
fga model transform '{ "schema_version": "1.1", "type_definitions":[{"type":"user"}] }' --input-format json
fga model transform --file=fga.mod
```

###### Response

```fga
model
  schema 1.1

type user

type document
  relations
    define can_view: [user]
```

##### Test an Authorization Model

Run a set of tests against a particular Authorization Model. If a model is provided the test will run in a built-in OpenFGA instance; otherwise it runs against the configured store.

###### Command

```
fga model test [flags]
```

###### Parameters

* `--allow-external-files`: Allow model_file, tuple_file and tuple_files references in the test file to resolve to paths outside the test file's directory. Only enable this for test files you trust.
* `--max-types-per-authorization-model`: Max allowed number of type definitions per authorization model
* `--model-id`: Model ID
* `--store-id`: Store ID
* `--suppress-summary`: Suppress the plain text summary output
* `--tests`: Path or glob of YAML test files
* `--verbose`: Print verbose JSON output

###### Example

```bash
fga model test --tests model.fga.yaml
fga model test --tests "tests/*.fga.yaml"
```

###### Response

```text
(FAILING) test-1: Checks (2/3 passing) | ListObjects (2/2 passing)
ⅹ Check(user=user:anne,relation=can_share,object=folder:1): expected=false, got=true
---
# Test Summary #
Tests 1/2 passing
Checks 3/4 passing
ListObjects 3/3 passing
```

#### Relationship Tuples

Read, write, delete, import and listen to changes in relationship tuples in a store.

##### Parameters

* `--store-id`: Store ID

##### Create Relationship Tuples

Add relationship tuples to the store. This command allows for the creation of relationship tuples either through direct command line arguments or by specifying a file. The file can be in JSON, YAML, or CSV format.

When using a CSV file, the file must adhere to a specific header structure for the command to correctly interpret the data. The required CSV header structure is as follows:
- "user_type":         Specifies the type of the user in the relationship tuple. (e.g. "team")
- "user_id":           The unique identifier of the user. (e.g. "marketing")
- "user_relation":     Defines the user relation forming a userset. (optional) (e.g. "member")
- "relation":          Defines the tuple relation. (e.g. "viewer")
- "object_type":       Specifies the type of the object in the relationship tuple. (e.g. "document")
- "object_id":         The unique identifier of the object. (e.g. "roadmap")
- "condition_name":    The name of the condition. (optional) (e.g. "inOfficeIP")
- "condition_context": The context of the condition as a json object. (optional) (e.g. "{""ip_addr"":""10.0.0.1""}")

For example, a valid CSV file might start with a row like:
user_type,user_id,user_relation,relation,object_type,object_id,condition_name,condition_context

This command is flexible in accepting data inputs, making it easier to add multiple relationship tuples in various convenient formats.

###### Command

```
fga tuple write <user> <relation> <object> [flags]
```

###### Parameters

* `--condition-context`: Condition Context (as a JSON string)
* `--condition-name`: Condition Name
* `--file`: Tuples file
* `--hide-imported-tuples`: Hide successfully imported tuples from output
* `--max-parallel-requests`: Max number of requests to issue to the server in parallel.
* `--max-rps`: The maximum requests per second.
* `--max-tuples-per-write`: Max tuples per write chunk.
* `--model-id`: Model ID
* `--on-duplicate`: Whether to ignore or error on duplicate tuples. Valid values are 'ignore' and 'error'. (default: 'ignore' when importing a file of tuples, 'error' otherwise)
* `--rampup-period-in-sec`: The period over which to ramp up the request rate.

###### Example

```bash
fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap --condition-name inOffice --condition-context '{"office_ip":"10.0.1.10"}'
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.json
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.yaml
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.csv
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.csv --max-tuples-per-write 10 --max-parallel-requests 5
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.csv --max-rps 10
  fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.csv --on-duplicate ignore
```

###### Response

```json
{
  "successful": [
    {
      "object":"document:roadmap",
      "relation":"writer",
      "user":"user:annie"
    }
  ],
  "failed": [],
  "failed_count": 0,
  "successful_count": 1,
  "total_count": 1
}
```

##### Delete Relationship Tuples

Delete relationship tuples from the store.

###### Command

```
fga tuple delete <user> <relation> <object> [flags]
```

###### Parameters

* `--file`: Tuples file
* `--hide-imported-tuples`: Hide successfully imported tuples from output
* `--max-parallel-requests`: Max number of requests to issue to the server in parallel.
* `--max-tuples-per-write`: Max tuples per write chunk.
* `--model-id`: Model ID
* `--on-missing`: Whether to ignore or error on missing tuples. Valid values are 'ignore' and 'error'. (default: 'ignore' when deleting a file of tuples, 'error' otherwise)

###### Example

```bash
fga tuple delete --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap
  fga tuple delete --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.csv --on-missing ignore
```

###### Response

```json
{}

Response when using --file:
{
  "successful": [
    {
      "object":"document:roadmap",
      "relation":"writer",
      "user":"user:annie"
    }
  ],
  "failed": []
}
```

##### Read Relationship Tuples

Read relationship tuples that exist in the system (does not evaluate).

###### Command

```
fga tuple read [flags]
```

###### Parameters

* `--consistency`: Consistency preference for the request. Valid options are HIGHER_CONSISTENCY and MINIMIZE_LATENCY.
* `--max-pages`: Max number of pages to get. Set to 0 to get all pages.
* `--object`: Object
* `--output-format`: Specifies the format for data presentation. Valid options: json, simple-json, csv, and yaml.
* `--page-size`: Number of tuples to return per page. Defaults to 100 when max-pages=0, or 50 otherwise. Max is 100.
* `--relation`: Relation
* `--user`: User

###### Example

```bash
fga tuple read --store-id=01H0H015178Y2V4CX10C2KGHF4 --user user:anne --relation can_view --object document:roadmap
```

###### Response

```json
{
  "tuples": [
    {
      "key": {
        "object": "document:roadmap",
        "relation": "can_view",
        "user": "user:anne"
      },
      "timestamp": "2023-07-06T15:12:55.080666875Z"
    }
  ]
}
```

##### Read Relationship Tuple Changes (Watch)

Get a list of relationship tuple changes (Writes and Deletes) across time.

###### Command

```
fga tuple changes [flags]
```

###### Parameters

* `--continuation-token`: Continuation token to start changes from.
* `--max-pages`: Max number of pages to get.
* `--start-time`: Time to return changes since.
* `--type`: Type to restrict the changes by.

###### Example

```bash
fga tuple changes --store-id=01H0H015178Y2V4CX10C2KGHF4 --type=document --continuation-token=M3w=
```

###### Response

```json
{
  "changes": [
    {
      "operation": "TUPLE_OPERATION_WRITE",
      "timestamp": "2023-07-06T15:12:40.294950382Z",
      "tuple_key": {
        "object": "document:roadmap",
        "relation": "can_view",
        "user": "user:anne"
      }
    }
  ],
  "continuation_token":"NHw="
}
```

#### Relationship Queries

Run queries (Check, Expand, ListObjects, ListRelations, ListUsers) that are evaluated according to a particular model.

##### Parameters

* `--consistency`: Consistency preference for the request. Valid options are HIGHER_CONSISTENCY and MINIMIZE_LATENCY.
* `--context`: Query context (as a JSON string)
* `--contextual-tuple`: Contextual Tuple, output: "user relation object"
* `--model-id`: Model ID
* `--store-id`: Store ID

##### Check

Check if a user has a particular relation with an object.

###### Command

```
fga query check <user> <relation> <object>
```

###### Example

```bash
fga query check --store-id="01H4P8Z95KTXXEP6Z03T75Q984" user:anne can_view document:roadmap --context '{"ip_address":"127.0.0.1"}' --consistency "HIGHER_CONSISTENCY"
```

###### Response

```json
{
  "allowed": true,
  "resolution": ""
}
```

##### Expand

Expands the relationships in userset tree format.

###### Command

```
fga query expand <relation> <object>
```

###### Example

```bash
fga query expand --store-id="01H4P8Z95KTXXEP6Z03T75Q984" can_view document:roadmap --consistency "HIGHER_CONSISTENCY"
```

###### Response

```json
{
  "tree": {
    "root": {
      "name": "document:roadmap#can_view",
      "union": {
        "nodes": [
          {
            "name": "document:roadmap#can_view",
            "leaf": {
              "users": {
                "users": [
                  "user:anne"
                ]
              }
            }
          }
        ]
      }
    }
  }
}
```

##### List Objects

List the objects of a certain type that a user has a particular relation to.

###### Command

```
fga query list-objects <user> <relation> <object-type>
```

###### Example

```bash
fga query list-objects --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document --contextual-tuple "user:anne can_view folder:product" --contextual-tuple "folder:product parent document:roadmap" --consistency "HIGHER_CONSISTENCY"
```

###### Response

```json
{
  "objects": [
    "document:roadmap"
  ]
}
```

##### List Relations

List relations that a user has with an object.

###### Command

```
fga query list-relations <user> <object> [flags]
```

###### Parameters

* `--relation`: Relation

###### Example

```bash
fga query list-relations --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne document:roadmap --relation can_view --consistency "HIGHER_CONSISTENCY"
```

###### Response

```json
{
  "relations": [
    "can_view"
  ]
}
```

##### List Users

List all users that have a certain relation with a particular object

###### Command

```
fga query list-users [flags]
```

###### Parameters

* `--object`: Object to list users for
* `--relation`: Relation to evaluate on
* `--user-filter`: Filter the responses can be in the formats <type> (to filter objects and typed public bound access) or <type>#<relation> (to filter usersets)

###### Example

```bash
fga query list-users --store-id=01H0H015178Y2V4CX10C2KGHF4 --object document:roadmap --relation can_view --consistency "HIGHER_CONSISTENCY"
```

###### Response

```json
{
  "users": [
    {
      "object": {
        "type": "user",
        "id": "anne"
      }
    }
  ]
}
```

#### Manage JSON-to-tuple mappings

Validate, test, and run JSON-to-tuple mapping files.

##### Scaffold a starter mapping file

Creates a new mapping YAML file with a sample rule and embedded test. Defaults to mapping.yaml.

###### Command

```
fga mapping init [mapping-file] [flags]
```

###### Parameters

* `--force`: Overwrite an existing file
* `--minimal`: Emit a skeleton without the example test block

###### Example

```bash
fga mapping init
  fga mapping init my-mapping.yaml
  fga mapping init --minimal mapping.yaml
  fga mapping init --force mapping.yaml
```

##### Validate a mapping file

Validates that a mapping file is syntactically correct and all expressions compile.
With --model-file, also checks that every tuple template is consistent with the
authorization model: object types, relations, and user types must exist and be valid.

###### Command

```
fga mapping validate [mapping-file] [flags]
```

###### Parameters

* `--format`: Output format: "text" or "json"
* `--model-file`: Path to FGA authorization model file (DSL, JSON, or modular)
* `--verbose`: Show per-rule validation status (text format only)

###### Example

```bash
fga mapping validate mapping.yaml
  fga mapping validate --format json mapping.yaml
  fga mapping validate --model-file model.fga mapping.yaml
```

##### Run the embedded tests in a mapping file

Compiles the mapping and runs its embedded test cases, reporting pass/fail per case.
Exits 1 when any test fails, 2 when the mapping file cannot be compiled.
Use --format to choose between human-readable text (default), JSON, or JUnit XML output.

###### Command

```
fga mapping test [mapping-file] [flags]
```

###### Parameters

* `--fail-fast`: Stop after the first failing test
* `--format`: Output format: "text", "json", or "junit"
* `--no-color`: Disable color in text output
* `--output-file`: Write output to a file instead of stdout
* `--run`: Run only tests whose name contains this substring (case-sensitive)
* `--verbose`: Show rule trace and tuple details for each test (text format only)

###### Example

```bash
fga mapping test mapping.yaml
  fga mapping test --format json mapping.yaml
  fga mapping test --format junit --output-file results.xml mapping.yaml
  fga mapping test --run anne mapping.yaml
  fga mapping test --fail-fast mapping.yaml
```

##### Evaluate a mapping against JSON input and emit tuple operations

Reads JSONL from stdin (or --input) and evaluates it against the mapping file.
Input is JSON Lines: one JSON object per line. Outputs tuple operations as JSONL (default)
or a JSON batch (--format json).
Runs entirely offline — no store reads, no credentials, no network.
Rules using tuple_filters cannot be expanded offline and are reported as warnings on stderr
(or under tuple_filter_operations in the --format json batch).

Default JSONL emits one operation per line with an "op" field (write or delete).
With --writes-only only write operations are emitted, in the bare ClientTupleKey shape
consumable directly by fga tuple write --file (JSONL lines, or a JSON array with --format json).

Each input record is evaluated and streamed in order. --format json instead collects the whole
run into a single document. --aggregate buffers all records and collapses them (cross-record
dedup of tuples and unresolved filters, plus write/delete conflict detection) before emitting;
a conflict is a runtime error. Both --format json and --aggregate buffer the whole run in memory
before emitting; the default streaming JSONL does not.
--continue-on-error skips any record that fails to parse or evaluate (warning to stderr) and
continues; the command still exits non-zero if any record was skipped.

###### Command

```
fga mapping run [mapping-file] [flags]
```

###### Parameters

* `--aggregate`: Buffer all records and collapse them (dedup tuples and filters, detect write/delete conflicts) before emitting
* `--continue-on-error`: Skip input records that fail to parse or evaluate (warn to stderr) and exit non-zero if any were skipped
* `--format`: Output format: "jsonl" or "json"
* `--input`: Path to JSONL input file, one JSON object per line (default: stdin)
* `--interactive`: Explore the mapping in a terminal loop: paste JSON documents and see the tuples they produce (requires a TTY)
* `--writes-only`: Emit only write operations in ClientTupleKey format (consumable by fga tuple write)

###### Example

```bash
echo '{"id":"anne","org":"acme"}' | fga mapping run mapping.yaml
  fga mapping run mapping.yaml --input event.json --format json
  fga mapping run --writes-only mapping.yaml > out.jsonl && fga tuple write --store-id $STORE_ID --file out.jsonl
  fga mapping run mapping.yaml -i
```

<!-- END_COMMANDS -->

## Contributing

See [CONTRIBUTING](https://github.com/openfga/.github/blob/main/CONTRIBUTING.md).

## Author

[OpenFGA](https://github.com/openfga)

## License

This project is licensed under the Apache-2.0 license. See the [LICENSE](https://github.com/openfga/cli/blob/main/LICENSE) file for more info.
