# StreamV2 Go 1.23 Iterator Integration Design

**Status**: Proposal
**Author**: StreamV2 Team
**Date**: September 2025
**Go Version**: 1.23+

## Executive Summary

This document explores integrating Go 1.23's new iterator features (`iter.Seq[T]`, `iter.Seq2[K,V]`, and range over func) into StreamV2 to provide a more idiomatic and ergonomic API while preserving StreamV2's core strengths.

## Current StreamV2 Architecture

### Core Function Type
```go
type Stream[T any] func() (T, error)
```

### Current Usage Pattern
```go
// Creating streams
numbers := stream.FromSlice([]int{1, 2, 3, 4, 5})

// Processing streams
filtered := stream.Where(func(x int) bool { return x > 2 })(numbers)

// Consuming streams
for {
    item, err := filtered()
    if err == stream.EOS {
        break // End of stream
    }
    if err != nil {
        return err // Handle error
    }
    fmt.Println(item)
}
```

### Current Strengths
- **Explicit error handling** - Every operation can fail gracefully
- **Lazy evaluation** - On-demand processing
- **Memory efficient** - No intermediate collections
- **Composable** - Functional pipeline design
- **Type safe** - Full generics support

### Current Pain Points
- **Verbose consumption** - Manual loop with error checking
- **Non-idiomatic** - Doesn't leverage Go's `range` syntax
- **Barrier to adoption** - Unfamiliar pattern for Go developers

## Go 1.23 Iterator Features

### Core Iterator Types
```go
// Single-value iterator
type Seq[V any] func(yield func(V) bool)

// Key-value iterator
type Seq2[K, V any] func(yield func(K, V) bool)
```

### Range Over Function
```go
// Natural range syntax
for v := range seq {
    fmt.Println(v)
}

for k, v := range seq2 {
    fmt.Printf("%v: %v\n", k, v)
}
```

### Standard Library Integration
```go
// New packages: iter, maps, slices with iterator support
slices.Values([]int{1,2,3})     // iter.Seq[int]
slices.All([]int{1,2,3})        // iter.Seq2[int, int] (index, value)
maps.Keys(map[string]int{...})  // iter.Seq[string]
```

## Proposed StreamV2 Iterator Integration

### Option 1: Dual API (Recommended)

Maintain both current API and new iterator API for maximum compatibility:

```go
// Current API (unchanged)
type Stream[T any] func() (T, error)

// New iterator-based API
type Iter[T any] iter.Seq[T]
type IterWithErrors[T any] iter.Seq2[T, error]

// Conversion methods
func (s Stream[T]) ToIter() Iter[T]
func (s Stream[T]) ToIterWithErrors() IterWithErrors[T]
func FromIter[T any](it iter.Seq[T]) Stream[T]
```

#### Usage Examples
```go
// Traditional (error-aware)
numbers := stream.FromSlice([]int{1, 2, 3, 4, 5})
for {
    item, err := numbers()
    if err == stream.EOS { break }
    if err != nil { return err }
    fmt.Println(item)
}

// New iterator style (simple)
for item := range stream.FromSlice([]int{1, 2, 3, 4, 5}).ToIter() {
    fmt.Println(item) // Much cleaner!
}

// New iterator style (with errors)
for item, err := range stream.FromSlice([]int{1, 2, 3, 4, 5}).ToIterWithErrors() {
    if err != nil { return err }
    fmt.Println(item)
}
```

### Option 2: Iterator-First Design

Replace Stream[T] with iterators as the primary type:

```go
// New primary type
type Stream[T any] iter.Seq[T]

// Error handling through separate mechanism
type StreamWithErrors[T any] iter.Seq2[T, error]
```

#### Usage Examples
```go
// Simple case
for item := range stream.FromSlice([]int{1, 2, 3, 4, 5}) {
    fmt.Println(item)
}

// With error handling
for item, err := range stream.FromSliceWithErrors(data) {
    if err != nil { return err }
    fmt.Println(item)
}
```

