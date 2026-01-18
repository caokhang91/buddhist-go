---
name: Tạo Tutorial Học Ngôn Ngữ Buddhist
overview: ""
todos:
  - id: create_tutorial_md
    content: Tạo file TUTORIAL.md với cấu trúc đầy đủ các phần học
    status: pending
  - id: add_getting_started
    content: "Viết phần 1: Bắt đầu - cài đặt và Hello World"
    status: pending
  - id: add_variables_types
    content: "Viết phần 2: Biến và kiểu dữ liệu với ví dụ"
    status: pending
  - id: add_operators
    content: "Viết phần 3: Toán tử và biểu thức"
    status: pending
  - id: add_control_flow
    content: "Viết phần 4: Cấu trúc điều khiển với ví dụ thực tế"
    status: pending
  - id: add_functions
    content: "Viết phần 5: Hàm với ví dụ Fibonacci, giai thừa"
    status: pending
  - id: add_data_structures
    content: "Viết phần 6: Mảng và Hash Maps"
    status: pending
  - id: add_practical_exercises
    content: "Viết phần 7: Bài tập thực hành, bao gồm tính thể tích hình cầu"
    status: pending
  - id: create_example_files
    content: Tạo thư mục examples/tutorial/ và file ví dụ sphere_volume.bl
    status: pending
---

# Tạo Tutorial Học Ngôn Ngữ Buddhist

Tạo file `TUTORIAL.md` với hướng dẫn từ cơ bản đến nâng cao, gồm ví dụ thực tế như tính thể tích hình cầu.

## Nội dung Tutorial

### Phần 1: Bắt đầu (Getting Started)

- Cài đặt và chạy chương trình đầu tiên
- Hello World
- Sử dụng REPL

### Phần 2: Biến và Kiểu Dữ Liệu (Variables & Types)

- Khai báo biến với `let` và `const`
- Các kiểu: integer, float, string, boolean, null
- Ví dụ: tính toán cơ bản

### Phần 3: Toán Tử và Biểu Thức (Operators & Expressions)

- Toán tử số học (+, -, *, /, %)
- Toán tử so sánh (>, <, ==, !=, >=, <=)
- Toán tử logic (&&, ||, !)
- Ví dụ: tính diện tích, chu vi

### Phần 4: Cấu Trúc Điều Khiển (Control Flow)

- `if/else` - cấu trúc lựa chọn
- `while` - vòng lặp với điều kiện
- `for` - vòng lặp với biến đếm
- `break` và `continue`
- Ví dụ: kiểm tra số nguyên tố

### Phần 5: Hàm (Functions)

- Định nghĩa hàm với `fn`
- Tham số và giá trị trả về
- Hàm ẩn danh (anonymous functions)
- Đóng (closures)
- Ví dụ: hàm tính giai thừa, Fibonacci

### Phần 6: Mảng và Hash Maps (Arrays & Hashes)

- Mảng tiêu chuẩn: `[1, 2, 3]`
- Mảng PHP-style với keys: `["name" => "value"]`
- Hash maps: `{"key": "value"}`
- Thao tác: truy cập, thêm, xóa
- Ví dụ: quản lý danh sách

### Phần 7: Bài Tập Thực Hành (Practical Exercises)

- Bài 1: Tính thể tích hình cầu từ diện tích mặt cầu
  - Sử dụng công thức: `V = (4π/3) * (√(S/4π))³`
  - Tích hợp `M_PI` (hoặc hằng π)
- Bài 2: Giải phương trình bậc 2
- Bài 3: Sắp xếp mảng
- Bài 4: Tìm phần tử lớn nhất/nhỏ nhất

### Phần 8: Concurrency (Tùy chọn, nâng cao)

- `spawn` - tạo goroutine
- Channels - giao tiếp giữa goroutines
- Ví dụ: xử lý song song

### Cấu trúc file

- File: `TUTORIAL.md` ở thư mục root
- Format: Markdown với code blocks
- Language syntax: `javascript` (cho highlighting)
- Ví dụ code: mỗi phần có code runnable

### Example code blocks

Mỗi phần sẽ có:

```markdown
### Ví dụ 1: [Tên]
\`\`\`javascript
// Code example
\`\`\`

**Kết quả:**
\`\`\`
Output here
\`\`\`
```

### Bài tập thực hành

- Tạo file ví dụ trong `examples/tutorial/`
- File ví dụ: `sphere_volume.bl` cho bài tập hình cầu
- File ví dụ: `exercises.bl` cho các bài tập khác