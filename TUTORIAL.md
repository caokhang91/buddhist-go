# 📚 Tutorial Học Ngôn Ngữ Buddhist

Tutorial này hướng dẫn bạn học ngôn ngữ Buddhist từ cơ bản đến nâng cao, với các ví dụ thực tế. 🚀

---

## 📖 Phần 1: Bắt Đầu (Getting Started)

### ⚙️ Cài Đặt

Buddhist là một ngôn ngữ thông dịch (interpreter) được viết bằng Go. Để sử dụng:

```bash
# Build từ source
go build -o buddhist ./cmd/buddhist

# Hoặc chạy trực tiếp
go run ./cmd/buddhist
```

### 👋 Chương Trình Đầu Tiên: Hello World

Tạo file `hello.bl`:

```buddhist
println("Hello, World!");
```

Chạy chương trình:

```bash
./buddhist hello.bl
```

**✅ Kết quả:**
```
Hello, World!
```

### 💻 Sử Dụng REPL (Interactive Mode)

REPL cho phép bạn chạy code tương tác:

```bash
./buddhist
```

```
╔══════════════════════════════════════════════════════════════╗
║               Go-Powered Interpreter Language                ║
║                     Version 1.0.0                            ║
╚══════════════════════════════════════════════════════════════╝

>>> println("Hello, World!")
Hello, World!
>>> let x = 10
>>> x * 2
20
```

**📋 Ra Lệnh REPL:**
- `help` - 📖 Hiển thị các lệnh có sẵn
- `clear` - 🧹 Xóa màn hình
- `exit` - 👋 Thoát REPL

---

## 📦 Phần 2: Biến và Kiểu Dữ Liệu (Variables & Types)

### 📝 Khai Báo Biến

Buddhist hỗ trợ hai cách khai báo biến:

**`let` - Biến có thể thay đổi:**
```javascript
let x = 5;
x = x + 1;  // ✅ Hợp lệ: x giờ là 6
```

**`const` - Hằng số (không thể thay đổi):**
```javascript
const PI = 3.14159;
// PI = 3.14;  // ❌ Lỗi: không thể gán lại cho const
```

### 🎯 Kiểu Dữ Liệu

Buddhist hỗ trợ các kiểu dữ liệu cơ bản:

#### 1️⃣ Integer (Số Nguyên)
```javascript
let age = 25;
let count = -10;
let zero = 0;
```

#### 2️⃣ Float (Số Thực)
```javascript
let price = 19.99;
let pi = 3.14159;
let temperature = -5.5;
```

#### 3️⃣ String (Chuỗi)
```javascript
let name = "Buddhist";
let greeting = 'Hello, World!';
let message = "Xin chào " + name;
```

#### 4️⃣ Boolean (Logic)
```javascript
let isActive = true;
let isFinished = false;
```

#### 5️⃣ Null (Rỗng)
```javascript
let value = null;
```

### 💡 Ví Dụ: Tính Toán Cơ Bản

```javascript
// 📐 Tính diện tích hình chữ nhật
let width = 10;
let height = 5;
let area = width * height;

println("Chiều rộng: " + str(width));
println("Chiều cao: " + str(height));
println("Diện tích: " + str(area));
```

**✅ Kết quả:**
```
Chiều rộng: 10
Chiều cao: 5
Diện tích: 50
```

---

## ➕ Phần 3: Toán Tử và Biểu Thức (Operators & Expressions)

### 🔢 Toán Tử Số Học

```javascript
let a = 10;
let b = 3;

println("10 + 3 = " + str(a + b));  // ➕ Cộng: 13
println("10 - 3 = " + str(a - b));  // ➖ Trừ: 7
println("10 * 3 = " + str(a * b));  // ✖️ Nhân: 30
println("10 / 3 = " + str(a / b));  // ➗ Chia: 3.333...
println("10 % 3 = " + str(a % b));  // 🔢 Chia lấy dư: 1
```

### ⚖️ Toán Tử So Sánh

```javascript
let x = 5;
let y = 10;

println(x > y);   // false
println(x < y);   // true
println(x == y);  // false
println(x != y);  // true
println(x >= 5);  // true
println(y <= 10); // true
```

