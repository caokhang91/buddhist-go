# Interpreter ngôn ngữ với Go Runtime

Tài liệu này mô tả định hướng xây dựng một interpreter tận dụng sức mạnh của Go (Golang), đặc biệt là khả năng xử lý song song thông qua Goroutines và Channels.

## Mục tiêu

- Thiết kế kiến trúc rõ ràng: Lexer → Parser → AST → Evaluator/Compiler → VM.
- Khai thác concurrency một cách tự nhiên bằng Goroutines và Channels.
- Tận dụng Go GC để giảm chi phí quản lý bộ nhớ thủ công.
- Xây dựng hệ thống kiểu dữ liệu động linh hoạt.

## Kiến trúc tổng quan

Thông thường một interpreter sẽ đi qua các bước: Lexer → Parser → Abstract Syntax Tree (AST) → Evaluator.
Nếu muốn hiệu năng cao hơn, bạn có thể bổ sung compiler để phát bytecode và chạy trên VM.

## Tận dụng Go Runtime

### Concurrency (Goroutines)

- Keyword `spawn` hoặc `go`: ánh xạ trực tiếp thao tác bất đồng bộ sang Goroutines.
- Channels làm công cụ giao tiếp: dùng `chan` để truyền dữ liệu giữa các script thay vì tự xây locks.

### Quản lý bộ nhớ (Garbage Collection)

- Khi bạn tạo một đối tượng trong ngôn ngữ (ví dụ `MyObject`), hãy biểu diễn bằng `struct` trong Go.
- Khi đối tượng không còn được tham chiếu, Go GC sẽ tự dọn dẹp.

### Type system và interfaces

- Dùng `interface{}` (hoặc `any` trong Go mới) để biểu diễn kiểu dữ liệu động.
- Type checking runtime trở nên linh hoạt hơn với cơ chế reflection hoặc type assertions.

## Lộ trình triển khai

1. Định nghĩa Token và Lexer
2. Xây dựng AST
3. Hiện thực Evaluator hoặc Compiler + VM

Ví dụ AST node cơ bản:

```go
type Node interface {
	TokenLiteral() string
	Eval() Object // Trả về kết quả thực thi
}
```

Ví dụ spawn với Goroutines:

```go
func evalSpawnExpression(node *ast.SpawnExpression, env *object.Environment) object.Object {
	go func() {
		Eval(node.Function, env)
	}()
	return &object.Null{}
}
```

## Cấu trúc thư mục đề xuất

```text
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

## Chi tiết vai trò từng package

### `pkg/code` (trái tim của bytecode)

- Định nghĩa các tập lệnh (Instructions) với Opcode (1 byte).
- Cung cấp hàm `Make` (mã hóa) và `Read` (giải mã) các toán hạng.

### `pkg/compiler`

- Duyệt AST và phát bytecode tương ứng.
- Quản lý constants pool và tối ưu hóa trước khi đưa vào VM.

### `pkg/vm`

- Thực thi bytecode theo mô hình stack-based.
- Khi gặp `OpSpawn`, khởi tạo Goroutine mới để chạy một frame độc lập.

## Ví dụ luồng xử lý

Mã nguồn: `let a = 1 + 2;`

| Bước | Package  | Kết quả |
| --- | --- | --- |
| 1 | `lexer` | `[LET, IDENT("a"), ASSIGN, INT(1), PLUS, INT(2)]` |
| 2 | `parser` | AST node gán với biểu thức cộng |
| 3 | `compiler` | `PUSH 1, PUSH 2, ADD, SET_VAR "a"` |
| 4 | `vm` | Lấy 1 và 2 từ stack, thực hiện `ADD`, lưu kết quả |

## Code mẫu: Định nghĩa Opcode (`pkg/code`)

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
	OpAdd:      {"OpAdd", []int{}},       // Không có toán hạng
}
```

## Gợi ý tiếp theo

- Xác định cú pháp cụ thể của ngôn ngữ.
- Chốt chiến lược thực thi: tree-walking hay bytecode VM.
- Thiết kế ví dụ đầu vào để kiểm thử (expr, function call, concurrency).
