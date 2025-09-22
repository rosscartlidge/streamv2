# StreamV2 Codelab: From Zero to Stream Processing Hero

Welcome to the StreamV2 codelab! This hands-on tutorial will teach you modern stream processing in Go using a simple, elegant API. You'll start with basic concepts and gradually build up to powerful data processing pipelines.

## What You'll Learn

- ✅ Core stream operations (Map, Filter, Reduce)
- ✅ Data aggregation and statistics
- ✅ Working with CSV, JSON, and other formats
- ✅ Building processing pipelines
- ✅ Real-world data analysis patterns

📚 **Advanced Topics**: For complex windowing, event-time processing, and real-time analytics, see the [Advanced Windowing Codelab](ADVANCED_WINDOWING_CODELAB.md)

## Prerequisites

- Basic Go knowledge
- Go 1.21+ installed

---

## Setup

```bash
# Clone the repository
git clone https://github.com/rosscartlidge/streamv2
cd streamv2

# Run the examples
go run examples/codelab/step1.go
```

---

# Step 1: Your First Stream

## What Are Streams?

Think of a stream as a **pipeline of data** that flows through your program. Instead of processing all your data at once in memory, streams let you process data **one piece at a time**, which is:

- **Memory efficient**: Handle millions of records without loading them all at once
- **Composable**: Chain operations together like building blocks
- **Lazy**: Nothing happens until you actually need the results

## The Stream Mindset

Traditional Go:
```go
// Process everything at once
numbers := []int{1, 2, 3, 4, 5}
doubled := make([]int, len(numbers))
for i, n := range numbers {
    doubled[i] = n * 2  // All in memory
}
```

StreamV2 Way:
```go
// Build a processing pipeline
numbers := []int{1, 2, 3, 4, 5}
doubled := stream.Map(func(n int) int { return n * 2 })(
    stream.FromSlice(numbers)
)
// Nothing processed yet! Just a pipeline definition.
```

## Your First Stream

Let's start with the simplest possible example - converting data to a stream and back:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Println("🚀 Welcome to StreamV2!")

    // Step 1: Start with regular Go data
    numbers := []int64{1, 2, 3, 4, 5}
    fmt.Println("📊 Original data:", numbers)

    // Step 2: Convert to a stream (creates a pipeline)
    numberStream := stream.FromSlice(numbers)
    fmt.Println("🔄 Created stream (no processing yet!)")

    // Step 3: Collect results (this triggers processing)
    result, err := stream.Collect(numberStream)
    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        return
    }

    fmt.Println("✅ Stream result:", result)
    fmt.Println("💡 Notice: Same data, but it went through a stream!")
}
```

## What Just Happened?

1. **`FromSlice(numbers)`**: Converts your regular Go slice into a stream
   - Think of this as "putting your data on a conveyor belt"
   - No processing happens yet - it's just preparation

2. **`Collect(numberStream)`**: Pulls all data through the stream
   - This is when processing actually happens
   - Converts the stream back to a regular Go slice

3. **Lazy Evaluation**: The magic of streams!
   - You build the pipeline first (`FromSlice`)
   - Processing only happens when you ask for results (`Collect`)

## Why This Matters

This might seem pointless (same input, same output), but you've just learned the **fundamental pattern** of stream processing:

```
Data → Stream → Operations → Collect → Results
```

Every StreamV2 program follows this pattern. Next, we'll add the "Operations" part to make it useful!

## Try It Yourself

Run the code above and observe:
- ✅ You get the same data back
- ✅ No errors or crashes
- ✅ You've successfully used streams!

This foundation will support everything else we build together.

---

# Step 2: Transform Data with Map

## The Power of Transformation

Now comes the fun part! `Map` is like having a **magical transformation machine**. You put data in, apply a function to each piece, and get transformed data out.

Think of it like an assembly line:
- **Input**: Raw materials (your data)
- **Machine**: Your transformation function
- **Output**: Finished products (transformed data)

## Real-World Analogy

Imagine you run a bakery:
- **Input**: `[flour, sugar, eggs, butter]`
- **Transform**: "bake into" function
- **Output**: `[bread, cake, cookies, pastry]`

That's exactly what `Map` does, but with data!

## Your First Transformation

Let's start with a simple math example:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Println("🔢 Learning Map - Data Transformation!")

    // Our input data - simple numbers
    numbers := []int64{1, 2, 3, 4, 5}
    fmt.Println("📊 Original numbers:", numbers)

    // Transform: Square each number
    fmt.Println("🔄 Squaring each number...")
    squares, err := stream.Collect(
        stream.Map(func(x int64) int64 {
            fmt.Printf("   Transforming %d → %d\n", x, x*x)
            return x * x
        })(stream.FromSlice(numbers)))

    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        return
    }

    fmt.Println("✅ Squares:", squares)
    fmt.Println()

    // Transform: Numbers to labels (change type!)
    fmt.Println("🏷️  Converting numbers to labels...")
    labels, err := stream.Collect(
        stream.Map(func(x int64) string {
            label := fmt.Sprintf("Item-%d", x)
            fmt.Printf("   Converting %d → %s\n", x, label)
            return label
        })(stream.FromSlice(numbers)))

    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        return
    }

    fmt.Println("✅ Labels:", labels)
}
```

## Breaking Down the Magic

Let's understand what's happening step by step:

### 1. The Map Function
```go
stream.Map(func(x int64) int64 {
    return x * x
})
```

This creates a transformation that says: "For each number `x`, return `x * x`"

### 2. The Pipeline Pattern
```go
stream.Map(...)(stream.FromSlice(numbers))
```

This might look weird at first! It's StreamV2's way of chaining operations:
- `stream.Map(...)` creates a transformation function
- `(stream.FromSlice(numbers))` applies it to your data stream

Think of it as: `transformation(data)`

### 3. Type Changes
```go
stream.Map(func(x int64) string {
    return fmt.Sprintf("Item-%d", x)
})
```

Notice how we go from `int64` → `string`. Map can change types completely!

## What You'll See

When you run this code:
```
🔢 Learning Map - Data Transformation!
📊 Original numbers: [1 2 3 4 5]
🔄 Squaring each number...
   Transforming 1 → 1
   Transforming 2 → 4
   Transforming 3 → 9
   Transforming 4 → 16
   Transforming 5 → 25
✅ Squares: [1 4 9 16 25]

🏷️  Converting numbers to labels...
   Converting 1 → Item-1
   Converting 2 → Item-2
   Converting 3 → Item-3
   Converting 4 → Item-4
   Converting 5 → Item-5
✅ Labels: [Item-1 Item-2 Item-3 Item-4 Item-5]
```

## Key Insights

1. **One-to-One**: Map always produces the same number of outputs as inputs
   - 5 numbers in → 5 numbers out
   - 5 numbers in → 5 strings out

2. **Type Safety**: Go's type system ensures your transformations are correct
   - If your function expects `int64`, you can't accidentally pass `string`

