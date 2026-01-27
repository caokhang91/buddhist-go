# Buddhist-go Makefile
# Run from repo root.
# GUI (faiface/pixel) requires -tags gui and CGO/OpenGL. Default build is without GUI for CI/headless.

.PHONY: address-manager address-manager-windows all-address-manager build build-gui

# Default interpreter build (no GUI; use for CI, govulncheck, headless)
build:
	go build -o buddhist-go .

# Interpreter with GUI (CGO + OpenGL; for running -g/--gui scripts)
build-gui:
	go build -tags gui -o buddhist-go .

# Native app: AddressManager (current OS), embedded script, needs GUI
address-manager:
	go build -tags gui -ldflags="-s -w" -o AddressManager ./cmd/address-manager

# Cross-compile for Windows (appNative.exe)
# Note: May fail on macOS due to CGO/OpenGL; build on Windows or in CI with windows-latest.
address-manager-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o AddressManager.exe ./cmd/address-manager

# Build both native and Windows
all-address-manager: address-manager address-manager-windows
