package main

import (
	"fmt"
	"log"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// BASIC STREAMV2 EXAMPLE - DEMONSTRATE CORE FUNCTIONALITY
// ============================================================================

func main() {
	fmt.Println("🚀 StreamV2 Basic Example")
	fmt.Println("=========================")

	demonstrateBasicOperations()
	demonstrateRecordProcessing()
	demonstrateTypeConversion()
}

// ============================================================================
// BASIC STREAM OPERATIONS
// ============================================================================

func demonstrateBasicOperations() {
	fmt.Println("\n📊 Basic Stream Operations")
	fmt.Println("--------------------------")

	// Create a stream of numbers using native Go types
	numbers := stream.FromSlice([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// Beautiful functional composition with full type safety
	evenSquares := stream.Chain(
		stream.Where(func(x int64) bool { return x%2 == 0 }),
		stream.Map(func(x int64) int64 { return x * x }),
	)(numbers)

	// Type-safe aggregation
	sum, err := stream.Sum(evenSquares)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sum of even squares: %d\n", sum) // 4 + 16 + 36 + 64 + 100 = 220

	// More operations
	allNumbers := stream.Range(1, 11, 1)
	count, _ := stream.Count(allNumbers)
	fmt.Printf("Count of numbers 1-10: %d\n", count)

	// Find maximum
	moreNumbers := stream.FromSlice([]int64{42, 17, 83, 9, 56, 31})
	max, _ := stream.Max(moreNumbers)
	fmt.Printf("Maximum value: %d\n", max)

	// Collect results
	processedNumbers := stream.Chain(
		stream.Where(func(x int64) bool { return x > 5 }),
		stream.Map(func(x int64) int64 { return x * 2 }),
		stream.Take[int64](3),
	)(stream.FromSlice([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))

	results, _ := stream.Collect(processedNumbers)
	fmt.Printf("Processed numbers (>5, doubled, first 3): %v\n", results)
}

// ============================================================================
// RECORD PROCESSING - NATIVE GO TYPES
// ============================================================================

func demonstrateRecordProcessing() {
	fmt.Println("\n📋 Record Processing")
	fmt.Println("--------------------")

	// Create records with native Go types
	users := stream.FromRecords([]stream.Record{
		stream.R("id", 1, "name", "Alice", "age", 25, "score", 95.5, "active", true),
		stream.R("id", 2, "name", "Bob", "age", 30, "score", 87.2, "active", false),
		stream.R("id", 3, "name", "Charlie", "age", 35, "score", 92.1, "active", true),
		stream.R("id", 4, "name", "Diana", "age", 28, "score", 88.7, "active", true),
	})

	fmt.Println("Original users:")
	printRecords(users)

	// Reset stream for processing
	users = stream.FromRecords([]stream.Record{
		stream.R("id", 1, "name", "Alice", "age", 25, "score", 95.5, "active", true),
		stream.R("id", 2, "name", "Bob", "age", 30, "score", 87.2, "active", false),
		stream.R("id", 3, "name", "Charlie", "age", 35, "score", 92.1, "active", true),
		stream.R("id", 4, "name", "Diana", "age", 28, "score", 88.7, "active", true),
	})

	// Beautiful record processing with type safety
	processed := stream.Chain(
		// Filter active users
		stream.Where(func(r stream.Record) bool {
			return stream.GetOr(r, "active", false)
		}),
		// Add computed fields
		stream.Update(func(r stream.Record) stream.Record {
			score := stream.GetOr(r, "score", 0.0)
			grade := getGrade(score)
			return r.Set("grade", grade).Set("bonus", score*0.1)
		}),
		// Select specific fields
		stream.Select("name", "score", "grade", "bonus"),
	)(users)

	fmt.Println("\nProcessed users (active only, with grades):")
	printRecords(processed)

	// Extract typed data for analysis
	users = stream.FromRecords([]stream.Record{
		stream.R("id", 1, "name", "Alice", "age", 25, "score", 95.5, "active", true),
		stream.R("id", 2, "name", "Bob", "age", 30, "score", 87.2, "active", false),
		stream.R("id", 3, "name", "Charlie", "age", 35, "score", 92.1, "active", true),
		stream.R("id", 4, "name", "Diana", "age", 28, "score", 88.7, "active", true),
	})

	scores := stream.ExtractField[float64]("score")(users)
	avgScore, _ := stream.Sum(scores)
	scoreCount, _ := stream.Count(stream.ExtractField[float64]("score")(users))

	fmt.Printf("\nScore Analysis: Average = %.2f (from %d users)\n",
		avgScore/float64(scoreCount), scoreCount)
}

// ============================================================================
// TYPE CONVERSION MAGIC
// ============================================================================

func demonstrateTypeConversion() {
	fmt.Println("\n🔄 Automatic Type Conversion")
	fmt.Println("-----------------------------")

	// Create a record with mixed types
	testRecord := stream.R(
		"string_number", "42",
		"int_bool", 1,
		"float_string", 3.14159,
		"bool_string", true,
	)

	fmt.Println("Original record:", testRecord)

	// Automatic type conversion magic
	fmt.Printf("String '42' as int64: %d\n", stream.GetOr(testRecord, "string_number", int64(0)))
	fmt.Printf("Int 1 as bool: %v\n", stream.GetOr(testRecord, "int_bool", false))
	fmt.Printf("Float 3.14159 as string: %s\n", stream.GetOr(testRecord, "float_string", ""))
	fmt.Printf("Bool true as string: %s\n", stream.GetOr(testRecord, "bool_string", ""))

	// Show type conversion in streams
	mixedData := stream.FromRecords([]stream.Record{
		stream.R("value", "100"),
		stream.R("value", 200),
		stream.R("value", 300.5),
		stream.R("value", "400"),
	})

	// Extract as numbers - automatic conversion happens
	numbers := stream.ExtractField[int64]("value")(mixedData)
	numberList, _ := stream.Collect(numbers)
	fmt.Printf("Mixed data converted to int64: %v\n", numberList)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func printRecords(recordStream stream.Stream[stream.Record]) {
	records, err := stream.Collect(recordStream)
	if err != nil {
		log.Printf("Error collecting records: %v", err)
		return
	}

	for i, record := range records {
		fmt.Printf("%d: %v\n", i+1, record)
	}
}

func getGrade(score float64) string {
	if score >= 90 {
		return "A"
	} else if score >= 80 {
		return "B"
	} else if score >= 70 {
		return "C"
	} else if score >= 60 {
		return "D"
	}
	return "F"
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
