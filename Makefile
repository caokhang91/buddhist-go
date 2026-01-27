# Buddhist-go Makefile
# Run from repo root.

.PHONY: address-manager address-manager-windows all-address-manager

# Native app: AddressManager (current OS), embedded script, -ldflags -s -w
address-manager:
	go build -ldflags="-s -w" -o AddressManager ./cmd/address-manager

# Cross-compile for Windows (appNative.exe)
# Note: May fail on macOS due to CGO/OpenGL; build on Windows or in CI with windows-latest.
address-manager-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o AddressManager.exe ./cmd/address-manager

# Build both native and Windows
all-address-manager: address-manager address-manager-windows