### Option 3: Hybrid Fluent API

Combine functional style with iterator consumption:

```go
// Fluent builder pattern
pipeline := stream.FromSlice([]int{1, 2, 3, 4, 5}).
    Where(func(x int) bool { return x > 2 }).
    Map(func(x int) string { return fmt.Sprintf("item_%d", x) })

// Iterator consumption
for item := range pipeline {
    fmt.Println(item)
}

// Or traditional consumption
for {
    item, err := pipeline.Next()
    if err == stream.EOS { break }
    if err != nil { return err }
    fmt.Println(item)
}
```

## Error Handling Strategies

### Strategy 1: Dual Iterators (Recommended)
```go
// Error-free iterator for simple cases
type Iter[T any] iter.Seq[T]

// Error-aware iterator for robust code
type IterWithErrors[T any] iter.Seq2[T, error]

// Example usage
for item, err := range stream.ReadCSV("data.csv").ToIterWithErrors() {
    if err != nil {
        log.Printf("Error reading CSV: %v", err)
        continue
    }
    process(item)
}
```

### Strategy 2: Error Channel
```go
func (s Stream[T]) ToIterWithErrorChan() (iter.Seq[T], <-chan error) {
    items := make(chan T)
    errors := make(chan error, 1)

    go func() {
        defer close(items)
        defer close(errors)
        for {
            item, err := s()
            if err == EOS { return }
            if err != nil {
                errors <- err
                return
            }
            items <- item
        }
    }()

    return slices.Values(items), errors
}

// Usage
items, errors := stream.ReadCSV("data.csv").ToIterWithErrorChan()
go func() {
    for err := range errors {
        log.Printf("Stream error: %v", err)
    }
}()

for item := range items {
    process(item)
}
```

### Strategy 3: Result Type
```go
type Result[T any] struct {
    Value T
    Error error
}

func (s Stream[T]) ToResults() iter.Seq[Result[T]] {
    return func(yield func(Result[T]) bool) {
        for {
            value, err := s()
            if err == EOS { return }
            if !yield(Result[T]{value, err}) { return }
            if err != nil { return }
        }
    }
}

// Usage
for result := range stream.ReadCSV("data.csv").ToResults() {
    if result.Error != nil {
        log.Printf("Error: %v", result.Error)
        continue
    }
    process(result.Value)
}
```

## Filter and Transformation API

### Current API
```go
// Functional composition
filtered := stream.Pipe(
    stream.Where(func(x int) bool { return x > 2 }),
    stream.Map(func(x int) string { return fmt.Sprintf("item_%d", x) }),
)(numbers)
```

### New Iterator API Options

#### Option A: Method Chaining
```go
// Fluent interface
type StreamBuilder[T any] struct {
    seq iter.Seq[T]
}

func FromSlice[T any](data []T) *StreamBuilder[T] {
    return &StreamBuilder[T]{seq: slices.Values(data)}
}

func (sb *StreamBuilder[T]) Where(predicate func(T) bool) *StreamBuilder[T] {
    return &StreamBuilder[T]{
        seq: func(yield func(T) bool) {
            for v := range sb.seq {
                if predicate(v) && !yield(v) {
                    return
                }
            }
        },
    }
}

func (sb *StreamBuilder[T]) Map[U any](mapper func(T) U) *StreamBuilder[U] {
    return &StreamBuilder[U]{
        seq: func(yield func(U) bool) {
            for v := range sb.seq {
                if !yield(mapper(v)) {
                    return
                }
            }
        },
    }
}

func (sb *StreamBuilder[T]) Iter() iter.Seq[T] {
    return sb.seq
}

// Usage
for item := range FromSlice([]int{1,2,3,4,5}).
    Where(func(x int) bool { return x > 2 }).
    Map(func(x int) string { return fmt.Sprintf("item_%d", x) }).
    Iter() {
    fmt.Println(item)
}
```