3. **Pure Functions**: Your transformation function should be predictable
   - Same input → same output (every time)
   - No side effects (don't modify global variables)

## Common Use Cases

- **Math transformations**: Squaring, doubling, converting units
- **String formatting**: Adding prefixes, changing case, templating
- **Data extraction**: Getting fields from complex objects
- **Type conversion**: Numbers to strings, strings to numbers

You've just learned the most fundamental stream operation! Next, we'll learn how to filter data to keep only what we want.

---

# Step 3: Filter Data with Where

## The Art of Selection

While `Map` transforms data, `Where` **selects** data. Think of it as having a **quality control inspector** on your assembly line who decides which items to keep and which to reject.

## Real-World Analogy

Imagine sorting fruit at a grocery store:
- **Input**: `[apple, orange, rotten_apple, banana, moldy_orange]`
- **Filter**: "Keep only fresh fruit"
- **Output**: `[apple, orange, banana]`

That's exactly what `Where` does - it applies a **test** to each item and keeps only those that pass!

## Your First Filter

Let's start with a simple example - finding even numbers:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Println("🔍 Learning Where - Data Filtering!")

    // Our test data - mixed numbers
    numbers := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    fmt.Println("📊 Original numbers:", numbers)

    // Filter: Keep only even numbers
    fmt.Println("🔄 Testing each number for evenness...")
    evens, err := stream.Collect(
        stream.Where(func(x int64) bool {
            isEven := x%2 == 0
            fmt.Printf("   Testing %d: even? %v → ", x, isEven)
            if isEven {
                fmt.Println("KEEP ✅")
            } else {
                fmt.Println("DISCARD ❌")
            }
            return isEven
        })(stream.FromSlice(numbers)))

    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        return
    }

    fmt.Println("✅ Even numbers:", evens)
    fmt.Println()

    // Filter: Keep only large numbers (> 5)
    fmt.Println("🔄 Testing each number for size...")
    bigNumbers, err := stream.Collect(
        stream.Where(func(x int64) bool {
            isBig := x > 5
            fmt.Printf("   Testing %d > 5: %v → ", x, isBig)
            if isBig {
                fmt.Println("KEEP ✅")
            } else {
                fmt.Println("DISCARD ❌")
            }
            return isBig
        })(stream.FromSlice(numbers)))

    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        return
    }

    fmt.Println("✅ Numbers > 5:", bigNumbers)
}
```

## Understanding the Filter Function

### The Predicate Function
```go
stream.Where(func(x int64) bool {
    return x%2 == 0  // This is your "test"
})
```

This function is called a **predicate** - it asks a yes/no question about each element:
- **Input**: One element from your stream
- **Output**: `true` (keep it) or `false` (discard it)

### Common Predicate Patterns

```go
// Numeric comparisons
stream.Where(func(x int64) bool { return x > 10 })     // Greater than
stream.Where(func(x int64) bool { return x <= 5 })     // Less than or equal
stream.Where(func(x int64) bool { return x == 42 })    // Exactly equal

// Mathematical tests
stream.Where(func(x int64) bool { return x%3 == 0 })   // Divisible by 3
stream.Where(func(x int64) bool { return x*x < 100 })  // Square less than 100

// Range tests
stream.Where(func(x int64) bool { return x >= 10 && x <= 20 }) // Between 10 and 20
```

## What You'll See

When you run the code:
```
🔍 Learning Where - Data Filtering!
📊 Original numbers: [1 2 3 4 5 6 7 8 9 10]
🔄 Testing each number for evenness...
   Testing 1: even? false → DISCARD ❌
   Testing 2: even? true → KEEP ✅
   Testing 3: even? false → DISCARD ❌
   Testing 4: even? true → KEEP ✅
   Testing 5: even? false → DISCARD ❌
   Testing 6: even? true → KEEP ✅
   Testing 7: even? false → DISCARD ❌
   Testing 8: even? true → KEEP ✅
   Testing 9: even? false → DISCARD ❌
   Testing 10: even? true → KEEP ✅
✅ Even numbers: [2 4 6 8 10]

🔄 Testing each number for size...
   Testing 1 > 5: false → DISCARD ❌
   Testing 2 > 5: false → DISCARD ❌
   Testing 3 > 5: false → DISCARD ❌
   Testing 4 > 5: false → DISCARD ❌
   Testing 5 > 5: false → DISCARD ❌
   Testing 6 > 5: true → KEEP ✅
   Testing 7 > 5: true → KEEP ✅
   Testing 8 > 5: true → KEEP ✅
   Testing 9 > 5: true → KEEP ✅
   Testing 10 > 5: true → KEEP ✅
✅ Numbers > 5: [6 7 8 9 10]
```

## Key Insights

1. **Order Preserved**: Filtered elements keep their original order
   - `[1,2,3,4,5,6,7,8,9,10]` → `[2,4,6,8,10]` (evens stay in order)

2. **Size Reduction**: Filtering usually makes your data smaller
   - 10 numbers → 5 even numbers
   - This saves memory and processing time

3. **All or Nothing**: Each element either passes or fails completely
   - No "partial" filtering - it's binary

4. **Pure Predicates**: Your test function should be predictable
   - Same input → same result (every time)
   - No side effects (don't modify data inside the test)

## Common Use Cases

- **Data cleaning**: Remove invalid or corrupt records
- **Search results**: Find items matching criteria
- **Validation**: Keep only items that pass quality checks
- **Security**: Filter out unauthorized requests
- **Performance**: Process only relevant data

## The Power of Filtering

Filtering is incredibly powerful because it:
- **Reduces noise**: Focus on what matters
- **Improves performance**: Process less data
- **Enables precision**: Target exactly what you need

Next, we'll learn how to combine Map and Where together for even more powerful data processing!

---

# Step 4: Combine Operations with Pipe

## Building Data Processing Pipelines

Now comes the **real magic**! You've learned `Map` (transform) and `Where` (filter) individually. But the true power of streams comes from **combining** operations to build sophisticated data processing pipelines.

## The Assembly Line Concept

Think of it like a car manufacturing assembly line:
1. **Station 1**: Install engine (transform)
2. **Station 2**: Quality check (filter)
3. **Station 3**: Paint car (transform)
4. **Station 4**: Final inspection (filter)

Each station does one job, but together they create something amazing!

## Your First Pipeline

Let's build a pipeline that finds the "big squares" - numbers that become large when squared:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Println("🔧 Building Your First Pipeline!")

    numbers := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    fmt.Println("📊 Starting with:", numbers)

    // BUILD THE PIPELINE
    fmt.Println("\n🏗️  Building pipeline: Square → Filter(>20)")

    // Method 1: Using Pipe (clean and readable)
    result, err := stream.Collect(
        stream.Pipe(
            stream.Map(func(x int64) int64 {
                fmt.Printf("   🔢 Squaring %d → %d\n", x, x*x)
                return x * x
            }),
            stream.Where(func(x int64) bool {
                keep := x > 20
                fmt.Printf("   🔍 Testing %d > 20: %v → ", x, keep)
                if keep {
                    fmt.Println("KEEP ✅")
                } else {
                    fmt.Println("DISCARD ❌")
                }
                return keep
            }),
        )(stream.FromSlice(numbers)))

    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        return
    }

    fmt.Printf("\n✅ Final result: %v\n", result)

    // STEP-BY-STEP ANALYSIS
    fmt.Println("\n📋 Let's trace this step by step:")

    // Step 1: Just the squares
    squares, _ := stream.Collect(
        stream.Map(func(x int64) int64 { return x * x })(
            stream.FromSlice(numbers)))

    fmt.Printf("   Step 1 (squares): %v\n", squares)
    fmt.Printf("   Step 2 (filter >20): %v\n", result)

    // ALTERNATIVE: Manual chaining (harder to read)
    fmt.Println("\n🔗 Alternative: Manual chaining")
    manualResult, _ := stream.Collect(
        stream.Where(func(x int64) bool { return x > 20 })(
            stream.Map(func(x int64) int64 { return x * x })(
                stream.FromSlice(numbers))))

    fmt.Printf("   Same result: %v\n", manualResult)
}
```

