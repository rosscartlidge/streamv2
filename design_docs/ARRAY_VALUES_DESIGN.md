# Array Values Architecture Design Document

## Executive Summary

This document analyzes the proposed architectural change to replace `Stream[T]` types in the `Value` interface with fixed-size arrays like `[4]int`, `[8]Record`, etc. This change would make all Value types copyable, enabling better concurrency, GPU acceleration, and performance optimization while maintaining the existing Stream processing model.

## Current Architecture

### Value Interface (Current)
```go
type Value interface {
    // Basic types (copyable)
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64 |
    ~bool | ~string | time.Time |
    Record |
    
    // Stream types (NOT copyable - function closures)
    Stream[int] | Stream[int8] | Stream[int16] | Stream[int32] | Stream[int64] |
    Stream[uint] | Stream[uint8] | Stream[uint16] | Stream[uint32] | Stream[uint64] |
    Stream[float32] | Stream[float64] |
    Stream[bool] | Stream[string] | Stream[time.Time] |
    Stream[Record]
}
```

### Current Problems
1. **Non-copyable Values**: Stream types contain function closures that cannot be safely copied
2. **Concurrency Issues**: Records containing streams cannot be safely shared between goroutines
3. **GPU Incompatibility**: Function closures cannot be transferred to GPU memory
4. **Serialization Challenges**: Streams cannot be marshaled/unmarshaled
5. **Memory Overhead**: Function closures have runtime overhead

## Proposed Architecture

### New Value Interface
```go
type Value interface {
    // Basic types (copyable)
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64 |
    ~bool | ~string | time.Time |
    Record |
    
    // Fixed-size arrays (copyable, GPU-compatible)
    [4]int | [8]int | [16]int | [32]int | [64]int |
    [4]int64 | [8]int64 | [16]int64 | [32]int64 | [64]int64 |
    [4]float32 | [8]float32 | [16]float32 | [32]float32 | [64]float32 |
    [4]float64 | [8]float64 | [16]float64 | [32]float64 | [64]float64 |
    [4]bool | [8]bool | [16]bool | [32]bool | [64]bool |
    [4]string | [8]string | [16]string | [32]string | [64]string |
    [4]time.Time | [8]time.Time | [16]time.Time | [32]time.Time | [64]time.Time |
    [4]Record | [8]Record | [16]Record | [32]Record | [64]Record
}
```

## Benefits Analysis

### 1. Concurrency and Thread Safety
**Current Problem:**
```go
// NOT SAFE - function closures cannot be copied
record := Record{
    "data": stream.Range(1, 1000, 1),  // Stream[int64]
}

// Sharing this record between goroutines is unsafe
go processRecord(record)  // ❌ Potential data races
```

**With Arrays:**
```go
// SAFE - arrays are value types
record := Record{
    "data": [32]int64{1, 2, 3, 4, 5...},  // Fully copyable
}

// Safe to share between goroutines
go processRecord(record)  // ✅ Safe copy
```

### 2. GPU Acceleration Readiness
**Current GPU Incompatibility:**
```go
// Cannot transfer function closures to GPU
type GPUKernel struct {
    data Stream[float64]  // ❌ Cannot serialize to GPU
}
```

**With Arrays (GPU-Ready):**
```go
// Perfect for GPU kernels
type GPUKernel struct {
    data [256]float64  // ✅ Direct GPU memory transfer
}

// CUDA kernel can process entire array in parallel
__global__ void processArray(float64* data, int size) {
    int idx = blockIdx.x * blockDim.x + threadIdx.x;
    if (idx < size) {
        data[idx] = sin(data[idx]) * cos(data[idx]);  // Vectorized
    }
}
```

### 3. Performance Improvements
**Memory Layout:**
```go
// Current: Scattered memory, function call overhead
stream := func() (int64, error) { /* closure */ }

// Proposed: Contiguous memory, cache-friendly
array := [64]int64{1, 2, 3, ...}  // Single memory block
```

**SIMD Vectorization:**
```go
// Arrays enable automatic vectorization
func sumArray(arr [64]float64) float64 {
    sum := 0.0
    for _, val := range arr {  // Compiler can vectorize this loop
        sum += val
    }
    return sum
}
```

### 4. Serialization and Persistence
**Current Challenge:**
```go
// Cannot marshal streams
record := Record{"stream": stream.Range(1, 100, 1)}
data, err := json.Marshal(record)  // ❌ Fails
```

**With Arrays:**
```go
// Perfect serialization
record := Record{"data": [32]int64{1, 2, 3...}}
data, err := json.Marshal(record)  // ✅ Works perfectly
```

## Trade-offs Analysis

### Advantages ✅

1. **Full Value Semantics**
   - All Value types become copyable
   - Safe concurrent access
   - Predictable memory behavior

