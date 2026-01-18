---
name: Thêm hỗ trợ keyword "then" cho if statements
overview: Thêm hỗ trợ keyword `then` (tùy chọn) để có cú pháp `if condition then { ... } else { ... }`, giúp ngôn ngữ dễ đọc hơn cho trẻ em. Vẫn giữ backward compatibility với cú pháp `if (condition) { ... }`.
todos:
  - id: add_then_token
    content: Thêm token type THEN vào pkg/token/token.go và keywords map
    status: completed
  - id: update_parser
    content: Sửa parseIfExpression() để hỗ trợ optional then keyword
    status: completed
  - id: update_syntax_highlighting
    content: Update syntax highlighting files để nhận diện then keyword
    status: completed
  - id: test_then_syntax
    content: Test cú pháp if-then-else với các trường hợp khác nhau
    status: completed
---

# Thêm hỗ trợ keyword "then" cho if statements

## Mục tiêu

Thêm hỗ trợ keyword `then` (tùy chọn) để có cú pháp `if condition then { ... } else { ... }`, làm cho ngôn ngữ dễ đọc hơn cho trẻ lớp 3. Vẫn giữ backward compatibility với cú pháp `if (condition) { ... }`.

## Các thay đổi cần thiết

### 1. Thêm token type THEN

- File: `pkg/token/token.go`
- Thêm `THEN TokenType = "THEN"` vào danh sách keywords (sau ELSE)
- Thêm `"then": THEN` vào keywords map

### 2. Sửa parser để hỗ trợ optional `then`

- File: `pkg/parser/parser.go`
- Sửa `parseIfExpression()` để:
  - Sau khi parse condition và đóng `)`, kiểm tra có `then` token không
  - Nếu có `then`, consume nó (optional)
  - Tiếp tục parse block statement như bình thường
  - Giữ backward compatibility: nếu không có `then`, vẫn hoạt động bình thường

### 3. Update syntax highlighting (nếu cần)

- File: `vscode-extension/syntaxes/buddhist.tmLanguage.json`
- File: IntelliJ plugin syntax highlighting
- Thêm `then` vào keywords pattern

### 4. Cú pháp hỗ trợ

**Cú pháp mới (với `then`):**

```javascript
if (condition) then {
    // ...
} else {
    // ...
}
```

**Cú pháp cũ (vẫn hoạt động - backward compatible):**

```javascript
if (condition) {
    // ...
} else {
    // ...
}
```

Cả hai đều hợp lệ và có thể dùng song song.

## Chi tiết implementation

### Parse logic

Sau `if (condition)`:

- Nếu next token là `then`: consume `then`, sau đó expect `{`
- Nếu next token là `{`: bỏ qua `then`, parse như bình thường

Điều này cho phép:

- `if (x > 5) then { ... }` ✅
- `if (x > 5) { ... }` ✅

## Testing

- Test với `if (x > 5) then { ... }`
- Test với `if (x > 5) { ... }` (backward compatibility)
- Test với `else if` + `then`
- Test với nested if statements

## Files cần sửa

1. `pkg/token/token.go` - Thêm THEN token type
2. `pkg/parser/parser.go` - Sửa parseIfExpression
3. `vscode-extension/syntaxes/buddhist.tmLanguage.json` - Update syntax highlighting
4. `pkg/lexer/lexer_test.go` - Thêm test case cho `then` keyword (optional)