# GPU/CUDA Integration Design for StreamV2

## Executive Summary

This document analyzes how GPU/CUDA acceleration will integrate with StreamV2's existing architecture, particularly in relation to the proposed Array Values change. The goal is to design a transparent GPU acceleration system that maintains API compatibility while providing significant performance improvements for suitable workloads.

## Current StreamV2 Architecture Analysis

### Existing CPU Architecture
```go
// Current Stream processing model
type Stream[T any] func() (T, error)
type Filter[T, U any] func(Stream[T]) Stream[U]

// Example usage
numbers := stream.FromSlice([]int64{1, 2, 3, 4, 5})
result := stream.Map(func(x int64) int64 { return x * x })(numbers)
squares, _ := stream.Collect(result)
```

### Value Constraint Impact
```go
// Current Value constraint
type Value interface {
    ~int | ~int64 | ~float32 | ~float64 | ~bool | ~string |
    Record | Stream[T] // <- Key decision point for GPU integration
}
```

## GPU Computing Fundamentals

### CUDA Programming Model
```cuda
// CUDA kernel operates on arrays, not individual elements
__global__ void square_kernel(float* input, float* output, int n) {
    int idx = blockIdx.x * blockDim.x + threadIdx.x;
    if (idx < n) {
        output[idx] = input[idx] * input[idx];
    }
}
```

### Key GPU Requirements
1. **Batch Processing**: GPUs need arrays/batches of data, not individual elements
2. **Contiguous Memory**: Data must be in contiguous memory blocks
3. **Type Compatibility**: Limited to numeric types (int32, int64, float32, float64)
4. **No Function Pointers**: Cannot transfer Go functions to GPU
5. **Synchronous Execution**: GPU kernels complete before returning results

## GPU Integration Architecture Options

### Option 1: Stream-Based with Auto-Batching
```go
// Transparent GPU acceleration without API changes
func Map[T, U GPUCompatible](fn func(T) U) Filter[T, U] {
    return func(input Stream[T]) Stream[U] {
        // Collect batch internally
        batch, err := collectBatch(input, optimalBatchSize())
        if err != nil || !shouldUseGPU(batch, fn) {
            return cpuMap(fn)(input) // Fallback to CPU
        }
        
        // Execute on GPU
        result := gpuMap(fn, batch)
        return FromSlice(result)
    }
}

// Usage remains identical
numbers := stream.FromSlice(data)
result := stream.Map(func(x float64) float64 { 
    return math.Sin(x) * math.Cos(x) 
})(numbers)
```

**Pros:**
- ✅ Zero API changes required
- ✅ Transparent acceleration
- ✅ Automatic CPU fallback

**Cons:**
- ❌ Hidden batching complexity
- ❌ Function serialization challenges
- ❌ Unpredictable memory behavior

### Option 2: Array-Based GPU-First Design
```go
// Array-based Value constraint enables direct GPU support
type Value interface {
    ~int32 | ~int64 | ~float32 | ~float64 | ~bool | ~string |
    Record |
    [256]int32 | [256]int64 | [256]float32 | [256]float64 | // GPU-ready arrays
    [1024]int32 | [1024]int64 | [1024]float32 | [1024]float64
}

// GPU-aware operations
func GPUMap[T, U GPUNumeric, N int](fn func(T) U) Filter[[N]T, [N]U] {
    return func(input Stream[[N]T]) Stream[[N]U] {
        return func() ([N]U, error) {
            batch, err := input()
            if err != nil {
                return [N]U{}, err
            }
            
            // Direct GPU execution
            return gpuMapKernel(fn, batch), nil
        }
    }
}
```

**Pros:**
- ✅ Direct GPU memory mapping
- ✅ Predictable performance
- ✅ Type-safe GPU operations

**Cons:**
- ❌ Requires Value constraint change
- ❌ Fixed batch sizes
- ❌ API modifications needed