2. **GPU Acceleration Enablement**
   - Direct memory transfer to GPU
   - SIMD/vectorization opportunities
   - Batch processing optimization

3. **Performance Improvements**
   - Cache-friendly memory layout
   - Reduced function call overhead
   - Predictable memory allocation

4. **Simplified Concurrency**
   - No need for synchronization primitives
   - Safe sharing between goroutines
   - Eliminates data race possibilities

5. **Serialization Support**
   - JSON/binary marshaling works
   - Network transfer capabilities
   - Persistence and caching support

### Disadvantages ❌

1. **Fixed Size Limitation**
   - Cannot represent infinite sequences
   - Must predetermine maximum size
   - Memory usage for partially filled arrays

2. **Flexibility Reduction**
   - Less dynamic than streams
   - Cannot handle variable-length data elegantly
   - May require padding or truncation

3. **Memory Overhead**
   - Arrays always allocate full size
   - Waste space for small datasets
   - Multiple size variants needed

4. **API Complexity**
   - Multiple array sizes in type constraints
   - Conversion functions needed
   - Size selection decisions

### Mitigation Strategies

#### 1. Size Selection Strategy
```go
// Use common sizes that cover most use cases
[4]T    // Small batches (4 elements)
[16]T   // Medium batches (16 elements)  
[64]T   // Large batches (64 elements)
[256]T  // Very large batches (256 elements)
```

#### 2. Dynamic Size Handling
```go
type SizedArray[T any, N int] struct {
    data [N]T
    size int  // Actual elements used
}

func (sa SizedArray[T, N]) Slice() []T {
    return sa.data[:sa.size]
}
```

#### 3. Infinite Stream Handling
```go
// Keep streams for infinite data, arrays for finite chunks
type Record map[string]any

// Finite data in records (arrays)
record["batch"] = [64]float64{...}

// Infinite data as separate streams
infiniteStream := stream.Generate(func() float64 { return rand.Float64() })
```

## Implementation Plan

### Phase 1: Foundation (Week 1-2)

#### 1.1 Update Value Interface
```go
// File: pkg/stream/stream.go
type Value interface {
    // ... existing basic types ...
    
    // Add array types
    [4]int | [8]int | [16]int | [32]int | [64]int |
    [4]int64 | [8]int64 | [16]int64 | [32]int64 | [64]int64 |
    [4]float32 | [8]float32 | [16]float32 | [32]float32 | [64]float32 |
    [4]float64 | [8]float64 | [16]float64 | [32]float64 | [64]float64 |
    [4]bool | [8]bool | [16]bool | [32]bool | [64]bool |
    [4]string | [8]string | [16]string | [32]string | [64]string |
    [4]time.Time | [8]time.Time | [16]time.Time | [32]time.Time | [64]time.Time |
    [4]Record | [8]Record | [16]Record | [32]Record | [64]Record |
    
    // Keep streams for backward compatibility (temporary)
    Stream[int] | Stream[int64] | Stream[float64] | Stream[bool] | 
    Stream[string] | Stream[time.Time] | Stream[Record]
}
```

#### 1.2 Add Array Helper Functions
```go
// File: pkg/stream/array_utils.go
package stream

// Stream to Array conversion
func StreamToArray4[T any](s Stream[T]) ([4]T, error) {
    var arr [4]T
    for i := 0; i < 4; i++ {
        val, err := s()
        if err != nil {
            return arr, err
        }
        arr[i] = val
    }
    return arr, nil
}

func StreamToArray16[T any](s Stream[T]) ([16]T, error) {
    var arr [16]T
    for i := 0; i < 16; i++ {
        val, err := s()
        if err != nil {
            return arr, err
        }
        arr[i] = val
    }
    return arr, nil
}

// Array to Stream conversion  
func Array4ToStream[T any](arr [4]T) Stream[T] {
    i := 0
    return func() (T, error) {
        if i >= 4 {
            var zero T
            return zero, EOS
        }
        val := arr[i]
        i++
        return val, nil
    }
}

// Dynamic sized arrays
type SizedArray[T any, N int] struct {
    Data [N]T
    Size int
}

func (sa SizedArray[T, N]) Slice() []T {
    return sa.Data[:sa.Size]
}

func NewSizedArray[T any, N int](slice []T) SizedArray[T, N] {
    var arr SizedArray[T, N]
    arr.Size = min(len(slice), N)
    copy(arr.Data[:arr.Size], slice)
    return arr
}
```