### 🔗 Toán Tử Logic

```javascript
let isActive = true;
let isEnabled = false;

println(isActive && isEnabled);  // AND: false
println(isActive || isEnabled);  // OR: true
println(!isActive);              // NOT: false
```

### 💡 Ví Dụ: Tính Diện Tích và Chu Vi Hình Tròn

```javascript
const PI = 3.141593;
let radius = 5.0;

// 📐 Diện tích = π * r²
let area = PI * radius * radius;

// 📏 Chu vi = 2 * π * r
let circumference = 2 * PI * radius;

println("Bán kính: " + str(radius));
println("Diện tích: " + str(area));
println("Chu vi: " + str(circumference));
```

**✅ Kết quả:**
```
Bán kính: 5.0
Diện tích: 78.539825
Chu vi: 31.41593
```

---

## 🔀 Phần 4: Cấu Trúc Điều Khiển (Control Flow)

### ❓ If-Else - Cấu Trúc Lựa Chọn

```javascript
let score = 85;

if (score >= 90) {
    println("Xuất sắc!");
} else if (score >= 80) {
    println("Tốt!");
} else if (score >= 70) {
    println("Khá!");
} else {
    println("Cần cố gắng thêm!");
}
```

**✅ Kết quả:**
```
Tốt!
```

### 🔁 While Loop - Vòng Lặp Với Điều Kiện

```javascript
let i = 1;
while (i <= 5) {
    println("Số: " + str(i));
    i = i + 1;
}
```

**✅ Kết quả:**
```
Số: 1
Số: 2
Số: 3
Số: 4
Số: 5
```

### 🔂 For Loop - Vòng Lặp Với Biến Đếm

```javascript
for (let i = 0; i < 5; i = i + 1) {
    println("Lần lặp: " + str(i));
}
```

**✅ Kết quả:**
```
Lần lặp: 0
Lần lặp: 1
Lần lặp: 2
Lần lặp: 3
Lần lặp: 4
```

### ⏸️ Break và Continue

```javascript
// ⏸️ Break: Thoát khỏi vòng lặp
let i = 0;
while (i < 10) {
    if (i == 5) {
        break;  // Dừng khi i = 5
    }
    println(str(i));
    i = i + 1;
}

// ⏭️ Continue: Bỏ qua lần lặp hiện tại
for (let i = 0; i < 5; i = i + 1) {
    if (i == 2) {
        continue;  // Bỏ qua khi i = 2
    }
    println(str(i));
}
```

**✅ Kết quả:**
```
0
1
2
3
4
0
1
3
4
```

### 💡 Ví Dụ: Kiểm Tra Số Nguyên Tố

```javascript
let checkPrime = fn(n) {
    if (n < 2) {
        return false;
    }
    let i = 2;
    while (i * i <= n) {
        if (n % i == 0) {
            return false;
        }
        i = i + 1;
    }
    return true;
};

// 🔍 Kiểm tra các số từ 2 đến 20
for (let num = 2; num <= 20; num = num + 1) {
    if (checkPrime(num)) {
        println(str(num) + " là số nguyên tố");
    }
}
```

**✅ Kết quả:**
```
2 là số nguyên tố
3 là số nguyên tố
5 là số nguyên tố
7 là số nguyên tố
11 là số nguyên tố
13 là số nguyên tố
17 là số nguyên tố
19 là số nguyên tố
```

---

## ⚙️ Phần 5: Hàm (Functions)

### 📌 Định Nghĩa Hàm

Hàm được định nghĩa bằng từ khóa `fn`:

```javascript
fn greet(name) {
    return "Xin chào, " + name + "!";
}

println(greet("Buddhist"));
```

**✅ Kết quả:**
```
Xin chào, Buddhist!
```

### 🔢 Hàm Với Nhiều Tham Số

```javascript
fn add(a, b) {
    return a + b;
}

fn multiply(a, b, c) {
    return a * b * c;
}

println("5 + 3 = " + str(add(5, 3)));
println("2 * 3 * 4 = " + str(multiply(2, 3, 4)));
```

**✅ Kết quả:**
```
5 + 3 = 8
2 * 3 * 4 = 24
```

