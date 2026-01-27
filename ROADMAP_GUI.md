# GUI Roadmap

Built-in GUI dùng [faiface/pixel](https://github.com/faiface/pixel) (pixelgl). Tài liệu ngắn về hiện trạng và hướng phát triển.

---

## Hiện trạng

### Đã có

| Thành phần | Mô tả |
|------------|--------|
| `gui_window(config)` | Tạo cửa sổ (title, width, height, vsync) |
| `gui_button(window, config)` | Nút với text, vị trí, kích thước, `onClick` |
| `gui_table(window, config)` | Bảng với headers, data, `onRowClick` |
| `gui_show(window)` | Đánh dấu cửa sổ để hiện khi `gui_run()` chạy |
| `gui_close(window)` | Đóng cửa sổ |
| `gui_run()` | Chạy event loop (block đến khi đóng hết cửa sổ) |

- Input: click chuột cho button và table row đã xử lý.
- Chạy script: `./buddhist-go -g <file.bl>` (hoặc `--gui`).

### Chưa vẽ thật

- **Button**: đã vẽ — rect (imdraw) + text (pixel/text).
- **Table**: `renderTable()` chưa vẽ header/cell/text/grid (mới có khung logic).

Tức là layout + event đã có, phần nhìn thấy (text + hình) chưa implement.

---

## Roadmap

### 1. Rendering cơ bản (ưu tiên)

- [x] **Button**: vẽ bằng `pixel/imdraw` (rect + màu) và `pixel/text` (label).
- [ ] **Table**: vẽ header + từng dòng, text mỗi ô bằng `pixel/text`, có thể dùng imdraw cho grid/background.

### 2. Widgets & API

- [ ] `gui_label(window, config)` — text tĩnh.
- [ ] `gui_textbox(window, config)` — ô nhập với `onChange` / `onSubmit`.
- [ ] `gui_list(window, config)` — danh sách chọn, `onSelect(index)`.
- [ ] Table: scroll khi nhiều dòng, resize cột (optional).

### 3. Layout & styling

- [x] Mốc tọa độ rõ ràng: script (0,0) = góc **trên-trái**; convert nội bộ sang pixel (góc dưới-trái) khi vẽ và hit-test.
- [x] Hỗ trợ màu qua config: `gui_window` dùng `backgroundColor` (hash `{"r", "g", "b"}` float 0–1 hoặc int 0–255); `gui_button` dùng `bgColor`, `textColor`; `gui_table` dùng `headerBg`, `headerTextColor`, `cellBg`, `selectedRowBg`, `textColor`. Chưa có `fontSize` / theme.
- [ ] Theme đơn giản: light/dark, màu nền/viền/màu chữ.

### 4. Ổn định & nền tảng

- [ ] macOS: hướng dẫn rõ khi gặp “Cocoa: Failed to find service port for display” (chạy từ Terminal.app, không SSH/headless).
- [ ] Windows/Linux: kiểm tra chạy ổn định với pixelgl.
- [ ] Thu thập lỗi/glitch khi resize cửa sổ, nhiều cửa sổ, nhiều table lớn.

### 5. Tài liệu & ví dụ

- [ ] README: phần GUI ngắn gọn + link tới `ROADMAP_GUI.md`.
- [ ] Ví dụ: `examples/gui_example.bl`, `examples/address_management.bl` giữ đơn giản, thêm ví dụ cho từng widget mới khi có.

---

## Ghi chú kỹ thuật

- **Thread**: `gui_run()` gọi `pixelgl.Run(eventLoop)` trên luồng hiện tại (block). Chạy bằng `-g` từ process bình thường thì đây là main thread, cửa sổ hiện đúng.
- **Tọa độ**: Script (x,y) dùng **(0,0) = góc trên-trái**; convert nội bộ sang pixel (0,0 = góc dưới-trái) khi vẽ và hit-test.
- **Event loop & lock**: Vòng lặp event giữ `guiStateMu.RLock()` khi duyệt cửa sổ và gọi `handleWindowInput`. **Không được** gọi `guiStateMu.Lock()` (hoặc code chạy VM/builtin cần Lock) trong đoạn đó — dễ deadlock. Mọi ghi state (vd. dismiss alert) và mọi callback VM đều đưa vào hàng đợi deferred, chạy **sau** khi RUnlock: trước `runDeferredStateUpdates()` (vd. xóa alert), rồi `runDeferredCallbacks()` (onClick / onRowClick).
- **Code**: `pkg/object/gui_builtins.go` — state, builtins, `renderWindowComponents` / `handleWindowInput`.

File này sẽ cập nhật khi làm xong từng mục roadmap hoặc khi đổi hướng thiết kế GUI.