#### 1.3 Add Array Processing Functions
```go
// File: pkg/stream/array_ops.go
package stream

// Map over arrays
func MapArray4[T, U any](fn func(T) U, arr [4]T) [4]U {
    var result [4]U
    for i, val := range arr {
        result[i] = fn(val)
    }
    return result
}

func MapArray16[T, U any](fn func(T) U, arr [16]T) [16]U {
    var result [16]U
    for i, val := range arr {
        result[i] = fn(val)
    }
    return result
}

// Filter arrays (returns slice since size varies)
func FilterArray4[T any](fn func(T) bool, arr [4]T) []T {
    var result []T
    for _, val := range arr {
        if fn(val) {
            result = append(result, val)
        }
    }
    return result
}

// Reduce arrays
func ReduceArray4[T, U any](fn func(U, T) U, initial U, arr [4]T) U {
    result := initial
    for _, val := range arr {
        result = fn(result, val)
    }
    return result
}

// Aggregate arrays
func SumArray4[T Numeric](arr [4]T) T {
    var sum T
    for _, val := range arr {
        sum += val
    }
    return sum
}

func SumArray16[T Numeric](arr [16]T) T {
    var sum T
    for _, val := range arr {
        sum += val
    }
    return sum
}
```

### Phase 2: Core Integration (Week 3-4)

#### 2.1 Update Record Operations
```go
// File: pkg/stream/stream.go
// Update Get/Set functions to handle arrays

func GetArray4[T any](record Record, key string) ([4]T, bool) {
    if val, exists := record[key]; exists {
        if arr, ok := val.([4]T); ok {
            return arr, true
        }
    }
    var zero [4]T
    return zero, false
}

func GetArrayOr4[T any](record Record, key string, defaultVal [4]T) [4]T {
    if arr, exists := GetArray4[T](record, key); exists {
        return arr
    }
    return defaultVal
}
```

#### 2.2 Update I/O Operations
```go
// File: pkg/stream/io.go
// Add array-based I/O functions

func FastTSVToArray16(reader io.Reader) ([][16]Record, error) {
    source := NewFastTSVSource(reader)
    stream := source.ToStream()
    
    var results [][16]Record
    var currentBatch [16]Record
    var batchIndex int
    
    for {
        record, err := stream()
        if err == EOS {
            break
        }
        if err != nil {
            return nil, err
        }
        
        currentBatch[batchIndex] = record
        batchIndex++
        
        if batchIndex == 16 {
            results = append(results, currentBatch)
            currentBatch = [16]Record{}
            batchIndex = 0
        }
    }
    
    // Handle partial batch
    if batchIndex > 0 {
        results = append(results, currentBatch)
    }
    
    return results, nil
}
```

### Phase 3: Advanced Features (Week 5-6)

#### 3.1 GPU-Ready Operations
```go
// File: pkg/stream/gpu_arrays.go
package stream

// GPU-compatible array operations
type GPUArray[T any, N int] struct {
    Data [N]T
    Size int
}

func (ga GPUArray[T, N]) IsGPUCompatible() bool {
    // Check if T is a GPU-compatible type
    switch any(*new(T)).(type) {
    case int32, int64, float32, float64:
        return true
    default:
        return false
    }
}

// Parallel array processing (CPU preparation for GPU)
func ParallelMapArray64[T, U any](workers int, fn func(T) U, arr [64]T) [64]U {
    var result [64]U
    
    if workers <= 1 {
        for i, val := range arr {
            result[i] = fn(val)
        }
        return result
    }
    
    chunkSize := 64 / workers
    var wg sync.WaitGroup
    
    for w := 0; w < workers; w++ {
        start := w * chunkSize
        end := start + chunkSize
        if w == workers-1 {
            end = 64
        }
        
        wg.Add(1)
        go func(start, end int) {
            defer wg.Done()
            for i := start; i < end; i++ {
                result[i] = fn(arr[i])
            }
        }(start, end)
    }
    
    wg.Wait()
    return result
}
```

#### 3.2 Array-Based Aggregations
```go
// File: pkg/stream/array_aggregators.go
package stream

// Array-specific aggregators for GroupBy
func ArraySumStream4[T Numeric](name string) AggregatorSpec[Record] {
    return AggregatorSpec[Record]{
        Name: name,
        Agg: Aggregator[Record, T, T]{
            Initial: func() T { var zero T; return zero },
            Accumulate: func(acc T, record Record) T {
                if arr, exists := GetArray4[T](record, name); exists {
                    return acc + SumArray4(arr)
                }
                return acc
            },
            Finalize: func(acc T) T { return acc },
        },
    }
}

func ArrayAvgStream16[T Numeric](name string) AggregatorSpec[Record] {
    return AggregatorSpec[Record]{
        Name: name,
        Agg: Aggregator[Record, [2]float64, float64]{
            Initial: func() [2]float64 { return [2]float64{0, 0} },
            Accumulate: func(acc [2]float64, record Record) [2]float64 {
                if arr, exists := GetArray16[T](record, name); exists {
                    sum := SumArray16(arr)
                    return [2]float64{acc[0] + float64(sum), acc[1] + 16}
                }
                return acc
            },
            Finalize: func(acc [2]float64) float64 {
                if acc[1] == 0 { return 0 }
                return acc[0] / acc[1]
            },
        },
    }
}
```

