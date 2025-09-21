# StreamV2 API Documentation

Complete reference for all exported functions in the StreamV2 package.

## Table of Contents

- [Core Types](#core-types)
- [Stream Constructors](#stream-constructors)
- [Core Filters](#core-filters)
- [Aggregators](#aggregators)
- [I/O Operations](#io-operations)
- [Advanced Windowing](#advanced-windowing)
- [Executor Architecture](#executor-architecture)

---

# Core Types

## Stream[T]
```go
type Stream[T any] func() (T, error)
```
The fundamental stream type. A function that returns the next element and an error. Returns `EOS` error when the stream is exhausted.

## Record
```go
type Record map[string]any
```
A record represents a row of data with named fields. Used for CSV, JSON, and structured data processing.

## Filter[T, U]
```go
type Filter[T, U any] func(Stream[T]) Stream[U]
```
A function that transforms one stream into another. The building block for all stream transformations.

---

# Stream Constructors

Functions for creating streams from various data sources.

## FromSlice
```go
func FromSlice[T any](slice []T) Stream[T]
```
Creates a stream from a slice of elements.

**Example:**
```go
stream := FromSlice([]int64{1, 2, 3, 4, 5})
```

## FromSliceAny
```go
func FromSliceAny[T any](slice []T) Stream[any]
```
Creates a stream of `any` type from a slice, useful for mixed-type processing.

**Example:**
```go
mixed := []any{"hello", 42, true}
stream := FromSliceAny(mixed)
```

## FromChannel
```go
func FromChannel[T any](ch <-chan T) Stream[T]
```
Creates a stream from a Go channel.

**Example:**
```go
ch := make(chan int64, 3)
ch <- 1
ch <- 2
ch <- 3
close(ch)
stream := FromChannel(ch)
```

## FromChannelAny
```go
func FromChannelAny[T any](ch <-chan T) Stream[any]
```
Creates a stream of `any` type from a channel.

## Generate
```go
func Generate[T any](generator func() T) Stream[T]
```
Creates an infinite stream using a generator function.

**Example:**
```go
counter := 0
stream := Generate(func() int64 {
    counter++
    return int64(counter)
})
```

## GenerateAny
```go
func GenerateAny[T any](generator func() T) Stream[any]
```
Creates an infinite stream of `any` type using a generator function.

## Range
```go
func Range(start, end int64) Stream[int64]
```
Creates a stream of integers from start (inclusive) to end (exclusive).

**Example:**
```go
stream := Range(1, 6) // Produces: 1, 2, 3, 4, 5
```

## Once
```go
func Once[T any](value T) Stream[T]
```
Creates a stream containing a single element.

**Example:**
```go
stream := Once("hello")
```

## OnceAny
```go
func OnceAny[T any](value T) Stream[any]
```
Creates a stream of `any` type containing a single element.

---

# Core Filters

Functions for transforming and filtering streams.

## Map
```go
func Map[T, U any](fn func(T) U) Filter[T, U]
```
Transforms each element in the stream using the provided function.

**Example:**
```go
doubled := Map(func(x int64) int64 { return x * 2 })
```

## Where
```go
func Where[T any](predicate func(T) bool) Filter[T, T]
```
Filters stream elements, keeping only those where the predicate returns true.

**Example:**
```go
evens := Where(func(x int64) bool { return x%2 == 0 })
```

## Take
```go
func Take[T any](n int) Filter[T, T]
```
Takes the first n elements from the stream.

**Example:**
```go
firstThree := Take[int64](3)
```

## Skip
```go
func Skip[T any](n int) Filter[T, T]
```
Skips the first n elements in the stream.

**Example:**
```go
skipTwo := Skip[int64](2)
```

## Pipe
```go
func Pipe[T, U, V any](f1 Filter[T, U], f2 Filter[U, V]) Filter[T, V]
func Pipe3[T, U, V, W any](f1 Filter[T, U], f2 Filter[U, V], f3 Filter[V, W]) Filter[T, W]
```
Combines multiple filters in sequence.

**Example:**
```go
pipeline := Pipe(
    Map(func(x int64) int64 { return x * x }),
    Where(func(x int64) bool { return x > 10 }),
)
```

## Select
```go
func Select[T, U any](selector func(T) U) Filter[T, U]
```
Alias for Map - selects/transforms elements.

## Update
```go
func Update[T any](updater func(T) T) Filter[T, T]
```
Updates elements in-place using the provided function.

**Example:**
```go
incrementer := Update(func(x int64) int64 { return x + 1 })
```

## ExtractField
```go
func ExtractField[T any](fieldName string) Filter[Record, T]
```
Extracts a specific field from Record objects.

**Example:**
```go
names := ExtractField[string]("name")
```

## Tee
```go
func Tee[T any](sideEffect func(T)) Filter[T, T]
```
Applies a side effect to each element without modifying the stream.

**Example:**
```go
logged := Tee(func(x int64) { fmt.Printf("Processing: %d\n", x) })
```

## FlatMap
```go
func FlatMap[T, U any](fn func(T) Stream[U]) Filter[T, U]
```
Maps each element to a stream and flattens the result.

## DotFlatten
```go
func DotFlatten[T any]() Filter[Record, Record]
```
Flattens nested records using dot notation for field names.

## CrossFlatten
```go
func CrossFlatten[T any]() Filter[Record, Record]
```
Flattens nested records using cross product expansion.

---

# Aggregators

Functions that consume streams and produce single values or collections.

## Sum
```go
func Sum[T Numeric](stream Stream[T]) (T, error)
```
Calculates the sum of all elements in the stream.

**Example:**
```go
total, err := Sum(FromSlice([]int64{1, 2, 3, 4, 5}))
// total = 15
```

## Count
```go
func Count[T any](stream Stream[T]) (int64, error)
```
Counts the number of elements in the stream.

**Example:**
```go
count, err := Count(FromSlice([]string{"a", "b", "c"}))
// count = 3
```

## Max
```go
func Max[T Ordered](stream Stream[T]) (T, error)
```
Finds the maximum element in the stream.

**Example:**
```go
max, err := Max(FromSlice([]int64{3, 1, 4, 1, 5}))
// max = 5
```

## Min
```go
func Min[T Ordered](stream Stream[T]) (T, error)
```
Finds the minimum element in the stream.

**Example:**
```go
min, err := Min(FromSlice([]int64{3, 1, 4, 1, 5}))
// min = 1
```

## Avg
```go
func Avg[T Numeric](stream Stream[T]) (float64, error)
```
Calculates the average of all elements in the stream.

**Example:**
```go
avg, err := Avg(FromSlice([]int64{1, 2, 3, 4, 5}))
// avg = 3.0
```

## Collect
```go
func Collect[T any](stream Stream[T]) ([]T, error)
```
Collects all elements from the stream into a slice.

**Example:**
```go
items, err := Collect(FromSlice([]int64{1, 2, 3}))
// items = [1, 2, 3]
```

## ForEach
```go
func ForEach[T any](action func(T)) func(Stream[T]) error
```
Applies an action to each element in the stream.

**Example:**
```go
err := ForEach(func(x int64) {
    fmt.Println(x)
})(FromSlice([]int64{1, 2, 3}))
```

## Custom Aggregators

### SumAggregator
```go
func SumAggregator[T any, U Numeric](extractor func(T) U) Aggregator[T, U]
```
Creates a sum aggregator that extracts numeric values using a function.

### MinAggregator
```go
func MinAggregator[T any, U Ordered](extractor func(T) U) Aggregator[T, U]
```
Creates a min aggregator that extracts comparable values.

### MaxAggregator
```go
func MaxAggregator[T any, U Ordered](extractor func(T) U) Aggregator[T, U]
```
Creates a max aggregator that extracts comparable values.

### AvgAggregator
```go
func AvgAggregator[T any, U Numeric](extractor func(T) U) Aggregator[T, float64]
```
Creates an average aggregator that extracts numeric values.

### CountAggregator
```go
func CountAggregator[T any]() Aggregator[T, int64]
```
Creates a count aggregator.

## Multiple Aggregations

### AggregateMultiple
```go
func AggregateMultiple[T, U, V any](stream Stream[T], agg1 Aggregator[T, U], agg2 Aggregator[T, V]) (U, V, error)
```
Runs two aggregators on the same stream simultaneously.

### Aggregates
```go
func Aggregates[T any](stream Stream[T], specs ...AggregatorSpec[T]) (Record, error)
```
Runs multiple named aggregators and returns results in a Record.

**Example:**
```go
results, err := Aggregates(stream,
    SumStream[int64]("total"),
    CountStream[int64]("count"),
    AvgStream[int64]("average"),
)
// results["total"], results["count"], results["average"]
```

### GroupBy
```go
func GroupBy(keyFields []string, aggregators ...AggregatorSpec[Record]) Filter[Record, Record]
```
Groups records by specified fields and applies aggregations to each group.

**Example:**
```go
// Basic grouping (only key fields)
grouped := GroupBy([]string{"department"})(users)

// Grouping with aggregations (explicit count required)
groupedWithStats := GroupBy([]string{"department"}, 
    CountField("count", "name"),
    SumField[int64]("total_salary", "salary"),
    AvgField[int64]("avg_salary", "salary"),
)(users)
```

### Aggregator Specifications

#### For Stream Types
```go
func SumStream[T Numeric](name string) AggregatorSpec[T]
func CountStream[T any](name string) AggregatorSpec[T]
func AvgStream[T Numeric](name string) AggregatorSpec[T]
func MinStream[T Comparable](name string) AggregatorSpec[T]
func MaxStream[T Comparable](name string) AggregatorSpec[T]
```

#### For Record Fields
```go
func SumField[T Numeric](name, fieldName string) AggregatorSpec[Record]
func CountField(name, fieldName string) AggregatorSpec[Record]
func AvgField[T Numeric](name, fieldName string) AggregatorSpec[Record]
func MinField[T Comparable](name, fieldName string) AggregatorSpec[Record]
func MaxField[T Comparable](name, fieldName string) AggregatorSpec[Record]
```

### Low-Level Aggregators

#### Generic Aggregators
```go
func SumAggregator[I any, T Numeric](extract func(I) T) Aggregator[I, T, T]
func CountAggregator[I any]() Aggregator[I, int64, int64]
func AvgAggregator[I any, T Numeric](extract func(I) T) Aggregator[I, [2]float64, float64]
func MinAggregator[I any, T Comparable](extract func(I) T) Aggregator[I, *T, T]
func MaxAggregator[I any, T Comparable](extract func(I) T) Aggregator[I, *T, T]
```

#### Field-Specific Aggregators
```go
func SumAggregatorField[T Numeric](fieldName string) Aggregator[Record, T, T]
func CountAggregatorField(fieldName string) Aggregator[Record, int64, int64]
func AvgAggregatorField[T Numeric](fieldName string) Aggregator[Record, [2]float64, float64]
func MinAggregatorField[T Comparable](fieldName string) Aggregator[Record, *T, T]
func MaxAggregatorField[T Comparable](fieldName string) Aggregator[Record, *T, T]
```

---

# I/O Operations

Functions for reading and writing data in various formats.

## CSV Operations

### NewCSVSource
```go
func NewCSVSource(reader io.Reader) *CSVSource
```
Creates a CSV source from an io.Reader.

**Methods:**
- `WithHeaders(headers []string) *CSVSource` - Set custom headers
- `WithoutHeaders() *CSVSource` - Disable header parsing
- `ToStream() Stream[Record]` - Convert to record stream

### CSVToStream
```go
func CSVToStream(reader io.Reader) Stream[Record]
```
Convenience function to create a record stream from CSV data.

**Example:**
```go
csvData := "name,age\nAlice,30\nBob,25"
stream := CSVToStream(strings.NewReader(csvData))
```

### NewCSVSink
```go
func NewCSVSink(writer io.Writer) *CSVSink
```
Creates a CSV sink for writing record streams.

**Methods:**
- `WithHeaders(headers []string) *CSVSink` - Set output headers
- `WriteStream(stream Stream[Record]) error` - Write stream to CSV
- `WriteRecords(records []Record) error` - Write record slice

### StreamToCSV
```go
func StreamToCSV(stream Stream[Record], writer io.Writer) error
```
Convenience function to write a record stream as CSV.

### File Operations
```go
func CSVToStreamFromFile(filename string) (Stream[Record], error)
func StreamToCSVFile(stream Stream[Record], filename string) error
```

## TSV Operations

Similar to CSV operations but for Tab-Separated Values:
- `NewTSVSource(reader io.Reader) *CSVSource`
- `TSVToStream(reader io.Reader) Stream[Record]`
- `NewTSVSink(writer io.Writer) *CSVSink`
- `StreamToTSV(stream Stream[Record], writer io.Writer) error`

## JSON Operations

### NewJSONSource
```go
func NewJSONSource(reader io.Reader) *JSONSource
```
Creates a JSON source from an io.Reader.

**Methods:**
- `WithFormat(format JSONFormat) *JSONSource` - Set JSON format
- `ToStream() Stream[Record]` - Convert to record stream

**JSON Formats:**
- `JSONLines` - One JSON object per line (default)
- `JSONArray` - Single array of JSON objects

### JSONToStream
```go
func JSONToStream(reader io.Reader) Stream[Record]
```
Convenience function to create a record stream from JSON data.

### NewJSONSink
```go
func NewJSONSink(writer io.Writer) *JSONSink
```
Creates a JSON sink for writing record streams.

**Methods:**
- `WithFormat(format JSONFormat) *JSONSink` - Set output format
- `WithPrettyPrint() *JSONSink` - Enable pretty printing
- `WriteStream(stream Stream[Record]) error` - Write stream to JSON
- `WriteRecords(records []Record) error` - Write record slice

### StreamToJSON
```go
func StreamToJSON(stream Stream[Record], writer io.Writer) error
```
Convenience function to write a record stream as JSON.

## Protocol Buffer Operations

### NewProtobufSource
```go
func NewProtobufSource(reader io.Reader, messageDesc protoreflect.MessageDescriptor) *ProtobufSource
```

### NewProtobufSink
```go
func NewProtobufSink(writer io.Writer, messageDesc protoreflect.MessageDescriptor) *ProtobufSink
```

---

# Advanced Windowing

Functions for time-based and count-based windowing operations.

## Window Builder

### Window
```go
func Window[T any]() *WindowBuilder[T]
```
Creates a new window builder for configuring advanced windowing.

**Methods:**
- `TriggerOnCount(count int) *WindowBuilder[T]` - Fire on element count
- `TriggerOnTime(duration time.Duration) *WindowBuilder[T]` - Fire on time
- `TriggerOnProcessingTime(interval time.Duration) *WindowBuilder[T]` - Fire on processing time
- `AllowLateness(lateness time.Duration) *WindowBuilder[T]` - Configure late data handling
- `AccumulationMode() *WindowBuilder[T]` - Accumulate late data
- `DiscardingMode() *WindowBuilder[T]` - Discard late data
- `Apply() Filter[T, Stream[T]]` - Create the windowing filter

**Example:**
```go
windowFilter := Window[int64]().
    TriggerOnCount(10).
    TriggerOnTime(5*time.Second).
    Apply()
```

## Session Windows

### SessionWindow
```go
func SessionWindow[T any](timeout time.Duration, activityDetector ActivityDetector[T]) Filter[T, Stream[T]]
```
Creates activity-based session windows.

**Example:**
```go
sessions := SessionWindow(30*time.Second, func(event Event) bool {
    return event.Type == "login" || event.Type == "purchase"
})
```

## Triggers

### NewAdvancedCountTrigger
```go
func NewAdvancedCountTrigger[T any](threshold int) *AdvancedCountTrigger[T]
```
Creates a trigger that fires when element count reaches threshold.

### NewAdvancedTimeTrigger
```go
func NewAdvancedTimeTrigger[T any](duration time.Duration) *AdvancedTimeTrigger[T]
```
Creates a trigger that fires after a time duration.

### NewAdvancedProcessingTimeTrigger
```go
func NewAdvancedProcessingTimeTrigger[T any](interval time.Duration) *AdvancedProcessingTimeTrigger[T]
```
Creates a trigger that fires at regular processing time intervals.

## Basic Windowing Functions

### CountWindow
```go
func CountWindow[T any](size int) Filter[T, Stream[T]]
```
Creates fixed-size windows based on element count.

### TimeWindow
```go
func TimeWindow[T any](duration time.Duration) Filter[T, Stream[T]]
```
Creates time-based windows.

### SlidingCountWindow
```go
func SlidingCountWindow[T any](size, step int) Filter[T, Stream[T]]
```
Creates sliding windows based on count.

## Streaming Aggregators

### StreamingSum
```go
func StreamingSum[T Numeric]() Filter[T, T]
```
Produces running sum of elements.

### StreamingCount
```go
func StreamingCount[T any]() Filter[T, int64]
```
Produces running count of elements.

### StreamingAvg
```go
func StreamingAvg[T Numeric]() Filter[T, float64]
```
Produces running average of elements.

### StreamingMax
```go
func StreamingMax[T Ordered]() Filter[T, T]
```
Produces running maximum of elements.

### StreamingMin
```go
func StreamingMin[T Ordered]() Filter[T, T]
```
Produces running minimum of elements.

### StreamingStats
```go
func StreamingStats[T Numeric]() Filter[T, Stats]
```
Produces running statistics (count, sum, avg, min, max).

---

# Executor Architecture

The executor architecture provides transparent acceleration for stream operations.

## ExecutorManager

### NewExecutorManager
```go
func NewExecutorManager() *ExecutorManager
```
Creates a new executor manager with automatic backend detection.

**Methods:**
- `AddExecutor(executor Executor)` - Add an executor
- `SelectBest(op Operation, ctx ExecutionContext) Executor` - Choose best executor

## Operation Builders

### NewMapOperation
```go
func NewMapOperation(dataType reflect.Type, size int64, complexity int) Operation
```
Creates metadata for a Map operation.

### NewFilterOperation
```go
func NewFilterOperation(dataType reflect.Type, size int64, complexity int) Operation
```
Creates metadata for a Filter operation.

### NewParallelOperation
```go
func NewParallelOperation(dataType reflect.Type, size int64, workers int, complexity int) Operation
```
Creates metadata for a Parallel operation.

## CPU Executor

### NewCPUExecutor
```go
func NewCPUExecutor() *CPUExecutor
```
Creates a new CPU executor for traditional processing.

## GPU Executor

### NewGPUExecutor
```go
func NewGPUExecutor() *GPUExecutor
```
Creates a new GPU executor if CUDA hardware is available (returns nil if not available).

---

# Type Constraints

## Numeric
```go
type Numeric interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64
}
```
Constraint for numeric types that support arithmetic operations.

## Ordered
```go
type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64 | ~string
}
```
Constraint for types that support comparison operations.

---

# Error Handling

## EOS Error
```go
var EOS = errors.New("end of stream")
```
Special error returned when a stream is exhausted. Always check for this error when iterating streams manually.

**Example:**
```go
for {
    item, err := stream()
    if err != nil {
        if err == EOS {
            break // Normal end of stream
        }
        return err // Actual error
    }
    // Process item
}
```

---

# Best Practices

## Error Handling
Always check errors from stream operations:
```go
result, err := Sum(stream)
if err != nil {
    return fmt.Errorf("failed to sum stream: %w", err)
}
```

## Pipeline Construction
Use `Pipe()` for readable multi-step operations:
```go
pipeline := Pipe(
    Map(func(x int64) int64 { return x * x }),
    Where(func(x int64) bool { return x > 25 }),
)
```

## Memory Management
Process large datasets in chunks:
```go
chunkSize := 1000
for i := 0; i < len(data); i += chunkSize {
    chunk := data[i:min(i+chunkSize, len(data))]
    result, _ := Sum(FromSlice(chunk))
    // Process chunk result
}
```

## Type Safety
Use safe type assertions with Records:
```go
name, ok := record["name"].(string)
if !ok {
    return errors.New("name field is not a string")
}
```

## Performance
- Use appropriate data types (int64, float64 for numbers)
- Avoid unnecessary type conversions
- Consider using streaming operations for large datasets
- Profile critical paths

---

# Examples

See the [StreamV2 Codelab](../STREAMV2_CODELAB.md) for comprehensive examples and tutorials.

Additional examples are available in the [examples directory](../stream_examples/).