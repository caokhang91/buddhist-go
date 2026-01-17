# Buddhist Go - Interpreter Language

Xây dựng một ngôn ngữ interpreter tận dụng sức mạnh của Go (Golang), khai thác khả năng xử lý song song (concurrency) thông qua Goroutines và Channels.

## Mục lục

- [Kiến trúc tổng quan](#kiến-trúc-tổng-quan)
- [Cách tận dụng Go Runtime](#cách-tận-dụng-go-runtime)
- [Các bước triển khai](#các-bước-triển-khai)
- [Cấu trúc Project](#cấu-trúc-project)
- [Thư viện hỗ trợ](#thư-viện-hỗ-trợ)
- [Lưu ý về hiệu năng](#lưu-ý-về-hiệu-năng)

---

## Kiến trúc tổng quan

Một interpreter thông thường sẽ đi qua các bước:

```
Lexer -> Parser -> Abstract Syntax Tree (AST) -> Evaluator
```

Để tận dụng Go Runtime, tập trung vào phần **Evaluator** (trình thực thi) và cách thiết kế AST nodes.

---

## Cách tận dụng Go Runtime

### Tận dụng Concurrency (Goroutines)

Nếu ngôn ngữ hỗ trợ các tác vụ bất đồng bộ hoặc chạy song song, hãy ánh xạ trực tiếp chúng vào Goroutines:

- **Keyword `spawn` hoặc `go`**: Định nghĩa một từ khóa trong ngôn ngữ để kích hoạt Goroutine trong Go.
- **Channels làm công cụ giao tiếp**: Sử dụng `chan` của Go để các script truyền dữ liệu cho nhau thay vì tự xây dựng cơ chế khóa phức tạp.

### Quản lý bộ nhớ (Garbage Collection)

Một trong những lợi thế lớn nhất là **không cần tự viết Garbage Collector (GC)**:

- Khi tạo một đối tượng (ví dụ: `MyObject`), hãy để nó là một `struct` trong Go.
- Khi biến không còn được tham chiếu trong interpreter, Go GC sẽ tự động dọn dẹp.

### Type System và Interfaces

Sử dụng `interface{}` (hoặc `any` trong các bản Go mới) để đại diện cho các kiểu dữ liệu động. Điều này giúp việc kiểm tra kiểu (Type Checking) ở runtime trở nên linh hoạt hơn.

---

## Các bước triển khai

### Bước 1: Định nghĩa Token và Lexer

Sử dụng `struct` để lưu trữ các token và dùng một vòng lặp để quét chuỗi đầu vào.

### Bước 2: Xây dựng AST

Mỗi node trong cây cú pháp nên là một interface:

```go
type Node interface {
    TokenLiteral() string
    Eval() Object // Trả về kết quả thực thi
}
```

### Bước 3: Hiện thực hóa Evaluator với Goroutines

Đây là nơi tận dụng Go:

```go
func evalSpawnExpression(node *ast.SpawnExpression, env *object.Environment) object.Object {
    go func() {
        Eval(node.Function, env)
    }()
    return &object.Null{}
}
```

---

## Cấu trúc Project

```
my-lang/
├── cmd/
│   └── mylang/          # Entry point (main.go) - CLI
├── pkg/
│   ├── lexer/           # Chuyển mã nguồn (string) thành Tokens
│   ├── ast/             # Định nghĩa cấu trúc cây cú pháp
│   ├── parser/          # Chuyển Tokens thành AST
│   ├── code/            # Định nghĩa Opcode và hướng dẫn mã hóa bytecode
│   ├── compiler/        # Chuyển AST thành Bytecode
│   ├── vm/              # Virtual Machine - Thực thi Bytecode
│   └── object/          # Hệ thống kiểu dữ liệu (Integer, Boolean, String, v.v.)
├── go.mod
└── main.go
```

### Chi tiết vai trò từng Package

| Package | Vai trò |
|---------|---------|
| `pkg/code` | Định nghĩa các tập lệnh (Instructions). Mỗi lệnh có một Opcode (1 byte). Ví dụ: `OpAdd`, `OpPush`, `OpJump` |
| `pkg/compiler` | Duyệt qua cây AST và phát ra các chỉ thị bytecode tương ứng |
| `pkg/vm` | Nơi thực thi quan trọng nhất - Stack-based VM với hỗ trợ Concurrency |

---

## Ví dụ luồng xử lý mã nguồn

Xử lý dòng lệnh: `let a = 1 + 2;`

| Bước | Package | Kết quả đầu ra |
|------|---------|----------------|
| 1 | `lexer` | `[LET, IDENT("a"), ASSIGN, INT(1), PLUS, INT(2)]` |
| 2 | `parser` | Cây AST (Node gán với biểu thức cộng) |
| 3 | `compiler` | Bytecode: `PUSH 1, PUSH 2, ADD, SET_VAR "a"` |
| 4 | `vm` | Lấy 1, 2 bỏ vào Stack, thực hiện ADD, lưu kết quả vào bộ nhớ |

---

## Code mẫu: Định nghĩa Opcode

```go
package code

type Opcode byte

const (
    OpConstant Opcode = iota // Đẩy hằng số vào stack
    OpAdd                    // Cộng 2 giá trị trên đỉnh stack
    OpPop                    // Lấy giá trị ra khỏi stack
)

// Definition mô tả cấu trúc của một Opcode
type Definition struct {
    Name          string
    OperandWidths []int // Độ dài của các toán hạng (tính bằng byte)
}

var definitions = map[Opcode]*Definition{
    OpConstant: {"OpConstant", []int{2}}, // 2-byte cho index của hằng số
    OpAdd:      {"OpAdd",      []int{}},  // Không có toán hạng
}
```

---

## Thư viện hỗ trợ

Nếu muốn đẩy nhanh tiến độ, có thể tham khảo:

- **[Participle](https://github.com/alecthomas/participle)**: Thư viện giúp xây dựng Parser bằng cách khai báo struct.
- **[Goyacc](https://pkg.go.dev/golang.org/x/tools/cmd/goyacc)**: Nếu muốn đi theo con đường truyền thống của yacc/lex.
- **[Yaegi](https://github.com/traefik/yaegi)**: Một interpreter cho chính ngôn ngữ Go, cực kỳ mạnh mẽ để học hỏi cách quản lý runtime.

---

## Lưu ý về hiệu năng

### Tree-walking vs Bytecode

- **Tree-walking**: Duyệt cây AST - dễ viết nhưng chậm
- **Bytecode**: Compile AST thành Bytecode và chạy trên Virtual Machine (VM) - nhanh hơn đáng kể

### Environment Sharing

Khi dùng Goroutines, cẩn thận với việc chia sẻ Environment (biến số):

- Sử dụng `sync.Map` hoặc `RWMutex` để tránh race conditions

---

## License

MIT License
