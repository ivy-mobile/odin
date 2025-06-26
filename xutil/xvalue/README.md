# xvalue 模块文档

## 概述
`xvalue` 模块提供了一个 `Value` 接口，用于方便地将任意类型的值转换为不同的基础类型和切片类型。该模块依赖 `github.com/ivy-mobile/odin/xconv` 包来完成具体的类型转换操作。

## 项目结构
```
/root/codes/github.com/ivy-mobile/odin/xvalue/
└── xvalue.go
```

## 核心结构体和接口
### `Value` 接口
`Value` 接口定义了一系列方法，用于将存储的值转换为不同的类型，包括整数、浮点数、布尔值、字符串等基础类型，以及这些类型的切片。
```go
type Value interface {
    Int() int
    Int8() int8
    // ... 其他方法
    Slice() []any
    Map() map[string]any
    Scan(pointer any) error
    Value() any
}
```

### `value` 结构体
`value` 结构体是 `Value` 接口的具体实现，它包含一个 `any` 类型的字段 `v`，用于存储实际的值。
```go
 type value struct {
    v any
}
```

## 核心函数
### `NewValue`
创建一个新的 `Value` 实例。如果没有提供参数，则存储的值为 `nil`。
```go
func NewValue(v ...any) Value {
    if len(v) == 0 {
        return &value{v: nil}
    }
    return &value{v: v[0]}
}
```

### 类型转换方法
`value` 结构体实现了 `Value` 接口的所有方法，这些方法通过调用 `xconv` 包中的函数来完成具体的类型转换。例如：
```go
func (v *value) Int() int {
    return xconv.Int(v.Value())
}
```

## 使用示例
```go
package main

import (
    "fmt"
    "github.com/ivy-mobile/odin/xvalue"
)

func main() {
    // 创建一个新的 Value 实例
    val := xvalue.NewValue(123)

    // 转换为不同的类型
    fmt.Println(val.Int())       // 输出: 123
    fmt.Println(val.String())    // 输出: "123"
    fmt.Println(val.Float64())   // 输出: 123.0
}
```