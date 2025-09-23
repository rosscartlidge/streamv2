# StreamV2 Advanced Features Codelab

Welcome to the StreamV2 Advanced Features codelab! This tutorial covers powerful, specialized operations that enable sophisticated stream processing patterns. You'll learn dynamic partitioning, stream composition, and advanced data flow control.

## What You'll Learn

- 🚀 **Dynamic Stream Splitting** - Partition data streams on-the-fly with `Split()`
- 🔄 **Stream Composition** - Build complex pipelines with `Chain()` and `Pipe()`
- ⚡ **Parallel Processing** - Explicit parallelization with `Parallel()`
- 🎯 **Context Control** - Resource management with `WithContext()`
- 🔧 **Advanced Patterns** - Real-world streaming architectures

## Prerequisites

- Complete the [main StreamV2 codelab](STREAMV2_CODELAB.md)
- Understanding of Go contexts and goroutines
- Go 1.21+ installed

## Chapter Navigation

**Fundamentals:** [Dynamic Splitting](#chapter-1-dynamic-stream-splitting-with-split) • [Stream Composition](#chapter-2-advanced-stream-composition) • [Parallel Processing](#chapter-3-explicit-parallel-processing)

**Advanced:** [Context Control](#chapter-4-context-and-resource-management) • [Real-World Patterns](#chapter-5-real-world-streaming-patterns)

---

# Chapter 1: Dynamic Stream Splitting with Split()

The `Split()` function is one of StreamV2's most powerful features - it dynamically partitions data streams based on key fields, creating separate substreams for each unique key combination.

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Printf("=== Dynamic Stream Splitting ===\n\n")

    // Simulate real-time log data from different services
    logs := []stream.Record{
        stream.NewRecord().String("service", "auth").String("level", "INFO").String("message", "User login").Build(),
        stream.NewRecord().String("service", "payment").String("level", "ERROR").String("message", "Payment failed").Build(),
        stream.NewRecord().String("service", "auth").String("level", "WARN").String("message", "Failed login attempt").Build(),
        stream.NewRecord().String("service", "inventory").String("level", "INFO").String("message", "Stock updated").Build(),
        stream.NewRecord().String("service", "payment").String("level", "INFO").String("message", "Payment processed").Build(),
        stream.NewRecord().String("service", "auth").String("level", "ERROR").String("message", "Authentication error").Build(),
    }

    fmt.Printf("Processing %d log entries from multiple services\n\n", len(logs))

    // 1. Split by service - create separate streams for each service
    fmt.Printf("=== Splitting by Service ===\n")
    logStream := stream.FromSlice(logs)
    serviceStreams := stream.Split([]string{"service"})(logStream)

    serviceCount := 0
    for {
        serviceStream, err := serviceStreams()
        if err != nil {
            break // No more services
        }
        serviceCount++

        // Process each service's logs separately
        serviceLogs, _ := stream.Collect(serviceStream)
        if len(serviceLogs) > 0 {
            service := stream.GetOr(serviceLogs[0], "service", "unknown")
            fmt.Printf("Service '%s': %d log entries\n", service, len(serviceLogs))

            // Analyze log levels for this service
            levels := make(map[string]int)
            for _, log := range serviceLogs {
                level := stream.GetOr(log, "level", "")
                levels[level]++
            }

            for level, count := range levels {
                fmt.Printf("  - %s: %d\n", level, count)
            }
            fmt.Println()
        }
    }

    fmt.Printf("Total services found: %d\n\n", serviceCount)

    // 2. Multi-key splitting - split by both service AND level
    fmt.Printf("=== Multi-Key Splitting (Service + Level) ===\n")
    logStream2 := stream.FromSlice(logs)
    multiKeyStreams := stream.Split([]string{"service", "level"})(logStream2)

    combinationCount := 0
    for {
        combinationStream, err := multiKeyStreams()
        if err != nil {
            break
        }
        combinationCount++

        combinationLogs, _ := stream.Collect(combinationStream)
        if len(combinationLogs) > 0 {
            firstLog := combinationLogs[0]
            service := stream.GetOr(firstLog, "service", "unknown")
            level := stream.GetOr(firstLog, "level", "unknown")
            fmt.Printf("%s/%s: %d entries\n", service, level, len(combinationLogs))
        }
    }

    fmt.Printf("\nTotal service/level combinations: %d\n", combinationCount)
}
```

**What this demonstrates:**
- **Dynamic partitioning** - Unknown number of services discovered at runtime
- **Multi-key splitting** - Partitioning by multiple fields creates fine-grained streams
- **Real-time processing** - Each substream can be processed independently
- **Zero configuration** - No need to pre-declare partitions or keys

---

# Chapter 2: Advanced Stream Composition

Build sophisticated processing pipelines using `Chain()` for same-type operations and `Pipe()` for type transformations.

```go
package main

import (
    "fmt"
    "strings"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Printf("=== Advanced Stream Composition ===\n\n")

    // E-commerce transaction data
    transactions := []stream.Record{
        stream.NewRecord().String("id", "tx1").String("product", "laptop").Float("amount", 1299.99).String("status", "pending").Build(),
        stream.NewRecord().String("id", "tx2").String("product", "mouse").Float("amount", 29.99).String("status", "completed").Build(),
        stream.NewRecord().String("id", "tx3").String("product", "keyboard").Float("amount", 89.99).String("status", "failed").Build(),
        stream.NewRecord().String("id", "tx4").String("product", "monitor").Float("amount", 449.99).String("status", "completed").Build(),
        stream.NewRecord().String("id", "tx5").String("product", "laptop").Float("amount", 1199.99).String("status", "completed").Build(),
    }

    // 1. Chain - Multiple same-type transformations
    fmt.Printf("=== Using Chain for Record Transformations ===\n")

    processedTransactions, _ := stream.Collect(
        stream.Chain(
            // Filter only completed transactions
            stream.Where(func(tx stream.Record) bool {
                return stream.GetOr(tx, "status", "") == "completed"
            }),
            // Add tax calculation
            stream.Update(func(tx stream.Record) stream.Record {
                amount := stream.GetOr(tx, "amount", 0.0)
                tax := amount * 0.08 // 8% tax
                return stream.SetField(tx, "tax", tax)
            }),
            // Add total with tax
            stream.Update(func(tx stream.Record) stream.Record {
                amount := stream.GetOr(tx, "amount", 0.0)
                tax := stream.GetOr(tx, "tax", 0.0)
                total := amount + tax
                return stream.SetField(tx, "total", total)
            }),
            // Normalize product names
            stream.Update(func(tx stream.Record) stream.Record {
                product := stream.GetOr(tx, "product", "")
                normalized := strings.Title(strings.ToLower(product))
                return stream.SetField(tx, "product", normalized)
            }),
        )(stream.FromSlice(transactions)))

    fmt.Printf("Processed %d completed transactions:\n", len(processedTransactions))
    for _, tx := range processedTransactions {
        id := stream.GetOr(tx, "id", "")
        product := stream.GetOr(tx, "product", "")
        amount := stream.GetOr(tx, "amount", 0.0)
        tax := stream.GetOr(tx, "tax", 0.0)
        total := stream.GetOr(tx, "total", 0.0)
        fmt.Printf("- %s: %s $%.2f (tax: $%.2f, total: $%.2f)\n", id, product, amount, tax, total)
    }

    // 2. Pipe - Type transformations in sequence
    fmt.Printf("\n=== Using Pipe for Type Transformations ===\n")

    productSales, _ := stream.Collect(
        stream.Pipe(
            // Record -> Extract product names
            stream.Map(func(tx stream.Record) string {
                return stream.GetOr(tx, "product", "")
            }),
            // string -> Filter non-empty
            stream.Where(func(product string) bool {
                return len(product) > 0
            }),
        )(stream.FromSlice(transactions)))

    fmt.Printf("Products sold: %v\n", productSales)

    // 3. Complex pipeline combining multiple techniques
    fmt.Printf("\n=== Complex Pipeline Example ===\n")

    summary, _ := stream.Aggregates(
        stream.Chain(
            stream.Where(func(tx stream.Record) bool {
                return stream.GetOr(tx, "status", "") == "completed"
            }),
            stream.Update(func(tx stream.Record) stream.Record {
                amount := stream.GetOr(tx, "amount", 0.0)
                if amount >= 1000 {
                    return stream.SetField(tx, "tier", "premium")
                }
                return stream.SetField(tx, "tier", "standard")
            }),
        )(stream.FromSlice(transactions)),
        stream.CountStream[stream.Record]("total_completed"),
        stream.SumField[float64]("total_revenue", "amount"),
        stream.AvgField[float64]("avg_transaction", "amount"),
    )

    fmt.Printf("Sales Summary:\n")
    fmt.Printf("- Total completed: %d\n", stream.GetOr(summary, "total_completed", int64(0)))
    fmt.Printf("- Total revenue: $%.2f\n", stream.GetOr(summary, "total_revenue", 0.0))
    fmt.Printf("- Average transaction: $%.2f\n", stream.GetOr(summary, "avg_transaction", 0.0))
}
```

**Key Composition Patterns:**
- **Chain()** - Same-type transformations (Record → Record)
- **Pipe()** - Type transformations (Record → string → filtered)
- **Nested operations** - Combining filters, maps, and aggregations
- **Business logic** - Tax calculations, product normalization, tier classification

---

# Chapter 3: Explicit Parallel Processing

Control parallelization explicitly for computationally intensive operations.

```go
package main

import (
    "fmt"
    "math"
    "time"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Printf("=== Explicit Parallel Processing ===\n\n")

    // Generate computationally expensive data
    data := make([]float64, 100)
    for i := range data {
        data[i] = float64(i + 1)
    }

    // Expensive computation function
    expensiveComputation := func(x float64) float64 {
        // Simulate expensive work
        result := 0.0
        for i := 0; i < 10000; i++ {
            result += math.Sin(x * float64(i)) * math.Cos(x * float64(i))
        }
        return result
    }

    // 1. Sequential processing
    fmt.Printf("=== Sequential Processing ===\n")
    start := time.Now()

    sequential, _ := stream.Collect(
        stream.Map(expensiveComputation)(
            stream.FromSlice(data)))

    sequentialTime := time.Since(start)
    fmt.Printf("Sequential: %d results in %v\n", len(sequential), sequentialTime)

    // 2. Parallel processing with 4 workers
    fmt.Printf("\n=== Parallel Processing (4 workers) ===\n")
    start = time.Now()

    parallel4, _ := stream.Collect(
        stream.Parallel(4, expensiveComputation)(
            stream.FromSlice(data)))

    parallel4Time := time.Since(start)
    fmt.Printf("Parallel (4 workers): %d results in %v\n", len(parallel4), parallel4Time)

    // 3. Parallel processing with 8 workers
    fmt.Printf("\n=== Parallel Processing (8 workers) ===\n")
    start = time.Now()

    parallel8, _ := stream.Collect(
        stream.Parallel(8, expensiveComputation)(
            stream.FromSlice(data)))

    parallel8Time := time.Since(start)
    fmt.Printf("Parallel (8 workers): %d results in %v\n", len(parallel8), parallel8Time)

    // Performance comparison
    fmt.Printf("\n=== Performance Analysis ===\n")
    fmt.Printf("Sequential:     %v\n", sequentialTime)
    fmt.Printf("4 workers:      %v (%.1fx speedup)\n", parallel4Time, float64(sequentialTime)/float64(parallel4Time))
    fmt.Printf("8 workers:      %v (%.1fx speedup)\n", parallel8Time, float64(sequentialTime)/float64(parallel8Time))

    // 4. Parallel processing in a pipeline
    fmt.Printf("\n=== Parallel Processing in Pipeline ===\n")

    complexPipeline, _ := stream.Collect(
        stream.Chain(
            stream.Where(func(x float64) bool { return x > 10 }),
            stream.Parallel(4, func(x float64) float64 {
                return expensiveComputation(x) * 2
            }),
            stream.Where(func(x float64) bool { return !math.IsNaN(x) }),
        )(stream.FromSlice(data)))

    fmt.Printf("Complex pipeline: %d results\n", len(complexPipeline))
}
```

**Parallel Processing Guidelines:**
- **CPU-bound operations** benefit most from parallelization
- **Worker count** should match available CPU cores
- **Small datasets** may not benefit due to overhead
- **Memory usage** increases with worker count

---

# Chapter 4: Context and Resource Management

Use `WithContext()` for proper resource management, cancellation, and timeouts.

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Printf("=== Context and Resource Management ===\n\n")

    // Simulate slow data source
    slowData := stream.Generate(func() (int, error) {
        time.Sleep(100 * time.Millisecond) // Simulate slow operation
        return 42, nil
    })

    // 1. Context with timeout
    fmt.Printf("=== Processing with Timeout ===\n")
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    contextStream := stream.WithContext(ctx, slowData)

    start := time.Now()
    results, err := stream.Collect(
        stream.Take(10)(contextStream)) // Try to take 10 items

    elapsed := time.Since(start)

    if err != nil {
        fmt.Printf("Operation timed out after %v\n", elapsed)
        fmt.Printf("Collected %d items before timeout\n", len(results))
    } else {
        fmt.Printf("Completed in %v with %d results\n", elapsed, len(results))
    }

    // 2. Manual cancellation
    fmt.Printf("\n=== Manual Cancellation ===\n")
    ctx2, cancel2 := context.WithCancel(context.Background())

    // Cancel after 300ms
    go func() {
        time.Sleep(300 * time.Millisecond)
        fmt.Printf("Triggering cancellation...\n")
        cancel2()
    }()

    contextStream2 := stream.WithContext(ctx2, slowData)

    start = time.Now()
    results2, err2 := stream.Collect(
        stream.Take(20)(contextStream2))

    elapsed2 := time.Since(start)

    if err2 != nil {
        fmt.Printf("Operation cancelled after %v\n", elapsed2)
        fmt.Printf("Collected %d items before cancellation\n", len(results2))
    }

    // 3. Context propagation in complex pipeline
    fmt.Printf("\n=== Context in Complex Pipeline ===\n")

    // Create data that respects context
    dataWithContext := func() stream.Stream[stream.Record] {
        return stream.Generate(func() (stream.Record, error) {
            time.Sleep(50 * time.Millisecond)
            return stream.NewRecord().
                Int("id", time.Now().UnixNano()).
                String("status", "active").
                Build(), nil
        })
    }

    ctx3, cancel3 := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel3()

    pipeline := stream.Chain(
        stream.WithContext(ctx3, dataWithContext()),
        stream.Where(func(r stream.Record) bool {
            return stream.GetOr(r, "status", "") == "active"
        }),
        stream.Update(func(r stream.Record) stream.Record {
            return stream.SetField(r, "processed_at", time.Now().Unix())
        }),
    )

    pipelineResults, pipelineErr := stream.Collect(
        stream.Take(15)(pipeline))

    if pipelineErr != nil {
        fmt.Printf("Pipeline stopped due to context: %d records processed\n", len(pipelineResults))
    } else {
        fmt.Printf("Pipeline completed: %d records processed\n", len(pipelineResults))
    }
}
```

**Context Best Practices:**
- **Always set timeouts** for external operations
- **Propagate context** through the entire pipeline
- **Handle cancellation gracefully** with proper cleanup
- **Use context for resource limits** and backpressure

---

# Chapter 5: Real-World Streaming Patterns

Combine all advanced features to solve real-world streaming challenges.

```go
package main

import (
    "context"
    "fmt"
    "math/rand"
    "time"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Printf("=== Real-World Streaming Pattern: Multi-Tenant Log Processor ===\n\n")

    // Simulate multi-tenant log stream
    logGenerator := func() stream.Stream[stream.Record] {
        tenants := []string{"tenant-a", "tenant-b", "tenant-c"}
        services := []string{"api", "database", "cache", "queue"}
        levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}

        return stream.Generate(func() (stream.Record, error) {
            tenant := tenants[rand.Intn(len(tenants))]
            service := services[rand.Intn(len(services))]
            level := levels[rand.Intn(len(levels))]

            return stream.NewRecord().
                String("tenant", tenant).
                String("service", service).
                String("level", level).
                String("message", fmt.Sprintf("Log from %s/%s", tenant, service)).
                Int("timestamp", time.Now().UnixNano()).
                Build(), nil
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Step 1: Create context-aware log stream
    logStream := stream.WithContext(ctx, logGenerator())

    // Step 2: Split by tenant for isolated processing
    fmt.Printf("=== Step 1: Splitting by Tenant ===\n")
    tenantStreams := stream.Split([]string{"tenant"})(
        stream.Take(50)(logStream)) // Process first 50 logs

    tenantProcessors := make(map[string][]stream.Record)

    // Process each tenant separately
    tenantCount := 0
    for {
        tenantStream, err := tenantStreams()
        if err != nil {
            break
        }
        tenantCount++

        // Collect tenant logs
        tenantLogs, _ := stream.Collect(tenantStream)
        if len(tenantLogs) > 0 {
            tenant := stream.GetOr(tenantLogs[0], "tenant", "unknown")
            tenantProcessors[tenant] = tenantLogs

            fmt.Printf("Tenant '%s': %d logs\n", tenant, len(tenantLogs))

            // Per-tenant processing: Split by service within tenant
            serviceStreams := stream.Split([]string{"service"})(
                stream.FromSlice(tenantLogs))

            serviceStats := make(map[string]map[string]int)
            for {
                serviceStream, err := serviceStreams()
                if err != nil {
                    break
                }

                serviceLogs, _ := stream.Collect(serviceStream)
                if len(serviceLogs) > 0 {
                    service := stream.GetOr(serviceLogs[0], "service", "unknown")

                    // Analyze log levels for this service
                    levelCounts := make(map[string]int)
                    for _, log := range serviceLogs {
                        level := stream.GetOr(log, "level", "")
                        levelCounts[level]++
                    }

                    serviceStats[service] = levelCounts
                }
            }

            // Report per-tenant service statistics
            for service, levels := range serviceStats {
                fmt.Printf("  %s service:\n", service)
                for level, count := range levels {
                    fmt.Printf("    %s: %d\n", level, count)
                }
            }
            fmt.Println()
        }
    }

    // Step 3: Global analysis across all tenants
    fmt.Printf("=== Step 2: Global Analysis ===\n")

    // Combine all tenant logs for global analysis
    allLogs := make([]stream.Record, 0)
    for _, logs := range tenantProcessors {
        allLogs = append(allLogs, logs...)
    }

    // Parallel processing of error analysis
    errorAnalysis, _ := stream.Aggregates(
        stream.Parallel(4, func(log stream.Record) stream.Record {
            level := stream.GetOr(log, "level", "")
            if level == "ERROR" {
                // Simulate expensive error analysis
                time.Sleep(1 * time.Millisecond)
                return stream.SetField(log, "analyzed", true)
            }
            return log
        })(stream.Chain(
            stream.Where(func(log stream.Record) bool {
                return stream.GetOr(log, "level", "") == "ERROR"
            }),
        )(stream.FromSlice(allLogs))),
        stream.CountStream[stream.Record]("total_errors"),
    )

    fmt.Printf("Global Statistics:\n")
    fmt.Printf("- Total tenants processed: %d\n", tenantCount)
    fmt.Printf("- Total logs processed: %d\n", len(allLogs))
    fmt.Printf("- Total errors analyzed: %d\n", stream.GetOr(errorAnalysis, "total_errors", int64(0)))

    // Step 4: Real-time alerting simulation
    fmt.Printf("\n=== Step 3: Real-Time Alerting ===\n")

    alerts := make([]string, 0)
    for tenant, logs := range tenantProcessors {
        errorCount := 0
        for _, log := range logs {
            if stream.GetOr(log, "level", "") == "ERROR" {
                errorCount++
            }
        }

        // Alert if error rate > 20%
        errorRate := float64(errorCount) / float64(len(logs))
        if errorRate > 0.2 {
            alert := fmt.Sprintf("HIGH ERROR RATE: %s (%.1f%% errors)", tenant, errorRate*100)
            alerts = append(alerts, alert)
        }
    }

    if len(alerts) > 0 {
        fmt.Printf("🚨 ALERTS GENERATED:\n")
        for _, alert := range alerts {
            fmt.Printf("- %s\n", alert)
        }
    } else {
        fmt.Printf("✅ All tenants operating normally\n")
    }
}
```

**Pattern Summary:**
- **Multi-level splitting** - First by tenant, then by service
- **Parallel processing** - CPU-intensive error analysis
- **Context management** - Timeout and resource control
- **Real-time analytics** - Streaming aggregation and alerting
- **Scalable architecture** - Independent per-tenant processing

---

## Summary

You've learned StreamV2's most advanced features:

- ✅ **Dynamic splitting** with `Split()` for runtime partitioning
- ✅ **Advanced composition** with `Chain()` and `Pipe()`
- ✅ **Explicit parallelization** with `Parallel()` for performance
- ✅ **Resource management** with `WithContext()` for production reliability
- ✅ **Real-world patterns** combining all features for complex systems

### Next Steps

- Explore the [Advanced Windowing Codelab](ADVANCED_WINDOWING_CODELAB.md) for time-based processing
- Check out the [API Documentation](api.md) for complete function references
- Build your own streaming applications using these advanced patterns

### Related Resources

- [StreamV2 Main Codelab](STREAMV2_CODELAB.md) - Core concepts and basic operations
- [Advanced Windowing Codelab](ADVANCED_WINDOWING_CODELAB.md) - Time-based stream processing
- [API Reference](api.md) - Complete function documentation
- [Examples](../examples/) - Real-world usage patterns

---

**Advanced features unlock StreamV2's full potential - use them to build scalable, high-performance streaming applications!** 🚀