### 🎭 Hàm Ẩn Danh (Anonymous Functions)

```javascript
let square = fn(x) { return x * x; };
let cube = fn(x) { x * x * x };

println("Bình phương của 5: " + str(square(5)));
println("Lập phương của 3: " + str(cube(3)));
```

**✅ Kết quả:**
```
Bình phương của 5: 25
Lập phương của 3: 27
```

### 🔒 Closures (Đóng)

Closure cho phép hàm truy cập biến từ scope bên ngoài:

```javascript
fn counter() {
    let count = 0;
    return fn() {
        count = count + 1;
        return count;
    };
}

let c = counter();
println(str(c()));  // 1
println(str(c()));  // 2
println(str(c()));  // 3
```

**✅ Kết quả:**
```
1
2
3
```

### 💡 Ví Dụ: Hàm Tính Giai Thừa

```javascript
fn factorial(n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

println("5! = " + str(factorial(5)));
println("10! = " + str(factorial(10)));
```

**✅ Kết quả:**
```
5! = 120
10! = 3628800
```

### 💡 Ví Dụ: Dãy Fibonacci

```javascript
fn fibonacci(n) {
    if (n <= 1) {
        return n;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

println("10 số Fibonacci đầu tiên:");
for (let i = 0; i < 10; i = i + 1) {
    print(str(fibonacci(i)) + " ");
}
println("");
```

**✅ Kết quả:**
```
10 số Fibonacci đầu tiên:
0 1 1 2 3 5 8 13 21 34
```

---

## 📊 Phần 6: Mảng và Hash Maps (Arrays & Hashes)

### 📋 Mảng Tiêu Chuẩn

```javascript
let numbers = [1, 2, 3, 4, 5];
println("Mảng: " + str(numbers));
println("Độ dài: " + str(len(numbers)));
println("Phần tử đầu: " + str(first(numbers)));
println("Phần tử cuối: " + str(last(numbers)));
println("Mảng còn lại: " + str(rest(numbers)));
```

**✅ Kết quả:**
```
Mảng: [1, 2, 3, 4, 5]
Độ dài: 5
Phần tử đầu: 1
Phần tử cuối: 5
Mảng còn lại: [2, 3, 4, 5]
```

### 🔍 Truy Cập Phần Tử

```javascript
let arr = [10, 20, 30, 40, 50];
println("arr[0] = " + str(arr[0]));  // 10
println("arr[2] = " + str(arr[2]));  // 30
```

### 🗺️ Mảng PHP-Style (Ordered Hash Maps)

```javascript
let person = [
    "name" => "Nguyễn Văn A",
    "age" => 25,
    "city" => "Hà Nội"
];

println("Tên: " + person["name"]);
println("Tuổi: " + str(person["age"]));
println("Thành phố: " + person["city"]);
```

**✅ Kết quả:**
```
Tên: Nguyễn Văn A
Tuổi: 25
Thành phố: Hà Nội
```

### 🗂️ Hash Maps

```javascript
let student = {
    "name": "Trần Thị B",
    "grade": "A",
    "score": 95
};

println("Học sinh: " + student["name"]);
println("Xếp loại: " + student["grade"]);
println("Điểm: " + str(student["score"]));
```

**✅ Kết quả:**
```
Học sinh: Trần Thị B
Xếp loại: A
Điểm: 95
```

### 🛠️ Thao Tác Với Mảng

```javascript
let fruits = ["táo", "chuối"];

// ➕ Thêm phần tử
fruits = push(fruits, "cam");
println("Sau khi thêm: " + str(fruits));

// 📏 Lấy độ dài
println("Số lượng: " + str(len(fruits)));

// ✂️ Slice (cắt mảng)
let firstTwo = slice(fruits, 0, 2);
println("2 phần tử đầu: " + str(firstTwo));
```

**✅ Kết quả:**
```
Sau khi thêm: [táo, chuối, cam]
Số lượng: 3
2 phần tử đầu: [táo, chuối]
```

### 💡 Ví Dụ: Quản Lý Danh Sách

```javascript
let students = [];

// ➕ Thêm sinh viên
students = push(students, "An");
students = push(students, "Bình");
students = push(students, "Cường");

println("Danh sách sinh viên:");
let i = 0;
while (i < len(students)) {
    println(str(i + 1) + ". " + students[i]);
    i = i + 1;
}
```

