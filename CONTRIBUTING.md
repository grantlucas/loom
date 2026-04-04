# Contributing to Loom

## Prerequisites

- [Go](https://go.dev/dl/) 1.26.1 or later
- [golangci-lint](https://golangci-lint.run/welcome/install/) (for linting)

## Getting Started

```bash
git clone https://github.com/grantlucas/loom.git
cd loom
```

## Install Dependencies

```bash
go mod download
```

## Build

```bash
make build
```

This compiles the binary to `bin/loom`.

## Run

```bash
make run
```

This builds and runs the binary in one step.

### Running the Binary Directly

After building, you can run the binary directly:

```bash
./bin/loom
```

Check the version:

```bash
./bin/loom --version
```

To build with a specific version string:

```bash
make build VERSION=1.0.0
./bin/loom --version
# loom 1.0.0
```

## Test

```bash
make test
```

Runs all tests across the project.

## Lint

```bash
make lint
```

Runs `golangci-lint` against the codebase.

## Clean

```bash
make clean
```

Removes the `bin/` build directory.

## Project Structure

```text
cmd/loom/       Entry point for the CLI application
internal/
  datasource/   Data source abstraction layer
  tui/          Terminal UI (Bubble Tea)
```
