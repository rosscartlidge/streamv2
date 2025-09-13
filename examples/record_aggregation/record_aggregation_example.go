package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// RECORD AGGREGATION EXAMPLE - AGGREGATING STREAM[RECORD] DIRECTLY
// ============================================================================

func main() {
	fmt.Println("📊 StreamV2 Record Aggregation Example")
	fmt.Println("======================================")

	demonstrateBasicRecordAggregation()
	demonstrateAdvancedRecordAnalytics()
	demonstrateComplexBusinessLogic()
}

// ============================================================================
// BASIC RECORD AGGREGATION
// ============================================================================

func demonstrateBasicRecordAggregation() {
	fmt.Println("\n📋 Basic Record Aggregation")
	fmt.Println("---------------------------")
	
	// Create employee data
	employees := []stream.Record{
		stream.NewRecord().String("name", "Alice").String("department", "engineering").Int("salary", 95000).Int("years", 5).Float("performance", 4.5).Build(),
		stream.NewRecord().String("name", "Bob").String("department", "engineering").Int("salary", 87000).Int("years", 3).Float("performance", 4.2).Build(),
		stream.NewRecord().String("name", "Charlie").String("department", "sales").Int("salary", 92000).Int("years", 7).Float("performance", 4.8).Build(),
		stream.NewRecord().String("name", "Diana").String("department", "sales").Int("salary", 88000).Int("years", 4).Float("performance", 4.1).Build(),
		stream.NewRecord().String("name", "Eve").String("department", "engineering").Int("salary", 91000).Int("years", 6).Float("performance", 4.6).Build(),
		stream.NewRecord().String("name", "Frank").String("department", "sales").Int("salary", 85000).Int("years", 2).Float("performance", 3.9).Build(),
		stream.NewRecord().String("name", "Grace").String("department", "marketing").Int("salary", 78000).Int("years", 3).Float("performance", 4.3).Build(),
	}
	
	fmt.Printf("Analyzing %d employee records\n", len(employees))
	
	// Department counter aggregator
	deptCountAgg := stream.Aggregator[stream.Record, map[string]int64, map[string]int64]{
		Initial: func() map[string]int64 { return make(map[string]int64) },
		Accumulate: func(acc map[string]int64, record stream.Record) map[string]int64 {
			if dept, ok := stream.Get[string](record, "department"); ok {
				acc[dept]++
			}
			return acc
		},
		Finalize: func(acc map[string]int64) map[string]int64 { return acc },
	}
	
	// Unique names collector
	uniqueNamesAgg := stream.Aggregator[stream.Record, map[string]bool, []string]{
		Initial: func() map[string]bool { return make(map[string]bool) },
		Accumulate: func(acc map[string]bool, record stream.Record) map[string]bool {
			if name, ok := stream.Get[string](record, "name"); ok {
				acc[name] = true
			}
			return acc
		},
		Finalize: func(acc map[string]bool) []string {
			names := make([]string, 0, len(acc))
			for name := range acc {
				names = append(names, name)
			}
			sort.Strings(names)
			return names
		},
	}
	
	recordStream := stream.FromRecords(employees)
	
	// Use Tee to split the stream for multiple aggregations
	streams := stream.Tee(recordStream, 3)
	
	// Run aggregations separately since they have different type signatures
	totalEmployees, err1 := stream.Count(streams[0])
	deptCounts, err2 := stream.AggregateWith(streams[1], deptCountAgg)
	employeeNames, err3 := stream.AggregateWith(streams[2], uniqueNamesAgg)
	
	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatal("Aggregation error")
	}
	
	fmt.Printf("  👥 Total Employees: %d\n", totalEmployees)
	fmt.Printf("  🏢 Department Breakdown: %v\n", deptCounts)
	fmt.Printf("  📝 All Employees: %v\n", employeeNames)
}

// ============================================================================
// ADVANCED RECORD ANALYTICS
// ============================================================================

