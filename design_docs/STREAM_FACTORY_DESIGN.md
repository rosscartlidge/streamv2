# Stream Factory Architecture Design Document

## Executive Summary

This document presents the Stream Factory solution for achieving Record copyability while maintaining stream processing flexibility. Stream Factories are functions that create fresh streams on demand, eliminating the shared state problem of traditional streams while preserving the full power of stream-based processing for CPU workloads.

## Problem Statement

### Current Architecture Challenges
```go
// Problem 1: Non-copyable Records
record := Record{
    "data": someStream,  // Stream[T] contains stateful closures
}
copy := record  // ❌ Unsafe - shared stream state

// Problem 2: One-time Stream Consumption  
stream1 := record["data"].(Stream[int64])
result1, _ := stream.Sum(stream1)  // Consumes stream

stream2 := record["data"].(Stream[int64])  
result2, _ := stream.Avg(stream2)  // ❌ Stream already consumed!

// Problem 3: GPU Integration Difficulty
// Streams cannot be directly transferred to GPU
gpuData := record["data"].(Stream[float64])  // ❌ Not GPU-compatible
```

### GPU Integration Requirements
- **Contiguous Memory**: GPU kernels need arrays, not function closures
- **Batch Processing**: GPUs work on fixed-size batches efficiently
- **Type Compatibility**: Limited to numeric types
- **Materialization**: Streams must be converted to arrays for GPU processing

## Stream Factory Solution

### Core Concept
Replace streams in Records with **Stream Factories** - functions that create fresh streams on demand.

```go
// Stream Factory: Creates fresh streams
type StreamFactory[T any] func() Stream[T]

// Usage: Copyable records with independent streams
record := Record{
    "data": func() Stream[int64] { return stream.Range(1, 100, 1) },
}

// Safe copying
copy := record  // ✅ Safe - no shared state

// Independent stream consumption
stream1 := record["data"].(StreamFactory[int64])()
stream2 := record["data"].(StreamFactory[int64])()  // Fresh stream
```

## Benefits Analysis

### 1. Record Copyability Achieved
```go
// Before: Non-copyable due to stateful streams
record1 := Record{"data": stream.Range(1, 100, 1)}
record2 := record1  // ❌ Shared state

// After: Fully copyable with independent streams  
factory := func() Stream[int64] { return stream.Range(1, 100, 1) }
record1 := Record{"data": factory}
record2 := record1  // ✅ Safe copy - independent factories

stream1 := record1["data"].(StreamFactory[int64])()
stream2 := record2["data"].(StreamFactory[int64])()  // Independent streams
```

### 2. Multiple Pipeline Processing
```go
record := Record{
    "numbers": func() Stream[int64] { 
        return stream.FromSlice([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
    },
}

// Pipeline 1: Calculate sum
stream1 := record["numbers"].(StreamFactory[int64])()
sum, _ := stream.Sum(stream1)  // Result: 55

// Pipeline 2: Calculate average (independent stream)
stream2 := record["numbers"].(StreamFactory[int64])()
avg, _ := stream.Avg(stream2)  // Result: 5.5

// Pipeline 3: Find maximum (another independent stream)
stream3 := record["numbers"].(StreamFactory[int64])()
max, _ := stream.Max(stream3)  // Result: 10
```

### 3. GPU Integration Flexibility
```go
// CPU Processing: Use streams directly
record := Record{
    "data": func() Stream[float64] { return generateMathData() },
}

cpuStream := record["data"].(StreamFactory[float64])()
cpuResult := stream.Map(func(x float64) float64 { 
    return math.Sin(x) * math.Cos(x) 
})(cpuStream)

// GPU Processing: Materialize to arrays
gpuStream := record["data"].(StreamFactory[float64])()
gpuArray, _ := stream.CollectToArray256(gpuStream)
gpuResult := gpuMapArray(func(x float64) float64 { 
    return math.Sin(x) * math.Cos(x) 
})(gpuArray)
```