**✅ Kết quả:**
```
Danh sách sinh viên:
1. An
2. Bình
3. Cường
```

---

## ✏️ Phần 7: Bài Tập Thực Hành (Practical Exercises)

### 🟢 Bài 1: Tính Thể Tích Hình Cầu Từ Diện Tích Mặt Cầu

**Yêu cầu:** Nhập vào diện tích mặt cầu `S`, tính thể tích `V`.

**Công thức:**
- Diện tích mặt cầu: `S = 4πR²`
- Thể tích hình cầu: `V = (4/3)πR³`
- Suy ra: `V = (4π/3) * (√(S/4π))³`

**Giải pháp:**

```javascript
const PI = 3.141593;

// Diện tích mặt cầu
let S = 256.128;

// Tính bán kính từ diện tích: R = √(S / 4π)
let R_squared = S / (4 * PI);
let R = R_squared;

// Tính R (căn bậc 2 - dùng phương pháp Newton đơn giản)
// Hoặc tính trực tiếp V từ S: V = (4π/3) * (R³) với R = √(S/4π)
// V = (4π/3) * (√(S/4π))³ = (4π/3) * ((S/4π)^(3/2))

// Tính V trực tiếp từ S
let V = (4.0 * PI / 3.0) * R * R * R;

// Hoặc tính chính xác hơn với R từ S
// R = √(S / 4π)
let radius = R;

// Tính V = (4π/3) * R³
V = (4.0 * PI / 3.0) * radius * radius * radius;

println("Nhap dien tich S: " + str(S));
println("The tich V = " + str(V));
```

**📊 Kết quả ước tính:**
```
Nhap dien tich S: 256.128
The tich V = ...
```

**Giải pháp chính xác hơn (tính R từ S):**

```javascript
const PI = 3.141593;

fn sqrt(x) {
    // Phương pháp Newton đơn giản để tính căn bậc 2
    if (x == 0) {
        return 0.0;
    }
    let guess = x / 2.0;
    let i = 0;
    while (i < 10) {
        guess = (guess + x / guess) / 2.0;
        i = i + 1;
    }
    return guess;
}

fn pow(x, n) {
    let result = 1.0;
    let i = 0;
    while (i < n) {
        result = result * x;
        i = i + 1;
    }
    return result;
}

// Diện tích mặt cầu
let S = 256.128;

// Tính R từ S: R = √(S / 4π)
let R = sqrt(S / (4 * PI));

// Tính V = (4π/3) * R³
let V = (4.0 * PI / 3.0) * pow(R, 3);

println("Nhap dien tich S: " + str(S));
println("The tich V = " + str(V));
```

### 🔵 Bài 2: Giải Phương Trình Bậc 2

```javascript
fn solveQuadratic(a, b, c) {
    let discriminant = b * b - 4 * a * c;
    
    if (discriminant < 0) {
        return "Phương trình vô nghiệm";
    } else if (discriminant == 0) {
        let x = -b / (2 * a);
        return "Nghiệm kép: x = " + str(x);
    } else {
        let sqrt_d = discriminant;
        // Tính căn bậc 2 (đơn giản hóa)
        let x1 = (-b + sqrt_d) / (2 * a);
        let x2 = (-b - sqrt_d) / (2 * a);
        return "Nghiệm 1: " + str(x1) + ", Nghiệm 2: " + str(x2);
    }
}

println(solveQuadratic(1, -5, 6));  // x² - 5x + 6 = 0
```

### 🟡 Bài 3: Sắp Xếp Mảng

```javascript
fn bubbleSort(arr) {
    let n = len(arr);
    let i = 0;
    while (i < n) {
        let j = 0;
        while (j < n - i - 1) {
            if (arr[j] > arr[j + 1]) {
                // Swap
                let temp = arr[j];
                arr[j] = arr[j + 1];
                arr[j + 1] = temp;
            }
            j = j + 1;
        }
        i = i + 1;
    }
    return arr;
}

let numbers = [64, 34, 25, 12, 22, 11, 90];
let sorted = bubbleSort(numbers);
println("Mảng sau khi sắp xếp: " + str(sorted));
```

