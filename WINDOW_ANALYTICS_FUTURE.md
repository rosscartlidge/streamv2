# Window Analytics Future - Design Document

This document outlines possibilities, trade-offs, and design considerations for adding SQL-style window functions to streamv2. This is a forward-looking analysis to inform future implementation decisions based on real-world usage patterns.

## Overview

Window functions (analytical functions) are one of the most powerful features in modern SQL, enabling sophisticated analytics like rankings, running totals, moving averages, and comparisons with adjacent rows. Adapting these to a streaming context requires careful consideration of memory constraints, ordering requirements, and the fundamental streaming nature of the system.

## Core Challenges

### 1. Streaming vs Materialization Tension
- **SQL Approach**: Materialize entire partitions, sort globally, apply window functions
- **Streaming Reality**: Infinite streams, bounded memory, real-time processing requirements
- **Trade-off**: Choose between SQL compatibility and streaming efficiency

### 2. Ordering Requirements
- **Global ordering**: Requires full materialization (incompatible with infinite streams)
- **Partition ordering**: More feasible but still challenging with memory constraints
- **Approximate ordering**: Sliding windows with local ordering

### 3. Memory Constraints
- **Unbounded windows**: `SUM() OVER (PARTITION BY dept)` needs entire partition
- **Bounded windows**: `SUM() OVER (ORDER BY date ROWS 10 PRECEDING)` more feasible
- **Partition cardinality**: High-cardinality partitions could exhaust memory

## Proposed Architecture

### Core Components

```go
// Window specification builder
type WindowBuilder interface {
    PartitionBy(fields ...string) WindowBuilder
    OrderBy(field string, direction SortDirection) WindowBuilder
    SlidingWindow(size int) WindowBuilder
    Frame(frameType FrameType, bounds FrameBounds) WindowBuilder
    
    // Analytical functions
    RowNumber() Filter[Record, Record]
    Rank() Filter[Record, Record]
    DenseRank() Filter[Record, Record]
    Lead(field string, offset int, defaultValue any) Filter[Record, Record]
    Lag(field string, offset int, defaultValue any) Filter[Record, Record]
    Sum(field string) Filter[Record, Record]
    Avg(field string) Filter[Record, Record]
    Count() Filter[Record, Record]
    Min(field string) Filter[Record, Record]
    Max(field string) Filter[Record, Record]
    FirstValue(field string) Filter[Record, Record]
    LastValue(field string) Filter[Record, Record]
    NthValue(field string, n int) Filter[Record, Record]
    PercentRank() Filter[Record, Record]
    CumeDist() Filter[Record, Record]
    Ntile(buckets int) Filter[Record, Record]
}

// Frame types and boundaries
type FrameType int
const (
    RowsFrame FrameType = iota
    RangeFrame
)

type FrameBounds struct {
    Start FrameBoundary
    End   FrameBoundary
}

type FrameBoundary struct {
    Type   BoundaryType  // UNBOUNDED_PRECEDING, CURRENT_ROW, etc.
    Offset int          // For N PRECEDING/FOLLOWING
}
```

### Usage Examples

```go
// Example 1: Row numbering within departments
pipeline := Window().
    PartitionBy("department").
    OrderBy("salary", DESC).
    SlidingWindow(1000).  // Keep last 1000 records per partition
    RowNumber()           // Add row_number field

// Example 2: Running total with sliding window
pipeline := Window().
    PartitionBy("account_id").
    OrderBy("transaction_date", ASC).
    SlidingWindow(100).   // Last 100 transactions
    Sum("amount")         // Running sum

// Example 3: Moving average
pipeline := Window().
    PartitionBy("stock_symbol").
    OrderBy("trade_time", ASC).
    Frame(RowsFrame, FrameBounds{
        Start: {Type: PRECEDING, Offset: 10},
        End:   {Type: CURRENT_ROW},
    }).
    Avg("price")          // 10-period moving average

// Example 4: Compare with previous value
pipeline := Window().
    OrderBy("timestamp", ASC).
    SlidingWindow(10000).
    Lag("value", 1, 0)    // Previous value, default 0

// Example 5: Ranking within sliding window
pipeline := Window().
    PartitionBy("category").
    OrderBy("score", DESC).
    SlidingWindow(500).
    DenseRank()           // Dense ranking within window
```

## Implementation Strategies

### Strategy 1: Sliding Window Approach (Recommended)

**Concept**: Maintain bounded buffers per partition, process within windows.

