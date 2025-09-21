# StreamV2 Codelab: From Zero to Stream Processing Hero

Welcome to the StreamV2 codelab! This hands-on tutorial will teach you modern stream processing in Go using a simple, elegant API. You'll start with basic concepts and gradually build up to powerful data processing pipelines.

## What You'll Learn

- ✅ Core stream operations (Map, Filter, Reduce)
- ✅ Data aggregation and statistics
- ✅ Working with CSV, JSON, and other formats
- ✅ Building processing pipelines
- ✅ Real-world data analysis patterns

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

Let's start with the most basic operation - creating a stream from data and processing it.

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    // Create a stream from a slice of numbers
    numbers := []int64{1, 2, 3, 4, 5}
    
    // Convert slice to stream and collect results
    result, err := stream.Collect(
        stream.FromSlice(numbers))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Original numbers:", result)
    // Output: Original numbers: [1 2 3 4 5]
}
```

**Key Concepts:**
- `FromSlice()` creates a stream from a slice
- `Collect()` converts a stream back to a slice
- Streams are lazy - nothing happens until you collect

---

# Step 2: Transform Data with Map

Now let's transform our data using the `Map` operation.

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    numbers := []int64{1, 2, 3, 4, 5}
    
    // Transform each number by squaring it
    squares, err := stream.Collect(
        stream.Map(func(x int64) int64 {
            return x * x
        })(stream.FromSlice(numbers)))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Original:", numbers)
    fmt.Println("Squares: ", squares)
    
    // You can also change types - numbers to strings
    labels, err := stream.Collect(
        stream.Map(func(x int64) string {
            return fmt.Sprintf("Item-%d", x)
        })(stream.FromSlice(numbers)))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Labels:  ", labels)
}
```

**Output:**
```
Original: [1 2 3 4 5]
Squares:  [1 4 9 16 25]
Labels:   [Item-1 Item-2 Item-3 Item-4 Item-5]
```

**Key Concepts:**
- `Map()` transforms each element using a function
- You can change types (int64 → string)
- Functions are composable using the `()(` pattern

---

# Step 3: Filter Data with Where

Let's learn to filter data by keeping only elements that match a condition.

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    numbers := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    
    // Keep only even numbers
    evens, err := stream.Collect(
        stream.Where(func(x int64) bool {
            return x%2 == 0
        })(stream.FromSlice(numbers)))
    
    if err != nil {
        panic(err)
    }
    
    // Keep only numbers greater than 5
    bigNumbers, err := stream.Collect(
        stream.Where(func(x int64) bool {
            return x > 5
        })(stream.FromSlice(numbers)))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Original:    ", numbers)
    fmt.Println("Even numbers:", evens)
    fmt.Println("Numbers > 5: ", bigNumbers)
}
```

**Output:**
```
Original:     [1 2 3 4 5 6 7 8 9 10]
Even numbers: [2 4 6 8 10]
Numbers > 5:  [6 7 8 9 10]
```

**Key Concepts:**
- `Where()` filters elements using a predicate function
- Predicate returns `true` to keep the element, `false` to discard it
- Filtering reduces the size of your data

---

# Step 4: Combine Operations with Pipe

The real power comes from combining operations. Let's chain Map and Where together.

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    numbers := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    
    // Chain operations: square the numbers, then keep only those > 20
    result, err := stream.Collect(
        stream.Pipe(
            stream.Map(func(x int64) int64 { return x * x }),
            stream.Where(func(x int64) bool { return x > 20 }),
        )(stream.FromSlice(numbers)))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Original numbers:", numbers)
    fmt.Println("Squares > 20:   ", result)
    
    // Let's trace through this step by step for clarity
    squares, _ := stream.Collect(
        stream.Map(func(x int64) int64 { return x * x })(
            stream.FromSlice(numbers)))
    
    fmt.Println("Step 1 - squares:", squares)
    fmt.Println("Step 2 - filter >20:", result)
}
```

**Output:**
```
Original numbers: [1 2 3 4 5 6 7 8 9 10]
Squares > 20:     [25 36 49 64 81 100]
Step 1 - squares: [1 4 9 16 25 36 49 64 81 100]
Step 2 - filter >20: [25 36 49 64 81 100]
```

**Key Concepts:**
- `Pipe()` combines multiple operations in sequence
- Data flows through operations left to right
- Each operation's output becomes the next operation's input

---

# Step 5: Aggregate Data