### 4. Deterministic Recreation
```go
// Reproducible random data with seeds
randomFactory := func() Stream[float64] {
    rng := rand.New(rand.NewSource(12345))  // Fixed seed
    return stream.Generate(func() float64 { 
        return rng.Float64() 
    }).Limit(1000)
}

record := Record{"random_data": randomFactory}

// Each call produces the same sequence (deterministic)
stream1 := record["random_data"].(StreamFactory[float64])()
stream2 := record["random_data"].(StreamFactory[float64])()
// Both streams produce identical sequences
```

## Trade-offs Analysis

### Advantages ✅

#### 1. **Full Copyability**
- Records become completely copyable value types
- No shared state between record copies
- Safe concurrent access to record copies

#### 2. **Stream Processing Flexibility**
- Maintains full stream processing capabilities
- Multiple independent consumers per factory
- Preserves lazy evaluation benefits

#### 3. **GPU Integration Ready**
- Factories can materialize data as arrays on demand
- Supports both CPU streams and GPU arrays from same source
- Type-safe conversion paths

#### 4. **Memory Efficiency Options**
- Can implement lazy factories that don't store data
- Caching strategies for expensive operations
- On-demand materialization reduces memory footprint

#### 5. **Deterministic Recreation**
- Reproducible streams with seed-based generators
- Consistent behavior across multiple invocations
- Testable and debuggable data generation

### Disadvantages ❌

#### 1. **API Complexity**
- Extra function call required to get stream
- Type assertions needed for factory access
- More complex Value interface

#### 2. **Potential Memory Duplication**
- Each factory call might recreate expensive data
- Without caching, could lead to memory bloat
- File I/O repeated for file-based factories

#### 3. **Performance Overhead**
- Function call overhead per stream creation
- Potential re-computation of derived data
- Cache invalidation complexity

#### 4. **Type Safety Challenges**
- Runtime type assertions required
- More complex generic constraints
- Potential for factory type mismatches

### Mitigation Strategies

#### 1. **Caching for Expensive Operations**
```go
func NewCachedFactory[T any](expensiveComputation func() []T) StreamFactory[T] {
    var cached []T
    var once sync.Once
    
    return func() Stream[T] {
        once.Do(func() {
            cached = expensiveComputation()
        })
        return stream.FromSlice(cached)
    }
}
```

#### 2. **Type-Safe Helpers**
```go
func (r Record) GetStream[T any](key string) (Stream[T], bool) {
    if val, exists := r[key]; exists {
        if factory, ok := val.(StreamFactory[T]); ok {
            return factory(), true
        }
    }
    return nil, false
}
```

#### 3. **Builder Pattern for Ergonomics**
```go
record := stream.NewRecord().
    Int("id", 1).
    String("name", "Alice").
    StreamFactory("data", dataFactory).
    Build()
```

## Implementation Details

### Core Types and Interface Updates

#### Updated Value Interface
```go
type Value interface {
    // Basic copyable types
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64 |
    ~bool | ~string | time.Time |
    Record |
    
    // Array types for GPU processing
    [64]int32 | [128]int32 | [256]int32 | [512]int32 | [1024]int32 |
    [64]int64 | [128]int64 | [256]int64 | [512]int64 | [1024]int64 |
    [64]float32 | [128]float32 | [256]float32 | [512]float32 | [1024]float32 |
    [64]float64 | [128]float64 | [256]float64 | [512]float64 | [1024]float64 |
    
    // Stream factory types for flexible CPU processing
    StreamFactory[int] | StreamFactory[int8] | StreamFactory[int16] |
    StreamFactory[int32] | StreamFactory[int64] |
    StreamFactory[uint] | StreamFactory[uint8] | StreamFactory[uint16] |
    StreamFactory[uint32] | StreamFactory[uint64] |
    StreamFactory[float32] | StreamFactory[float64] |
    StreamFactory[bool] | StreamFactory[string] | StreamFactory[time.Time] |
    StreamFactory[Record]
}
```

### Factory Creation Utilities