**Pros**:
- Bounded memory usage
- Real-time processing
- Compatible with infinite streams
- Natural fit for streaming analytics

**Cons**:
- Not fully SQL-compatible (limited window scope)
- Approximate results for some functions
- Window size tuning required

**Implementation**:
```go
type PartitionWindow struct {
    partitionKey string
    records      *CircularBuffer[Record]  // Fixed-size sliding buffer
    windowSpec   WindowSpec
    sortedView   []Record                // Sorted within window
}
```

**Memory Complexity**: O(W × P) where W = window size, P = active partitions

### Strategy 2: Micro-batch Processing

**Concept**: Collect small batches, process with full SQL semantics, emit results.

**Pros**:
- Full SQL compatibility within batches
- Predictable memory usage
- Easier to implement complex functions

**Cons**:
- Introduces latency (batch collection time)
- May miss cross-batch patterns
- Requires careful batch boundary handling

**Implementation**:
```go
type MicroBatch struct {
    records     []Record
    partitions  map[string][]Record
    windowFuncs []WindowFunction
}
```

### Strategy 3: Hybrid Approach

**Concept**: Different strategies for different function types.

**Simple Functions** (ROW_NUMBER, LAG/LEAD): Sliding window
**Aggregate Functions** (SUM, AVG): Maintain running state
**Complex Functions** (RANK, percentiles): Micro-batch or approximation

**Pros**:
- Optimal strategy per function type
- Balances performance and accuracy
- Incremental implementation possible

**Cons**:
- Complex implementation
- Inconsistent behavior across functions
- Harder to compose multiple functions

## Function Categories & Trade-offs

### Category 1: Positional Functions
**Functions**: ROW_NUMBER, RANK, DENSE_RANK, NTILE
**Challenge**: Require ordering within partition
**Streaming Solution**: Sliding window with internal sorting
**Accuracy**: Good within window bounds
**Memory**: O(window_size × partitions)

### Category 2: Offset Functions  
**Functions**: LAG, LEAD, FIRST_VALUE, LAST_VALUE, NTH_VALUE
**Challenge**: Need access to other rows
**Streaming Solution**: Circular buffer per partition
**Accuracy**: Perfect within window/offset range
**Memory**: O(max_offset × partitions)

### Category 3: Aggregate Functions
**Functions**: SUM, AVG, COUNT, MIN, MAX
**Challenge**: Frame boundaries and running state
**Streaming Solution**: Incremental computation with sliding window
**Accuracy**: Perfect for sliding windows
**Memory**: O(1) for running aggregates, O(frame_size) for sliding frames

### Category 4: Statistical Functions
**Functions**: PERCENT_RANK, CUME_DIST, PERCENTILE
**Challenge**: Need entire population for accurate calculation
**Streaming Solution**: Approximate algorithms (quantiles, sketches) or micro-batches
**Accuracy**: Approximate
**Memory**: O(sketch_size) or O(batch_size)

## Design Decisions & Trade-offs

### Memory Management

**Option A: Fixed Global Limit**
```go
WindowConfig{
    MaxTotalMemory: 1GB,      // Global limit across all windows
    EvictionPolicy: LRU,      // How to handle limit exceeded
}
```

**Option B: Per-Window Configuration**
```go
Window().
    PartitionBy("department").
    MaxPartitions(1000).      // Limit active partitions
    WindowSize(100).          // Records per partition
    RowNumber()
```

**Option C: Dynamic Adaptation**
- Monitor memory usage
- Automatically reduce window sizes under pressure
- Emit warnings when accuracy is compromised

### Ordering Guarantees

**Pre-sorted Input**: Assume input is ordered, cheaper to process
```go
Window().
    AssumeOrderedBy("timestamp").  // Input contract
    PartitionBy("user_id").
    SlidingWindow(100).
    RowNumber()
```

**Internal Sorting**: Sort within window, more expensive but flexible
```go
Window().
    PartitionBy("user_id").
    OrderBy("score", DESC).        // Sort within window
    SlidingWindow(100).
    Rank()
```

**Hybrid**: Pre-sorted for frame, internal sort for analytics
```go
Window().
    AssumeOrderedBy("timestamp").  // Frame ordering
    PartitionBy("user_id").
    OrderBy("score", DESC).        // Analytics ordering
    SlidingWindow(100).
    Rank()
```

### Result Emission Strategies

**Immediate Emission**: Emit results as soon as computed
- Pro: Low latency
- Con: Results may change as window slides

