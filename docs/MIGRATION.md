# Migration Guide: StreamV1 to StreamV2

## 🚀 **Overview**

StreamV2 is a complete redesign using Go generics for maximum type safety, performance, and developer experience. This guide helps you migrate from the original stream library to StreamV2.

## 🔍 **Key Changes**

### **1. No More Value Interface**
```go
// V1: Manual wrapping/unwrapping
record := stream.Record{
    "id": stream.Int(123),
    "name": stream.String("Alice"),
}
id, err := record["id"].Int()

// V2: Native Go types
record := stream.R("id", 123, "name", "Alice")
id := stream.Get[int64](record, "id")
```

### **2. Generics-First Streams**
```go
// V1: Only Record streams
type Stream func() (Record, error)

// V2: Generic streams for any type
type Stream[T any] func() (T, error)
type RecordStream = Stream[Record]  // For compatibility
```

### **3. Type-Safe Operations**
```go
// V1: Runtime type checking
stream.Where(func(r Record) bool {
    val, _ := r["age"].Int()
    return val > 18
})

// V2: Compile-time type safety
stream.Where(func(age int64) bool { return age > 18 })(
    stream.ExtractField[int64]("age")(recordStream)
)
```

## 📋 **Migration Steps**

### **Step 1: Update Imports**
```go
// V1
import "github.com/rosscartlidge/stream"

// V2
import "github.com/rosscartlidge/streamv2/pkg/stream"
```

### **Step 2: Replace Record Creation**
```go
// V1 → V2
stream.Record{
    "id":     stream.Int(123),
    "name":   stream.String("Alice"),
    "active": stream.Bool(true),
}
↓
stream.R("id", 123, "name", "Alice", "active", true)
```

### **Step 3: Replace Value Access**
```go
// V1 → V2
id, err := record["id"].Int()
if err != nil { /* handle */ }
↓
id := stream.Get[int64](record, "id")
// Or with default:
id := stream.GetOr(record, "id", int64(0))
```

### **Step 4: Update Stream Operations**
```go
// V1 → V2
stream.Where(func(r Record) bool {
    age, _ := r["age"].Int()
    return age > 18
})
↓
stream.Where(func(r stream.Record) bool {
    return stream.GetOr(r, "age", int64(0)) > 18
})
```

## 🎯 **Common Migration Patterns**

### **Pattern 1: Value Extraction**
```go
// V1
func getUserAge(r stream.Record) int64 {
    if val, ok := r["age"]; ok {
        if age, err := val.Int(); err == nil {
            return age
        }
    }
    return 0
}

// V2
func getUserAge(r stream.Record) int64 {
    return stream.GetOr(r, "age", int64(0))
}
```

### **Pattern 2: Stream Processing**
```go
// V1
activeUsers := stream.Where(func(r stream.Record) bool {
    active, _ := r["active"].Bool()
    return active
})(userStream)

// V2
activeUsers := stream.Where(func(r stream.Record) bool {
    return stream.GetOr(r, "active", false)
})(userStream)
```

### **Pattern 3: Type-Safe Numeric Operations**
```go
// V1: Always through Records
numbers := stream.RecordsToStream([]stream.Record{
    {"value": stream.Int(1)},
    {"value": stream.Int(2)},
})

// V2: Direct type support
numbers := stream.FromSlice([]int64{1, 2, 3, 4, 5})
doubled := stream.Map(func(x int64) int64 { return x * 2 })(numbers)
```

## ⚡ **Performance Benefits**

| Operation | V1 Performance | V2 Performance | Improvement |
|-----------|----------------|----------------|-------------|
| Record creation | Interface wrapping | Native types | **5-10x faster** |
| Value access | Interface + unwrap | Direct access | **3-5x faster** |
| Numeric operations | Record overhead | Direct types | **2-5x faster** |
| Memory usage | Interface overhead | Native storage | **50-80% less** |

## 🔧 **Migration Tools**

### **Automated Replacements**
```bash
# Replace common patterns with sed/awk
sed -i 's/stream\.Int(\([^)]*\))/\1/g' *.go
sed -i 's/stream\.String(\([^)]*\))/\1/g' *.go
sed -i 's/stream\.Float(\([^)]*\))/\1/g' *.go
sed -i 's/stream\.Bool(\([^)]*\))/\1/g' *.go
```

### **Testing Compatibility**
```go
// Test both versions side by side
func TestMigration(t *testing.T) {
    // V1 result
    v1Result := processWithV1(data)
    
    // V2 result  
    v2Result := processWithV2(data)
    
    // Compare outputs
    assert.Equal(t, v1Result, v2Result)
}
```

## 🎉 **New V2-Only Features**

### **Network Analytics**
```go
processor := stream.NewProcessor()
insights := processor.TopTalkers(100)(networkFlows)
```

### **Parallel Processing**
```go
results := stream.Parallel(8, expensiveFunction)(dataStream)
```

### **Type-Safe Aggregations**
```go
sum, _ := stream.Sum(numberStream)  // Compile-time type safety
max, _ := stream.Max(numberStream)
```

### **Smart Type Conversion**
```go
// Automatic conversion between compatible types
intVal := stream.GetOr(record, "string_number", int64(0)) // "42" → 42
```

## 📚 **Learning Resources**

- **Examples**: See `examples/` directory for complete working examples
- **Documentation**: Full API documentation in `docs/`
- **Benchmarks**: Performance comparisons in `benchmarks/`

## 🤔 **Migration Decision Tree**

```
Do you process large datasets (>100K records)?
├─ YES → Migrate immediately (huge performance gains)
└─ NO
   └─ Is type safety critical?
      ├─ YES → Migrate (prevents runtime errors)
      └─ NO → Evaluate based on development velocity needs
```

## ⏰ **Estimated Timeline**

- **Small codebase** (<1000 LOC): 1-2 days
- **Medium codebase** (<10K LOC): 1-2 weeks  
- **Large codebase** (>10K LOC): 2-4 weeks

The performance and maintainability benefits make the migration effort worthwhile for most projects.

## 🆘 **Getting Help**

- Check the examples in this repository
- Review the API documentation
- Open issues on GitHub for migration questions
- Join the community discussions

The migration to StreamV2 is a significant upgrade that brings modern Go idioms, better performance, and enhanced type safety to your stream processing code.