## Understanding Pipeline Flow

Let's trace exactly what happens to each number:

```
Input: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

Pipeline: Square → Filter(>20)

Number 1: 1 → 1² = 1  → 1 > 20? NO  → DISCARD
Number 2: 2 → 2² = 4  → 4 > 20? NO  → DISCARD
Number 3: 3 → 3² = 9  → 9 > 20? NO  → DISCARD
Number 4: 4 → 4² = 16 → 16 > 20? NO → DISCARD
Number 5: 5 → 5² = 25 → 25 > 20? YES → KEEP
Number 6: 6 → 6² = 36 → 36 > 20? YES → KEEP
...and so on...

Final Output: [25, 36, 49, 64, 81, 100]
```

## Two Ways to Build Pipelines

### Method 1: Using `Pipe` (Recommended)
```go
stream.Pipe(
    stream.Map(func(x int64) int64 { return x * x }),
    stream.Where(func(x int64) bool { return x > 20 }),
)(stream.FromSlice(numbers))
```

**Advantages**:
- ✅ Easy to read left-to-right
- ✅ Clear operation sequence
- ✅ Easy to add/remove steps

### Method 2: Manual Chaining
```go
stream.Where(func(x int64) bool { return x > 20 })(
    stream.Map(func(x int64) int64 { return x * x })(
        stream.FromSlice(numbers)))
```

**Disadvantages**:
- ❌ Reads right-to-left (confusing)
- ❌ Hard to modify
- ❌ Nested parentheses get messy

## Key Pipeline Concepts

1. **Sequential Processing**: Operations happen in order
   - First ALL numbers get squared
   - Then ALL squares get filtered
   - (Not one-by-one, but conceptually in sequence)

2. **Type Safety**: Each step must match the next
   - `Map(int64→int64)` → `Where(int64→bool)` ✅
   - `Map(int64→string)` → `Where(int64→bool)` ❌ (Type mismatch!)

3. **Efficient**: StreamV2 optimizes the pipeline
   - No intermediate arrays created
   - Memory efficient processing

## What You'll See

When you run the code:
```
🔧 Building Your First Pipeline!
📊 Starting with: [1 2 3 4 5 6 7 8 9 10]

🏗️  Building pipeline: Square → Filter(>20)
   🔢 Squaring 1 → 1
   🔍 Testing 1 > 20: false → DISCARD ❌
   🔢 Squaring 2 → 4
   🔍 Testing 4 > 20: false → DISCARD ❌
   🔢 Squaring 3 → 9
   🔍 Testing 9 > 20: false → DISCARD ❌
   🔢 Squaring 4 → 16
   🔍 Testing 16 > 20: false → DISCARD ❌
   🔢 Squaring 5 → 25
   🔍 Testing 25 > 20: true → KEEP ✅
   🔢 Squaring 6 → 36
   🔍 Testing 36 > 20: true → KEEP ✅
   ...

✅ Final result: [25 36 49 64 81 100]

📋 Let's trace this step by step:
   Step 1 (squares): [1 4 9 16 25 36 49 64 81 100]
   Step 2 (filter >20): [25 36 49 64 81 100]
```

## Pipeline Best Practices

1. **Start Simple**: Begin with 2-3 operations, add more as needed
2. **Use Pipe**: Much more readable than manual chaining
3. **Name Operations**: Create variables for complex functions
4. **Test Each Step**: Verify intermediate results during development

You've now learned the foundation of stream processing! Next, we'll learn about aggregation - turning streams of data into single summary values.

---

# Step 5: Aggregate Data

## From Many to One: The Power of Aggregation

So far you've learned to transform streams (Map) and filter streams (Where). Now let's learn **aggregation** - taking a stream of many values and computing a single summary result.

## Real-World Analogy

Think of aggregation like a **teacher grading exams**:
- **Input**: Stack of test papers `[85, 92, 78, 95, 88]`
- **Aggregations**:
  - Average score: `87.6`
  - Highest score: `95`
  - Lowest score: `78`
  - Total points: `438`
  - Number of students: `5`

That's exactly what stream aggregation does - it **summarizes** your data!

## Understanding Aggregation Types

There are two main categories:

### 1. **Mathematical Aggregations**
- `Sum`: Add all values together
- `Avg`: Calculate the average (mean)
- `Count`: How many items exist

### 2. **Comparison Aggregations**
- `Max`: Find the largest value
- `Min`: Find the smallest value

## Your First Aggregation

Let's analyze some sales data:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Println("📊 Learning Aggregation - Sales Data Analysis!")

    // Daily sales figures for a week (in dollars)
    dailySales := []int64{1200, 1800, 950, 2100, 1650, 2400, 1900}
    fmt.Println("📈 Daily sales this week:", dailySales)

    fmt.Println("\n🧮 Calculating business metrics...")

    // Calculate total revenue
    totalRevenue, err := stream.Sum(stream.FromSlice(dailySales))
    if err != nil {
        fmt.Printf("❌ Error calculating sum: %v\n", err)
        return
    }
    fmt.Printf("   💰 Total weekly revenue: $%d\n", totalRevenue)

    // Count number of days
    dayCount, err := stream.Count(stream.FromSlice(dailySales))
    if err != nil {
        fmt.Printf("❌ Error counting days: %v\n", err)
        return
    }
    fmt.Printf("   📅 Number of business days: %d\n", dayCount)

    // Calculate average daily sales
    avgSales, err := stream.Avg(stream.FromSlice(dailySales))
    if err != nil {
        fmt.Printf("❌ Error calculating average: %v\n", err)
        return
    }
    fmt.Printf("   📊 Average daily sales: $%.0f\n", avgSales)

    // Find best day
    bestDay, err := stream.Max(stream.FromSlice(dailySales))
    if err != nil {
        fmt.Printf("❌ Error finding maximum: %v\n", err)
        return
    }
    fmt.Printf("   🏆 Best sales day: $%d\n", bestDay)

    // Find worst day
    worstDay, err := stream.Min(stream.FromSlice(dailySales))
    if err != nil {
        fmt.Printf("❌ Error finding minimum: %v\n", err)
        return
    }
    fmt.Printf("   📉 Lowest sales day: $%d\n", worstDay)

    // ADVANCED: Combine with transformations
    fmt.Println("\n🔄 Advanced: Aggregating after transformation...")

    // Question: What if we gave a 10% bonus on each day?
    bonusTotal, err := stream.Sum(
        stream.Map(func(sales int64) int64 {
            bonus := sales + (sales / 10) // Add 10% bonus
            fmt.Printf("   💡 $%d + 10%% bonus = $%d\n", sales, bonus)
            return bonus
        })(stream.FromSlice(dailySales)))

    if err != nil {
        fmt.Printf("❌ Error calculating bonus total: %v\n", err)
        return
    }

    fmt.Printf("\n✅ Total with 10%% bonuses: $%d\n", bonusTotal)
    fmt.Printf("📈 That's $%d more than regular total!\n", bonusTotal-totalRevenue)
}
```

## How Aggregation Works

Unlike `Map` and `Where` which return streams, aggregation functions **consume** the entire stream:

```
Stream Processing:    [Data] → Map → Where → [More Data]
Aggregation:         [Data] → Sum → Single Number
```

Here's what happens internally:

```go
// For Sum([10, 20, 30, 40, 50]):
result = 0
result = result + 10  // = 10
result = result + 20  // = 30
result = result + 30  // = 60
result = result + 40  // = 100
result = result + 50  // = 150
return 150
```

## Combining Aggregation with Pipelines

The real power comes from **aggregating after transformation**:

```go
// Question: What's the sum of all even numbers squared?
result := stream.Sum(
    stream.Map(func(x int64) int64 { return x * x })(      // Square each
        stream.Where(func(x int64) bool { return x%2 == 0 })( // Keep evens
            stream.FromSlice([]int64{1,2,3,4,5,6,7,8,9,10}))))

