# StreamV2 Documentation

Welcome to the StreamV2 documentation! This directory contains comprehensive guides for using StreamV2 effectively.

## Quick Start

New to StreamV2? Start with the [Codelab](../STREAMV2_CODELAB.md) for a hands-on introduction.

## Documentation Files

### Core Documentation
- **[API Reference](api.md)** - Complete function reference for all 131+ exported functions
- **[Performance Guide](performance.md)** - Optimization strategies and best practices
- **[Codelab Tutorial](../STREAMV2_CODELAB.md)** - Progressive 10-step learning path

### Development Documentation  
- **[Test Failure Analysis](../TEST_FAILURE_ANALYSIS.md)** - Debugging guide for current test failures
- **[Examples](../stream_examples/)** - Practical code examples

## Getting Started

### Installation
```bash
go get github.com/rosscartlidge/streamv2
```

### Basic Usage
```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    numbers := []int64{1, 2, 3, 4, 5}
    
    // Transform and filter data
    result, err := stream.Collect(
        stream.Pipe(
            stream.Map(func(x int64) int64 { return x * x }),
            stream.Where(func(x int64) bool { return x > 10 }),
        )(stream.FromSlice(numbers)))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result) // [16 25]
}
```

## Key Features

### ✅ **Type-Safe Streaming**
- Generic types for compile-time safety
- No runtime type errors
- Clear function signatures

### ✅ **Lazy Evaluation** 
- Operations are lazy by default
- Memory efficient for large datasets
- Early termination support

### ✅ **Rich Operation Set**
- **Transformations**: Map, FlatMap, Select
- **Filters**: Where, Limit, Offset  
- **Aggregations**: Sum, Count, Max, Min, Avg
- **I/O**: CSV, JSON, TSV, Protocol Buffers
- **Windowing**: Time-based and count-based windows

### ✅ **Multiple Data Formats**
- CSV with automatic type detection
- JSON (Lines and Array formats)
- TSV (Tab-Separated Values)
- Protocol Buffers (planned)

### ✅ **Performance Optimized**
- Auto-parallelization for complex operations
- Memory-efficient streaming
- CPU and GPU execution backends (planned)

### ✅ **Production Ready**
- Comprehensive error handling
- Extensive test coverage (63%+ with 131+ functions tested)
- Performance benchmarks
- Memory profiling support

## Architecture Overview

StreamV2 is built around these core concepts:

### Stream[T]
```go
type Stream[T any] func() (T, error)
```
The fundamental abstraction - a function that returns elements one at a time.

### Filter[T, U]  
```go
type Filter[T, U any] func(Stream[T]) Stream[U]
```
Functions that transform streams - the building blocks of processing pipelines.

### Record
```go
type Record map[string]any
```
Structured data representation for CSV, JSON, and database-like operations.

## Common Patterns

### Data Processing Pipeline
```go
result, err := stream.Collect(
    stream.Pipe(
        stream.Map(transform),        // Transform each element
        stream.Where(filter),         // Filter elements  
        stream.Limit[T](limit),      // Limit results (SQL LIMIT)
    )(sourceStream))
```

### CSV Analysis
```go
records, err := stream.Collect(
    stream.Where(func(r stream.Record) bool {
        return r["department"].(string) == "Engineering"
    })(stream.CSVToStream(file)))

avgSalary, err := stream.Avg(
    stream.Map(func(r stream.Record) float64 {
        return r["salary"].(float64)
    })(stream.FromSlice(records)))
```

### Streaming Aggregation
```go
err := stream.ForEach(func(record stream.Record) {
    processRecord(record)  // Process each record as it arrives
})(stream.CSVToStream(largeFile))
```

## Learning Path

1. **[Start with the Codelab](../STREAMV2_CODELAB.md)** - Hands-on tutorial with examples
2. **[Read the API docs](api.md)** - Understand available functions  
3. **[Check Performance Guide](performance.md)** - Optimize for your use case
4. **[Explore Examples](../stream_examples/)** - See real-world usage patterns

## Performance Considerations

### Memory Efficiency
- Use streaming operations for large datasets
- Avoid unnecessary `Collect()` calls
- Process data in chunks when needed

### CPU Optimization
- StreamV2 auto-parallelizes complex operations
- Simple operations stay sequential for efficiency
- Use `Pipe()` for readable, efficient pipelines

### I/O Performance
- Stream CSV/JSON for memory efficiency
- Batch process small files for speed
- Use buffered I/O for large files

## Error Handling

Always check for errors:
```go
result, err := stream.Sum(myStream)
if err != nil {
    if err == stream.EOS {
        // Normal end of stream
        return nil
    }
    return fmt.Errorf("processing failed: %w", err)
}
```

## Contributing

See the main repository for contribution guidelines and development setup.

## Support

- **Issues**: Report bugs and feature requests on GitHub
- **Discussions**: Join community discussions  
- **Examples**: Check the examples directory for usage patterns

---

StreamV2 makes stream processing in Go simple, efficient, and type-safe. Happy streaming! 🚀