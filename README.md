# OpenFGA CLI

A cross-platform CLI to interact with an OpenFGA server

[![Go Reference](https://pkg.go.dev/badge/github.com/openfga/cli.svg)](https://pkg.go.dev/github.com/openfga/cli)
[![Release](https://img.shields.io/github/v/release/openfga/cli?sort=semver&color=green)](https://github.com/openfga/cli/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](./LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenfga%2Fcli.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenfga%2Fcli?ref=badge_shield)
[![Discord Server](https://img.shields.io/discord/759188666072825867?color=7289da&logo=discord "Discord Server")](https://discord.com/channels/759188666072825867/930524706854031421)
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
      - [Get a Store](#get-store)
      - [Delete a Store](#delete-store)
    - [Authorization Models](#authorization-models)
      - [Read Authorization Models](#read-authorization-models)
      - [Write Authorization Model](#write-authorization-model)
      - [Read a Single Authorization Model](#read-a-single-authorization-model)
      - [Read the Latest Authorization Model](#read-the-latest-authorization-model)
      - [Validate an Authorization Model](#validate-an-authorization-model)
    - [Relationship Tuples](#relationship-tuples)
      - [Read Relationship Tuple Changes (Watch)](#read-relationship-tuple-changes-watch)
      - [Read Relationship Tuples](#read-relationship-tuples)
      - [Create Relationship Tuples](#create-relationship-tuples)
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
:TODO:

## Building from Source

Make sure you have Go 1.20 or later installed. See the [Go downloads](https://go.dev/dl/) page.

1. Clone the repo to a local directory, and navigate to that directory:

   ```bash
   git clone https://github.com/openfga/cli.git && cd cli
   ```

2. Then use the build command:

   ```bash
   go build -o ./fga main.go
   ```

   or if you have `make` installed, just run:

   ```bash
   make build
   ```

3. Run the OpenFGA CLI with:

   ```bash
   ./fga 
   ```

## Usage

### Configuration

For any command that interacts with an OpenFGA server, these configuration values can be passed (where applicable)

| Name                   | Flag                 | CLI                    |
|------------------------|----------------------|------------------------|
| Server Url             | `--server-url`       | `FGA_SERVER_URL`       |
| Shared Secret          | `--api-token`        | `FGA_API_TOKEN`        |
| Client ID              | `--client-id`        | `FGA_CLIENT_ID`        |
| Client Secret          | `--client-secret`    | `FGA_CLIENT_SECRET`    |
| Token Issuer           | `--api-token-issuer` | `FGA_API_TOKEN_ISSUER` |
| Token Audience         | `--api-audience`     | `FGA_API_AUDIENCE`     |
| Store ID               | `--store-id`         | `FGA_STORE_ID`         |
| Authorization Model ID | `--model-id`         | `FGA_MODEL_ID`         |

### Commands

#### Stores
| Description                     | command  | parameters   | example                                                  |
|---------------------------------|----------|--------------|----------------------------------------------------------|
| [Create a Store](#create-store) | `create` | `--name`     | `fga store create --name="FGA Demo Store"`               |
| [List Stores](#list-stores)     | `list`   |              | `fga store list`                                         |
| [Get a Store](#get-store)       | `get`    | `--store-id` | `fga store get --store-id=01H0H015178Y2V4CX10C2KGHF4`    |
| [Delete a Store](#delete-store) | `delete` | `--store-id` | `fga store delete --store-id=01H0H015178Y2V4CX10C2KGHF4` |

##### Create Store

###### Command
fga store **create**

###### Parameters
* `--name`: Specifies the name of the store to be created

###### Example
`fga store create --name "FGA Demo Store"`

###### JSON Response
```json
{
    "id": "01H0H015178Y2V4CX10C2KGHF4",
    "name": "FGA Demo Store",
    "created_at": "2023-05-19T16:10:07.637585677Z",
    "updated_at": "2023-05-19T16:10:07.637585677Z"
}
```

##### List Stores

###### Command
fga store **list**

###### Parameters
* `--max-pages`: Max number of pages to retrieve (default: 20)

###### Example
`fga store list`

###### JSON Response
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

###### JSON Response
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

###### JSON Response
```json
{}
```

#### Authorization Models

* `model`

| Description                                                             | command | parameters                 | example                                                                                                     |
|-------------------------------------------------------------------------|---------|----------------------------|-------------------------------------------------------------------------------------------------------------|
| [Read Authorization Models](#read-authorization-models)                 | `list`  | `--store-id`               | `fga model list --store-id=01H0H015178Y2V4CX10C2KGHF4`                                                      |
| [Write Authorization Model ](#write-authorization-model)                | `write` | `--store-id`               | `fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 '{"schema_version":"1.1","type_definitions":[...]}'` |
| [Read a Single Authorization Model](#read-a-single-authorization-model) | `get`   | `--store-id`, `--model-id` | `fga model get --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1`                 |

##### Read Authorization Models 

###### Command
fga model **list**

###### Parameters
* `--store-id`: Specifies the store id
* `--max-pages`: Max number of pages to retrieve (default: 20)

###### Example
`fga model list --store-id=01H0H015178Y2V4CX10C2KGHF4`

###### JSON Response
```json5
[{
    "schema_version": "1.1",
    "id": "01GXSA8YR785C4FYS3C0RTG7B1",
    "type_definitions": [
      {"type": "user"},
      // { ... }
    ],
},
// { ... }
]
```

##### Write Authorization Model 

###### Command
fga model **write**

###### Parameters
* `--store-id`: Specifies the store id

###### Example
`fga model write --store-id=01H0H015178Y2V4CX10C2KGHF4 '{"type_definitions": [ { "type": "user" }, { "type": "document", "relations": { "can_view": { "this": {} } }, "metadata": { "relations": { "can_view": { "directly_related_user_types": [ { "type": "user" } ] }}}} ], "schema_version": "1.1"}'`

###### JSON Response
```json5
{
    "schema_version": "1.1",
    "id": "01GXSA8YR785C4FYS3C0RTG7B1",
    "type_definitions": [
      {"type": "user"},
      // { ... }
    ],
}
```

##### Read a Single Authorization Model 

###### Command
fga model **get**

###### Parameters
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id

###### Example
`fga model get --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1`

###### JSON Response
```json5
{
    "schema_version": "1.1",
    "id": "01GXSA8YR785C4FYS3C0RTG7B1",
    "type_definitions": [
      {"type": "user"},
      // { ... }
    ],
}
```

##### Read the Latest Authorization Model 

If `model-id` is not specified when using the `get` command, the latest authorization model will be returned.

###### Command
fga model **get**

###### Parameters
* `--store-id`: Specifies the store id

###### Example
`fga model get --store-id=01H0H015178Y2V4CX10C2KGHF4`

###### JSON Response
```json5
{
    "schema_version": "1.1",
    "id": "01GXSA8YR785C4FYS3C0RTG7B1",
    "type_definitions": [
      {"type": "user"},
      // { ... }
    ],
}
```

##### Validate an Authorization Model

###### Command
fga model **validate**

###### Parameters

###### Example
`fga model validate '{"schema_version":"1.1,"type_definitions":[{"type":"user"}]}'`

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

#### Relationship Tuples

* `tuple`

| Description                                                                       | command   | parameters                 | example                                                                                                 |
|-----------------------------------------------------------------------------------|-----------|----------------------------|---------------------------------------------------------------------------------------------------------|
| [Write Relationship Tuples](#write-relationship-tuples)                           | `write`   | `--store-id`, `--model-id` | `fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 '{"schema_version":"1.1","type_definitions":[...]}'` |
| [Delete Relationship Tuples](#delete-relationship-tuples)                         | `delete`  | `--store-id`, `--model-id` | `fga tuple delete --store-id=01H0H015178Y2V4CX10C2KGHF4`                                                |
| [Read Relationship Tuples](#read-relationship-tuples)                             | `read`    | `--store-id`, `--model-id` | `fga tuple read --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1`            |
| [Read Relationship Tuple Changes (Watch)](#read-relationship-tuple-changes-watch) | `changes` | `--store-id`, `--model-id` | `fga tuple changes --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1`         |
| [Import Relationship Tuples](#import-relationship-tuplesl)                        | `import`  | `--store-id`, `--model-id` | `fga tuple import --store-id=01H0H015178Y2V4CX10C2KGHF4 --model-id=01GXSA8YR785C4FYS3C0RTG7B1`          |

##### Write Relationship Tuples

###### Command
fga tuple **write** <user> <relation> <object> --store-id=<store-id>

###### Parameters
* `<user>`: User
* `<relation>`: Relation
* `<object>`: Object
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id to target (optional)

###### Example
`fga tuple write --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap`

###### JSON Response
```json5
{}
```

##### Delete Relationship Tuples

###### Command
fga tuple **delete** <user> <relation> <object> --store-id=<store-id>

###### Parameters
* `<user>`: User
* `<relation>`: Relation
* `<object>`: Object
* `--store-id`: Specifies the store id

###### Example
`fga tuple delete --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap`

###### JSON Response
```json5
{}
```

##### Read Relationship Tuples

###### Command
fga tuple **read** [--user=<user>] [--relation=<relation>] [--object=<object>]  --store-id=<store-id>

###### Parameters
* `--store-id`: Specifies the store id
* `--user`: User
* `--relation`: Relation
* `--object`: Object

###### Example
`fga tuple read --store-id=01H0H015178Y2V4CX10C2KGHF4 --user user:anne --relation can_view --object document:roadmap`

###### JSON Response
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

##### Read Relationship Tuple Changes (Watch)

###### Command
fga tuple **changes** --type <type> --store-id=<store-id>

###### Parameters
* `--store-id`: Specifies the store id
* `--type`: restrict to a specific type (optional)
* `--max-pages`: Max number of pages to retrieve (default: 20)

###### Example
`fga tuple changes --store-id=01H0H015178Y2V4CX10C2KGHF4 --type document`

###### JSON Response
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
fga query **check** <user> <relation> <object> [--contextual-tuple "<user> <relation> <object>"]* --store-id=<store-id> [--model-id=<model-id>]

###### Parameters
* `--store-id`: Specifies the store id
* `--model-id`: Specifies the model id to target (optional)
* `--contextual-tuple`: Contextual tuples

###### Example
`fga query check --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document:roadmap --contextual-tuple "user:anne can_view folder:product" --contextual-tuple "folder:product parent document:roadmap"`

###### JSON Response
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

###### Example
`fga query list-objects --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne can_view document --contextual-tuple "user:anne can_view folder:product" --contextual-tuple "folder:product parent document:roadmap"`

###### JSON Response
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

###### Example
`fga query list-relations --store-id=01H0H015178Y2V4CX10C2KGHF4 user:anne document:roadmap --relation can_view`

###### JSON Response
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

###### JSON Response
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
