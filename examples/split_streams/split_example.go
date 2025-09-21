package main

import (
	"fmt"
	"log"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// STREAMING SPLIT EXAMPLE - ZERO-BUFFERING SUBSTREAM PROCESSING
// ============================================================================

func main() {
	fmt.Println("🔀 StreamV2 Split Example - Stream of Substreams")
	fmt.Println("================================================")

	demonstrateBasicSplit()
	demonstrateAdvancedProcessing()
	demonstrateLazyEvaluation()
}

// ============================================================================
// BASIC SPLIT DEMONSTRATION
// ============================================================================

func demonstrateBasicSplit() {
	fmt.Println("\n📊 Basic Split by Department")
	fmt.Println("----------------------------")

	// Employee data with mixed departments
	employees := []stream.Record{
		stream.NewRecord().String("name", "Alice").String("dept", "engineering").Int("salary", 95000).Build(),
		stream.NewRecord().String("name", "Bob").String("dept", "sales").Int("salary", 85000).Build(),
		stream.NewRecord().String("name", "Charlie").String("dept", "engineering").Int("salary", 87000).Build(),
		stream.NewRecord().String("name", "Diana").String("dept", "marketing").Int("salary", 78000).Build(),
		stream.NewRecord().String("name", "Eve").String("dept", "engineering").Int("salary", 92000).Build(),
		stream.NewRecord().String("name", "Frank").String("dept", "sales").Int("salary", 83000).Build(),
	}

	fmt.Printf("Processing %d employees across departments...\n", len(employees))

	employeeStream, err := stream.FromRecords(employees)
	if err != nil {
		panic(err)
	}
	
	// Split into department substreams - emits substreams as we discover new departments
	departmentStreams := stream.Split([]string{"dept"})(employeeStream)

	fmt.Println("\nProcessing each department substream:")

	// Process each department substream independently and collect results
	deptResults, _ := stream.Collect(
		stream.Map(func(deptStream stream.Stream[stream.Record]) string {
			// Each deptStream contains all employees from one department
			deptRecords, err := stream.Collect(deptStream)
			if err != nil {
				return "Error processing department"
			}

			if len(deptRecords) == 0 {
				return "Empty department"
			}

			// Get department name from first record
			deptName := stream.GetOr(deptRecords[0], "dept", "unknown")
			
			// Calculate department stats
			totalSalary := int64(0)
			for _, record := range deptRecords {
				salary := stream.GetOr(record, "salary", int64(0))
				totalSalary += salary
			}
			avgSalary := totalSalary / int64(len(deptRecords))

			return fmt.Sprintf("  %s: %d people, avg salary $%d", 
				deptName, len(deptRecords), avgSalary)
		})(departmentStreams),
	)

	for _, result := range deptResults {
		fmt.Println(result)
	}

	// Consume the results to trigger processing
	employeeStream2, err := stream.FromRecords(employees)
	if err != nil {
		panic(err)
	}
	departmentStreams2 := stream.Split([]string{"dept"})(employeeStream2)
	results, _ := stream.Collect(
		stream.Map(func(deptStream stream.Stream[stream.Record]) string {
			deptRecords, _ := stream.Collect(deptStream)
			if len(deptRecords) == 0 {
				return ""
			}
			deptName := stream.GetOr(deptRecords[0], "dept", "unknown")
			return deptName
		})(departmentStreams2),
	)

	fmt.Printf("\nDiscovered departments: %v\n", results)
}

// ============================================================================
// ADVANCED PROCESSING WITH DIFFERENT OPERATIONS PER GROUP
// ============================================================================

func demonstrateAdvancedProcessing() {
	fmt.Println("\n🔧 Advanced Processing - Different Logic Per Group")
	fmt.Println("--------------------------------------------------")

	// Order data with different statuses
	orders := []stream.Record{
		stream.NewRecord().Int("id", 1).String("status", "pending").Float("amount", 150.0).Build(),
		stream.NewRecord().Int("id", 2).String("status", "completed").Float("amount", 200.0).Build(),
		stream.NewRecord().Int("id", 3).String("status", "pending").Float("amount", 75.0).Build(),
		stream.NewRecord().Int("id", 4).String("status", "failed").Float("amount", 300.0).Build(),
		stream.NewRecord().Int("id", 5).String("status", "completed").Float("amount", 125.0).Build(),
		stream.NewRecord().Int("id", 6).String("status", "pending").Float("amount", 250.0).Build(),
	}

	orderStream, err := stream.FromRecords(orders)
	if err != nil {
		panic(err)
	}
	statusStreams := stream.Split([]string{"status"})(orderStream)

	fmt.Println("Processing orders by status with custom logic:")

	// Apply different processing based on order status
	results, _ := stream.Collect(
		stream.Map(func(statusStream stream.Stream[stream.Record]) string {
			// Get first record to determine status
			firstRecord, err := statusStream()
			if err != nil {
				return "Empty status group"
			}
			
			status := stream.GetOr(firstRecord, "status", "unknown")
			
			// Create stream that includes the first record we consumed
			fullStream := func() stream.Stream[stream.Record] {
				firstReturned := false
				return func() (stream.Record, error) {
					if !firstReturned {
						firstReturned = true
						return firstRecord, nil
					}
					return statusStream()
				}
			}()

			switch status {
			case "pending":
				// For pending orders: count and sum
				count, _ := stream.Count(fullStream)
				amounts := stream.ExtractField[float64]("amount")(fullStream)
				total, _ := stream.Sum(amounts)
				return fmt.Sprintf("  📋 Pending: %d orders, $%.0f total (needs processing)", count, total)
				
			case "completed":
				// For completed orders: calculate metrics
				amounts := stream.ExtractField[float64]("amount")(fullStream)
				stats, _ := stream.Aggregates(amounts,
					stream.CountStream[float64]("count"),
					stream.AvgStream[float64]("average"),
				)
				return fmt.Sprintf("  ✅ Completed: %d orders, $%.0f avg (revenue recognized)", 
					stats["count"], stats["average"])
					
			case "failed":
				// For failed orders: just count for analysis
				count, _ := stream.Count(fullStream)
				return fmt.Sprintf("  ❌ Failed: %d orders (needs investigation)", count)
				
			default:
				return fmt.Sprintf("  ❓ Unknown status '%s'", status)
			}
		})(statusStreams),
	)

	for _, result := range results {
		fmt.Println(result)
	}
}

// ============================================================================
// DEMONSTRATE LAZY EVALUATION AND STREAMING NATURE
// ============================================================================

func demonstrateLazyEvaluation() {
	fmt.Println("\n⚡ Lazy Evaluation - Streaming Without Full Buffering")
	fmt.Println("----------------------------------------------------")

	// Create a large dataset that would be expensive to buffer
	fmt.Println("Creating stream of 1000 records across 3 categories...")
	
	var records []stream.Record
	categories := []string{"A", "B", "C"}
	for i := 0; i < 1000; i++ {
		records = append(records, stream.NewRecord().
			Int("id", int64(i)).
			String("category", categories[i%3]).
			Float("value", float64(i*10)).
			Build())
	}

	recordStream, err := stream.FromRecords(records)
	if err != nil {
		panic(err)
	}
	categoryStreams := stream.Split([]string{"category"})(recordStream)

	fmt.Println("Processing each category stream independently (streaming, no buffering):")

	// Process only the first few records from each category to show streaming nature
	results, _ := stream.Collect(
		stream.Map(func(categoryStream stream.Stream[stream.Record]) string {
			category := ""
			total := 0.0
			count := 0
			
			// Take only first 5 records from this category stream
			limited := stream.Take[stream.Record](5)(categoryStream)
			
			err := stream.ForEach(func(record stream.Record) {
				if category == "" {
					category = stream.GetOr(record, "category", "unknown")
				}
				value := stream.GetOr(record, "value", 0.0)
				total += value
				count++
			})(limited)
			
			if err != nil {
				return "Error processing category"
			}

			return fmt.Sprintf("  Category %s: processed %d records, sum=%.0f (streaming!)", 
				category, count, total)
		})(categoryStreams),
	)

	for _, result := range results {
		fmt.Println(result)
	}

	fmt.Println("\n✨ Each substream was processed independently without buffering the entire dataset!")
}


func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}