// Pipeline: [1,2,3,4,5,6,7,8,9,10] → [2,4,6,8,10] → [4,16,36,64,100] → 220
```

## What You'll See

When you run the sales analysis code:
```
📊 Learning Aggregation - Sales Data Analysis!
📈 Daily sales this week: [1200 1800 950 2100 1650 2400 1900]

🧮 Calculating business metrics...
   💰 Total weekly revenue: $12000
   📅 Number of business days: 7
   📊 Average daily sales: $1714
   🏆 Best sales day: $2400
   📉 Lowest sales day: $950

🔄 Advanced: Aggregating after transformation...
   💡 $1200 + 10% bonus = $1320
   💡 $1800 + 10% bonus = $1980
   💡 $950 + 10% bonus = $1045
   💡 $2100 + 10% bonus = $2310
   💡 $1650 + 10% bonus = $1815
   💡 $2400 + 10% bonus = $2640
   💡 $1900 + 10% bonus = $2090

✅ Total with 10% bonuses: $13200
📈 That's $1200 more than regular total!
```

## Key Aggregation Insights

1. **Terminal Operations**: Aggregations end your stream pipeline
   - You can't chain more operations after aggregation
   - They return concrete values, not streams

2. **Type Safety**: Aggregations work with appropriate types
   - `Sum`, `Max`, `Min` need comparable/numeric types
   - `Count` works with any type

3. **Empty Streams**: Handle empty data gracefully
   - `Count` returns 0
   - Other aggregations may return errors

4. **Performance**: Aggregations are optimized
   - Single pass through your data
   - Memory efficient processing

## Common Aggregation Patterns

```go
// Business metrics
totalSales := stream.Sum(salesStream)
customerCount := stream.Count(customerStream)
averageOrderValue := stream.Avg(orderValueStream)

// Quality control
highestScore := stream.Max(testScoreStream)
lowestTemperature := stream.Min(temperatureStream)

// Data analysis
outlierCount := stream.Count(
    stream.Where(func(x float64) bool { return x > threshold })(dataStream))
```

You've now learned how to summarize and analyze your stream data! Next, we'll work with real-world data formats like CSV files.

---

# Step 6: Working with Real Data - CSV Processing

## From Spreadsheets to Streams

Now let's apply everything you've learned to **real-world data**! CSV (Comma-Separated Values) files are everywhere in business - from sales reports to customer databases to financial records.

## What Are CSV Files?

CSV files are like **digital spreadsheets** stored as plain text:

```
name,age,salary,department
Alice,30,75000,Engineering
Bob,25,65000,Sales
Charlie,35,85000,Engineering
```

Each row is a **record** (like a database row), and each column is a **field**. StreamV2 converts these into `Record` objects that you can process with all the operations you've learned!

## Understanding StreamV2 Records

When StreamV2 reads CSV data, it creates `Record` objects:

```go
// CSV row: "Alice,30,75000,Engineering"
// Becomes: stream.Record{
//   "name": "Alice",
//   "age": 30,
//   "salary": 75000,
//   "department": "Engineering"
// }
```

## Your First CSV Analysis

Let's analyze employee data from a company's HR system:

```go
package main

import (
    "fmt"
    "strings"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Println("📊 Employee Data Analysis with CSV!")

    // Sample employee data (in real life, this would come from a .csv file)
    csvData := `name,age,salary,department
Alice,30,75000,Engineering
Bob,25,65000,Sales
Charlie,35,85000,Engineering
Diana,28,70000,Marketing
Eve,32,80000,Engineering
Frank,29,72000,Sales
Grace,26,68000,Marketing`

    fmt.Println("📂 Loading employee data from CSV...")

    // Parse CSV data into Record objects
    records, err := stream.Collect(
        stream.CSVToStream(strings.NewReader(csvData)))

    if err != nil {
        fmt.Printf("❌ Error reading CSV: %v\n", err)
        return
    }

    fmt.Printf("✅ Loaded %d employee records\n", len(records))

    // Let's explore what a Record looks like
    fmt.Println("\n🔍 Examining the first record:")
    firstEmployee := records[0]
    fmt.Printf("   Record structure: %+v\n", firstEmployee)
    fmt.Printf("   Name: %v (type: %T)\n", firstEmployee["name"], firstEmployee["name"])
    fmt.Printf("   Age: %v (type: %T)\n", firstEmployee["age"], firstEmployee["age"])

    // ANALYSIS 1: Extract all employee names
    fmt.Println("\n👥 Analysis 1: All Employee Names")
    names, err := stream.Collect(
        stream.Map(func(record stream.Record) string {
            // Extract the "name" field and convert to string
            name := stream.GetOr(record, "name", "Unknown")
            fmt.Printf("   📝 Extracted name: %s\n", name)
            return name
        })(stream.FromSlice(records)))

    if err != nil {
        fmt.Printf("❌ Error extracting names: %v\n", err)
        return
    }

    fmt.Printf("✅ All employees: %v\n", names)

    // ANALYSIS 2: Find all Engineering employees
    fmt.Println("\n🔧 Analysis 2: Engineering Department")
    engineers, err := stream.Collect(
        stream.Where(func(record stream.Record) bool {
            dept := stream.GetOr(record, "department", "")
            isEngineer := dept == "Engineering"
            fmt.Printf("   🔍 %s works in %s → Engineer? %v\n",
                      stream.GetOr(record, "name", "Unknown"), dept, isEngineer)
            return isEngineer
        })(stream.FromSlice(records)))

    if err != nil {
        fmt.Printf("❌ Error filtering engineers: %v\n", err)
        return
    }

    fmt.Printf("\n✅ Found %d engineers:\n", len(engineers))
    for _, eng := range engineers {
        name := stream.GetOr(eng, "name", "Unknown")
        age := stream.GetOr(eng, "age", 0)
        salary := stream.GetOr(eng, "salary", 0)
        fmt.Printf("   👨‍💻 %s: age %v, salary $%v\n", name, age, salary)
    }

    // ANALYSIS 3: Calculate average engineering salary
    fmt.Println("\n💰 Analysis 3: Engineering Salary Analysis")
    avgEngSalary, err := stream.Avg(
        stream.Map(func(record stream.Record) float64 {
            salary := stream.GetOr(record, "salary", float64(0))
            fmt.Printf("   💵 Engineering salary: $%.0f\n", salary)
            return salary
        })(stream.FromSlice(engineers)))

    if err != nil {
        fmt.Printf("❌ Error calculating average: %v\n", err)
        return
    }

    fmt.Printf("✅ Average engineering salary: $%.0f\n", avgEngSalary)
}
```

## Understanding Record Access

StreamV2 provides safe ways to access Record fields:

### Method 1: Direct Access (Risky)
```go
name := record["name"].(string)  // Crashes if field missing or wrong type!
```

### Method 2: Safe Access (Recommended)
```go
name := stream.GetOr(record, "name", "Unknown")  // Returns default if missing
```

The `GetOr` function is your friend - it prevents crashes and provides sensible defaults!

## What You'll See

When you run the employee analysis:
```
📊 Employee Data Analysis with CSV!
📂 Loading employee data from CSV...
✅ Loaded 7 employee records

