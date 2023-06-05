# OpenFGA CLI

A cross-platform CLI to interact with an OpenFGA server

## Table of Contents
- [About OpenFGA](#about)
- [Resources](#resources)
- [Installation](#installation)
- [Usage](#usage)
  - [Configuration](#configuration)
  - [Commands](#commands)
    - [Stores](#stores)
      - [List All Stores](#list-stores)
      - [Create a Store](#create-store)
      - [Get a Store](#get-store)
      - [Delete a Store](#delete-store)
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

## Usage

### Configuration

For any command that interacts with an OpenFGA server, these configuration values can be passed (where applicable)

| Name           | Flag                 | CLI                    |
|----------------|----------------------|------------------------|
| Server Url     | `--server-url`       | `FGA_SERVER_URL`       |
| Shared Secret  | `--api-token`        | `FGA_API_TOKEN`        |
| Client ID      | `--client-id`        | `FGA_CLIENT_ID`        |
| Client Secret  | `--client-secret`    | `FGA_CLIENT_SECRET`    |
| Token Issuer   | `--api-token-issuer` | `FGA_API_TOKEN_ISSUER` |
| Token Audience | `--api-audience`     | `FGA_API_AUDIENCE`     |

### Commands

#### Stores
| Description                     | command  | parameters   | example                                                   |
|---------------------------------|----------|--------------|-----------------------------------------------------------|
| [Create a Store](#create-store) | `create` | `--name`     | `fga stores create --name="FGA Demo Store"`               |
| [List Stores](#list-stores)     | `list`   |              | `fga stores list`                                         |
| [Get a Store](#get-store)       | `get`    | `--store_id` | `fga stores get --store_id=01H0H015178Y2V4CX10C2KGHF4`    |
| [Delete a Store](#delete-store) | `delete` | `--store_id` | `fga stores delete --store_id=01H0H015178Y2V4CX10C2KGHF4` |

##### Create Store

###### Command
fga stores **create**

###### Parameters
* `--name`: Specifies the name of the store to be created

###### Example
`fga stores create "FGA Demo Store"`

###### JSON Response
```json
{
    "id": "01H0H015178Y2V4CX10C2KGHF4",
    "name": "FGA Demo Store",
    "created_at": "2023-05-19T16:10:07.637585677Z",
    "updated_at": "2023-05-19T16:10:07.637585677Z"
}
```

###### Text Response
```text
Store <STORE_NAME> with ID: <ID> has been successfully created.
```

##### List Stores

###### Command
fga stores **list**

###### Example
`fga stores list`

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

###### Text Response

```text
id, name, created_at, updated_at, deleted_at
<store_id>, <store_name>, <created_at>, <updated_at>, <deleted_at>
<store_id>, <store_name>, <created_at>, <updated_at>, <deleted_at>
<store_id>, <store_name>, <created_at>, <updated_at>, <deleted_at>
```

##### Get Store

###### Command
fga stores **get**

###### Parameters
* `--store_id`: Specifies the store id to get

###### Example
`fga stores get --store_id=01H0H015178Y2V4CX10C2KGHF4`

###### JSON Response
```json
{
    "id": "01H0H015178Y2V4CX10C2KGHF4",
    "name": "FGA Demo Store",
    "created_at": "2023-05-19T16:10:07.637585677Z",
    "updated_at": "2023-05-19T16:10:07.637585677Z"
}
```

###### Text Response

```text
id: <store_id>
name: <store_name>
created_at: <created_at>
updated_at: <updated_at>
deleted_at: <deleted_at>
```

##### Delete Store

###### Command
fga stores **delete**

###### Parameters
* `--store_id`: Specifies the store id to delete

###### Example
`fga stores delete --store_id=01H0H015178Y2V4CX10C2KGHF4`

###### JSON Response
```json
{}
```

###### Text Response

```text
Store with ID: <ID> has been successfully deleted.
```

## Contributing

See [CONTRIBUTING](https://github.com/openfga/.github/blob/main/CONTRIBUTING.md).

## Author

[OpenFGA](https://github.com/openfga)

## License

This project is licensed under the Apache-2.0 license. See the [LICENSE](https://github.com/openfga/fga-cli/blob/main/LICENSE) file for more info.
