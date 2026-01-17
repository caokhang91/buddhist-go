# Cloud Agent Instructions

This repository is a Go project. Use this file for standard build and test
commands when operating in the cloud environment.

## Project entrypoint

- CLI source: `cmd/buddhist`
- Binary name: `buddhist`

## Build

- Build all packages:
  - `go build ./...`
- Build the CLI:
  - `go build -o buddhist ./cmd/buddhist`
- Cross-compile Windows amd64 (no CGO):
  - `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o buddhist.exe ./cmd/buddhist`

## Tests

- Unit tests:
  - `go test ./...`

## Manual checks (optional)

- Run a sample script:
  - `go run ./cmd/buddhist examples/hello.bl`

## Formatting

- Go format:
  - `gofmt -w <files>`