🔍 Examining the first record:
   Record structure: map[age:30 department:Engineering name:Alice salary:75000]
   Name: Alice (type: string)
   Age: 30 (type: int64)

👥 Analysis 1: All Employee Names
   📝 Extracted name: Alice
   📝 Extracted name: Bob
   📝 Extracted name: Charlie
   📝 Extracted name: Diana
   📝 Extracted name: Eve
   📝 Extracted name: Frank
   📝 Extracted name: Grace
✅ All employees: [Alice Bob Charlie Diana Eve Frank Grace]

🔧 Analysis 2: Engineering Department
   🔍 Alice works in Engineering → Engineer? true
   🔍 Bob works in Sales → Engineer? false
   🔍 Charlie works in Engineering → Engineer? true
   🔍 Diana works in Marketing → Engineer? false
   🔍 Eve works in Engineering → Engineer? true
   🔍 Frank works in Sales → Engineer? false
   🔍 Grace works in Marketing → Engineer? false

✅ Found 3 engineers:
   👨‍💻 Alice: age 30, salary $75000
   👨‍💻 Charlie: age 35, salary $85000
   👨‍💻 Eve: age 32, salary $80000

💰 Analysis 3: Engineering Salary Analysis
   💵 Engineering salary: $75000
   💵 Engineering salary: $85000
   💵 Engineering salary: $80000
✅ Average engineering salary: $80000
```

## Key CSV Processing Concepts

1. **Automatic Parsing**: `CSVToStream` handles the parsing for you
   - Headers become field names
   - Data gets appropriate Go types (string, int64, float64)

2. **Record Structure**: Each CSV row becomes a `Record`
   - Think of it as `map[string]any`
   - Access fields by column name: `record["salary"]`

3. **Type Safety**: Use `GetOr` for safe field access
   - Prevents crashes from missing fields
   - Provides sensible defaults

4. **Real Data**: CSV processing opens the door to real-world applications
   - Sales reports, customer databases, sensor data
   - Any spreadsheet can become a stream!

## Reading From Actual Files

In real applications, you'd read from files:

```go
// Read from a real CSV file
file, _ := os.Open("employees.csv")
defer file.Close()

records, _ := stream.Collect(stream.CSVToStream(file))
```

You've just learned how to process real-world data with StreamV2! Next, we'll build more sophisticated analysis pipelines.

---

# Step 7: Data Analysis Pipeline

## Building Professional Analytics

Now let's combine everything you've learned into a **complete data analysis pipeline**! This is where StreamV2 really shines - building sophisticated analytics that would take dozens of lines in traditional Go.

## Real-World Scenario: HR Analytics

Imagine you're a data analyst at a tech company. HR wants insights about:
- 💰 Average salaries by department
- 🏆 High-performing engineers (experience + salary)
- 📊 Company-wide salary distribution
- 🎯 Which departments to focus recruitment on

Let's build this analysis step by step!

```go
package main

import (
    "fmt"
    "strings"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    csvData := `name,age,salary,department,years_experience
Alice,30,75000,Engineering,5
Bob,25,65000,Sales,3
Charlie,35,85000,Engineering,10
Diana,28,70000,Marketing,4
Eve,32,80000,Engineering,7
Frank,29,68000,Sales,4
Grace,33,82000,Engineering,8
Henry,26,62000,Marketing,2
Iris,31,77000,Engineering,6
Jack,27,64000,Sales,3`

    records, err := stream.Collect(
        stream.CSVToStream(strings.NewReader(csvData)))
    if err != nil {
        panic(err)
    }

    fmt.Printf("=== Employee Data Analysis ===\n")
    fmt.Printf("Total employees: %d\n\n", len(records))

    // 1. Average salary by department
    depts := []string{"Engineering", "Sales", "Marketing"}
    
    for _, dept := range depts {
        deptEmployees, _ := stream.Collect(
            stream.Where(func(r stream.Record) bool {
                return r["department"].(string) == dept
            })(stream.FromSlice(records)))
        
        avgSalary, _ := stream.Avg(
            stream.Map(func(r stream.Record) float64 {
                return float64(r["salary"].(int64))
            })(stream.FromSlice(deptEmployees)))
        
        fmt.Printf("%s: %d employees, avg salary: $%.0f\n", 
                   dept, len(deptEmployees), avgSalary)
    }
    
    fmt.Println()

    // 2. High performers: Engineers with >6 years experience and >$75k salary
    highPerformers, _ := stream.Collect(
        stream.Pipe(
            stream.Where(func(r stream.Record) bool {
                return r["department"].(string) == "Engineering"
            }),
            stream.Where(func(r stream.Record) bool {
                return r["years_experience"].(int64) > 6
            }),
            stream.Where(func(r stream.Record) bool {
                return r["salary"].(int64) > 75000
            }),
        )(stream.FromSlice(records)))

    fmt.Printf("High-performing engineers (>6 years, >$75k): %d\n", len(highPerformers))
    for _, hp := range highPerformers {
        fmt.Printf("- %s: %d years, $%d\n", 
                   hp["name"], hp["years_experience"], hp["salary"])
    }
    
    fmt.Println()

    // 3. Salary efficiency: salary per year of experience
    efficiencyData, _ := stream.Collect(
        stream.Map(func(r stream.Record) map[string]any {
            salary := float64(r["salary"].(int64))
            experience := float64(r["years_experience"].(int64))
            efficiency := salary / experience
            
            return map[string]any{
                "name": r["name"],
                "department": r["department"],
                "salary_per_year": efficiency,
            }
        })(stream.FromSlice(records)))

    // Find most efficient employee
    maxEfficiency, _ := stream.Max(
        stream.Map(func(r map[string]any) float64 {
            return r["salary_per_year"].(float64)
        })(stream.FromSlice(efficiencyData)))

    mostEfficient, _ := stream.Collect(
        stream.Where(func(r map[string]any) bool {
            return r["salary_per_year"].(float64) == maxEfficiency
        })(stream.FromSlice(efficiencyData)))

    fmt.Printf("Most salary-efficient employee:\n")
    emp := mostEfficient[0]
    fmt.Printf("- %s (%s): $%.0f per year of experience\n", 
               emp["name"], emp["department"], emp["salary_per_year"])
}
```

**Output:**
```
=== Employee Data Analysis ===
Total employees: 10