#### Option B: Functional with Iterator Return
```go
// Keep functional style, return iterators
func Where[T any](predicate func(T) bool) func(iter.Seq[T]) iter.Seq[T] {
    return func(seq iter.Seq[T]) iter.Seq[T] {
        return func(yield func(T) bool) {
            for v := range seq {
                if predicate(v) && !yield(v) {
                    return
                }
            }
        }
    }
}

// Usage - familiar functional composition
pipeline := stream.Pipe(
    stream.Where(func(x int) bool { return x > 2 }),
    stream.Map(func(x int) string { return fmt.Sprintf("item_%d", x) }),
)(slices.Values([]int{1,2,3,4,5}))

for item := range pipeline {
    fmt.Println(item)
}
```

## SQL-Style Operations with Iterators

### Current SQL-Style API
```go
// SELECT name, salary FROM employees WHERE active = true LIMIT 10 OFFSET 5
result := stream.Collect(
    stream.Limit[stream.Record](10)(
        stream.Offset[stream.Record](5)(
            stream.Select("name", "salary")(
                stream.Where(func(r stream.Record) bool {
                    return stream.GetOr(r, "active", false)
                })(employeeStream)))))
```

### Iterator-Enhanced SQL Style
```go
// Method chaining approach
for employee := range FromRecords(employees).
    Where(func(r Record) bool { return GetOr(r, "active", false) }).
    Select("name", "salary").
    Offset(5).
    Limit(10).
    Iter() {
    fmt.Printf("Employee: %v\n", employee)
}

// Or collect to slice
employees := slices.Collect(
    FromRecords(employees).
        Where(func(r Record) bool { return GetOr(r, "active", false) }).
        Select("name", "salary").
        Offset(5).
        Limit(10).
        Iter())
```

## I/O Integration

### Current I/O API
```go
// Reading CSV
csvStream, err := stream.CSVToStream("data.csv", stream.CSVConfig{HasHeaders: true})
if err != nil { return err }

// Processing
for {
    record, err := csvStream()
    if err == stream.EOS { break }
    if err != nil { return err }
    process(record)
}
```

### Iterator-Enhanced I/O
```go
// Simple reading
for record := range stream.ReadCSV("data.csv").Iter() {
    process(record) // Errors handled internally or ignored
}

// Error-aware reading
for record, err := range stream.ReadCSV("data.csv").IterWithErrors() {
    if err != nil {
        log.Printf("CSV error: %v", err)
        continue
    }
    process(record)
}

// Fluent pipeline
for processed := range stream.ReadCSV("data.csv").
    Where(func(r Record) bool { return GetOr(r, "active", false) }).
    Map(func(r Record) Record { return AddField(r, "processed_at", time.Now()) }).
    Iter() {
    fmt.Printf("Processed: %v\n", processed)
}
```

## Migration Strategy

### Phase 1: Add Iterator Support (v2.1)
- Add `ToIter()` and `ToIterWithErrors()` methods to existing Stream[T]
- Add `FromIter()` constructor
- Maintain 100% backward compatibility
- Add iterator-based examples

### Phase 2: Enhanced Iterator API (v2.2)
- Add fluent StreamBuilder API
- Add iterator-based I/O methods
- Add more iterator utilities
- Performance optimizations

### Phase 3: Iterator-First API (v3.0) - Optional
- Make iterators the primary API
- Provide compatibility layer for Stream[T]
- Comprehensive documentation and migration guide
- Performance benchmarks and optimizations

## Performance Considerations

### Iterator Overhead
```go
// Current: Direct function call
for {
    item, err := stream()  // Single function call
    // ...
}

// Iterator: Function call + yield callback
for item := range iterator {   // Function call + yield(item) callback
    // ...
}
```

### Memory Allocation
- Iterators may create more temporary closures
- Need benchmarks to measure actual impact
- Potential optimizations through compiler inlining

