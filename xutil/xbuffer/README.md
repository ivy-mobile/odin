# xbuffer 模块文档

## 概述
`xbuffer` 模块提供了一系列用于处理二进制数据的读写、缓冲和无拷贝操作的工具。它支持多种数据类型的读写，包括布尔值、整数、浮点数、字符串等，并且提供了缓冲池和无拷贝缓冲区的功能，以提高性能。

## 项目结构
```
buffer.go
buffer_test.go
error.go
nocopy_buffer.go
nocopy_node.go
reader.go
reader_test.go
writer.go
writer_pool.go
```

## 核心结构体和接口
### `Buffer` 接口 <mcfile name="buffer.go" path="/root/codes/github.com/ivy-mobile/odin/xbuffer/buffer.go"></mcfile>
`Buffer` 接口定义了一系列用于操作缓冲区的方法，包括获取长度、获取字节数据、挂载数据、分配内存、迭代和释放资源等。
```go
package xbuffer

type Buffer interface {
    // Len 获取字节长度
    Len() int
    // Bytes 获取所有字节（性能较低，不推荐使用）
    Bytes() []byte
    // Mount 挂载数据到Buffer上
    Mount(block any, whence ...Whence) 
    // Malloc 分配一块内存给Writer
    Malloc(cap int, whence ...Whence) *Writer
    // Range 迭代
    Range(fn func(node *NocopyNode) bool)
    // Release 释放
    Release()
}
```

### `Reader` 结构体 <mcfile name="reader.go" path="/root/codes/github.com/ivy-mobile/odin/xbuffer/reader.go"></mcfile>
`Reader` 结构体用于从二进制数据中读取各种类型的数据。它支持多种数据类型的读取，并且可以设置字节序。
```go
package xbuffer

import (
    "encoding/binary"
    "errors"
    "io"
    "math"
)

type Reader struct {
    buf []byte
    off int
}

// 示例方法：读取int32值
func (r *Reader) ReadInt32(order binary.ByteOrder) (int32, error) {
    buf, err := r.slice(b32)
    if err != nil {
        return 0, err
    }

    return int32(order.Uint32(buf)), nil
}
```

### `Writer` 结构体 <mcfile name="writer.go" path="/root/codes/github.com/ivy-mobile/odin/xbuffer/writer.go"></mcfile>
`Writer` 结构体用于将各种类型的数据写入二进制缓冲区。它支持多种数据类型的写入，并且可以设置字节序。
```go
package xbuffer

import (
    "encoding/binary"
    "math"
)

type Writer struct {
    buf []byte
    off int
}

// 示例方法：写入int64值
func (w *Writer) WriteInt64s(order binary.ByteOrder, values ...int64) {
    w.grow(b64 * len(values))
    for _, v := range values {
        order.PutUint64(w.buf[w.off:w.off+b64], uint64(v))
        w.off += b64
    }
}
```

## 核心函数
### 初始化函数
- `NewReader(data []byte) *Reader`：创建一个新的 `Reader` 实例。
- `NewWriter(cap ...int) *Writer`：创建一个新的 `Writer` 实例。
- `NewWriterPool(capacities []int) *WriterPool`：创建一个新的 `Writer` 缓冲池。

### 读写函数
- `ReadBool() (bool, error)`：从 `Reader` 中读取一个布尔值。
- `WriteBools(values ...bool)`：将多个布尔值写入 `Writer`。

## 无拷贝缓冲区
`NocopyBuffer` 结构体提供了无拷贝缓冲区的功能，可以高效地管理和操作多个数据块。它支持将数据块挂载到缓冲区的头部或尾部，并且可以迭代缓冲区中的所有数据块。
```go
package xbuffer

func (b *NocopyBuffer) Mount(block any, whence ...Whence) {
    // 实现代码
}
```

## 测试
项目中包含多个测试文件，用于验证各个结构体和函数的功能。可以使用以下命令运行测试：
```sh
go test -v ./xbuffer/...
```

## 使用示例
### 读写数据
```go
package main

import (
    "encoding/binary"
    "fmt"
    "github.com/ivy-mobile/odin/xbuffer"
)

func main() {
    // 创建一个Writer并写入数据
    writer := xbuffer.NewWriter(0)
    writer.WriteInt64s(binary.BigEndian, 123456789)

    // 获取写入的数据
    data := writer.Bytes()

    // 创建一个Reader并读取数据
    reader := xbuffer.NewReader(data)
    value, _ := reader.ReadInt64(binary.BigEndian)

    fmt.Println(value) // 输出: 123456789
}
```

### 使用无拷贝缓冲区
```go
package main

import (
    "encoding/binary"
    "github.com/ivy-mobile/odin/xbuffer"
)

func main() {
    // 创建一个无拷贝缓冲区
    buf := xbuffer.NewNocopyBuffer()

    // 分配内存并写入数据
    writer := buf.Malloc(8)
    writer.WriteInt64s(binary.BigEndian, 123456789)

    // 输出缓冲区的字节数据
    fmt.Println(buf.Bytes())
}
```