func demonstrateAdvancedRecordAnalytics() {
	fmt.Println("\n📈 Advanced Record Analytics")
	fmt.Println("----------------------------")
	
	// Same employee data
	employees := []stream.Record{
		stream.NewRecord().String("name", "Alice").String("department", "engineering").Int("salary", 95000).Int("years", 5).Float("performance", 4.5).Build(),
		stream.NewRecord().String("name", "Bob").String("department", "engineering").Int("salary", 87000).Int("years", 3).Float("performance", 4.2).Build(),
		stream.NewRecord().String("name", "Charlie").String("department", "sales").Int("salary", 92000).Int("years", 7).Float("performance", 4.8).Build(),
		stream.NewRecord().String("name", "Diana").String("department", "sales").Int("salary", 88000).Int("years", 4).Float("performance", 4.1).Build(),
		stream.NewRecord().String("name", "Eve").String("department", "engineering").Int("salary", 91000).Int("years", 6).Float("performance", 4.6).Build(),
		stream.NewRecord().String("name", "Frank").String("department", "sales").Int("salary", 85000).Int("years", 2).Float("performance", 3.9).Build(),
		stream.NewRecord().String("name", "Grace").String("department", "marketing").Int("salary", 78000).Int("years", 3).Float("performance", 4.3).Build(),
	}
	
	// Salary statistics per department
	type SalaryStats struct {
		Total   int64
		Count   int64
		Average float64
		Min     int64
		Max     int64
	}
	
	salaryByDeptAgg := stream.Aggregator[stream.Record, map[string]*SalaryStats, map[string]*SalaryStats]{
		Initial: func() map[string]*SalaryStats { return make(map[string]*SalaryStats) },
		Accumulate: func(acc map[string]*SalaryStats, record stream.Record) map[string]*SalaryStats {
			dept, deptOk := stream.Get[string](record, "department")
			salary, salaryOk := stream.Get[int](record, "salary")
			
			if deptOk && salaryOk {
				sal64 := int64(salary)
				if stats, exists := acc[dept]; exists {
					stats.Total += sal64
					stats.Count++
					if sal64 < stats.Min {
						stats.Min = sal64
					}
					if sal64 > stats.Max {
						stats.Max = sal64
					}
					stats.Average = float64(stats.Total) / float64(stats.Count)
				} else {
					acc[dept] = &SalaryStats{
						Total:   sal64,
						Count:   1,
						Average: float64(sal64),
						Min:     sal64,
						Max:     sal64,
					}
				}
			}
			return acc
		},
		Finalize: func(acc map[string]*SalaryStats) map[string]*SalaryStats { return acc },
	}
	
	// High performers (performance > 4.5)
	highPerformersAgg := stream.Aggregator[stream.Record, []string, []string]{
		Initial: func() []string { return []string{} },
		Accumulate: func(acc []string, record stream.Record) []string {
			name, nameOk := stream.Get[string](record, "name")
			perf, perfOk := stream.Get[float64](record, "performance")
			
			if nameOk && perfOk && perf > 4.5 {
				acc = append(acc, name)
			}
			return acc
		},
		Finalize: func(acc []string) []string { 
			sort.Strings(acc)
			return acc 
		},
	}
	
	recordStream := stream.FromRecords(employees)
	
	// Use Tee to split the stream for multiple aggregations
	streams := stream.Tee(recordStream, 2)
	
	// Run aggregations separately
	salaryStats, err1 := stream.AggregateWith(streams[0], salaryByDeptAgg)
	highPerformers, err2 := stream.AggregateWith(streams[1], highPerformersAgg)
	
	if err1 != nil || err2 != nil {
		log.Fatal("Aggregation error")
	}
	
	fmt.Println("💰 Salary Statistics by Department:")
	for dept, stats := range salaryStats {
		fmt.Printf("  %s:\n", dept)
		fmt.Printf("    Total Payroll: $%d\n", stats.Total)
		fmt.Printf("    Employees: %d\n", stats.Count)
		fmt.Printf("    Average: $%.0f\n", stats.Average)
		fmt.Printf("    Range: $%d - $%d\n", stats.Min, stats.Max)
	}
	
	fmt.Printf("\n🌟 High Performers (>4.5): %v\n", highPerformers)
}

// ============================================================================
// COMPLEX BUSINESS LOGIC
// ============================================================================

