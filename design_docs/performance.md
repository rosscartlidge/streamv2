# StreamV2 Performance Guide

This guide covers performance considerations and optimization strategies for StreamV2.

## Table of Contents

- [Performance Characteristics](#performance-characteristics)
- [Memory Management](#memory-management)
- [CPU Optimization](#cpu-optimization)
- [I/O Performance](#io-performance)
- [Benchmarking](#benchmarking)
- [Best Practices](#best-practices)

---

# Performance Characteristics

## Stream Operations

### Low Overhead Operations
- **FromSlice**: O(1) setup, lazy evaluation
- **Map**: O(n) with function call overhead
- **Where**: O(n) with predicate evaluation
- **Limit/Offset**: O(1) setup, early termination

### Medium Overhead Operations
- **Aggregations** (Sum, Count, Max, Min): O(n) single pass
- **Collect**: O(n) memory allocation
- **Pipe**: Minimal composition overhead

### Higher Overhead Operations
- **Sorting**: O(n log n) 
- **GroupBy**: O(n) with hash table overhead
- **Windowing**: O(n) with buffer management

## Memory Usage

### Lazy Evaluation Benefits
```go
// This doesn't allocate intermediate arrays
result := Pipe(
    Map(func(x int64) int64 { return x * x }),
    Where(func(x int64) bool { return x > 100 }),
    Take[int64](10),
)(FromSlice(largeData))
```

### Memory Allocation Patterns
- **Streaming operations**: Constant memory usage
- **Collect operations**: O(n) memory allocation
- **Windowing**: Bounded memory based on window size

---

# Memory Management

## Chunked Processing

For large datasets, process in chunks to control memory usage:

```go
func ProcessLargeDataset(data []int64, chunkSize int) (int64, error) {
    totalSum := int64(0)
    
    for i := 0; i < len(data); i += chunkSize {
        end := min(i+chunkSize, len(data))
        chunk := data[i:end]
        
        chunkSum, err := Sum(
            Map(func(x int64) int64 { return x * x })(
                FromSlice(chunk)))
        
        if err != nil {
            return 0, err
        }
        
        totalSum += chunkSum
    }
    
    return totalSum, nil
}
```

## Streaming vs Batch Processing

### Use Streaming When:
- Processing infinite or very large datasets
- Memory usage is a concern
- Real-time processing requirements
- Early termination is possible (Take, Where with low selectivity)

```go
// Streaming: constant memory
err := ForEach(func(record Record) {
    processRecord(record)
})(CSVToStream(largeFile))
```

### Use Batch When:
- Small to medium datasets
- Need random access to results
- Sorting or complex aggregations required

```go
// Batch: loads all into memory
records, err := Collect(CSVToStream(file))
sort.Slice(records, func(i, j int) bool {
    return records[i]["timestamp"].(time.Time).Before(
        records[j]["timestamp"].(time.Time))
})
```

---

# CPU Optimization

## Function Complexity

### Simple Functions (Optimized for CPU)
```go
// Simple arithmetic - stays on CPU, vectorizable
doubled := Map(func(x int64) int64 { return x * 2 })

// Simple comparisons - efficient branching
evens := Where(func(x int64) bool { return x%2 == 0 })
```

### Complex Functions (Consider Parallelization)
```go
// Math-heavy operations - may benefit from parallelization
complex := Map(func(x float64) float64 {
    return math.Sin(x) * math.Cos(x) * math.Sqrt(x+1)
})

// String processing - CPU intensive
normalized := Map(func(s string) string {
    return strings.ToLower(strings.TrimSpace(s))
})
```

## Auto-Parallelization

StreamV2 automatically chooses between sequential and parallel execution based on:

### Factors for Parallelization
- **Data size**: Large datasets benefit from parallelization
- **Function complexity**: Complex operations get parallel treatment
- **Available cores**: Respects `GOMAXPROCS`
- **Memory constraints**: Avoids parallelization if memory-bound

### Override Automatic Decisions
```go
// Force sequential processing for simple operations
result := Map(func(x int64) int64 { return x + 1 })(stream)

// Force parallel processing for complex operations
result := Parallel(8, func(x float64) float64 {
    return expensiveComputation(x)
})(stream)
```

## Pipeline Optimization

### Efficient Pipeline Construction
```go
// Good: Single pipeline, minimal intermediate allocations
result := Pipe(
    Map(func(x int64) int64 { return x * x }),
    Where(func(x int64) bool { return x > 100 }),
    Take[int64](10),
)(stream)

// Less efficient: Multiple passes
squares, _ := Collect(Map(func(x int64) int64 { return x * x })(stream))
filtered, _ := Collect(Where(func(x int64) bool { return x > 100 })(FromSlice(squares)))
result, _ := Collect(Take[int64](10)(FromSlice(filtered)))
```

### Early Termination
```go
// Take advantage of early termination
firstLargeSquare := Take[int64](1)(
    Where(func(x int64) bool { return x > 1000000 })(
        Map(func(x int64) int64 { return x * x })(
            Range(1, 10000))))
```

---

# I/O Performance

## CSV Processing

### Streaming CSV (Memory Efficient)
```go
// Process CSV without loading entire file
err := ForEach(func(record Record) {
    processRecord(record)
})(CSVToStream(file))
```

### Batch CSV (Faster for Small Files)
```go
// Load entire CSV for multiple passes
records, err := Collect(CSVToStream(file))
if err != nil {
    return err
}

// Multiple operations on same data
avgSalary, _ := Avg(Map(func(r Record) float64 {
    return r["salary"].(float64)
})(FromSlice(records)))

maxAge, _ := Max(Map(func(r Record) int64 {
    return r["age"].(int64)
})(FromSlice(records)))
```

### Custom CSV Parsing
```go
// Use custom headers for better performance
source := NewCSVSource(file).WithHeaders([]string{
    "id", "name", "salary", "department",
})
```

## JSON Processing

### JSON Lines vs JSON Array
```go
// JSON Lines: streaming, memory efficient
stream := NewJSONSource(file).WithFormat(JSONLines).ToStream()

// JSON Array: batch, faster for small files
stream := NewJSONSource(file).WithFormat(JSONArray).ToStream()
```

## File I/O Best Practices

### Buffered I/O
```go
file, err := os.Open("large-file.csv")
if err != nil {
    return err
}
defer file.Close()

// Use buffered reader for better performance
buffered := bufio.NewReader(file)
stream := CSVToStream(buffered)
```

### Concurrent I/O
```go
// Process multiple files concurrently
var wg sync.WaitGroup
results := make(chan Result, len(files))

for _, filename := range files {
    wg.Add(1)
    go func(file string) {
        defer wg.Done()
        
        result, err := processFile(file)
        results <- Result{File: file, Data: result, Error: err}
    }(filename)
}

wg.Wait()
close(results)
```

---

# Benchmarking

## Built-in Benchmarks

Run performance tests:
```bash
go test -bench=. ./pkg/stream
go test -bench=BenchmarkMap -benchmem ./pkg/stream
```

## Custom Benchmarks

```go
func BenchmarkPipelineVsMultiPass(b *testing.B) {
    data := make([]int64, 10000)
    for i := range data {
        data[i] = int64(i + 1)
    }
    
    b.Run("Pipeline", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, _ = Collect(
                Pipe(
                    Map(func(x int64) int64 { return x * x }),
                    Where(func(x int64) bool { return x > 100 }),
                )(FromSlice(data)))
        }
    })
    
    b.Run("MultiPass", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            squares, _ := Collect(Map(func(x int64) int64 { return x * x })(FromSlice(data)))
            _, _ = Collect(Where(func(x int64) bool { return x > 100 })(FromSlice(squares)))
        }
    })
}
```

## Memory Profiling

```bash
# Generate memory profile
go test -bench=BenchmarkLargeData -memprofile=mem.prof ./pkg/stream

# Analyze memory usage
go tool pprof mem.prof
```

## CPU Profiling

```bash
# Generate CPU profile
go test -bench=BenchmarkCPUIntensive -cpuprofile=cpu.prof ./pkg/stream

# Analyze CPU usage
go tool pprof cpu.prof
```

---

# Best Practices

## Do's

### ✅ Use Appropriate Data Types
```go
// Use int64 for integers, float64 for decimals
numbers := []int64{1, 2, 3, 4, 5}  // Not []int

// Use specific types when possible
type UserID int64
type Salary float64
```

### ✅ Pipeline Operations
```go
// Combine operations for efficiency
result := Pipe(
    Map(transform),
    Where(filter),
    Take[T](limit),
)(stream)
```

### ✅ Early Termination
```go
// Use Take to limit processing
firstTen := Take[int64](10)(largeStream)

// Use Where for selective processing  
relevant := Where(func(x Record) bool {
    return x["status"] == "active"
})(allRecords)
```

### ✅ Measure Performance
```go
start := time.Now()
result, err := processData(stream)
duration := time.Since(start)
log.Printf("Processed %d items in %v", len(result), duration)
```

## Don'ts

### ❌ Avoid Unnecessary Collect Operations
```go
// Bad: Collects intermediate results
squares, _ := Collect(Map(square)(stream))
filtered, _ := Collect(Where(filter)(FromSlice(squares)))

// Good: Streaming pipeline
result := Pipe(Map(square), Where(filter))(stream)
```

### ❌ Don't Ignore Errors
```go
// Bad: Ignoring errors
result, _ := Sum(stream)

// Good: Handle errors
result, err := Sum(stream)
if err != nil {
    return fmt.Errorf("calculation failed: %w", err)
}
```

### ❌ Avoid Over-Parallelization
```go
// Bad: Parallel for simple operations
result := Parallel(16, func(x int64) int64 { return x + 1 })(stream)

// Good: Let auto-parallelization decide
result := Map(func(x int64) int64 { return x + 1 })(stream)
```

## Performance Monitoring

### Metrics to Track
- **Processing rate**: items/second
- **Memory usage**: peak and average
- **CPU utilization**: per core
- **I/O throughput**: MB/second
- **Error rates**: failures/attempts

### Logging Performance
```go
func ProcessWithMetrics(stream Stream[Record]) error {
    start := time.Now()
    count := int64(0)
    
    err := ForEach(func(record Record) {
        count++
        if count%10000 == 0 {
            rate := float64(count) / time.Since(start).Seconds()
            log.Printf("Processed %d records (%.0f/sec)", count, rate)
        }
        
        processRecord(record)
    })(stream)
    
    if err != nil {
        return err
    }
    
    duration := time.Since(start)
    rate := float64(count) / duration.Seconds()
    log.Printf("Completed: %d records in %v (%.0f/sec)", count, duration, rate)
    
    return nil
}
```

---

# Scaling Guidelines

## Small Data (< 1MB)
- Use simple operations and Collect freely
- Batch processing is often faster than streaming
- Memory usage is not a concern

## Medium Data (1MB - 100MB)  
- Use streaming for memory efficiency
- Consider chunked processing
- Pipeline operations for better performance

## Large Data (> 100MB)
- Always use streaming operations
- Process in chunks
- Monitor memory usage
- Consider parallel processing for CPU-intensive operations

## Very Large Data (> 1GB)
- Mandatory streaming
- Chunked processing with progress monitoring
- Parallel processing where appropriate
- Consider external sorting/processing tools for extreme cases