#### Basic Factory Constructors
```go
// Create factory from immutable slice
func NewSliceFactory[T any](data []T) StreamFactory[T] {
    // Capture slice data (safe since slices are copied)
    return func() Stream[T] {
        return FromSlice(data)
    }
}

// Create factory from file (re-readable source)
func NewFileFactory(filename string) StreamFactory[Record] {
    return func() Stream[Record] {
        stream, _ := FastTSVToStreamFromFile(filename)
        return stream
    }
}

// Create factory with deterministic generator
func NewGeneratorFactory[T any](generator func() T, count int) StreamFactory[T] {
    return func() Stream[T] {
        return Generate(generator).Limit(count)
    }
}

// Create factory with seeded random generation
func NewSeededRandomFactory(seed int64, count int) StreamFactory[float64] {
    return func() Stream[float64] {
        rng := rand.New(rand.NewSource(seed))
        return Generate(func() float64 { 
            return rng.Float64() 
        }).Limit(count)
    }
}
```

#### Advanced Factory Types
```go
// Cached factory for expensive computations
type CachedFactory[T any] struct {
    computer func() []T
    cached   []T
    once     sync.Once
}

func NewCachedFactory[T any](computer func() []T) StreamFactory[T] {
    cf := &CachedFactory[T]{computer: computer}
    
    return func() Stream[T] {
        cf.once.Do(func() {
            cf.cached = cf.computer()
        })
        return FromSlice(cf.cached)
    }
}

// Dual-format factory for CPU/GPU flexibility
type DualFormatFactory[T GPUNumeric, N int] struct {
    dataSource func() []T
}

func NewDualFormatFactory[T GPUNumeric, N int](dataSource func() []T) *DualFormatFactory[T, N] {
    return &DualFormatFactory[T, N]{dataSource: dataSource}
}

func (df *DualFormatFactory[T, N]) Stream() Stream[T] {
    return FromSlice(df.dataSource())
}

func (df *DualFormatFactory[T, N]) Array() ([N]T, int) {
    data := df.dataSource()
    var arr [N]T
    actualSize := min(len(data), N)
    copy(arr[:actualSize], data)
    return arr, actualSize
}
```

### Record Helper Functions

#### Type-Safe Factory Access
```go
// Safe factory extraction and stream creation
func GetStreamFactory[T any](record Record, key string) (Stream[T], bool) {
    if val, exists := record[key]; exists {
        if factory, ok := val.(StreamFactory[T]); ok {
            return factory(), true
        }
    }
    return nil, false
}

// Get stream or return empty stream
func GetStreamOr[T any](record Record, key string) Stream[T] {
    if stream, exists := GetStreamFactory[T](record, key); exists {
        return stream
    }
    return Empty[T]()
}

// Extract factory without calling it
func GetFactory[T any](record Record, key string) (StreamFactory[T], bool) {
    if val, exists := record[key]; exists {
        if factory, ok := val.(StreamFactory[T]); ok {
            return factory, true
        }
    }
    return nil, false
}
```

#### Fluent Record Builder
```go
type RecordBuilder struct {
    data Record
}

func NewRecord() *RecordBuilder {
    return &RecordBuilder{data: make(Record)}
}

func (rb *RecordBuilder) Int(key string, value int64) *RecordBuilder {
    rb.data[key] = value
    return rb
}

func (rb *RecordBuilder) String(key string, value string) *RecordBuilder {
    rb.data[key] = value
    return rb
}

func (rb *RecordBuilder) StreamFactory[T any](key string, factory StreamFactory[T]) *RecordBuilder {
    rb.data[key] = factory
    return rb
}

func (rb *RecordBuilder) SliceFactory[T any](key string, data []T) *RecordBuilder {
    rb.data[key] = NewSliceFactory(data)
    return rb
}

func (rb *RecordBuilder) CachedFactory[T any](key string, computer func() []T) *RecordBuilder {
    rb.data[key] = NewCachedFactory(computer)
    return rb
}

func (rb *RecordBuilder) Build() Record {
    return rb.data
}
```

## Usage Examples