### Phase 4: Migration and Optimization (Week 7-8)

#### 4.1 Backward Compatibility Layer
```go
// File: pkg/stream/migration.go
package stream

// Legacy stream support with deprecation warnings
func StreamToRecord[T any](stream Stream[T], key string, size int) Record {
    record := make(Record)
    
    switch size {
    case 4:
        if arr, err := StreamToArray4(stream); err == nil {
            record[key] = arr
        }
    case 16:
        if arr, err := StreamToArray16(stream); err == nil {
            record[key] = arr
        }
    // ... other sizes
    }
    
    return record
}

// Automatic migration helper
func MigrateStreamRecord(oldRecord Record) Record {
    newRecord := make(Record)
    
    for key, value := range oldRecord {
        switch v := value.(type) {
        case Stream[int64]:
            if arr, err := StreamToArray16(v); err == nil {
                newRecord[key] = arr
            }
        case Stream[float64]:
            if arr, err := StreamToArray16(v); err == nil {
                newRecord[key] = arr
            }
        // ... handle other stream types
        default:
            newRecord[key] = value
        }
    }
    
    return newRecord
}
```

#### 4.2 Performance Benchmarks
```go
// File: pkg/stream/array_bench_test.go
package stream

func BenchmarkArrayVsStream(b *testing.B) {
    data := make([]int64, 10000)
    for i := range data {
        data[i] = int64(i)
    }
    
    b.Run("Stream", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            stream := FromSlice(data)
            sum, _ := Sum(stream)
            _ = sum
        }
    })
    
    b.Run("Array16", func(b *testing.B) {
        // Convert to arrays of 16
        arrays := make([][16]int64, len(data)/16)
        for i := 0; i < len(arrays); i++ {
            copy(arrays[i][:], data[i*16:(i+1)*16])
        }
        
        for i := 0; i < b.N; i++ {
            var total int64
            for _, arr := range arrays {
                total += SumArray16(arr)
            }
            _ = total
        }
    })
}
```

### Phase 5: Documentation and Examples (Week 9)

#### 5.1 Update Documentation
- Update API documentation for array types
- Add migration guide from streams to arrays
- Document GPU readiness features
- Add performance comparison charts

#### 5.2 Add Examples
```go
// File: examples/array_processing/array_example.go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    // Create records with array values
    records := []stream.Record{
        {
            "id": 1,
            "scores": [4]float64{85.5, 90.0, 87.5, 92.0},
            "tags": [8]string{"math", "science", "history", "english", "", "", "", ""},
        },
        {
            "id": 2, 
            "scores": [4]float64{78.0, 85.5, 90.0, 88.5},
            "tags": [8]string{"math", "art", "music", "english", "", "", "", ""},
        },
    }
    
    // Process array values
    for _, record := range records {
        if scores, exists := stream.GetArray4[float64](record, "scores"); exists {
            avg := stream.SumArray4(scores) / 4.0
            fmt.Printf("ID %d: Average score = %.2f\n", record["id"], avg)
            
            // Parallel processing ready
            improved := stream.MapArray4(func(score float64) float64 {
                return score * 1.1 // 10% bonus
            }, scores)
            
            fmt.Printf("       Improved scores: %v\n", improved)
        }
    }
}
```

## Migration Strategy

### Backward Compatibility Approach
1. **Phase 1**: Add array types to Value interface alongside existing streams
2. **Phase 2**: Provide conversion utilities between streams and arrays  
3. **Phase 3**: Update examples and documentation to prefer arrays
4. **Phase 4**: Deprecate stream values (compile warnings)
5. **Phase 5**: Remove stream values from Value interface (breaking change)

### Migration Timeline
- **Weeks 1-4**: Implementation foundation
- **Weeks 5-6**: Advanced features and optimization
- **Weeks 7-8**: Migration tools and compatibility
- **Week 9**: Documentation and examples
- **Week 10**: Beta testing and feedback
- **Week 11-12**: Production release

## Conclusion

The proposed change to replace Stream types with fixed-size arrays in the Value interface represents a significant architectural improvement that:

1. **Enables true value semantics** for all Record contents
2. **Prepares StreamV2 for GPU acceleration** with minimal additional work
3. **Improves performance** through better memory layout and vectorization
4. **Simplifies concurrency** by eliminating non-copyable types
5. **Enables serialization** for persistence and network transfer

While there are trade-offs in terms of flexibility and fixed sizing, the benefits significantly outweigh the costs, especially considering the strategic direction toward GPU acceleration and high-performance computing.

The implementation plan provides a gradual migration path that maintains backward compatibility while moving toward a more performant and GPU-ready architecture.

**Recommendation: Proceed with implementation** following the phased approach outlined above.