### Option 3: Hybrid CPU/GPU with Smart Dispatch
```go
// Dual execution paths with automatic selection
type Executor interface {
    CanAccelerate(operation Operation, dataSize int) bool
    Execute(operation Operation, data interface{}) (interface{}, error)
}

type GPUExecutor struct {
    deviceID int
    memoryMB int
}

func (e GPUExecutor) CanAccelerate(op Operation, size int) bool {
    return op.IsGPUCompatible() && 
           size >= e.MinBatchSize() && 
           e.HasAvailableMemory(size)
}

// Auto-dispatching Map
func Map[T, U any](fn func(T) U) Filter[T, U] {
    return func(input Stream[T]) Stream[U] {
        // Peek at data characteristics
        sample, _ := peekStream(input, 100)
        
        executor := selectExecutor(fn, sample)
        return executor.Map(fn)(input)
    }
}
```

## Recommended GPU Integration Design

### Architecture: Hybrid Array-Based with Transparent Dispatch

#### Core GPU Integration Components

#### 1. GPU-Compatible Type System
```go
// GPU-compatible numeric types
type GPUNumeric interface {
    ~int32 | ~int64 | ~float32 | ~float64
}

// GPU-ready array types (fixed sizes for optimal memory layout)
type GPUArray[T GPUNumeric] interface {
    [64]T | [128]T | [256]T | [512]T | [1024]T
}

// Updated Value constraint
type Value interface {
    // CPU types
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64 | ~bool | ~string | time.Time |
    Record |
    
    // GPU-ready arrays
    [64]int32 | [128]int32 | [256]int32 | [512]int32 | [1024]int32 |
    [64]int64 | [128]int64 | [256]int64 | [512]int64 | [1024]int64 |
    [64]float32 | [128]float32 | [256]float32 | [512]float32 | [1024]float32 |
    [64]float64 | [128]float64 | [256]float64 | [512]float64 | [1024]float64
}
```

#### 2. Transparent GPU Operations
```go
// GPU-aware Map with automatic dispatch
func Map[T, U any](fn func(T) U) Filter[T, U] {
    return func(input Stream[T]) Stream[U] {
        // Check if GPU acceleration is beneficial
        if isGPUCompatible[T, U]() && shouldUseGPU(input) {
            return gpuMap(fn)(input)
        }
        return cpuMap(fn)(input)
    }
}

// GPU-specific implementation (hidden from users)
func gpuMap[T, U GPUNumeric](fn func(T) U) Filter[T, U] {
    return func(input Stream[T]) Stream[U] {
        // Auto-batch incoming stream
        batches := autoBatch(input, optimalGPUBatchSize[T]())
        
        return func() (U, error) {
            batch, err := batches()
            if err != nil {
                return *new(U), err
            }
            
            // Execute GPU kernel
            result := executeGPUKernel(fn, batch)
            return unbatch(result)() // Convert back to stream
        }
    }
}
```

#### 3. Function Serialization for GPU
```go
// Supported GPU operations (compiled to CUDA kernels)
type GPUOperation interface {
    // Mathematical operations
    Sin | Cos | Tan | Sqrt | Exp | Log |
    // Arithmetic operations  
    Add | Subtract | Multiply | Divide | Power |
    // Comparison operations
    GreaterThan | LessThan | Equal | NotEqual
}

// Function analysis and GPU compilation
func analyzeFunction[T, U GPUNumeric](fn func(T) U) (GPUOperation, bool) {
    // Use runtime reflection or AST analysis to determine operation type
    // Return corresponding GPU operation if supported
}

// Example GPU kernel mapping
var gpuKernels = map[GPUOperation]string{
    Sin: `
        __global__ void sin_kernel(float* input, float* output, int n) {
            int idx = blockIdx.x * blockDim.x + threadIdx.x;
            if (idx < n) output[idx] = sinf(input[idx]);
        }
    `,
    Multiply: `
        __global__ void multiply_kernel(float* input, float* output, float factor, int n) {
            int idx = blockIdx.x * blockDim.x + threadIdx.x;
            if (idx < n) output[idx] = input[idx] * factor;
        }
    `,
}
```

