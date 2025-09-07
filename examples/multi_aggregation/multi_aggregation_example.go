package main

import (
	"fmt"
	"log"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// MULTI-AGGREGATION EXAMPLE - SOLVING THE SINGLE-PASS PROBLEM
// ============================================================================

func main() {
	fmt.Println("📊 StreamV2 Multi-Aggregation Example")
	fmt.Println("======================================")

	demonstrateBasicMultiAggregation()
	demonstrateStreamSplitting()
	demonstrateGroupedAnalytics()
}

// ============================================================================
// BASIC MULTI-AGGREGATION
// ============================================================================

func demonstrateBasicMultiAggregation() {
	fmt.Println("\n🔢 Basic Multi-Aggregation")
	fmt.Println("--------------------------")

	// Create sample data
	scores := stream.FromSlice([]int64{95, 87, 92, 88, 91, 76, 84, 89, 93, 78})

	fmt.Println("Challenge: Streams are consumed once, so multiple aggregations need special handling:")
	fmt.Println("  sum, _ := stream.Sum(scores)    // ✅ Works")
	fmt.Println("  count, _ := stream.Count(scores) // ❌ Fails - stream already consumed!")

	fmt.Println("\nSolution 1: Generalized Aggregates - All stats in one pass")

	// Reset stream
	scores = stream.FromSlice([]int64{95, 87, 92, 88, 91, 76, 84, 89, 93, 78})
	stats, err := stream.Aggregates(scores,
		stream.CountSpec[int64]("count"),
		stream.SumSpec[int64]("sum"), 
		stream.MinSpec[int64]("min"),
		stream.MaxSpec[int64]("max"),
		stream.AvgSpec[int64]("avg"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("  📈 Count: %d\n", stats["count"])
	fmt.Printf("  📈 Sum:   %d\n", stats["sum"])
	fmt.Printf("  📈 Min:   %d\n", stats["min"])
	fmt.Printf("  📈 Max:   %d\n", stats["max"])
	fmt.Printf("  📈 Avg:   %.2f\n", stats["avg"])

	fmt.Println("\nSolution 2: Two specific aggregations")
	scores2 := stream.FromSlice([]int64{95, 87, 92, 88, 91, 76, 84, 89, 93, 78})
	results, err := stream.Aggregates(scores2,
		stream.SumSpec[int64]("total"),
		stream.CountSpec[int64]("count"),
	)
	if err != nil {
		log.Fatal(err)
	}
	sum := results["total"].(int64)
	count := results["count"].(int64)
	fmt.Printf("  📈 Sum and Count: %d, %d (avg: %.2f)\n", sum, count, float64(sum)/float64(count))

	fmt.Println("\nSolution 3: Single Average")
	scores3 := stream.FromSlice([]int64{95, 87, 92, 88, 91, 76, 84, 89, 93, 78})
	avgResult, err := stream.Aggregates(scores3, stream.AvgSpec[int64]("average"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  📈 Average: %.2f\n", avgResult["average"])
}

// ============================================================================
// STREAM SPLITTING WITH TEE
// ============================================================================

func demonstrateStreamSplitting() {
	fmt.Println("\n🔀 Stream Splitting with Tee")
	fmt.Println("----------------------------")

	fmt.Println("Problem: Need multiple different aggregations from same stream")
	fmt.Println("Solution: Tee splits one stream into multiple independent streams")

	scores := stream.FromSlice([]int64{95, 87, 92, 88, 91, 76, 84, 89, 93, 78})

	// Split into 4 independent streams
	streams := stream.Tee(scores, 4)
	fmt.Printf("Created %d independent streams from one source\n", len(streams))

	// Now we can do different aggregations on each stream
	sum, _ := stream.Sum(streams[0])
	count, _ := stream.Count(streams[1])
	min, _ := stream.Min(streams[2])
	max, _ := stream.Max(streams[3])

	fmt.Printf("  📈 Sum:   %d (from stream 1)\n", sum)
	fmt.Printf("  📈 Count: %d (from stream 2)\n", count)
	fmt.Printf("  📈 Min:   %d (from stream 3)\n", min)
	fmt.Printf("  📈 Max:   %d (from stream 4)\n", max)
	fmt.Printf("  📈 Avg:   %.2f (computed: sum/count)\n", float64(sum)/float64(count))
}

// ============================================================================
// GROUPED ANALYTICS WITH MULTI-AGGREGATION
// ============================================================================

func demonstrateGroupedAnalytics() {
	fmt.Println("\n📊 Grouped Analytics")
	fmt.Println("--------------------")

	// Create employee data
	employees := []struct {
		name   string
		dept   string
		salary int64
		years  int64
	}{
		{"Alice", "Engineering", 95000, 5},
		{"Bob", "Engineering", 87000, 3},
		{"Charlie", "Sales", 92000, 7},
		{"Diana", "Sales", 88000, 4},
		{"Eve", "Engineering", 91000, 6},
		{"Frank", "Sales", 85000, 2},
		{"Grace", "Engineering", 98000, 8},
	}

	fmt.Printf("Analyzing %d employees across departments\n", len(employees))

	// Group manually for this demo
	deptData := make(map[string][]int64)
	for _, emp := range employees {
		deptData[emp.dept] = append(deptData[emp.dept], emp.salary)
	}

	fmt.Println("\nDepartment Salary Statistics:")
	for dept, salaries := range deptData {
		salaryStream := stream.FromSlice(salaries)
		stats, err := stream.Aggregates(salaryStream,
			stream.CountSpec[int64]("count"),
			stream.SumSpec[int64]("total"),
			stream.AvgSpec[int64]("average"),
			stream.MinSpec[int64]("minimum"),
			stream.MaxSpec[int64]("maximum"),
		)
		if err != nil {
			continue
		}

		fmt.Printf("\n  %s Department:\n", dept)
		fmt.Printf("    👥 Employees: %d\n", stats["count"])
		fmt.Printf("    💰 Total Pay: $%d\n", stats["total"])
		fmt.Printf("    💰 Avg Salary: $%.0f\n", stats["average"])
		fmt.Printf("    💰 Min Salary: $%d\n", stats["minimum"])
		fmt.Printf("    💰 Max Salary: $%d\n", stats["maximum"])
		
		min := stats["minimum"].(int64)
		max := stats["maximum"].(int64)
		fmt.Printf("    💰 Range: $%d\n", max-min)
	}

	// Demonstrate with experience years using Tee
	fmt.Println("\nExperience Analysis using Tee:")
	allYears := make([]int64, len(employees))
	for i, emp := range employees {
		allYears[i] = emp.years
	}

	yearsStream := stream.FromSlice(allYears)
	yearStats, _ := stream.Aggregates(yearsStream,
		stream.AvgSpec[int64]("average"),
		stream.MinSpec[int64]("minimum"), 
		stream.MaxSpec[int64]("maximum"),
	)

	fmt.Printf("  📅 Average Experience: %.1f years\n", yearStats["average"])
	fmt.Printf("  📅 Experience Range: %d - %d years\n", yearStats["minimum"], yearStats["maximum"])
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}