Engineering: 5 employees, avg salary: $77800
Sales: 3 employees, avg salary: $65667
Marketing: 2 employees, avg salary: $66000

High-performing engineers (>6 years, >$75k): 3
- Charlie: 10 years, $85000
- Eve: 7 years, $80000
- Grace: 8 years, $82000

Most salary-efficient employee:
- Henry (Marketing): $31000 per year of experience
```

**Key Concepts:**
- Complex pipelines combining multiple Where clauses
- Data transformation creating new calculated fields
- Multi-step analysis using intermediate results
- Type conversions for mathematical operations

---

# Step 8: Working with JSON Data

Let's process JSON data, which is common in modern applications.

```go
package main

import (
    "fmt"
    "strings"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    // JSON Lines format (one JSON object per line)
    jsonData := `{"product": "Widget A", "price": 19.99, "category": "Tools", "units_sold": 150}
{"product": "Gadget B", "price": 29.99, "category": "Electronics", "units_sold": 89}
{"product": "Tool C", "price": 39.99, "category": "Tools", "units_sold": 67}
{"product": "Device D", "price": 49.99, "category": "Electronics", "units_sold": 203}
{"product": "Widget E", "price": 24.99, "category": "Tools", "units_sold": 124}`

    // Parse JSON data
    products, err := stream.Collect(
        stream.JSONToStream(strings.NewReader(jsonData)))
    if err != nil {
        panic(err)
    }

    fmt.Printf("=== Product Sales Analysis ===\n")
    fmt.Printf("Total products: %d\n\n", len(products))

    // 1. Calculate total revenue per product
    revenues, err := stream.Collect(
        stream.Map(func(p stream.Record) map[string]any {
            price := p["price"].(float64)
            units := p["units_sold"].(int64)
            revenue := price * float64(units)
            
            return map[string]any{
                "product": p["product"],
                "category": p["category"],
                "revenue": revenue,
                "units_sold": units,
            }
        })(stream.FromSlice(products)))

    if err != nil {
        panic(err)
    }

    // 2. Find best-selling product by revenue
    maxRevenue, _ := stream.Max(
        stream.Map(func(p map[string]any) float64 {
            return p["revenue"].(float64)
        })(stream.FromSlice(revenues)))

    bestSeller, _ := stream.Collect(
        stream.Where(func(p map[string]any) bool {
            return p["revenue"].(float64) == maxRevenue
        })(stream.FromSlice(revenues)))

    fmt.Printf("Best-selling product by revenue:\n")
    top := bestSeller[0]
    fmt.Printf("- %s: $%.2f revenue (%d units)\n\n", 
               top["product"], top["revenue"], top["units_sold"])

    // 3. Category analysis with GroupBy
    // Convert to Records for structured data processing
    recordStream := stream.Map(func(p map[string]any) stream.Record {
        return stream.Record(p)  // Convert map to Record
    })(stream.FromSlice(revenues))

    // Group by category and calculate aggregations
    categoryStats, _ := stream.Collect(
        stream.GroupBy([]string{"category"}, 
            stream.SumField[float64]("total_revenue", "revenue"),
            stream.SumField[int64]("total_units", "units_sold"),
            stream.CountField("product_count", "product"),
        )(recordStream))

    fmt.Printf("Revenue by category:\n")
    for _, stat := range categoryStats {
        category := stream.GetOr(stat, "category", "")
        revenue := stream.GetOr(stat, "total_revenue", 0.0)
        units := stream.GetOr(stat, "total_units", int64(0))
        products := stream.GetOr(stat, "product_count", int64(0))
        
        fmt.Printf("- %s: $%.2f revenue, %d units sold (%d products)\n", 
                   category, revenue, units, products)
    }

    // 4. Export results back to JSON
    var jsonOutput strings.Builder
    err = stream.StreamToJSON(stream.FromSlice(revenues), &jsonOutput)
    if err != nil {
        panic(err)
    }

    fmt.Printf("\n=== Exported JSON (first 200 chars) ===\n")
    output := jsonOutput.String()
    if len(output) > 200 {
        fmt.Printf("%s...\n", output[:200])
    } else {
        fmt.Println(output)
    }
}
```

**Output:**
```
=== Product Sales Analysis ===
Total products: 5

Best-selling product by revenue:
- Device D: $10147.97 revenue (203 units)

Revenue by category:
- Tools: $8646.33 revenue, 341 units sold (3 products)
- Electronics: $12046.08 revenue, 292 units sold (2 products)

