# AddressManager — Native app

Ứng dụng native chạy script Address Management (GUI) với script nhúng sẵn. Mặc định dùng embedded script; có thể override bằng `-script <file>`.

## Build (từ repo root)

```bash
# macOS / Linux (binary: AddressManager)
go build -ldflags="-s -w" -o AddressManager ./cmd/address-manager

# Hoặc dùng Makefile
make address-manager
```

## Cross-compile Windows (appNative)

```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o AddressManager.exe ./cmd/address-manager

# Hoặc
make address-manager-windows
```

Kết quả: `AddressManager.exe` — chạy trên Windows, mặc định mở GUI với script nhúng.

**Lưu ý:** Build Windows từ macOS có thể lỗi do pixel/OpenGL dùng CGO và build constraints theo OS. Nên build `AddressManager.exe` trên máy Windows hoặc trong CI (vd. GitHub Actions với `runs-on: windows-latest`).

## Chạy

```bash
# Chạy với embedded script (native app mode)
./AddressManager

# Chạy với script ngoài
./AddressManager -script /path/to/other.bl

# -gui (mặc định true, native app luôn chạy GUI)
./AddressManager -gui=true
```

## Cấu trúc

- `main.go` — entry, embed `script.bl`, flag `-script` / `-gui`, gọi VM.
- `script.bl` — script Address Management nhúng (có thể sync từ `examples/address_management.bl` khi cần).
