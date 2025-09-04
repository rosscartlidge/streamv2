package main

import (
	"fmt"
	"log"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// NESTED STREAMS EXAMPLE - STREAMS AS RECORD FIELDS
// ============================================================================

func main() {
	fmt.Println("🔄 StreamV2 Nested Streams Example")
	fmt.Println("===================================")

	demonstrateBasicNestedStreams()
	demonstrateStreamExpansion()
	demonstrateGroupingWithStreams()
	demonstrateFlatMapping()
}

// ============================================================================
// BASIC NESTED STREAMS
// ============================================================================

func demonstrateBasicNestedStreams() {
	fmt.Println("\n📊 Basic Nested Streams")
	fmt.Println("-----------------------")

	// Create records with nested streams
	records := []stream.Record{
		createRecordWithStreams("user1", []int64{10, 20, 30}, []string{"a", "b"}),
		createRecordWithStreams("user2", []int64{40, 50}, []string{"c", "d", "e"}),
		createRecordWithStreams("user3", []int64{60, 70, 80, 90}, []string{"f"}),
	}

	fmt.Printf("Created %d records with nested streams\n", len(records))

	// Process each record's nested streams
	for i, record := range records {
		fmt.Printf("\nRecord %d (%s):\n", i+1, record["name"])

		// Extract and process nested number stream
		if numberStream, ok := stream.GetStream[int64](record, "numbers"); ok {
			sum, _ := stream.Sum(numberStream)
			fmt.Printf("  Sum of numbers: %d\n", sum)
		}

		// Extract and process nested string stream
		if stringStream, ok := stream.GetStream[string](record, "strings"); ok {
			strings, _ := stream.Collect(stringStream)
			fmt.Printf("  Strings: %v\n", strings)
		}

		// Show metadata
		fmt.Printf("  Record keys: %v\n", record.Keys())
	}
}

func createRecordWithStreams(name string, numbers []int64, strings []string) stream.Record {
	record := stream.R("name", name, "id", len(name))

	// Add streams to record
	stream.SetTypedStream(record, "numbers", stream.FromSlice(numbers))
	stream.SetTypedStream(record, "strings", stream.FromSlice(strings))

	return record
}

// ============================================================================
// STREAM EXPANSION
// ============================================================================

func demonstrateStreamExpansion() {
	fmt.Println("\n🔄 Stream Expansion")
	fmt.Println("-------------------")

	// Create records with nested streams of different lengths
	recordsWithStreams := stream.FromRecords([]stream.Record{
		createRecordWithStreams("dataset1", []int64{100, 200}, []string{"x", "y", "z"}),
		createRecordWithStreams("dataset2", []int64{300, 400, 500}, []string{"a", "b"}),
	})

	fmt.Println("Original records with nested streams:")
	printRecordsWithStreams(recordsWithStreams)

	// Reset stream
	recordsWithStreams = stream.FromRecords([]stream.Record{
		createRecordWithStreams("dataset1", []int64{100, 200}, []string{"x", "y", "z"}),
		createRecordWithStreams("dataset2", []int64{300, 400, 500}, []string{"a", "b"}),
	})

	// Expand streams into individual records using dot product (element-wise)
	expandedRecordsDot := stream.ExpandStreamsDot("numbers", "strings")(recordsWithStreams)

	fmt.Println("\nDot Product Expansion (element-wise pairing):")
	printRecords(expandedRecordsDot)
	
	// Reset stream for cross product test
	recordsWithStreams = stream.FromRecords([]stream.Record{
		createRecordWithStreams("dataset1", []int64{100, 200}, []string{"x", "y", "z"}),
		createRecordWithStreams("dataset2", []int64{300, 400, 500}, []string{"a", "b"}),
	})
	
	// Expand streams using cross product (Cartesian product)
	expandedRecordsCross := stream.ExpandStreamsCross("numbers", "strings")(recordsWithStreams)
	
	fmt.Println("\nCross Product Expansion (Cartesian product):")
	printRecords(expandedRecordsCross)
	
	// Demonstrate auto-detection (no field arguments)
	fmt.Println("\nAuto-Detection Expansion (finds all stream fields automatically):")
	recordWithMixedStreams := stream.R("name", "auto_test", "category", "demo")
	stream.SetTypedStream(recordWithMixedStreams, "values", stream.FromSlice([]int64{10, 20}))
	stream.SetTypedStream(recordWithMixedStreams, "tags", stream.FromSlice([]string{"fast", "reliable"}))
	stream.SetTypedStream(recordWithMixedStreams, "scores", stream.FromSlice([]float64{98.5, 87.3}))
	
	autoStream := stream.FromRecords([]stream.Record{recordWithMixedStreams})
	autoExpanded := stream.ExpandStreamsCross()(autoStream) // No arguments - auto-detects all streams!
	
	autoResults, _ := stream.Collect(autoExpanded)
	fmt.Printf("Auto-detected %d stream fields, generated %d combinations:\n", 3, len(autoResults))
	for i, r := range autoResults {
		fmt.Printf("  %d: %v\n", i+1, r)
		if i >= 7 { // Show first 8
			if len(autoResults) > 8 {
				fmt.Printf("  ... (%d more combinations)\n", len(autoResults)-8)
			}
			break
		}
	}
}

// ============================================================================
// GROUPING WITH STREAMS
// ============================================================================

func demonstrateGroupingWithStreams() {
	fmt.Println("\n📋 Grouping with Streams")
	fmt.Println("------------------------")

	// Create sample data for grouping
	users := stream.FromRecords([]stream.Record{
		stream.R("name", "Alice", "department", "engineering", "score", 95),
		stream.R("name", "Bob", "department", "engineering", "score", 87),
		stream.R("name", "Charlie", "department", "sales", "score", 92),
		stream.R("name", "Diana", "department", "sales", "score", 88),
		stream.R("name", "Eve", "department", "engineering", "score", 91),
	})

	// Group by department, creating streams of grouped records
	groupedByDept := stream.GroupByWithNestedStreams([]string{"department"})(users)

	groupedRecords, _ := stream.Collect(groupedByDept)
	fmt.Printf("Grouped into %d departments\n", len(groupedRecords))

	// Process each group's nested stream
	for _, group := range groupedRecords {
		dept := group["department"]
		count := group["group_count"]
		fmt.Printf("\nDepartment: %s (%d people)\n", dept, count)

		// Extract grouped records stream
		if groupStream, ok := stream.GetStream[stream.Record](group, "grouped_records"); ok {
			// Process each person in the group
			people, _ := stream.Collect(groupStream)
			for _, person := range people {
				fmt.Printf("  - %s (score: %v)\n", person["name"], person["score"])
			}

			// Calculate department statistics using multi-aggregation from the stream
			if groupStream2, ok := stream.GetStream[stream.Record](group, "grouped_records"); ok {
				// Extract scores and compute multiple statistics in one pass
				scoreStream := stream.ExtractField[int]("score")(groupStream2)
				stats, err := stream.MultiAggregate(scoreStream)
				if err == nil && stats.Count > 0 {
					fmt.Printf("  Average score: %.1f (min=%d, max=%d)\n", 
						stats.Avg, stats.Min, stats.Max)
				}
			}
		}
	}
}

// ============================================================================
// FLAT MAPPING WITH NESTED STREAMS
// ============================================================================

func demonstrateFlatMapping() {
	fmt.Println("\n🔀 Flat Mapping")
	fmt.Println("---------------")

	// Create a stream of records, each containing a stream of numbers
	recordsWithNumbers := stream.FromRecords([]stream.Record{
		createRecordWithStreams("batch1", []int64{1, 2, 3}, []string{}),
		createRecordWithStreams("batch2", []int64{4, 5}, []string{}),
		createRecordWithStreams("batch3", []int64{6, 7, 8, 9}, []string{}),
	})

	// FlatMap to extract all numbers from all nested streams
	allNumbers := stream.FlatMap(func(record stream.Record) stream.Stream[int64] {
		if numberStream, ok := stream.GetStream[int64](record, "numbers"); ok {
			return numberStream
		}
		return stream.FromSlice([]int64{}) // Empty stream if no numbers
	})(recordsWithNumbers)

	// Collect all flattened numbers
	numbers, _ := stream.Collect(allNumbers)
	fmt.Printf("Flattened numbers from all records: %v\n", numbers)

	// Calculate total
	allNumbers2 := stream.FlatMap(func(record stream.Record) stream.Stream[int64] {
		if numberStream, ok := stream.GetStream[int64](record, "numbers"); ok {
			return numberStream
		}
		return stream.FromSlice([]int64{})
	})(stream.FromRecords([]stream.Record{
		createRecordWithStreams("batch1", []int64{1, 2, 3}, []string{}),
		createRecordWithStreams("batch2", []int64{4, 5}, []string{}),
		createRecordWithStreams("batch3", []int64{6, 7, 8, 9}, []string{}),
	}))

	total, _ := stream.Sum(allNumbers2)
	fmt.Printf("Total sum: %d\n", total)

	// Demonstrate processing nested streams in parallel
	fmt.Println("\nParallel processing of nested streams:")

	recordsWithNumbers3 := stream.FromRecords([]stream.Record{
		createRecordWithStreams("dataset1", []int64{10, 20, 30}, []string{}),
		createRecordWithStreams("dataset2", []int64{40, 50, 60}, []string{}),
		createRecordWithStreams("dataset3", []int64{70, 80, 90}, []string{}),
	})

	// Process each record's nested stream in parallel
	processedRecords := stream.Parallel(2, func(record stream.Record) stream.Record {
		name := record["name"].(string)

		if numberStream, ok := stream.GetStream[int64](record, "numbers"); ok {
			// Process the nested stream
			doubled := stream.Map(func(x int64) int64 { return x * 2 })(numberStream)
			doubledList, _ := stream.Collect(doubled)

			// Return processed record
			return stream.R(
				"name", name+"_processed",
				"original_name", name,
				"doubled_numbers", doubledList,
			)
		}

		return record
	})(recordsWithNumbers3)

	processed, _ := stream.Collect(processedRecords)
	for _, record := range processed {
		fmt.Printf("  %s: %v\n", record["name"], record["doubled_numbers"])
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func printRecordsWithStreams(recordStream stream.RecordStream) {
	records, err := stream.Collect(recordStream)
	if err != nil {
		log.Printf("Error collecting records: %v", err)
		return
	}

	for i, record := range records {
		fmt.Printf("%d: %s (keys: %v)\n", i+1, record["name"], record.Keys())
	}
}

func printRecords(recordStream stream.RecordStream) {
	records, err := stream.Collect(recordStream)
	if err != nil {
		log.Printf("Error collecting records: %v", err)
		return
	}

	for i, record := range records {
		fmt.Printf("%d: %v\n", i+1, record)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