func demonstrateComplexBusinessLogic() {
	fmt.Println("\n🔧 Complex Business Logic")
	fmt.Println("-------------------------")
	
	// Transaction data
	transactions := []stream.Record{
		stream.NewRecord().String("id", "T001").String("customer", "Alice").Float("amount", 1250.50).String("category", "software").String("date", "2024-01-15").String("status", "completed").Build(),
		stream.NewRecord().String("id", "T002").String("customer", "Bob").Float("amount", 2100.00).String("category", "hardware").String("date", "2024-01-16").String("status", "completed").Build(),
		stream.NewRecord().String("id", "T003").String("customer", "Alice").Float("amount", 850.25).String("category", "software").String("date", "2024-01-17").String("status", "pending").Build(),
		stream.NewRecord().String("id", "T004").String("customer", "Charlie").Float("amount", 3200.75).String("category", "consulting").String("date", "2024-01-18").String("status", "completed").Build(),
		stream.NewRecord().String("id", "T005").String("customer", "Bob").Float("amount", 1800.00).String("category", "hardware").String("date", "2024-01-19").String("status", "failed").Build(),
		stream.NewRecord().String("id", "T006").String("customer", "Diana").Float("amount", 950.00).String("category", "software").String("date", "2024-01-20").String("status", "completed").Build(),
	}
	
	fmt.Printf("Analyzing %d transaction records\n", len(transactions))
	
	// Complex customer analysis
	type CustomerProfile struct {
		Name           string
		TotalSpent     float64
		TransactionCount int64
		CompletedCount   int64
		FailedCount      int64
		PendingCount     int64
		Categories       map[string]bool
		AverageAmount    float64
		Status          string // "VIP", "Regular", "At Risk"
	}
	
	customerAnalysisAgg := stream.Aggregator[stream.Record, map[string]*CustomerProfile, map[string]*CustomerProfile]{
		Initial: func() map[string]*CustomerProfile { return make(map[string]*CustomerProfile) },
		Accumulate: func(acc map[string]*CustomerProfile, record stream.Record) map[string]*CustomerProfile {
			customer, customerOk := stream.Get[string](record, "customer")
			amount, amountOk := stream.Get[float64](record, "amount")
			category, categoryOk := stream.Get[string](record, "category")
			status, statusOk := stream.Get[string](record, "status")
			
			if customerOk && amountOk && categoryOk && statusOk {
				profile, exists := acc[customer]
				if !exists {
					profile = &CustomerProfile{
						Name:       customer,
						Categories: make(map[string]bool),
					}
					acc[customer] = profile
				}
				
				profile.TransactionCount++
				profile.Categories[category] = true
				
				switch status {
				case "completed":
					profile.CompletedCount++
					profile.TotalSpent += amount
				case "failed":
					profile.FailedCount++
				case "pending":
					profile.PendingCount++
				}
				
				// Calculate average and status
				if profile.CompletedCount > 0 {
					profile.AverageAmount = profile.TotalSpent / float64(profile.CompletedCount)
				}
				
				// Determine status
				if profile.TotalSpent > 3000 {
					profile.Status = "VIP"
				} else if profile.FailedCount > 0 {
					profile.Status = "At Risk"
				} else {
					profile.Status = "Regular"
				}
			}
			return acc
		},
		Finalize: func(acc map[string]*CustomerProfile) map[string]*CustomerProfile { return acc },
	}
	
	// Revenue by category (completed only)
	revenueByCategoryAgg := stream.Aggregator[stream.Record, map[string]float64, map[string]float64]{
		Initial: func() map[string]float64 { return make(map[string]float64) },
		Accumulate: func(acc map[string]float64, record stream.Record) map[string]float64 {
			category, categoryOk := stream.Get[string](record, "category")
			amount, amountOk := stream.Get[float64](record, "amount")
			status, statusOk := stream.Get[string](record, "status")
			
			if categoryOk && amountOk && statusOk && status == "completed" {
				acc[category] += amount
			}
			return acc
		},
		Finalize: func(acc map[string]float64) map[string]float64 { return acc },
	}
	
	recordStream := stream.FromRecords(transactions)
	
	// Use Tee to split the stream for multiple aggregations
	streams := stream.Tee(recordStream, 2)
	
	// Run aggregations separately
	profiles, err1 := stream.AggregateWith(streams[0], customerAnalysisAgg)
	revenue, err2 := stream.AggregateWith(streams[1], revenueByCategoryAgg)
	
	if err1 != nil || err2 != nil {
		log.Fatal("Aggregation error")
	}
	
	// Display customer profiles
	fmt.Println("\n👤 Customer Profiles:")
	for _, profile := range profiles {
		categories := make([]string, 0, len(profile.Categories))
		for cat := range profile.Categories {
			categories = append(categories, cat)
		}
		sort.Strings(categories)
		
		fmt.Printf("  %s (%s):\n", profile.Name, profile.Status)
		fmt.Printf("    Total Spent: $%.2f\n", profile.TotalSpent)
		fmt.Printf("    Transactions: %d (✅%d ❌%d ⏳%d)\n", 
			profile.TransactionCount, profile.CompletedCount, profile.FailedCount, profile.PendingCount)
		fmt.Printf("    Average Amount: $%.2f\n", profile.AverageAmount)
		fmt.Printf("    Categories: %v\n", categories)
	}
	
	// Display revenue breakdown
	fmt.Println("\n💰 Revenue by Category:")
	var total float64
	for category, amount := range revenue {
		fmt.Printf("  %s: $%.2f\n", category, amount)
		total += amount
	}
	fmt.Printf("  Total Revenue: $%.2f\n", total)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}