### Benchmark Plan
```go
func BenchmarkCurrentAPI(b *testing.B) {
    // Benchmark current Stream[T] performance
}

func BenchmarkIteratorAPI(b *testing.B) {
    // Benchmark new iterator performance
}

func BenchmarkConversionOverhead(b *testing.B) {
    // Benchmark Stream[T] → Iterator conversion cost
}
```

## Standard Library Integration

### Leveraging New iter Package
```go
import (
    "iter"
    "slices"
    "maps"
)

// Integration with standard library iterators
func FromSliceValues[T any](slice []T) iter.Seq[T] {
    return slices.Values(slice)
}

func FromMapKeys[K comparable, V any](m map[K]V) iter.Seq[K] {
    return maps.Keys(m)
}

// Conversion to standard collections
results := slices.Collect(
    stream.FromSlice(data).
        Where(predicate).
        Map(transform).
        Iter())
```

### Interoperability Benefits
- Seamless integration with standard library
- Works with any library using iter.Seq[T]
- Future-proof as ecosystem adopts iterators

## Open Questions

### 1. Error Handling Philosophy
- **Question**: Should simple iteration hide errors for ergonomics?
- **Options**:
  - Dual API (simple + error-aware)
  - Always explicit errors
  - Configurable error handling
- **Decision needed**: User research on error handling preferences

### 2. API Consistency
- **Question**: Maintain functional style or adopt method chaining?
- **Options**:
  - Keep functional: `Pipe(Where(x), Map(y))(stream)`
  - Method chaining: `stream.Where(x).Map(y)`
  - Both approaches
- **Decision needed**: Developer experience research

### 3. Performance Requirements
- **Question**: What performance degradation is acceptable?
- **Target**: <5% overhead for iterator conversion
- **Need**: Comprehensive benchmarks

### 4. Go Version Support
- **Question**: When to require Go 1.23+?
- **Options**:
  - Immediate (new major version)
  - Gradual (feature flag or build tags)
  - Separate module
- **Decision needed**: User base analysis

## Next Steps

### Immediate (Week 1-2)
1. **Prototype implementation** of Option 1 (Dual API)
2. **Basic benchmarks** comparing current vs iterator performance
3. **Simple examples** demonstrating iterator usage

### Short Term (Week 3-4)
1. **Complete API design** for core operations
2. **Error handling strategy** implementation and testing
3. **I/O integration** prototype
4. **Performance optimization** initial round

### Medium Term (Month 2-3)
1. **Comprehensive benchmarks** across all operations
2. **User feedback** collection and iteration
3. **Documentation** and migration guides
4. **Production testing** with real workloads

## Success Criteria

### Developer Experience
- **Reduced boilerplate**: 50% less code for simple stream consumption
- **Natural syntax**: Leverages Go's `range` statement
- **Smooth migration**: Existing code continues to work

### Performance
- **Minimal overhead**: <5% performance impact for iterator conversion
- **Memory efficiency**: No significant memory usage increase
- **Lazy evaluation**: Preserved completely

### Ecosystem Integration
- **Standard library**: Seamless integration with iter, slices, maps packages
- **Third-party libraries**: Works with any iter.Seq[T] compatible code
- **Future-proof**: Ready for Go ecosystem evolution

---

## Conclusion

Go 1.23 iterators present a significant opportunity to make StreamV2 more idiomatic and user-friendly while preserving its core strengths. The recommended approach is a dual API that maintains backward compatibility while offering iterator-based ergonomics.

The key to success will be:
1. **Careful error handling design** that balances simplicity and robustness
2. **Performance optimization** to minimize iterator overhead
3. **Gradual migration strategy** that doesn't break existing users
4. **Comprehensive testing** with real-world workloads

This design positions StreamV2 to be the premier streaming library for modern Go development while respecting its existing user base and proven architecture.