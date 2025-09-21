# StreamV2 - Modern Stream Processing for Go

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**StreamV2** is a powerful, type-safe stream processing library for Go. Process data with functional programming patterns while keeping your code simple and readable.

## 🚀 **Key Features**

- **🔥 Type-Safe** - Full generics support with compile-time safety
- **⚡ High Performance** - Auto-parallel processing and GPU acceleration
- **📊 Rich I/O** - CSV, JSON, TSV, Protocol Buffers with streaming support
- **🎯 Simple API** - Clean, composable functions that just work

## 📦 **Installation**

```bash
go get github.com/rosscartlidge/streamv2
```

## 🎯 **Quick Start**

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    // Create a stream from any slice
    numbers := stream.FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
    
    // Chain operations with full type safety
    result, _ := stream.Sum(
        stream.Map(func(x int) int { return x * x })(
            stream.Where(func(x int) bool { return x%2 == 0 })(numbers)))
    
    fmt.Printf("Sum of even squares: %d\n", result) // Output: 220
}
```

## 🔥 **Why StreamV2?**

### **Traditional Go**
```go
// Verbose and error-prone
var sum int
for _, item := range data {
    if item.Value > threshold {
        sum += item.Value * item.Value
    }
}
```

### **StreamV2**
```go
// Clean and type-safe
sum, _ := stream.Sum(
    stream.Map(func(x int) int { return x * x })(
        stream.Where(func(x int) bool { return x > threshold })(data)))
```

## 📊 **Working with Structured Data**

```go
// Create records with type-safe builders
users := []stream.Record{
    stream.NewRecord().String("name", "Alice").Int("age", 30).Float("salary", 75000).Build(),
    stream.NewRecord().String("name", "Bob").Int("age", 25).Float("salary", 65000).Build(),
    stream.NewRecord().String("name", "Carol").Int("age", 35).Float("salary", 85000).Build(),
}

// Group and aggregate with explicit operations
results, _ := stream.Collect(
    stream.GroupBy([]string{"department"}, 
        stream.CountField("count", "name"),
        stream.AvgField[float64]("avg_salary", "salary"),
    )(stream.FromSlice(users)))

// Access results safely
for _, result := range results {
    count := stream.GetOr(result, "count", int64(0))
    avgSalary := stream.GetOr(result, "avg_salary", 0.0)
    fmt.Printf("Average salary: $%.0f (%d people)\n", avgSalary, count)
}
```

## 🔧 **Data Processing Made Simple**

### **CSV Processing**
```go
// Read CSV file
data, _ := stream.CSVToStreamFromFile("sales.csv")

// Process and write results  
processed := stream.Map(func(r stream.Record) stream.Record {
    total := stream.GetOr(r, "price", 0.0) * stream.GetOr(r, "quantity", 0.0)
    return r.Set("total", total)
})(data)

stream.StreamToCSVFile(processed, "results.csv")
```

### **Multiple Aggregations**
```go
// Get multiple statistics at once
stats, _ := stream.Aggregates(numbers,
    stream.SumStream[int]("total"),
    stream.CountStream[int]("count"), 
    stream.AvgStream[int]("average"),
    stream.MinStream[int]("minimum"),
    stream.MaxStream[int]("maximum"),
)

fmt.Printf("Stats: %+v\n", stats)
```

### **Parallel Processing**
```go
// Automatic parallelization for expensive operations
results := stream.Map(expensiveFunction)(largeDataset) // Auto-parallel
simple := stream.Map(func(x int) int { return x * 2 })(smallDataset) // Sequential
```

## 💡 **Real-World Examples**

### **Log Analysis**
```go
// Process server logs
logs, _ := stream.CSVToStreamFromFile("access.log")

errorCounts, _ := stream.Collect(
    stream.GroupBy([]string{"status_code"}, 
        stream.CountField("errors", "request_id"),
    )(stream.Where(func(r stream.Record) bool {
        return stream.GetOr(r, "status_code", 0) >= 400
    })(logs)))
```

### **Financial Data**
```go
// Calculate trading statistics
trades, _ := stream.CSVToStreamFromFile("trades.csv")

summary, _ := stream.Aggregates(
    stream.ExtractField[float64]("price")(trades),
    stream.SumStream[float64]("total_volume"),
    stream.AvgStream[float64]("avg_price"),
    stream.MaxStream[float64]("high"),
    stream.MinStream[float64]("low"),
)
```

## 🏗️ **Core Concepts**

StreamV2 has just three main types:
- **`Stream[T]`** - A sequence of values of type T
- **`Filter[T, U]`** - Transforms Stream[T] to Stream[U] 
- **`Record`** - Structured data (like a database row)

Everything else builds on these simple foundations.

## 📚 **Learn More**

- **[Getting Started Guide](STREAMV2_CODELAB.md)** - Step-by-step tutorial
- **[API Reference](docs/api.md)** - Complete function documentation  
- **[Examples](examples/)** - Real-world usage patterns

## 🎯 **Perfect For**

- **Data Analytics** - ETL pipelines and transformations
- **Log Processing** - Parse and analyze structured logs
- **CSV/JSON Processing** - Type-safe data manipulation
- **Real-time Analytics** - Process streaming data efficiently

## 🤝 **Contributing**

StreamV2 is currently stabilizing its API. See [CONTRIBUTING.md](CONTRIBUTING.md) for how to provide feedback and help with testing.

## 📄 **License**

MIT License - see [LICENSE](LICENSE) file for details.

---

**StreamV2** makes stream processing in Go simple, type-safe, and fast. Start building better data pipelines today!