### Basic Stream Factory Usage
```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    // Create record with stream factory
    record := stream.NewRecord().
        Int("id", 1).
        String("name", "Alice").
        SliceFactory("scores", []float64{85.5, 90.0, 87.5, 92.0}).
        SliceFactory("grades", []string{"A", "A-", "B+", "A"}).
        Build()
    
    // Record is fully copyable
    recordCopy := record  // ✅ Safe copy
    
    // Multiple independent stream processing
    
    // Calculate average score
    if scoreStream, exists := stream.GetStreamFactory[float64](record, "scores"); exists {
        avg, _ := stream.Avg(scoreStream)
        fmt.Printf("Average score: %.2f\n", avg)
    }
    
    // Find maximum score (independent stream)
    if scoreStream, exists := stream.GetStreamFactory[float64](recordCopy, "scores"); exists {
        max, _ := stream.Max(scoreStream)
        fmt.Printf("Maximum score: %.2f\n", max)
    }
    
    // Process grades (yet another independent stream)
    if gradeStream, exists := stream.GetStreamFactory[string](record, "grades"); exists {
        count, _ := stream.Count(gradeStream)
        fmt.Printf("Number of grades: %d\n", count)
    }
}
```

### GroupBy with Stream Factories
```go
func groupByExample() {
    // Create records with stream factories
    records := []stream.Record{
        stream.NewRecord().
            String("department", "Engineering").
            SliceFactory("salaries", []int64{75000, 85000, 95000, 105000}).
            Build(),
        
        stream.NewRecord().
            String("department", "Sales").
            SliceFactory("salaries", []int64{65000, 70000, 80000, 90000}).
            Build(),
        
        stream.NewRecord().
            String("department", "Marketing").
            SliceFactory("salaries", []int64{60000, 65000, 75000, 85000}).
            Build(),
    }
    
    // GroupBy safely copies records with factories
    grouped := stream.GroupBy([]string{"department"})(stream.FromSlice(records))
    
    // Process each group independently
    groupResults, _ := stream.Collect(grouped)
    for _, group := range groupResults {
        dept := stream.GetOr(group, "department", "Unknown")
        
        // Each group processes its salary factory independently
        if salaryStream, exists := stream.GetStreamFactory[int64](group, "salaries"); exists {
            avgSalary, _ := stream.Avg(salaryStream)
            fmt.Printf("%s average salary: $%.2f\n", dept, avgSalary)
        }
    }
}
```

### CPU/GPU Dual Processing
```go
func dualProcessingExample() {
    // Large dataset factory
    dataFactory := stream.NewCachedFactory(func() []float64 {
        // Simulate expensive data computation
        data := make([]float64, 10000)
        for i := range data {
            data[i] = rand.Float64() * 100
        }
        return data
    })
    
    record := stream.NewRecord().
        Int("id", 1).
        CachedFactory("measurements", dataFactory).
        Build()
    
    // CPU processing - use as stream
    if dataStream, exists := stream.GetStreamFactory[float64](record, "measurements"); exists {
        // Complex stream processing
        processed := stream.Pipe(
            stream.Where(func(x float64) bool { return x > 50.0 }),
            stream.Map(func(x float64) float64 { return math.Sin(x) }),
        )(dataStream)
        
        avg, _ := stream.Avg(processed)
        fmt.Printf("CPU processing average: %.4f\n", avg)
    }
    
    // GPU processing - materialize to array
    if dataStream, exists := stream.GetStreamFactory[float64](record, "measurements"); exists {
        // Convert to GPU-ready array
        data, _ := stream.Collect(dataStream)
        if len(data) >= 256 {
            gpuArray := sliceTo256Array(data[:256])
            
            // GPU processing (simulated)
            result := gpuMapArray(func(x float64) float64 { 
                return math.Sin(x) 
            }, gpuArray)
            
            gpuAvg := sumArray256(result) / 256.0
            fmt.Printf("GPU processing average: %.4f\n", gpuAvg)
        }
    }
}

// Helper function to convert slice to array
func sliceTo256Array(data []float64) [256]float64 {
    var arr [256]float64
    copy(arr[:], data)
    return arr
}
```

