package main

import (
	"fmt"
	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Printf("=== StreamV2 Sorting Operations ===\n\n")

	// Basic sorting with numbers
	numbers := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Printf("Original numbers: %v\n", numbers)

	// Sort ascending
	ascending, _ := stream.Collect(
		stream.SortAsc[int]()(
			stream.FromSlice(numbers)))
	fmt.Printf("Ascending: %v\n", ascending)

	// Sort descending
	descending, _ := stream.Collect(
		stream.SortDesc[int]()(
			stream.FromSlice(numbers)))
	fmt.Printf("Descending: %v\n\n", descending)

	// Sorting records by multiple fields
	students := []stream.Record{
		stream.NewRecord().String("name", "Alice").Int("grade", 85).String("class", "Math").Build(),
		stream.NewRecord().String("name", "Bob").Int("grade", 92).String("class", "Science").Build(),
		stream.NewRecord().String("name", "Alice").Int("grade", 78).String("class", "History").Build(),
		stream.NewRecord().String("name", "Charlie").Int("grade", 96).String("class", "Math").Build(),
	}

	fmt.Printf("Students sorted by name, then grade:\n")
	sorted, _ := stream.Collect(
		stream.SortBy("name", "grade")(
			stream.FromSlice(students)))

	for _, student := range sorted {
		name := stream.GetOr(student, "name", "")
		grade := stream.GetOr(student, "grade", int64(0))
		class := stream.GetOr(student, "class", "")
		fmt.Printf("- %s: %d (%s)\n", name, grade, class)
	}

	// Top-K example for streaming scenarios
	fmt.Printf("\nTop 3 highest grades:\n")
	allGrades := []int{85, 92, 78, 96, 81, 88, 94, 76, 90, 87}

	top3, _ := stream.Collect(
		stream.TopK(3, func(a, b int) int {
			return a - b // ascending comparison for max-heap
		})(stream.FromSlice(allGrades)))

	fmt.Printf("All grades: %v\n", allGrades)
	fmt.Printf("Top 3: %v\n", top3)

	// Windowed sorting for batch processing
	fmt.Printf("\nWindowed sorting (batches of 4):\n")
	windowed, _ := stream.Collect(
		stream.SortCountWindow(4, func(a, b int) int {
			return a - b // ascending
		})(stream.FromSlice(allGrades)))

	fmt.Printf("Original: %v\n", allGrades)
	fmt.Printf("Sorted in windows: %v\n", windowed)
}