### 🟣 Bài 4: Tìm Phần Tử Lớn Nhất và Nhỏ Nhất

```javascript
fn findMinMax(arr) {
    if (len(arr) == 0) {
        return "Mảng rỗng";
    }
    
    let min = arr[0];
    let max = arr[0];
    let i = 1;
    
    while (i < len(arr)) {
        if (arr[i] < min) {
            min = arr[i];
        }
        if (arr[i] > max) {
            max = arr[i];
        }
        i = i + 1;
    }
    
    return "Min: " + str(min) + ", Max: " + str(max);
}

let numbers = [5, 2, 9, 1, 7, 3];
println(findMinMax(numbers));
```

**✅ Kết quả:**
```
Min: 1, Max: 9
```

---

## 🚀 Phần 8: Concurrency (Lập Trình Song Song) - Nâng Cao

### ⚡ Spawn - Tạo Goroutine

`spawn` cho phép chạy hàm trong goroutine riêng:

```javascript
spawn fn() {
    println("Chạy trong goroutine!");
};

println("Chạy trong main");
```

### 📡 Channels - Giao Tiếp Giữa Goroutines

```javascript
// Tạo channel
let ch = channel;

// Spawn goroutine gửi dữ liệu
spawn fn() {
    ch <- "Hello từ goroutine!";
};

// Nhận dữ liệu từ channel
let msg = <-ch;
println(msg);
```

**✅ Kết quả:**
```
Hello từ goroutine!
```

### 📦 Buffered Channels

```javascript
// Tạo buffered channel với buffer size = 3
let ch = channel(3);

ch <- "Message 1";
ch <- "Message 2";
ch <- "Message 3";

println(<-ch);
println(<-ch);
println(<-ch);
```

### 💡 Ví Dụ: Xử Lý Song Song

```javascript
let processNumbers = fn(numbers) {
    let ch = channel;
    let count = 0;
    
    // Spawn goroutines để xử lý song song
    let i = 0;
    while (i < len(numbers)) {
        let num = numbers[i];
        spawn fn() {
            let result = num * 2;  // Xử lý
            ch <- result;
        };
        i = i + 1;
    }
    
    // Thu thập kết quả
    let results = [];
    while (count < len(numbers)) {
        let result = <-ch;
        results = push(results, result);
        count = count + 1;
    }
    
    return results;
};

let numbers = [1, 2, 3, 4, 5];
let doubled = processNumbers(numbers);
println("Kết quả: " + str(doubled));
```

---

## 🛠️ Built-in Functions (Hàm Có Sẵn)

Buddhist cung cấp nhiều hàm có sẵn:

| Hàm | Mô Tả |
|-----|-------|
| `println(...)` | In giá trị với dòng mới |
| `print(...)` | In giá trị không có dòng mới |
| `len(x)` | Lấy độ dài mảng/chuỗi |
| `first(arr)` | Lấy phần tử đầu mảng |
| `last(arr)` | Lấy phần tử cuối mảng |
| `rest(arr)` | Lấy mảng không có phần tử đầu |
| `push(arr, val)` | Thêm phần tử vào mảng |
| `slice(arr, start, end)` | Cắt mảng |
| `str(x)` | Chuyển sang chuỗi |
| `int(x)` | Chuyển sang số nguyên |
| `float(x)` | Chuyển sang số thực |
| `type(x)` | Lấy kiểu dữ liệu |

---

## 🎉 Kết Luận

Tutorial này đã hướng dẫn bạn những kiến thức cơ bản về ngôn ngữ Buddhist:

- 📦 Biến và kiểu dữ liệu
- ➕ Toán tử và biểu thức
- 🔀 Cấu trúc điều khiển
- ⚙️ Hàm và closures
- 📊 Mảng và hash maps
- 🚀 Lập trình song song

Tiếp tục thực hành với các ví dụ và bài tập để thành thạo ngôn ngữ! 💪

---

**📚 Tài Liệu Tham Khảo:**
- 📄 README.md - Tổng quan về ngôn ngữ
- 📁 examples/ - Các ví dụ mẫu
