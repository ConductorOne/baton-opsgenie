![Baton Logo](./docs/images/baton-logo.png)

# `baton-opsgenie` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-opsgenie.svg)](https://pkg.go.dev/github.com/conductorone/baton-opsgenie) ![main ci](https://github.com/conductorone/baton-opsgenie/actions/workflows/main.yaml/badge.svg)

`baton-opsgenie` is a connector for Opsgenie built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It communicates with the Opsgenie API to sync data about teams, roles, and users.

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-opsgenie
baton-opsgenie
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_API_KEY=apiKey ghcr.io/conductorone/baton-opsgenie:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-opsgenie/cmd/baton-opsgenie@main

BATON_API_KEY=apiKey
baton resources
```

# Data Model

`baton-opsgenie` will pull down information about the following Opsgenie resources:
- Teams
- Users
- Roles

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-opsgenie` Command Line Usage

```
baton-opsgenie

Usage:
  baton-opsgenie [flags]
  baton-opsgenie [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --api-key string         required: Opsgenie API Key ($BATON_API_KEY)
      --client-id string       The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string   The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string            The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                   help for baton-opsgenie
      --log-format string      The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string       The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning           This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --skip-full-sync         This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --ticketing              This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                version for baton-opsgenie

Use "baton-opsgenie [command] --help" for more information about a command.

```
