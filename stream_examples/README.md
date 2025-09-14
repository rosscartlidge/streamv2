# Stream Examples

This directory contains comprehensive examples demonstrating the key features of the streamv2 library.

## Examples

### Flatten Examples
**Location**: `flatten_examples/main.go`

Demonstrates the powerful data transformation capabilities of DotFlatten and CrossFlatten filters:

- **DotFlatten**: Converts nested records to flat structures using dot notation
- **CrossFlatten**: Expands stream fields into multiple records using cartesian products
- **Selective Processing**: Both functions support field-specific operations
- **Pipeline Integration**: Shows how to combine flatten operations with other filters
- **Real-world Scenarios**: E-commerce analytics and complex data transformations

**Run Example**:
```bash
go run stream_examples/flatten_examples/main.go
```

### Join Examples  
**Location**: `join_examples/main.go`

Demonstrates all four SQL-style join operations with hash join algorithm:

- **InnerJoin**: Only matching records from both streams
- **LeftJoin**: All left records + matching right records  
- **RightJoin**: All right records + matching left records
- **FullJoin**: All records from both streams
- **Field Conflict Resolution**: Configurable prefixes for handling name collisions
- **Performance**: Large dataset examples showing O(R + L) time complexity
- **Pipeline Integration**: Multiple joins and aggregation workflows

**Run Example**:
```bash
go run stream_examples/join_examples/main.go
```

## Key Features Demonstrated

### Data Transformation
- Nested record flattening with customizable separators
- Stream field expansion with cartesian products
- Selective field processing for performance optimization

### Data Joining
- All SQL join types with hash join algorithm
- Field conflict resolution strategies
- Multi-table join pipelines
- Performance considerations for large datasets

### Functional Composition
- Pipeline construction with `Pipe()`, `Pipe3()`
- Filter chaining for complex data workflows
- Integration with `FlatMap()` for stream-level processing

### Real-world Applications
- E-commerce order processing
- Customer analytics
- Employee data enrichment
- Performance optimization patterns

## Performance Notes

### Flatten Operations
- **DotFlatten**: O(N) time, processes records lazily
- **CrossFlatten**: O(N × M) where M is cartesian product size
- **Memory**: Minimal memory usage, preserves streaming nature

### Join Operations
- **Hash Join**: O(R + L) time, O(R) space complexity
- **Right Stream**: Collected into memory (must be finite)
- **Left Stream**: Processed lazily, can be infinite
- **Optimal Use**: Large left stream with smaller right lookup table

## Getting Started

1. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

2. **Run Examples**:
   ```bash
   # Flatten operations
   go run stream_examples/flatten_examples/main.go
   
   # Join operations  
   go run stream_examples/join_examples/main.go
   ```

3. **Experiment**: Modify the examples to test different scenarios and data structures.

## Architecture Integration

These examples showcase how the new Flatten and Join filters integrate seamlessly with the existing streamv2 architecture:

- **Type Safety**: Full generic type support
- **Functional Design**: Composable filter functions
- **Error Handling**: Consistent error patterns
- **Memory Efficiency**: Lazy evaluation and streaming processing
- **Performance**: Optimized algorithms for real-world usage

## Next Steps

After exploring these examples, consider:

1. **Custom Filters**: Create domain-specific filters using the same patterns
2. **Pipeline Optimization**: Experiment with filter ordering for performance
3. **Integration**: Combine with existing I/O operations (CSV, JSON, etc.)
4. **Testing**: Use patterns from the test files for your own data processing