**Stabilized Emission**: Emit only when results won't change
- Pro: Stable results
- Con: Higher latency, requires larger buffers

**Checkpoint Emission**: Emit at regular intervals or checkpoints
- Pro: Predictable behavior
- Con: May miss real-time requirements

## Integration with Existing Architecture

### With Join Operations
```go
// Window functions before join
pipeline := Pipe(
    Window().PartitionBy("department").RowNumber(),
    InnerJoin(profileStream, "emp_id", "emp_id"),
)(employeeStream)

// Window functions after join
pipeline := Pipe(
    InnerJoin(profileStream, "emp_id", "emp_id"),
    Window().PartitionBy("department").Avg("salary"),
)(employeeStream)
```

### With Flatten Operations
```go
// Flatten then window
pipeline := Pipe(
    DotFlatten("."),
    Window().PartitionBy("user.department").Sum("transaction.amount"),
)(complexRecordStream)
```

### With Parallel Processing
Window functions often require ordering, which conflicts with parallelization:
- **Solution**: Partition-level parallelization
- **Trade-off**: Limited parallelism by partition count

## Performance Characteristics

### Time Complexity
- **Per record**: O(log W) for sorted insertion in sliding window
- **Window function**: O(1) to O(W) depending on function
- **Overall**: O(N log W) where N = stream size, W = window size

### Space Complexity
- **Per partition**: O(W) where W = window size
- **Total**: O(W × P) where P = active partitions
- **Worst case**: High-cardinality partitioning could exhaust memory

### Throughput Impact
- **Sliding window**: 10-50% overhead depending on window size
- **Complex functions**: 50-100% overhead for ranking/percentiles
- **Memory pressure**: Severe degradation when limits exceeded

## Implementation Phases

### Phase 1: Foundation (Minimal Viable Product)
- Basic window infrastructure
- Sliding window implementation
- ROW_NUMBER, LAG, LEAD functions
- Single partition key support

### Phase 2: Core Analytics
- SUM, AVG, COUNT, MIN, MAX functions
- Multiple partition keys
- Frame specifications (ROWS BETWEEN)
- Memory management

### Phase 3: Advanced Functions
- RANK, DENSE_RANK functions  
- FIRST_VALUE, LAST_VALUE, NTH_VALUE
- Performance optimizations
- Integration testing

### Phase 4: Statistical Functions
- PERCENT_RANK, CUME_DIST
- NTILE functions
- Approximate algorithms for large datasets
- Production hardening

## Production Considerations

### Monitoring & Observability
- Window memory usage per partition
- Function computation latencies  
- Accuracy metrics for approximate functions
- Partition cardinality tracking

### Configuration Management
- Window size auto-tuning based on available memory
- Function-specific optimizations
- Graceful degradation under resource pressure

### Error Handling
- Partition memory limit exceeded
- Unsupported function combinations
- Ordering constraint violations
- Recovery from partial failures

## Alternative Approaches

### External Processing
- Offload window functions to external systems (Apache Flink, etc.)
- Pro: Leverage existing optimized implementations
- Con: Breaks streaming model, adds operational complexity

### Approximation Algorithms
- Use sketches and probabilistic data structures
- Pro: Bounded memory, good performance
- Con: Approximate results, limited function support

### Materialized Views
- Pre-compute common window function results
- Pro: Fast queries, exact results
- Con: Storage overhead, freshness issues

## Recommendation

Start with **Strategy 1 (Sliding Window)** for these reasons:

1. **Natural fit**: Aligns with streaming philosophy
2. **Bounded resources**: Predictable memory usage
3. **Real-time**: Low latency results
4. **Incremental**: Can add functions progressively
5. **User control**: Window size tuning allows performance/accuracy trade-offs

Focus initial implementation on:
- ROW_NUMBER (simplest, high value)
- LAG/LEAD (useful for time series)
- SUM/AVG with sliding frames (common analytics)

Avoid initially:
- RANK/DENSE_RANK (complex, requires sorting)
- Statistical functions (approximation challenges)
- Complex frame specifications (implementation complexity)

## Future Research Areas

- **Adaptive window sizing** based on data patterns and memory pressure
- **Distributed window functions** across multiple nodes
- **Approximate window functions** with error bounds
- **Window function optimization** through query planning
- **Integration with time-series databases** for specialized workloads

---

This document should evolve based on real-world usage patterns and performance requirements. The key insight is that streaming window functions require fundamentally different trade-offs than traditional SQL, and the design should embrace rather than fight these constraints.