#### 4. Memory Management
```go
// GPU memory pool for efficient allocation
type GPUMemoryPool struct {
    deviceID int
    pools    map[reflect.Type]*sync.Pool
}

func (p *GPUMemoryPool) Get[T GPUNumeric](size int) (*GPUBuffer[T], error) {
    pool := p.pools[reflect.TypeOf(*new(T))]
    if pool == nil {
        pool = &sync.Pool{
            New: func() interface{} {
                return allocateGPUBuffer[T](size)
            },
        }
        p.pools[reflect.TypeOf(*new(T))] = pool
    }
    
    return pool.Get().(*GPUBuffer[T]), nil
}

// CUDA memory wrapper
type GPUBuffer[T GPUNumeric] struct {
    devicePtr unsafe.Pointer
    hostPtr   []T
    size      int
    deviceID  int
}

func (b *GPUBuffer[T]) CopyToDevice() error {
    return cudaMemcpy(b.devicePtr, unsafe.Pointer(&b.hostPtr[0]), 
                     b.size*int(unsafe.Sizeof(*new(T))), cudaMemcpyHostToDevice)
}

func (b *GPUBuffer[T]) CopyToHost() error {
    return cudaMemcpy(unsafe.Pointer(&b.hostPtr[0]), b.devicePtr,
                     b.size*int(unsafe.Sizeof(*new(T))), cudaMemcpyDeviceToHost)
}
```

## Detailed API Design

### 1. GPU-Aware Stream Operations

#### Map Operation
```go
// Current CPU-only Map
func Map[T, U any](fn func(T) U) Filter[T, U]

// New GPU-aware Map (same signature, transparent acceleration)
func Map[T, U any](fn func(T) U) Filter[T, U] {
    return func(input Stream[T]) Stream[U] {
        if gpuOp, isGPUCompatible := analyzeFunction(fn); isGPUCompatible {
            return createGPUStream(gpuOp, input)
        }
        return createCPUStream(fn, input)
    }
}

// Usage (unchanged)
numbers := stream.FromSlice([]float64{1, 2, 3, 4, 5})
result := stream.Map(func(x float64) float64 { 
    return math.Sin(x) 
})(numbers) // Automatically uses GPU if beneficial
```

#### Array-Based Operations
```go
// Explicit GPU array operations for maximum performance
func GPUMapArray[T, U GPUNumeric, N int](fn func(T) U) func([N]T) [N]U {
    return func(input [N]T) [N]U {
        // Direct GPU kernel execution
        return executeKernel(compileKernel(fn), input)
    }
}

// Batch processing with arrays
func ProcessArrayBatches[T, U GPUNumeric](
    batches Stream[[256]T], 
    operation func(T) U,
) Stream[[256]U] {
    return func() ([256]U, error) {
        batch, err := batches()
        if err != nil {
            return [256]U{}, err
        }
        
        return GPUMapArray(operation)(batch), nil
    }
}
```

### 2. Aggregation Operations

#### GPU-Accelerated Aggregations
```go
// Sum with GPU acceleration for large datasets
func Sum[T Numeric](input Stream[T]) (T, error) {
    // Collect data and determine execution strategy
    data, err := Collect(input)
    if err != nil {
        return *new(T), err
    }
    
    if len(data) > 10000 && isGPUNumeric[T]() {
        return gpuSum(data), nil
    }
    
    return cpuSum(data), nil
}

// GPU sum implementation using reduction kernels
func gpuSum[T GPUNumeric](data []T) T {
    // Pad to optimal size and transfer to GPU
    padded := padToOptimalSize(data)
    buffer := allocateGPUBuffer(padded)
    defer buffer.Release()
    
    // Execute parallel reduction kernel
    return executeReductionKernel[T](buffer, AddOperation)
}
```