### File-Based Stream Factory
```go
func fileProcessingExample() {
    // Create factory that re-reads file for each stream
    fileFactory := stream.NewFileFactory("data.tsv")
    
    record := stream.NewRecord().
        String("source", "data.tsv").
        StreamFactory("data", fileFactory).
        Build()
    
    // First processing: calculate total records
    if dataStream, exists := stream.GetStreamFactory[stream.Record](record, "data"); exists {
        count, _ := stream.Count(dataStream)
        fmt.Printf("Total records: %d\n", count)
    }
    
    // Second processing: calculate average age (independent file read)
    if dataStream, exists := stream.GetStreamFactory[stream.Record](record, "data"); exists {
        ages := stream.Map(func(r stream.Record) float64 {
            return stream.GetOr(r, "age", 0.0)
        })(dataStream)
        
        avgAge, _ := stream.Avg(ages)
        fmt.Printf("Average age: %.2f\n", avgAge)
    }
    
    // Third processing: filter and count (another independent file read)
    if dataStream, exists := stream.GetStreamFactory[stream.Record](record, "data"); exists {
        filtered := stream.Where(func(r stream.Record) bool {
            age := stream.GetOr(r, "age", 0.0)
            return age >= 18 && age <= 65
        })(dataStream)
        
        workingAge, _ := stream.Count(filtered)
        fmt.Printf("Working age population: %d\n", workingAge)
    }
}
```

## Migration Strategy

### Phase 1: Foundation (Weeks 1-2)
1. **Implement Core Types**
   - Add StreamFactory[T] type definition
   - Update Value interface to include factory types
   - Create basic factory constructors

2. **Helper Functions**
   - Implement GetStreamFactory and GetStreamOr
   - Add type-safe factory extraction utilities
   - Create fluent RecordBuilder with factory support

### Phase 2: Factory Utilities (Weeks 3-4)  
1. **Factory Constructors**
   - Implement NewSliceFactory, NewFileFactory
   - Add NewGeneratorFactory and NewSeededRandomFactory
   - Create NewCachedFactory for expensive operations

2. **Advanced Factory Types**
   - Implement DualFormatFactory for CPU/GPU support
   - Add factory composition utilities
   - Create factory transformation helpers

### Phase 3: Integration (Weeks 5-6)
1. **Existing Operation Updates**
   - Update GroupBy to work seamlessly with factories
   - Ensure aggregation functions handle factory-based records
   - Add factory support to I/O operations

2. **GPU Integration Prep**
   - Add array materialization helpers
   - Implement factory-to-array conversion utilities
   - Create GPU-ready factory types

### Phase 4: Optimization and Documentation (Weeks 7-8)
1. **Performance Optimization**
   - Add caching strategies for expensive factories
   - Implement factory pooling for memory efficiency
   - Add lazy evaluation optimizations

2. **Documentation and Examples**
   - Create comprehensive usage examples
   - Add migration guide from stream-based records
   - Document performance characteristics and best practices

## Conclusion

The Stream Factory architecture provides an elegant solution that:

1. **✅ Achieves full Record copyability** while preserving stream processing flexibility
2. **✅ Enables multiple independent consumers** from the same data source
3. **✅ Supports both CPU stream processing and GPU array processing** from unified factories
4. **✅ Maintains backward compatibility** with existing stream operations
5. **✅ Provides type safety** through generic factory constraints

**Key Advantages:**
- Solves the copyability problem without sacrificing flexibility
- Enables efficient GPU integration when needed
- Maintains the stream processing paradigm that makes StreamV2 powerful
- Provides deterministic, testable data generation

**Trade-offs:**
- Slightly more complex API due to factory indirection
- Potential memory usage concerns without proper caching
- Requires careful design of factory lifecycles

**Recommendation:** This architecture provides the best balance between flexibility, performance, and architectural cleanliness. It positions StreamV2 to handle both current CPU-focused workloads and future GPU-accelerated processing while maintaining the elegant stream processing model.

The implementation should proceed with careful attention to caching strategies and ergonomic helper functions to minimize the API complexity while maximizing the benefits of the factory pattern.