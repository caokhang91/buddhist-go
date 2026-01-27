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

- **Button**: `renderButton()` chưa vẽ chữ/rect (mới có khung logic).
- **Table**: `renderTable()` chưa vẽ header/cell/text/grid (mới có khung logic).

Tức là layout + event đã có, phần nhìn thấy (text + hình) chưa implement.

---

## Roadmap

### 1. Rendering cơ bản (ưu tiên)

- [ ] **Button**: vẽ bằng `pixel/imdraw` (rect + màu) và `pixel/text` (label).
- [ ] **Table**: vẽ header + từng dòng, text mỗi ô bằng `pixel/text`, có thể dùng imdraw cho grid/background.

### 2. Widgets & API

- [ ] `gui_label(window, config)` — text tĩnh.
- [ ] `gui_textbox(window, config)` — ô nhập với `onChange` / `onSubmit`.
- [ ] `gui_list(window, config)` — danh sách chọn, `onSelect(index)`.
- [ ] Table: scroll khi nhiều dòng, resize cột (optional).

### 3. Layout & styling

- [ ] Mốc tọa độ rõ ràng (ví dụ: góc trên-trái = (0,0) hoặc hỗ trợ cả hai).
- [ ] Hỗ trợ font/size/color qua config (ví dụ trong `gui_button`, `gui_label`, `gui_table`).
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
- **Tọa độ**: pixel (0,0) ở góc dưới-trái.
- **Code**: `pkg/object/gui_builtins.go` — state, builtins, `renderWindowComponents` / `handleWindowInput`.

File này sẽ cập nhật khi làm xong từng mục roadmap hoặc khi đổi hướng thiết kế GUI.
