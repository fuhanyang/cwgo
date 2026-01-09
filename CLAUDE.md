# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

**This repository is archived.** cwgo is no longer actively maintained. Users should migrate to using Kitex and Hertz tools directly.

## Overview

cwgo is an all-in-one code generation tool for CloudWeGo that integrates Kitex (RPC) and Hertz (HTTP) tools. It generates MVC project layouts, server/client code, and database CRUD code for Go microservices.

## Architecture

The codebase follows a modular CLI structure:

- **`cwgo.go`**: Entry point that initializes plugin modes for Hz, Kitex, and MongoDB
- **`cmd/static/`**: CLI command definitions and flag parsing using urfave/cli/v2
  - `cmd.go`: Main command router with subcommands (server, client, model, doc, job, api_list)
  - `*_flags.go`: Flag definitions for each command
- **`config/`**: Configuration structures for CLI arguments (ServerArgument, ClientArgument, ModelArgument, etc.)
- **`pkg/`**: Core business logic packages
  - `server/`: Server code generation (Kitex RPC and Hertz HTTP)
  - `client/`: Client code generation with encapsulation
  - `model/`: Relational database CRUD code generation (GORM-based)
  - `curd/doc/`: Document database code generation (MongoDB)
  - `api_list/`: Hertz route analysis tool
  - `job/`: Batch job code generation
  - `config_generator/`: Configuration file generation
  - `fallback/`: Fallback to native Kitex/Hz tools
- **`tpl/`**: Embedded templates for Kitex and Hertz code generation using Go templates

## Build and Run

```bash
# Build the cwgo binary
go build -o cwgo cwgo.go

# Run tests across all modules
make test

# Run vet checks across all modules
make vet
```

The `hack/tools.sh` script handles multi-module testing and vetting by finding all Go modules in the repository and running tests/vet on each module with race detection and coverage reporting.

## Development Workflow

1. **Multi-module repository**: This repo contains multiple Go modules (main module + test modules in `pkg/api_list/internal/tests/`)
2. **Plugin modes**: cwgo operates as a plugin for both Hz (`app.PluginMode()`) and Kitex (`kitexPluginMode()`)
3. **Template system**: Uses embedded filesystems (`//go:embed` directives) for Kitex/Hertz templates, extracted to temp directories at runtime
4. **Custom template functions**: Sprig v3 functions are registered plus a custom `ToCamel` function for template rendering

## Key Dependencies

- `github.com/cloudwego/kitex`: RPC framework
- `github.com/cloudwego/hertz/cmd/hz`: HTTP tool
- `github.com/urfave/cli/v2`: CLI framework
- `gorm.io/gorm`: ORM for model generation
- `github.com/Masterminds/sprig/v3`: Template function library

## Testing

Tests are located in:
- `pkg/api_list/api_list_test.go`
- `pkg/config_generator/config_generator_test.go`, `sdk_test.go`, `yaml2go_test.go`
- `pkg/job/job_test.go`

Run a single test:
```bash
go test -v ./pkg/api_list -run TestName
```