### 3. I/O Integration

#### GPU-Ready Data Loading
```go
// Fast TSV to GPU arrays
func FastTSVToGPUArrays[N int](reader io.Reader) Stream[[N]float64] {
    source := NewFastTSVSource(reader)
    return BatchToArrays[N](source.ToStream())
}

// Batch conversion utility
func BatchToArrays[N int](input Stream[Record]) Stream[[N]float64] {
    return func() ([N]float64, error) {
        var batch [N]float64
        
        for i := 0; i < N; i++ {
            record, err := input()
            if err != nil {
                return batch, err
            }
            
            // Extract numeric field for GPU processing
            if val, exists := GetNumeric(record, "value"); exists {
                batch[i] = val
            }
        }
        
        return batch, nil
    }
}
```

### 4. Configuration and Control

#### GPU Settings and Fallback
```go
// Global GPU configuration
type GPUConfig struct {
    Enabled          bool
    DeviceID         int
    MinBatchSize     int
    MemoryLimitMB    int
    FallbackOnError  bool
}

var DefaultGPUConfig = GPUConfig{
    Enabled:         true,
    DeviceID:        0,
    MinBatchSize:    1024,
    MemoryLimitMB:   1024,
    FallbackOnError: true,
}

// Per-operation GPU control
func MapWithGPU[T, U GPUNumeric](fn func(T) U, forceGPU bool) Filter[T, U] {
    return func(input Stream[T]) Stream[U] {
        if forceGPU {
            return gpuMap(fn)(input)
        }
        return Map(fn)(input) // Auto-dispatch
    }
}

// GPU availability check
func IsGPUAvailable() bool {
    return cudaGetDeviceCount() > 0
}

func GetGPUInfo() []GPUDevice {
    // Return available GPU devices and capabilities
}
```

## Implementation Strategy

### Phase 1: Foundation (Weeks 1-3)
1. **CUDA/CGO Integration Setup**
   - Create CUDA C wrapper functions
   - Set up CGO bindings for memory management
   - Implement basic kernel execution framework

2. **GPU Memory Management**
   - Implement GPUBuffer type with automatic memory management
   - Create memory pool for efficient allocation/deallocation
   - Add host-device memory transfer utilities

3. **Function Analysis Framework**
   - Create function signature analysis for GPU compatibility
   - Implement basic mathematical operation detection
   - Build kernel template system

### Phase 2: Core Operations (Weeks 4-6)
1. **GPU-Aware Map Implementation**
   - Add transparent GPU dispatch to existing Map function
   - Implement auto-batching for stream-to-array conversion
   - Add fallback mechanisms for unsupported operations

2. **Array-Based Operations**
   - Implement direct array processing functions
   - Add fixed-size array types to Value constraint
   - Create conversion utilities between streams and arrays

3. **Basic Aggregations**
   - Implement GPU-accelerated Sum, Count, Min, Max
   - Add parallel reduction kernels
   - Ensure numerical precision matches CPU results

### Phase 3: Advanced Features (Weeks 7-9)
1. **Complex Mathematical Operations**
   - Add support for trigonometric functions (sin, cos, tan)
   - Implement exponential and logarithmic functions
   - Add custom mathematical expression compilation

2. **Multi-GPU Support**
   - Implement device selection and load balancing
   - Add cross-device memory management
   - Create parallel execution across multiple GPUs

3. **Performance Optimization**
   - Implement kernel fusion for chained operations
   - Add memory coalescing optimization
   - Create adaptive batch sizing based on GPU memory

### Phase 4: Integration and Testing (Weeks 10-12)
1. **Integration with Existing Operations**
   - Update GroupBy for GPU-accelerated aggregations
   - Modify I/O operations for GPU-ready data formats
   - Ensure compatibility with existing examples

2. **Comprehensive Testing**
   - Add GPU/CPU result comparison tests
   - Implement performance benchmarks
   - Create fallback testing for non-GPU environments

