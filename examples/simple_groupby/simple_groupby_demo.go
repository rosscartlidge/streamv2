package main

import (
	"fmt"
	"log"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Println("🧪 Testing New Simplified GroupBy")
	fmt.Println("=================================")

	// Create test data
	users, _ := stream.FromRecords([]stream.Record{
		stream.NewRecord().String("name", "Alice").String("department", "engineering").Int("salary", 95000).Build(),
		stream.NewRecord().String("name", "Bob").String("department", "engineering").Int("salary", 87000).Build(),
		stream.NewRecord().String("name", "Charlie").String("department", "sales").Int("salary", 92000).Build(),
		stream.NewRecord().String("name", "Diana").String("department", "sales").Int("salary", 88000).Build(),
		stream.NewRecord().String("name", "Eve").String("department", "engineering").Int("salary", 91000).Build(),
	})

	fmt.Println("\n📊 Basic GroupBy (count only):")
	basicGrouped := stream.GroupBy([]string{"department"})(users)
	basicResults, _ := stream.Collect(basicGrouped)
	
	for _, result := range basicResults {
		fmt.Printf("  %s: %d people\n", 
			result["department"], 
			result["group_count"])
	}

	// Reset data for next test
	users, _ = stream.FromRecords([]stream.Record{
		stream.NewRecord().String("name", "Alice").String("department", "engineering").Int("salary", 95000).Build(),
		stream.NewRecord().String("name", "Bob").String("department", "engineering").Int("salary", 87000).Build(),
		stream.NewRecord().String("name", "Charlie").String("department", "sales").Int("salary", 92000).Build(),
		stream.NewRecord().String("name", "Diana").String("department", "sales").Int("salary", 88000).Build(),
		stream.NewRecord().String("name", "Eve").String("department", "engineering").Int("salary", 91000).Build(),
	})

	fmt.Println("\n💰 GroupBy with Salary Aggregations:")
	groupedWithStats := stream.GroupBy([]string{"department"}, 
		stream.FieldAvgSpec[int64]("avg_salary", "salary"),
		stream.FieldMinSpec[int64]("min_salary", "salary"),
		stream.FieldMaxSpec[int64]("max_salary", "salary"),
		stream.FieldSumSpec[int64]("total_salary", "salary"),
	)(users)

	statsResults, _ := stream.Collect(groupedWithStats)
	
	for _, result := range statsResults {
		fmt.Printf("  %s Department:\n", result["department"])
		fmt.Printf("    👥 People: %d\n", result["group_count"])
		fmt.Printf("    💰 Avg Salary: $%.0f\n", result["avg_salary"])
		fmt.Printf("    📊 Range: $%d - $%d\n", result["min_salary"], result["max_salary"])
		fmt.Printf("    💸 Total Payroll: $%d\n\n", result["total_salary"])
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}