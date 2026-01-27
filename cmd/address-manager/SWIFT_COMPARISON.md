# Buddhist-go (AddressManager) vs Swift

So sánh ngắn trong bối cảnh **app native có GUI** (vd. Address Management).

---

## Tổng quan

| Khía cạnh | Buddhist-go (AddressManager) | Swift (SwiftUI / AppKit) |
|-----------|-----------------------------|---------------------------|
| **Ngôn ngữ** | Script `.bl`, interpreted (bytecode VM) | Compiled, static typing |
| **Runtime** | Interpreter Go + pixel/OpenGL | Native (LLVM), Apple frameworks |
| **GUI** | Built-in `gui_window`, `gui_button`, `gui_table`, `gui_alert` (pixel) | SwiftUI / UIKit / AppKit |
| **Nền tảng** | macOS, Linux, Windows (cùng codebase Go) | macOS / iOS / watchOS / tvOS (đôi khi Linux qua SwiftPM) |

---

## Ngôn ngữ & mô hình

| | Buddhist script | Swift |
|---|----------------|-------|
| **Kiểu** | Dynamic (runtime), một số type hint | Static, strong typing, type inference |
| **Cú pháp** | `place x = 5`, `fn add(a,b) { return a+b }`, `gui_window({...})` | `let x = 5`, `func add(a: Int, b: Int) -> Int { a + b }` |
| **OOP** | Hash + methods qua builtin / convention | struct/class, protocol, inheritance |
| **Concurrency** | `spawn fn() { }`, channels | async/await, actors, Grand Central Dispatch |
| **Script vs binary** | Script chạy qua VM; có thể embed script trong binary (AddressManager) | Compile thành binary, không “chạy script” trừ khi embed interpreter |

Buddhist phù hợp khi muốn **script nhanh, sửa logic không cần build lại**; Swift phù hợp khi cần **hiệu năng tối đa, tooling IDE, và tích hợp sâu với hệ điều hành**.

---

## GUI & app native

| | Buddhist-go AddressManager | Swift (SwiftUI) |
|---|---------------------------|-----------------|
| **Khai báo UI** | Hash + builtin: `gui_button(window, {"text":"Add", "onClick": fn(){ }})` | Declarative: `Button("Add") { ... }` |
| **Layout** | Tọa độ (x,y) top-left, width/height | Stack, Grid, frame, Auto Layout / constraints |
| **Data binding** | Biến global/closure, cập nhật thủ công | `@State`, `@Binding`, `ObservableObject` — reactive |
| **Look & feel** | Tự vẽ (pixel), đơn giản | Native (HIG), Material, tùy biến |
| **Deploy** | Một binary (Go) + embed script | App bundle (macOS/iOS), App Store / notarization |

SwiftUI cho trải nghiệm **native, responsive, accessibility** tốt hơn; Buddhist cho **prototype hoặc tool nội bộ** với ít phụ thuộc toolchain Apple.

---

## Build & phân phối

| | Buddhist-go | Swift |
|---|-------------|-------|
| **Build** | `go build -o AddressManager ./cmd/address-manager` | `swift build` hoặc Xcode |
| **Cross-compile** | `GOOS=windows GOARCH=amd64 go build ...` (CGO/OpenGL có thể gây lỗi từ macOS) | Thường build trên từng nền (macOS → Mac, Xcode → iOS) |
| **Kích thước** | Một binary ~10–20MB+ (VM + runtime + pixel) | App bundle, tùy SwiftUI/frameworks |
| **Phụ thuộc** | Go toolchain; pixel/OpenGL/GLFW | Xcode / Swift toolchain; Cocoa frameworks |

Buddhist dễ “một lệnh build ra app” cho nhiều OS (nếu không vướng CGO); Swift gắn chặt với ecosystem Apple và store.

---

## Hiệu năng

| | Buddhist-go | Swift |
|---|------------|-------|
| **Execution** | Bytecode VM, không JIT | Compiled native, tối ưu LLVM |
| **Khởi động** | Load script → parse → compile → run | Load binary, khởi động nhanh |
| **GUI / vẽ** | Cùng CGO/pixel/OpenGL → nhanh cho app nhỏ | Metal/GPU, Core Animation — tối ưu cho app phức tạp |

Swift vượt trội khi cần CPU/GPU cao; Buddhist đủ cho script, form, bảng đơn giản.

---

## Khi nào dùng gì?

- **Chọn Buddhist-go (AddressManager)** khi:
  - Cần script chỉnh sửa nhanh, embed trong một binary.
  - Tool nội bộ, form/bảng đơn giản, chạy đa nền (macOS/Linux/Windows nếu build được).
  - Không bắt buộc phải “app store” hay look-and-feel chuẩn Apple.

- **Chọn Swift** khi:
  - App gửi lên Mac App Store / iOS App Store.
  - Cần UI phức tạp, accessibility, localization, sandbox.
  - Ưu tiên hiệu năng và tích hợp OS (notifications, iCloud, Shortcuts, v.v.).

---

## Tóm tắt một dòng

- **Buddhist-go AddressManager**: script nhúng trong binary Go, GUI qua pixel — phù hợp prototype / tool nội bộ, đa nền, ít phụ thuộc Apple.
- **Swift (SwiftUI)**: ngôn ngữ compiled, GUI native — phù hợp app ship cho người dùng cuối trên nền tảng Apple.
