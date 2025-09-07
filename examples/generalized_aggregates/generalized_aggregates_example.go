package main

import (
	"fmt"
	"log"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// GENERALIZED AGGREGATES EXAMPLE - MIXING STANDARD & CUSTOM AGGREGATORS
// ============================================================================

func main() {
	fmt.Println("🔢 StreamV2 Generalized Aggregates Example")
	fmt.Println("==========================================")

	demonstrateBasicAggregates()
	demonstrateCustomAggregates()
	demonstrateRealWorldExample()
}

// ============================================================================
// BASIC AGGREGATES WITH NAMED RESULTS
// ============================================================================

func demonstrateBasicAggregates() {
	fmt.Println("\n📊 Basic Aggregates with Named Results")
	fmt.Println("--------------------------------------")
	
	scores := []int64{95, 87, 92, 88, 91, 76, 84, 89}
	fmt.Printf("Test scores: %v\n", scores)
	
	stream1 := stream.FromSlice(scores)
	
	results, err := stream.Aggregates(stream1,
		stream.SumSpec[int64]("total_points"),
		stream.CountSpec[int64]("student_count"),
		stream.AvgSpec[int64]("class_average"),
		stream.MinSpec[int64]("lowest_score"),
		stream.MaxSpec[int64]("highest_score"),
	)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("  📈 Total Points: %v\n", results["total_points"])
	fmt.Printf("  📈 Students: %v\n", results["student_count"])
	fmt.Printf("  📈 Class Average: %.1f\n", results["class_average"])
	fmt.Printf("  📈 Grade Range: %v - %v\n", results["lowest_score"], results["highest_score"])
}

// ============================================================================
// MIXING STANDARD AND CUSTOM AGGREGATORS
// ============================================================================

func demonstrateCustomAggregates() {
	fmt.Println("\n🔧 Mixing Standard & Custom Aggregators")
	fmt.Println("---------------------------------------")
	
	numbers := []int64{2, 3, 4, 5}
	fmt.Printf("Numbers: %v\n", numbers)
	
	// Create custom aggregators
	productAgg := stream.Aggregator[int64, int64, int64]{
		Initial:    func() int64 { return 1 },
		Accumulate: func(acc int64, val int64) int64 { return acc * val },
		Finalize:   func(acc int64) int64 { return acc },
	}
	
	// Sum of squares aggregator
	sumOfSquaresAgg := stream.Aggregator[int64, int64, int64]{
		Initial:    func() int64 { return 0 },
		Accumulate: func(acc int64, val int64) int64 { return acc + (val * val) },
		Finalize:   func(acc int64) int64 { return acc },
	}
	
	stream1 := stream.FromSlice(numbers)
	
	results, err := stream.Aggregates(stream1,
		stream.SumSpec[int64]("sum"),                              // Standard
		stream.CountSpec[int64]("count"),                          // Standard
		stream.CustomSpec("product", productAgg),                  // Custom
		stream.CustomSpec("sum_of_squares", sumOfSquaresAgg),     // Custom
		stream.AvgSpec[int64]("mean"),                            // Standard
	)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("  📊 Sum: %v\n", results["sum"])
	fmt.Printf("  📊 Count: %v\n", results["count"])
	fmt.Printf("  📊 Product: %v\n", results["product"])
	fmt.Printf("  📊 Sum of Squares: %v\n", results["sum_of_squares"])
	fmt.Printf("  📊 Mean: %.1f\n", results["mean"])
	
	// Calculate variance using the results
	mean := results["mean"].(float64)
	sumSquares := float64(results["sum_of_squares"].(int64))
	count := float64(results["count"].(int64))
	variance := (sumSquares / count) - (mean * mean)
	fmt.Printf("  📊 Variance: %.2f (computed from results)\n", variance)
}

// ============================================================================
// REAL-WORLD EXAMPLE: SALES ANALYTICS
// ============================================================================

func demonstrateRealWorldExample() {
	fmt.Println("\n💼 Real-World Example: Sales Analytics")
	fmt.Println("--------------------------------------")
	
	// Sales data for the month
	sales := []float64{1250.50, 2100.00, 850.25, 3200.75, 1800.00, 950.00, 2750.25}
	fmt.Printf("Daily sales this week: $%.2f, $%.2f, $%.2f, $%.2f, $%.2f, $%.2f, $%.2f\n", 
		sales[0], sales[1], sales[2], sales[3], sales[4], sales[5], sales[6])
	
	// Custom aggregators for business metrics
	
	// Commission calculator (5% of total)
	commissionAgg := stream.Aggregator[float64, float64, float64]{
		Initial:    func() float64 { return 0 },
		Accumulate: func(acc float64, val float64) float64 { return acc + val },
		Finalize:   func(acc float64) float64 { return acc * 0.05 }, // 5% commission
	}
	
	// Above-target counter (target: $1500/day)
	aboveTargetAgg := stream.Aggregator[float64, int64, int64]{
		Initial:    func() int64 { return 0 },
		Accumulate: func(acc int64, val float64) int64 {
			if val > 1500.0 {
				return acc + 1
			}
			return acc
		},
		Finalize: func(acc int64) int64 { return acc },
	}
	
	salesStream := stream.FromSlice(sales)
	
	results, err := stream.Aggregates(salesStream,
		stream.SumSpec[float64]("total_revenue"),
		stream.CountSpec[float64]("sales_days"),
		stream.AvgSpec[float64]("daily_average"),
		stream.MinSpec[float64]("worst_day"),
		stream.MaxSpec[float64]("best_day"),
		stream.CustomSpec("commission_earned", commissionAgg),
		stream.CustomSpec("days_above_target", aboveTargetAgg),
	)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("\n📈 Weekly Sales Report:\n")
	fmt.Printf("  💰 Total Revenue: $%.2f\n", results["total_revenue"])
	fmt.Printf("  📅 Sales Days: %v\n", results["sales_days"])
	fmt.Printf("  💸 Daily Average: $%.2f\n", results["daily_average"])
	fmt.Printf("  📉 Worst Day: $%.2f\n", results["worst_day"])
	fmt.Printf("  📈 Best Day: $%.2f\n", results["best_day"])
	fmt.Printf("  💵 Commission Earned: $%.2f\n", results["commission_earned"])
	fmt.Printf("  🎯 Days Above Target ($1500): %v/7\n", results["days_above_target"])
	
	// Calculate performance metrics
	targetDays := results["days_above_target"].(int64)
	totalDays := results["sales_days"].(int64)
	performance := float64(targetDays) / float64(totalDays) * 100
	fmt.Printf("  📊 Performance: %.1f%% of days above target\n", performance)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}