3. **Documentation and Examples**
   - Create GPU programming guide for StreamV2
   - Add performance optimization recommendations
   - Document GPU memory requirements and limitations

## Performance Expectations

### Expected Speedups
```go
// Mathematical operations (sin, cos, sqrt, etc.)
// CPU:  50MB/s   -> GPU: 500-2000MB/s   (10-40x speedup)

// Large aggregations (sum, average over millions of elements)
// CPU:  100MB/s  -> GPU: 800-3000MB/s   (8-30x speedup)

// Complex analytics (multiple operations per element)
// CPU:  30MB/s   -> GPU: 400-1500MB/s   (13-50x speedup)
```

### Memory Requirements
- **Minimum**: 2GB GPU memory for basic operations
- **Recommended**: 8GB+ for optimal batch sizes
- **Optimal**: 16GB+ for very large datasets

### GPU Hardware Support
- **Minimum**: CUDA Compute Capability 6.0 (Pascal architecture)
- **Recommended**: RTX 30xx/40xx series or Tesla V100/A100
- **Automatic fallback**: CPU execution when GPU unavailable

## Impact on Value Constraint Decision

### Strong Recommendation: Adopt Array-Based Values

The GPU integration analysis strongly supports the Array Values architecture change:

#### 1. **Memory Layout Compatibility**
```go
// GPU kernels require contiguous memory - arrays provide this naturally
[256]float64  // Perfect GPU memory layout
Stream[float64]  // Requires expensive batching and copying
```

#### 2. **Type Safety for GPU Operations**
```go
// Arrays enable compile-time GPU compatibility checking
func GPUProcess[T GPUNumeric, N int](data [N]T) [N]T // ✅ Type-safe

// Streams require runtime checking
func GPUProcess(data Stream[any]) Stream[any] // ❌ Runtime type checking needed
```

#### 3. **Performance Predictability**
```go
// Array-based: Predictable GPU memory usage
records := []Record{{"data": [256]float64{...}}}  // Known size: 256 * 8 bytes

// Stream-based: Unpredictable memory requirements  
records := []Record{{"data": someStream}}  // Unknown size until consumption
```

#### 4. **Simplified API**
```go
// With arrays: Direct GPU operation
result := GPUMapArray(mathFunction)(data)

// With streams: Hidden complexity
result := Map(mathFunction)(data)  // Secretly batches, transfers, unbatches
```

### Recommended Value Constraint
```go
type Value interface {
    // CPU-optimized types
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64 | ~bool | ~string | time.Time |
    Record |
    
    // GPU-optimized array types
    [64]int32 | [128]int32 | [256]int32 | [512]int32 | [1024]int32 |
    [64]int64 | [128]int64 | [256]int64 | [512]int64 | [1024]int64 |
    [64]float32 | [128]float32 | [256]float32 | [512]float32 | [1024]float32 |
    [64]float64 | [128]float64 | [256]float64 | [512]float64 | [1024]float64
}
```

## Conclusion

The GPU/CUDA integration design demonstrates that:

1. **Array-based Values are essential** for optimal GPU performance
2. **Transparent acceleration is possible** while maintaining API compatibility
3. **Significant performance gains** (10-50x) are achievable for suitable workloads
4. **The architecture supports both CPU and GPU** execution paths seamlessly

**Strong Recommendation**: Proceed with the Array Values architecture change as it provides the foundation for efficient GPU acceleration while maintaining the flexibility and ease-of-use that makes StreamV2 successful.

The proposed hybrid approach enables:
- **Transparent GPU acceleration** for existing code
- **Explicit GPU operations** for maximum performance
- **Automatic fallback** to CPU when GPU is unavailable
- **Type-safe GPU programming** through compile-time constraints

This positions StreamV2 as a truly modern, GPU-ready stream processing library that can compete with and exceed the performance of traditional distributed systems for many use cases.