=== Exported JSON (first 200 chars) ===
{"category":"Tools","product":"Widget A","revenue":2998.5,"units_sold":150}
{"category":"Electronics","product":"Gadget B","revenue":2669.11,"units_sold":89}
{"ca...
```

**Key Concepts:**
- `JSONToStream()` and `StreamToJSON()` for JSON processing
- JSON numbers become float64, integers become int64
- `GroupBy()` with explicit aggregations (no automatic count)
- Modern aggregation functions: `SumField`, `CountField`, etc.
- Type-safe record access with `GetOr()`
- Round-trip data processing (JSON → analysis → JSON)

---

# Step 9: Advanced Stream Operations

Let's explore some more advanced stream operations for powerful data processing.

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    // Sample transaction data
    transactions := []map[string]any{
        {"id": 1, "amount": 100.50, "type": "credit", "user_id": "u1"},
        {"id": 2, "amount": 50.25, "type": "debit", "user_id": "u1"},
        {"id": 3, "amount": 200.00, "type": "credit", "user_id": "u2"},
        {"id": 4, "amount": 75.00, "type": "debit", "user_id": "u2"},
        {"id": 5, "amount": 300.00, "type": "credit", "user_id": "u1"},
        {"id": 6, "amount": 25.75, "type": "debit", "user_id": "u3"},
    }

    fmt.Printf("=== Advanced Stream Operations ===\n")
    fmt.Printf("Total transactions: %d\n\n", len(transactions))

    // 1. Take and Skip operations
    firstThree, _ := stream.Collect(
        stream.Take[map[string]any](3)(
            stream.FromSlice(transactions)))

    fmt.Printf("First 3 transactions:\n")
    for _, tx := range firstThree {
        fmt.Printf("- ID %v: $%.2f %s\n", tx["id"], tx["amount"], tx["type"])
    }
    fmt.Println()

    skipTwo, _ := stream.Collect(
        stream.Skip[map[string]any](2)(
            stream.FromSlice(transactions)))

    fmt.Printf("After skipping first 2 transactions: %d remaining\n\n", len(skipTwo))

    // 2. Multiple aggregations at once
    creditAmounts, _ := stream.Collect(
        stream.Map(func(tx map[string]any) float64 {
            if tx["type"].(string) == "credit" {
                return tx["amount"].(float64)
            }
            return 0
        })(stream.FromSlice(transactions)))

    debitAmounts, _ := stream.Collect(
        stream.Map(func(tx map[string]any) float64 {
            if tx["type"].(string) == "debit" {
                return tx["amount"].(float64)
            }
            return 0
        })(stream.FromSlice(transactions)))

    totalCredits, _ := stream.Sum(stream.FromSlice(creditAmounts))
    totalDebits, _ := stream.Sum(stream.FromSlice(debitAmounts))
    netBalance := totalCredits - totalDebits

    fmt.Printf("Financial Summary:\n")
    fmt.Printf("- Total Credits: $%.2f\n", totalCredits)
    fmt.Printf("- Total Debits:  $%.2f\n", totalDebits)
    fmt.Printf("- Net Balance:   $%.2f\n\n", netBalance)

    // 3. Complex filtering and transformation
    largeCredits, _ := stream.Collect(
        stream.Pipe(
            stream.Where(func(tx map[string]any) bool {
                return tx["type"].(string) == "credit"
            }),
            stream.Where(func(tx map[string]any) bool {
                return tx["amount"].(float64) >= 200.0
            }),
            stream.Map(func(tx map[string]any) string {
                return fmt.Sprintf("User %s received $%.2f", 
                                   tx["user_id"], tx["amount"])
            }),
        )(stream.FromSlice(transactions)))

    fmt.Printf("Large credit transactions (>= $200):\n")
    for _, msg := range largeCredits {
        fmt.Printf("- %s\n", msg)
    }
    fmt.Println()

    // 4. User account balances
    users := []string{"u1", "u2", "u3"}
    
    fmt.Printf("User Account Balances:\n")
    for _, userID := range users {
        userTxs, _ := stream.Collect(
            stream.Where(func(tx map[string]any) bool {
                return tx["user_id"].(string) == userID
            })(stream.FromSlice(transactions)))

        balance := 0.0
        for _, tx := range userTxs {
            amount := tx["amount"].(float64)
            if tx["type"].(string) == "credit" {
                balance += amount
            } else {
                balance -= amount
            }
        }

        fmt.Printf("- User %s: $%.2f (%d transactions)\n", 
                   userID, balance, len(userTxs))
    }
}
```

**Output:**
```
=== Advanced Stream Operations ===
Total transactions: 6

First 3 transactions:
- ID 1: $100.50 credit
- ID 2: $50.25 debit
- ID 3: $200.00 credit

After skipping first 2 transactions: 4 remaining

Financial Summary:
- Total Credits: $600.50
- Total Debits:  $151.00
- Net Balance:   $449.50

Large credit transactions (>= $200):
- User u2 received $200.00
- User u1 received $300.00

User Account Balances:
- User u1: $350.25 (3 transactions)
- User u2: $125.00 (2 transactions)
- User u3: $-25.75 (1 transactions)
```

**Key Concepts:**
- `Take()` and `Skip()` for pagination and limiting results
- Multiple aggregations using separate pipelines
- Complex multi-step filtering and transformation
- Business logic implementation using stream operations

---

# Step 10: Best Practices and Performance Tips

## Writing Production-Ready Stream Code

Congratulations! You've learned all the core StreamV2 operations. Now let's cover the **best practices** that will help you write clean, efficient, and maintainable stream processing code.

## 1. Always Handle Errors Properly

StreamV2 operations can fail, especially when reading files or processing invalid data. **Always check errors**:

```go
// ❌ BAD: Ignoring errors
result, _ := stream.Sum(stream.FromSlice(numbers))

// ✅ GOOD: Proper error handling
result, err := stream.Sum(stream.FromSlice(numbers))
if err != nil {
    log.Printf("Failed to calculate sum: %v", err)
    return
}
```

**Why this matters**: In production, ignoring errors leads to silent failures and debugging nightmares.

## 2. Use Pipe for Readable Pipelines

When chaining multiple operations, use `Pipe` instead of nested function calls:

```go
// ❌ BAD: Hard to read (right-to-left)
result := stream.Where(func(x int) bool { return x > 5 })(
    stream.Map(func(x int) int { return x * x })(
        stream.FromSlice(data)))

// ✅ GOOD: Easy to read (left-to-right)
result := stream.Pipe(
    stream.Map(func(x int) int { return x * x }),
    stream.Where(func(x int) bool { return x > 5 }),
)(stream.FromSlice(data))
```

**Why this matters**: Your future self (and your teammates) will thank you for readable code.

## 3. Use Multiple Aggregations for Efficiency

When you need several statistics, compute them in one pass:

```go
// ❌ BAD: Multiple passes through data
sum, _ := stream.Sum(stream.FromSlice(numbers))
count, _ := stream.Count(stream.FromSlice(numbers))  // Reads data again!
avg, _ := stream.Avg(stream.FromSlice(numbers))      // Reads data again!

// ✅ GOOD: Single pass with multiple aggregations
stats, err := stream.Aggregates(stream.FromSlice(numbers),
    stream.SumStream[int64]("total"),
    stream.CountStream[int64]("count"),
    stream.AvgStream[int64]("average"),
)
```

**Why this matters**: Reading data once is much faster than reading it multiple times.

## 4. Use GetOr for Safe Record Access

When working with Records (from CSV, JSON), always use safe accessors:

```go
// ❌ BAD: Can crash if field missing
name := record["name"].(string)
salary := record["salary"].(int64)

// ✅ GOOD: Safe with sensible defaults
name := stream.GetOr(record, "name", "Unknown")
salary := stream.GetOr(record, "salary", int64(0))
```

**Why this matters**: Real-world data is messy. Missing fields shouldn't crash your program.

## 5. Name Complex Operations

For complex transformations, create named functions:

```go
// ❌ BAD: Inline complexity
stream.Map(func(r stream.Record) float64 {
    base := stream.GetOr(r, "salary", 0.0)
    bonus := stream.GetOr(r, "bonus", 0.0)
    return (base + bonus) * 1.15 // What does 1.15 mean?
})(records)

// ✅ GOOD: Named functions with clear intent
calculateTotalCompensation := func(r stream.Record) float64 {
    baseSalary := stream.GetOr(r, "salary", 0.0)
    bonusAmount := stream.GetOr(r, "bonus", 0.0)
    const TAX_MULTIPLIER = 1.15
    return (baseSalary + bonusAmount) * TAX_MULTIPLIER
}

stream.Map(calculateTotalCompensation)(records)
```

**Why this matters**: Self-documenting code reduces bugs and improves maintainability.

## 6. Test Your Stream Pipelines

Write tests for your stream processing logic:

```go
func TestSalaryAnalysis(t *testing.T) {
    // Arrange: Create test data
    testData := []stream.Record{
        stream.NewRecord().String("name", "Alice").Int("salary", 75000).Build(),
        stream.NewRecord().String("name", "Bob").Int("salary", 65000).Build(),
    }

    // Act: Run your stream pipeline
    avgSalary, err := stream.Avg(
        stream.Map(func(r stream.Record) float64 {
            return float64(stream.GetOr(r, "salary", 0))
        })(stream.FromSlice(testData)))

    // Assert: Verify results
    assert.NoError(t, err)
    assert.Equal(t, 70000.0, avgSalary)
}
```

**Why this matters**: Stream pipelines can be complex. Tests ensure they work correctly.

## 7. Performance Guidelines

### For Small Data (< 10,000 records)
- Use simple operations
- Don't worry about optimization
- Focus on code clarity

### For Large Data (> 100,000 records)
- Consider processing in chunks
- Use memory-efficient operations
- Monitor performance

```go
// Example: Processing large datasets efficiently
func processLargeDataset(data []stream.Record) {
    const CHUNK_SIZE = 10000

    for i := 0; i < len(data); i += CHUNK_SIZE {
        end := i + CHUNK_SIZE
        if end > len(data) {
            end = len(data)
        }

        chunk := data[i:end]
        result := stream.Map(expensiveTransformation)(
            stream.FromSlice(chunk))

        // Process chunk result...
    }
}
```

## 8. Common Patterns to Remember

### Data Cleaning
```go
cleanData := stream.Pipe(
    stream.Where(isValidRecord),      // Remove bad data
    stream.Map(normalizeFields),      // Standardize formats
    stream.Where(hasRequiredFields),  // Final validation
)
```

### Analysis Pipeline
```go
analysis := stream.Pipe(
    stream.Where(matchesCriteria),    // Filter relevant data
    stream.Map(extractMetrics),       // Transform for analysis
    // Then aggregate or collect
)
```

### Data Export
```go
export := stream.Pipe(
    stream.Map(formatForExport),      // Prepare for output format
    stream.Where(includeInExport),    // Final filtering
    // Then write to file/database
)
```

## Summary: Your StreamV2 Journey

You've learned how to:
- ✅ **Transform** data with Map
- ✅ **Filter** data with Where
- ✅ **Combine** operations with Pipe
- ✅ **Aggregate** data for insights
- ✅ **Process** real-world CSV and JSON
- ✅ **Build** sophisticated analysis pipelines
- ✅ **Write** production-ready code

### Key Principles to Remember:
1. **Start Simple**: Begin with basic operations, add complexity gradually
2. **Error Handling**: Always check errors in production code
3. **Readability**: Use Pipe and named functions for clarity
4. **Testing**: Write tests for your stream processing logic
5. **Performance**: Optimize when needed, but clarity first

You're now ready to build powerful data processing applications with StreamV2!

    // 3. Memory efficiency - process in chunks for large data
    largeData := make([]int64, 10000)
    for i := range largeData {
        largeData[i] = int64(i + 1)
    }
    
    start := time.Now()
    
    // Process in smaller chunks rather than all at once
    chunkSize := 1000
    totalSum := int64(0)
    
    for i := 0; i < len(largeData); i += chunkSize {
        end := i + chunkSize
        if end > len(largeData) {
            end = len(largeData)
        }
        
        chunk := largeData[i:end]
        chunkSum, _ := stream.Sum(
            stream.Map(func(x int64) int64 { return x * x })(
                stream.FromSlice(chunk)))
        
        totalSum += chunkSum
    }
    
    duration := time.Since(start)
    fmt.Printf("✅ Chunked processing of %d items: sum=%d in %v\n", 
               len(largeData), totalSum, duration)

    // 4. Type safety tips
    records := []stream.Record{
        {"name": "Alice", "age": 30, "salary": 75000},
        {"name": "Bob", "age": 25, "salary": 65000},
    }
    
    // Always use type assertions safely
    names, _ := stream.Collect(
        stream.Map(func(r stream.Record) string {
            // Safe type assertion with check
            if name, ok := r["name"].(string); ok {
                return name
            }
            return "Unknown"
        })(stream.FromSlice(records)))
    
    fmt.Printf("✅ Safe type assertions: %v\n", names)

    // 5. Reusable stream operations
    // Define reusable operations as variables
    doubleNumbers := stream.Map(func(x int64) int64 { return x * 2 })
    onlyEvens := stream.Where(func(x int64) bool { return x%2 == 0 })
    
    // Combine reusable operations
    processed, _ := stream.Collect(
        stream.Pipe(doubleNumbers, onlyEvens)(
            stream.FromSlice([]int64{1, 2, 3, 4, 5})))
    
    fmt.Printf("✅ Reusable operations: %v\n", processed)

    // 6. Performance measurement
    fmt.Printf("\n=== Performance Tips ===\n")
    
    // Measure operation performance
    testData := make([]int64, 1000)
    for i := range testData {
        testData[i] = int64(i + 1)
    }
    
    start = time.Now()
    simpleSum, _ := stream.Sum(stream.FromSlice(testData))
    simpleDuration := time.Since(start)
    
    start = time.Now()
    complexSum, _ := stream.Sum(
        stream.Map(func(x int64) int64 { return x })(  // Identity map
            stream.FromSlice(testData)))
    complexDuration := time.Since(start)
    
    fmt.Printf("Simple sum: %d in %v\n", simpleSum, simpleDuration)
    fmt.Printf("With map:   %d in %v\n", complexSum, complexDuration)
    fmt.Printf("Overhead:   %v (%.1fx slower)\n", 
               complexDuration-simpleDuration,
               float64(complexDuration)/float64(simpleDuration))

    fmt.Printf("\n=== Summary ===\n")
    fmt.Printf("✅ Always handle errors\n")
    fmt.Printf("✅ Use Pipe() for readable pipelines\n")
    fmt.Printf("✅ Process large data in chunks\n")
    fmt.Printf("✅ Use safe type assertions\n")
    fmt.Printf("✅ Create reusable operations\n")
    fmt.Printf("✅ Measure performance when needed\n")
    fmt.Printf("\nHappy streaming! 🚀\n")
}
```

**Key Takeaways:**
- Always check errors from stream operations
- Use `Pipe()` for readable multi-step operations
- Process large datasets in chunks to manage memory
- Use safe type assertions with records
- Create reusable stream operations as variables
- Measure performance for critical paths

---

# What's Next?

Congratulations! You've learned the fundamentals of stream processing with StreamV2. You can now:

- ✅ Create and transform streams
- ✅ Filter and aggregate data
- ✅ Process CSV and JSON files
- ✅ Build complex data analysis pipelines
- ✅ Apply best practices for performance and maintainability

## Further Exploration

- **🕐 [Advanced Windowing Codelab](ADVANCED_WINDOWING_CODELAB.md)**: Master event-time processing, session windows, watermarks, and real-time analytics
- **Parallel processing**: Automatic parallelization for large datasets
- **Custom aggregators**: Build your own aggregation functions
- **Integration patterns**: Connect streams to databases, APIs, and message queues

## Resources

- [API Documentation](./docs/api.md)
- [Performance Guide](./docs/performance.md)
- [Examples Repository](./examples/)
- [Contributing Guide](./CONTRIBUTING.md)

Start building amazing data processing applications with StreamV2! 🎉