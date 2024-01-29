# OpenFGA CLI

A cross-platform CLI to interact with an OpenFGA server

[![Go Reference](https://pkg.go.dev/badge/github.com/openfga/cli.svg)](https://pkg.go.dev/github.com/openfga/cli)
[![Release](https://img.shields.io/github/v/release/openfga/cli?sort=semver&color=green)](https://github.com/openfga/cli/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](./LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenfga%2Fcli.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenfga%2Fcli?ref=badge_shield)
[![Discord Server](https://img.shields.io/discord/759188666072825867?color=7289da&logo=discord "Discord Server")](https://discord.gg/8naAwJfWN6)
[![Twitter](https://img.shields.io/twitter/follow/openfga?color=%23179CF0&logo=twitter&style=flat-square "@openfga on Twitter")](https://twitter.com/openfga)

## Table of Contents
- [About OpenFGA](#about)
- [Resources](#resources)
- [Installation](#installation)
- [Building from Source](#building-from-source)
- [Usage](#usage)
  - [Configuration](#configuration)
  - [Commands](#commands)
    - [Stores](#stores)
      - [List All Stores](#list-stores)
      - [Create a Store](#create-store)
      - [Import a Store](#import-store)
      - [Get a Store](#get-store)
      - [Delete a Store](#delete-store)
    - [Authorization Models](#authorization-models)
      - [Read Authorization Models](#read-authorization-models)
      - [Write Authorization Model](#write-authorization-model)
      - [Read a Single Authorization Model](#read-a-single-authorization-model)
      - [Read the Latest Authorization Model](#read-the-latest-authorization-model)
      - [Validate an Authorization Model](#validate-an-authorization-model)
      - [Run Tests on an Authorization Model](#run-tests-on-an-authorization-model)
      - [Transform an Authorization Model](#transform-an-authorization-model)
    - [Relationship Tuples](#relationship-tuples)
      - [Read Relationship Tuple Changes (Watch)](#read-relationship-tuple-changes-watch)
      - [Read Relationship Tuples](#read-relationship-tuples)
      - [Write Relationship Tuples](#write-relationship-tuples)
      - [Delete Relationship Tuples](#delete-relationship-tuples)
    - [Relationship Queries](#relationship-queries)
      - [Check](#check)
      - [Expand](#expand)
      - [List Objects](#list-objects)
      - [List Relations](#list-relations)
- [Contributing](#contributing)
- [License](#license)


## About
[OpenFGA](https://openfga.dev) is an open source Fine-Grained Authorization solution inspired by [Google's Zanzibar paper](https://research.google/pubs/pub48190/). It was created by the FGA team at [Auth0](https://auth0.com) based on [Auth0 Fine-Grained Authorization (FGA)](https://fga.dev), available under [a permissive license (Apache-2)](https://github.com/openfga/rfcs/blob/main/LICENSE) and welcomes community contributions.

OpenFGA is designed to make it easy for application builders to model their permission layer, and to add and integrate fine-grained authorization into their applications. OpenFGA’s design is optimized for reliability and low latency at a high scale.

## Resources

- [OpenFGA Documentation](https://openfga.dev/docs)
- [OpenFGA API Documentation](https://openfga.dev/api/service)
- [Twitter](https://twitter.com/openfga)
- [OpenFGA Discord Community](https://discord.gg/8naAwJfWN6)
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

### Go

```shell
go install github.com/openfga/cli/cmd/fga@latest
```

### Manually
Download the pre-compiled binaries from the [releases page](https://github.com/openfga/cli/releases).

## Building from Source

Make sure you have Go 1.20 or later installed. See the [Go downloads](https://go.dev/dl/) page.

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

If you are authenticating with a shared secret, you should specify the API Token value. If you are authenticating using OAuth, you should specify the Client ID, Client Secret, API Audience and Token Issuer. For example:

```
# Note: This example is for Auth0 FGA
api-url: https://api.us1.fga.dev
client-id: 4Zb..UYjaHreLKOJuU8
client-secret: J3...2pBwiauD
api-audience: https://api.us1.fga.dev/
api-token-issuer: fga.us.auth0.com
store-id: 01H0H015178Y2V4CX10C2KGHF4
```

### Commands

#### Stores
| Description                     | command  | parameters   | example                                                  |
|---------------------------------|----------|--------------|----------------------------------------------------------|
| [Create a Store](#create-store) | `create` | `--name`     | `fga store create --name="FGA Demo Store"`               |
| [Import a Store](#import-store) | `import` | `--file`     | `fga store import --file store.fga.yaml`                 |
| [List Stores](#list-stores)     | `list`   |              | `fga store list`                                         |
| [Get a Store](#get-store)       | `get`    | `--store-id` | `fga store get --store-id=01H0H015178Y2V4CX10C2KGHF4`    |
| [Delete a Store](#delete-store) | `delete` | `--store-id` | `fga store delete --store-id=01H0H015178Y2V4CX10C2KGHF4` |

##### Create Store

###### Command
fga store **create**

###### Parameters
* `--name`: Name of the store to be created. If the `model` parameter is specified, the model file name will be used as the default store name. 
* `--model`: File with the authorization model. Can be in JSON or OpenFGA format (optional).
* `--format` : Authorization model input format. Can be "fga" or "json" (optional, defaults to the model file extension if present).

###### Example
`fga store create --name "FGA Demo Store"`

###### Response
```json
{
    "id": "01H0H015178Y2V4CX10C2KGHF4",
    "name": "FGA Demo Store",
    "created_at": "2023-05-19T16:10:07.637585677Z",
    "updated_at": "2023-05-19T16:10:07.637585677Z"
}
```

`fga store create --model Model.fga`

###### Response
```json
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

To automatically set the created store id as an environment variable that will then be used by the CLI, you can use the following command:

```bash
export FGA_STORE_ID=$(fga store create --model model.fga | jq -r .store.id)
```
##### Import Store

###### Command
fga store **import**

###### Parameters
* `--file`: File containing the store.
* `--store-id`: Specifies the store id to import into
* `--max-tuples-per-write`: Max tuples to send in a single write (optional, default=1)
* `--max-parallel-requests`: Max requests to send in parallel (optional, default=4)

###### Example
`fga store import --file model.fga.yaml`

###### Response
```json
{}
```

##### List Stores

###### Command
fga store **list**

###### Parameters
* `--max-pages`: Max number of pages to retrieve (default: 20)

###### Example
`fga store list`

###### Response
```json
{
  "stores": [{
    "id": "..",
    "name": "..",
    "created_at": "",
    "updated_at": "",
    "deleted_at": ""
  }, { .. }]
}
```

##### Get Store

###### Command
fga store **get**

###### Parameters
* `--store-id`: Specifies the store id to get

###### Example
`fga store get --store-id=01H0H015178Y2V4CX10C2KGHF4`

###### Response
```json
{
    "id": "01H0H015178Y2V4CX10C2KGHF4",
    "name": "FGA Demo Store",
    "created_at": "2023-05-19T16:10:07.637585677Z",
    "updated_at": "2023-05-19T16:10:07.637585677Z"
}
```

##### Delete Store

###### Command
fga store **delete**

###### Parameters
* `--store-id`: Specifies the store id to delete

###### Example
`fga store delete --store-id=01H0H015178Y2V4CX10C2KGHF4`

###### Response
```json
{}
```

#### Authorization Models

* `model`

| Description                                                                 | command     | parameters                 | example                                                                                     |
|-----------------------------------------------------------------------------|-------------|----------------------------|---------------------------------------------------------------------------------------------|
| [Read Authorization Models](#read-authorization-models)                     | `list`      | `--store-id`               | `fga model list --store-id=01H0H015178Y2V4CX10C2KGHF4`                                      |
| [Write Authorization Model ](#write-authorization-model)                    | `write`     | `--store-id`, `--file`     | `fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file model.fga`                    |
| [Read a Single Authorization Model](#read-a-single-authorization-model)     | `get`       | `--store-id`, `--model-id` | `fga model get --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1` |
| [Validate an Authorization Model](#validate-an-authorization-model)         | `validate`  | `--file`, `--format`       | `fga model validate --file model.fga`                                                       |
| [Run Tests on an Authorization Model](#run-tests-on-an-authorization-model) | `test`      | `--tests`, `--verbose`     | `fga model test --tests tests.fga.yaml`                                                     |
| [Transform an Authorization Model](#transform-an-authorization-model)       | `transform` | `--file`, `--input-format` | `fga model transform --file model.json`                                                     |


##### Read Authorization Models 

List all authorization models for a store, in descending order by creation date. The first model in the list is the latest one.

###### Command
fga model **list**

###### Parameters
* `--store-id`: Specifies the store id
* `--max-pages`: Max number of pages to retrieve (default: 20)
* `--field`: Fields to display. Choices are: id, created_at and model. Default are id, created_at.

###### Example
`fga model list --store-id=01H0H015178Y2V4CX10C2KGHF4`

###### Response
```json5
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

##### Write Authorization Model 

###### Command
fga model **write**

###### Parameters
* `--store-id`: Specifies the store id
* `--file`: File containing the authorization model.
* `--format`: Authorization model input format. Can be "fga" or "json". Defaults to the file extension if provided (optional)

###### Example
* `fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file model.fga`
* `fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 '{"type_definitions": [ { "type": "user" }, { "type": "document", "relations": { "can_view": { "this": {} } }, "metadata": { "relations": { "can_view": { "directly_related_user_types": [ { "type": "user" } ] }}}} ], "schema_version": "1.1"}' --format json`

###### Response
```json5
{
  "authorization_model_id":"01GXSA8YR785C4FYS3C0RTG7B1"
}
```

##### Read a Single Authorization Model 

###### Command
fga model **get**

###### Parameters
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id
* `--format`: Authorization model output format. Can be "fga" or "json" (default fga).
* `--field`: Fields to display, choices are: `id`, `created_at` and `model`. Default is `model`.

###### Example
`fga model get --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1`

###### Response
```python
model
  schema 1.1

type user

type document
  relations
    define can_view: [user]
```

##### Read the Latest Authorization Model 

If `model-id` is not specified when using the `get` command, the latest authorization model will be returned.

###### Command
fga model **get**

###### Parameters
* `--store-id`: Specifies the store id

###### Example
`fga model get --store-id=01H0H015178Y2V4CX10C2KGHF4`

###### Response
```python
model
  schema 1.1

type user

type document
  relations
    define can_view: [user]
```

##### Validate an Authorization Model

###### Command
fga model **validate**

###### Parameters
* `--file`: File containing the authorization model.
* `--format`: Authorization model input format. Can be "fga" or "json". Defaults to the file extension if provided (optional)

###### Example
`fga model validate --file model.json`

###### JSON Response
* Valid model with an ID
```json5
{"id":"01GPGWB8R33HWXS3KK6YG4ETGH","created_at":"2023-01-11T16:59:22Z","is_valid":true}
```
* Valid model without an ID
```json5
{"is_valid":true}
```
* Invalid model with an ID
```json5
{"id":"01GPGTVEH5NYTQ19RYFQKE0Q4Z","created_at":"2023-01-11T16:33:15Z","is_valid":false,"error":"invalid schema version"}
```
* Invalid model without an ID
```json5
{"is_valid":false,"error":"the relation type 'employee' on 'member' in object type 'group' is not valid"}
```

##### Run Tests on an Authorization Model

Given a model, and a set of tests (tuples, check and list objects requests, and expected results) report back on any tests that do not return the same results as expected.

###### Command
fga model **test**

###### Parameters

* `--tests`: Name of the tests file. Must be in yaml format (see below)
* `--verbose`: Outputs the results in JSON

If a model is provided, the test will run in a built-in OpenFGA instance (you do not need a separate server). Otherwise, the test will be run against the configured store of your OpenFGA instance. When running against a remote instance, the tuples will be sent as contextual tuples, and will have to abide by the OpenFGA server limits (20 contextual tuples per request).

The tests file should be in yaml and have the following format:

```yaml
---
name: Store Name # store name, optional
# model_file: ./model.fga # a global model that would apply to all tests, optional
# model can be used instead of model_file, optional
model: |
  model
    schema 1.1
  type user
  type folder
    relations
      define owner: [user] 
      define parent: [folder]
      define can_view: owner or can_view from parent
      define can_write: owner or can_write from parent
      define can_share: owner

# tuple_file: ./tuples.yaml # global tuples that would apply to all tests, optional
tuples: # global tuples that would apply to all tests, optional
  - user: folder:1
    relation: parent
    object: folder:2
tests: # required
  - name: test-1
    description: testing that the model works # optional
    # tuple_file: ./tuples.json # tuples that would apply per test
    tuples:
      - user: user:anne
        relation: owner
        object: folder:1
    check: # a set of checks to run
      - user: user:anne
        object: folder:1
        assertions:
          # a set of expected results for each relation
          can_view: true
          can_write: true
          can_share: false
    list_objects: # a set of list objects to run
      - user: user:anne
        type: folder
        assertions:
          # a set of expected results for each relation
          can_view:
            - folder:1
            - folder:2
          can_write:
            - folder:1
            - folder:2
  - name: test-2
    description: another test
    tuples:
      - user: user:anne
        relation: owner
        object: folder:1
    check:
      - user: user:anne
        object: folder:1
        assertions:
          # a set of expected results for each relation
          can_view: true
    list_objects:
      - user: user:anne
        type: folder
        assertions:
          # a set of expected results for each relation
          can_view:
            - folder:1
            - folder:2
```

###### Example
`fga model test --tests tests.fga.yaml`

###### Response

```shell
(FAILING) test-1: Checks (2/3 passing) | ListObjects (2/2 passing)
✓ Check(user=user:anne,relation=can_write,object=folder:1)
ⅹ Check(user=user:anne,relation=can_share,object=folder:1): expected=false, got=true, error=<nil>
✓ Check(user=user:anne,relation=can_view,object=folder:1)
---
(PASSING) test-2: Checks (1/1 passing) | ListObjects (1/1 passing)
```

##### Transform an Authorization Model

###### Command
fga model **transform** 

###### Parameters
* `--file`: File containing the authorization model
* `--input-format`: Authorization model input format. Can be "fga" or "json". Defaults to the file extension if provided (optional)

###### Example
`fga model transform --file model.json`

###### Response
```python
model
  schema 1.1

type user

type document
  relations
    define can_view: [user]
```

#### Relationship Tuples

* `tuple`

| Description                                                                       | command   | parameters                           | example                                                                                                           |
|-----------------------------------------------------------------------------------|-----------|--------------------------------------|-------------------------------------------------------------------------------------------------------------------|
| [Write Relationship Tuples](#write-relationship-tuples)                           | `write`   | `--store-id`, `--model-id`           | `fga tuple write user:anne can_view document:roadmap --store-id=01H0H015178Y2V4CX10C2KGHF4`        |
| [Delete Relationship Tuples](#delete-relationship-tuples)                         | `delete`  | `--store-id`, `--model-id`           | `fga tuple delete user:anne can_view document:roadmap --store-id=01H0H015178Y2V4CX10C2KGHF4`                                                          |
| [Read Relationship Tuples](#read-relationship-tuples)                             | `read`    | `--store-id`, `--model-id`           | `fga tuple read --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1`                      |
| [Read Relationship Tuple Changes (Watch)](#read-relationship-tuple-changes-watch) | `changes` | `--store-id`, `--type`, `--continuation-token`,           | `fga tuple changes --store-id=01H0H015178Y2V4CX10C2KGHF4 --type=document --continuation-token=M3w=`                   |
| [Import Relationship Tuples](#import-relationship-tuples)                        | `import`  | `--store-id`, `--model-id`, `--file` | `fga tuple import --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1 --file tuples.json` |

##### Write Relationship Tuples

###### Command
fga tuple **write** <user> <relation> <object> --store-id=<store-id>

###### Parameters
* `<user>`: User
* `<relation>`: Relation
* `<object>`: Object
* `--condition-name`: Condition name (optional)
* `--condition-context`: Condition context (optional)
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id to target (optional)
* `--file`: Specifies the file name, `json`, `yaml` and `csv` files are supported
* `--max-tuples-per-write`: Max tuples to send in a single write (optional, default=1)
* `--max-parallel-requests`: Max requests to send in parallel (optional, default=4)

###### Example (with arguments)
- `fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap`
- `fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap --condition-name inOffice --condition-context '{"office_ip":"10.0.1.10"}'`

###### Response
```json5
{
  "successful": [
    {
      "object":"document:roadmap",
      "relation":"writer",
      "user":"user:annie"
    }
  ],
}
```

###### Example (with file)
`fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.json`

If using a `csv` file, the format should be:

```csv
user_type,user_id,user_relation,relation,object_type,object_id,condition_name,condition_context
folder,product,,parent,folder,product-2021,inOfficeIP,"{""ip_addr"":""10.0.0.1""}"
```


If using a `yaml` file, the format should be:

```yaml
- user: folder:5
  relation: parent
  object: folder:product-2021
- user: folder:product-2021
  relation: parent
  object: folder:product-2021Q1
```

If using a `json` file, the format should be:

```json
[
  {
    "user": "user:anne",
    "relation": "owner",
    "object": "folder:product"
  },
  {
    "user": "folder:product",
    "relation": "parent",
    "object": "folder:product-2021"
  },
  {
    "user": "user:beth",
    "relation": "viewer",
    "object": "folder:product-2021"
  }
]
```

###### Response
```json5
{
  "successful": [
    {
      "object":"document:roadmap",
      "relation":"writer",
      "user":"user:annie"
    }
  ],
  "failed": [
    {
      "tuple_key": {
        "object":"document:roadmap",
        "relation":"writer",
        "user":"carl"
      },
      "reason":"Write validation error ..."
    }
  ]
}
```

##### Delete Relationship Tuples

###### Command
fga tuple **delete** <user> <relation> <object> --store-id=<store-id>

###### Parameters
* `<user>`: User
* `<relation>`: Relation
* `<object>`: Object
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id to target (optional)
* `--file`: Specifies the file name, `yaml` and `json` files are supported
* `--max-tuples-per-write`: Max tuples to send in a single write (optional, default=1)
* `--max-parallel-requests`: Max requests to send in parallel (optional, default=4)

###### Example (with arguments)
`fga tuple delete --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap`

###### Response
```json5
{}
```

###### Example (with file)
`fga tuple delete --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.json`

###### Response
```json5
{
  "successful": [
    {
      "object":"document:roadmap",
      "relation":"writer",
      "user":"user:annie"
    }
  ],
  "failed": [
    {
      "tuple_key": {
        "object":"document:roadmap",
        "relation":"writer",
        "user":"carl"
      },
      "reason":"Write validation error ..."
    }
  ]
}
```

If you want to delete all the tuples in a store, you can use the following code:

```
fga tuple read --simple-output --max-pages=0 > tuples.json
fga tuple delete --file tuples.json
```

##### Read Relationship Tuples

###### Command
fga tuple **read** [--user=<user>] [--relation=<relation>] [--object=<object>]  --store-id=<store-id>

###### Parameters
* `--store-id`: Specifies the store id
* `--user`: User
* `--relation`: Relation
* `--object`: Object
* `--max-pages`: Max number of pages to get. (default 20)
* `--simple-output`: Output simpler JSON version. (It can be used by write and delete commands)

###### Example
`fga tuple read --store-id=01H0H015178Y2V4CX10C2KGHF4 --user user:anne --relation can_view --object document:roadmap`

###### Response
```json5
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
###### Response (--simple-output)
```json5
[
  {
    "object": "document:roadmap",
    "relation": "can_view",
    "user": "user:anne"
  }
]
```


If you want to transform this output in a way that can be then imported using the `fga tuple import` you can run

```
fga tuple read --simple-output --max-pages 0 > tuples.json
fga tuple import --file tuples.json
```

##### Read Relationship Tuple Changes (Watch)

###### Command
fga tuple **changes** --type <type> --store-id=<store-id>

###### Parameters
* `--store-id`: Specifies the store id
* `--type`: Restrict to a specific type (optional)
* `--max-pages`: Max number of pages to retrieve (default: 20)
* `--continuation-token`: Continuation token to start changes from

###### Example
`fga tuple changes --store-id=01H0H015178Y2V4CX10C2KGHF4 --type=document --continuation-token=M3w=`

###### Response
```json5
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

##### Import Relationship Tuples

###### Command
fga tuple **import** --store-id=<store-id> [--model-id=<model-id>] --file <filename> [--max-tuples-per-write=<num>] [--max-parallel-requests=<num>]

###### Parameters
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id to target (optional)
* `--file`: Specifies the file name, `yaml` and `json` files are supported
* `--max-tuples-per-write`: Max tuples to send in a single write (optional, default=1)
* `--max-parallel-requests`: Max requests to send in parallel (optional, default=4)

File format should be:
In YAML:
```yaml
- user: user:anne
  relation: can_view
  object: document:roadmap
- user: user:beth
  relation: can_view
  object: document:roadmap
```

In JSON:

```json
[{
  "user": "user:anne",
  "relation": "can_view",
  "object": "document:roadmap"
}, {
  "user": "user:beth",
  "relation": "can_view",
  "object": "document:roadmap"
}]
```

###### Example
`fga tuple import --store-id=01H0H015178Y2V4CX10C2KGHF4 --file tuples.json`

###### Response
```json5
{
  "successful": [
    {
      "object":"document:roadmap",
      "relation":"writer",
      "user":"user:annie"
    }
  ],
  "failed": [
    {
      "tuple_key": {
        "object":"document:roadmap",
        "relation":"writer",
        "user":"carl"
      },
      "reason":"Write validation error ..."
    }
  ]
}
```

#### Relationship Queries

- `query`

| Description                       | command          | parameters                 | example                                                                                     |
|-----------------------------------|------------------|----------------------------|---------------------------------------------------------------------------------------------|
| [Check](#check)                   | `check`          | `--store-id`, `--model-id` | `fga query check --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap` |
| [List Objects](#list-objects)     | `list-objects`   | `--store-id`, `--model-id` | `fga query list-objects --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document`  |
| [List Relations](#list-relations) | `list-relations` | `--store-id`, `--model-id` | `fga query list-relations --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne document`         |
| [Expand](#expand)                 | `expand`         | `--store-id`, `--model-id` | `fga query expand --store-id=01H0H015178Y2V4CX10C2KGHF4 can_view document:roadmap`          |

##### Check

###### Command
fga query **check** <user> <relation> <object> [--condition] [--contextual-tuple "\<user\> \<relation\> \<object\>"]* --store-id=<store-id> [--model-id=<model-id>]

###### Parameters
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id to target (optional)
* `--contextual-tuple`: Contextual tuples (optional)
* `--context`: Condition context (optional)

###### Example
- `fga query check --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap --contextual-tuple "user:anne can_view folder:product" --contextual-tuple "folder:product parent document:roadmap"`
- `fga query check --store-id="01H4P8Z95KTXXEP6Z03T75Q984" user:anne can_view document:roadmap --context '{"ip_address":"127.0.0.1"}'`


###### Response
```json5
{
    "allowed": true,
}
```

##### List Objects

###### Command
fga query **list-objects** <user> <relation> <object_type> [--contextual-tuple "<user> <relation> <object>"]* --store-id=<store-id> [--model-id=<model-id>]

###### Parameters
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id to target (optional)
* `--contextual-tuple`: Contextual tuples (optional) (can be multiple)
* `--context`: Condition context (optional)

###### Example
- `fga query list-objects --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document --contextual-tuple "user:anne can_view folder:product" --contextual-tuple "folder:product parent document:roadmap"`
- `fga query list-objects --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document --context '{"ip_address":"127.0.0.1"}`

###### Response
```json5
{
    "objects": [
      "document:roadmap",
      "document:budget"
    ],
}
```

##### List Relations

###### Command
fga query **list-objects** <user> <object> [--relation <relation>]* [--contextual-tuple "<user> <relation> <object>"]* --store-id=<store-id> [--model-id=<model-id>]

###### Parameters
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id to target (optional)
* `--contextual-tuple`: Contextual tuples (optional) (can be multiple)
* `--context`: Condition context (optional)

###### Example
`fga query list-relations --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne document:roadmap --relation can_view`
`fga query list-relations --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne document:roadmap --relation can_view --contextual-tuple "user:anne can_view folder:product"`
`fga query list-relations --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne document:roadmap --relation can_view --context '{"ip_address":"127.0.0.1"}`

###### Response
```json5
{
    "relations": [
      "can_view"
    ],
}
```

##### Expand

###### Command
fga query **expand** <relation> <object> --store-id=<store-id> [--model-id=<model-id>]

###### Parameters
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id to target (optional)

###### Example
`fga query expand --store-id=01H0H015178Y2V4CX10C2KGHF4 can_view document:roadmap`

###### Response
```json5
{
  "tree": {
    "root": {
      "name": "repo:openfga/openfga#reader",
      "union": {
        "nodes": [{
          "leaf": {
            "users": {
              "users": ["user:anne"]
            }
          },
          "name": "repo:openfga/openfga#reader"
        }]
      }
    }
  }
}
```

## Contributing

See [CONTRIBUTING](https://github.com/openfga/.github/blob/main/CONTRIBUTING.md).

## Author

[OpenFGA](https://github.com/openfga)

## License

This project is licensed under the Apache-2.0 license. See the [LICENSE](https://github.com/openfga/cli/blob/main/LICENSE) file for more info.