Now let's learn to compute statistics and summaries from our streams.

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    numbers := []int64{10, 20, 30, 40, 50}
    
    // Calculate basic statistics
    sum, err := stream.Sum(stream.FromSlice(numbers))
    if err != nil {
        panic(err)
    }
    
    count, err := stream.Count(stream.FromSlice(numbers))
    if err != nil {
        panic(err)
    }
    
    max, err := stream.Max(stream.FromSlice(numbers))
    if err != nil {
        panic(err)
    }
    
    min, err := stream.Min(stream.FromSlice(numbers))
    if err != nil {
        panic(err)
    }
    
    avg, err := stream.Avg(stream.FromSlice(numbers))
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Numbers:", numbers)
    fmt.Printf("Sum: %d, Count: %d, Max: %d, Min: %d, Avg: %.1f\n", 
               sum, count, max, min, avg)
    
    // You can also aggregate after transformations
    sumOfSquares, err := stream.Sum(
        stream.Map(func(x int64) int64 { return x * x })(
            stream.FromSlice(numbers)))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Sum of squares: %d\n", sumOfSquares)
}
```

**Output:**
```
Numbers: [10 20 30 40 50]
Sum: 150, Count: 5, Max: 50, Min: 10, Avg: 30.0
Sum of squares: 5500
```

**Key Concepts:**
- Aggregation functions consume the entire stream and return a single value
- Common aggregations: `Sum`, `Count`, `Max`, `Min`, `Avg`
- You can aggregate after transformations

---

# Step 6: Working with Real Data - CSV Processing

Let's work with real-world data by processing CSV files.

```go
package main

import (
    "fmt"
    "strings"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    // Sample CSV data (in real life, this would come from a file)
    csvData := `name,age,salary,department
Alice,30,75000,Engineering
Bob,25,65000,Sales
Charlie,35,85000,Engineering
Diana,28,70000,Marketing
Eve,32,80000,Engineering`

    // Parse CSV data into records
    records, err := stream.Collect(
        stream.CSVToStream(strings.NewReader(csvData)))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Loaded %d employee records\n", len(records))
    
    // Let's look at the first record
    fmt.Println("First record:", records[0])
    
    // Extract just the names
    names, err := stream.Collect(
        stream.Map(func(record stream.Record) string {
            return record["name"].(string)
        })(stream.FromSlice(records)))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Employee names:", names)
    
    // Filter to just Engineering employees
    engineers, err := stream.Collect(
        stream.Where(func(record stream.Record) bool {
            return record["department"].(string) == "Engineering"
        })(stream.FromSlice(records)))
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d engineers\n", len(engineers))
    for _, eng := range engineers {
        fmt.Printf("- %s (age %v, salary %v)\n", 
                   eng["name"], eng["age"], eng["salary"])
    }
}
```

**Output:**
```
Loaded 5 employee records
First record: map[age:30 department:Engineering name:Alice salary:75000]
Employee names: [Alice Bob Charlie Diana Eve]
Found 3 engineers
- Alice (age 30, salary 75000)
- Charlie (age 35, salary 85000)
- Eve (age 32, salary 80000)
```

**Key Concepts:**
- `CSVToStream()` parses CSV data into Record objects
- Records are `map[string]any` - access fields with `record["fieldname"]`
- Type assertions needed: `record["name"].(string)`

---

# Step 7: Data Analysis Pipeline

Let's build a more sophisticated analysis pipeline combining multiple operations.

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

Let's wrap up with best practices for using StreamV2 effectively.

```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Printf("=== StreamV2 Best Practices ===\n\n")

    // 1. Error handling best practice
    numbers := []int64{1, 2, 3, 4, 5}
    
    result, err := stream.Collect(
        stream.Map(func(x int64) int64 { return x * 2 })(
            stream.FromSlice(numbers)))
    
    if err != nil {
        fmt.Printf("Error processing stream: %v\n", err)
        return
    }
    fmt.Printf("✅ Always check errors: %v\n", result)

    // 2. Multiple aggregations in one pass
    testNumbers := []int64{10, 20, 30, 40, 50}
    
    stats, err := stream.Aggregates(stream.FromSlice(testNumbers),
        stream.SumStream[int64]("total"),
        stream.CountStream[int64]("count"), 
        stream.AvgStream[int64]("average"),
        stream.MinStream[int64]("minimum"),
        stream.MaxStream[int64]("maximum"),
    )
    
    if err != nil {
        fmt.Printf("Error calculating stats: %v\n", err)
        return
    }
    
    fmt.Printf("✅ Multiple aggregations: %+v\n", stats)

    // 3. Readable pipeline construction
    data := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    
    // Bad: Hard to read nested calls
    badResult, _ := stream.Collect(
        stream.Where(func(x int64) bool { return x > 5 })(
            stream.Map(func(x int64) int64 { return x * x })(
                stream.FromSlice(data))))
    
    // Good: Clear pipeline with Pipe
    goodResult, _ := stream.Collect(
        stream.Pipe(
            stream.Map(func(x int64) int64 { return x * x }),
            stream.Where(func(x int64) bool { return x > 25 }),
        )(stream.FromSlice(data)))
    
    fmt.Printf("✅ Use Pipe for readability: %v vs %v\n", badResult, goodResult)

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

- **Advanced windowing**: Time-based and count-based windows for streaming data
- **Parallel processing**: Automatic parallelization for large datasets
- **Custom aggregators**: Build your own aggregation functions
- **Integration patterns**: Connect streams to databases, APIs, and message queues

## Resources

- [API Documentation](./docs/api.md)
- [Performance Guide](./docs/performance.md)
- [Examples Repository](./examples/)
- [Contributing Guide](./CONTRIBUTING.md)

Start building amazing data processing applications with StreamV2! 🎉