package stream

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSortAsc(t *testing.T) {
	data := []int{5, 2, 8, 1, 9, 3}

	result, err := Collect(SortAsc[int]()(FromSlice(data)))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := []int{1, 2, 3, 5, 8, 9}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

func TestSortDesc(t *testing.T) {
	data := []int{5, 2, 8, 1, 9, 3}

	result, err := Collect(SortDesc[int]()(FromSlice(data)))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := []int{9, 8, 5, 3, 2, 1}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

func TestSortBy(t *testing.T) {
	data := []Record{
		NewRecord().String("name", "Charlie").Int("age", 25).Build(),
		NewRecord().String("name", "Alice").Int("age", 30).Build(),
		NewRecord().String("name", "Bob").Int("age", 20).Build(),
	}

	result, err := Collect(SortBy("name")(FromSlice(data)))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := []string{"Alice", "Bob", "Charlie"}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, record := range result {
		name := GetOr(record, "name", "")
		if name != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, name)
		}
	}
}

func TestTopK(t *testing.T) {
	data := []int{5, 2, 8, 1, 9, 3, 7, 4, 6}

	result, err := Collect(TopK(3, func(a, b int) int {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	})(FromSlice(data)))

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should get top 3 elements in descending order: [9, 8, 7]
	expected := []int{9, 8, 7}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

func TestSortCountWindow(t *testing.T) {
	data := []int{5, 2, 8, 1, 9, 3}

	result, err := Collect(SortCountWindow(3, func(a, b int) int {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	})(FromSlice(data)))

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should get two windows: [2, 5, 8] and [1, 3, 9]
	expected := []int{2, 5, 8, 1, 3, 9}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

func TestSortCustomComparison(t *testing.T) {
	data := []string{"apple", "Banana", "cherry", "Date"}

	// Case-insensitive sort
	result, err := Collect(Sort(func(a, b string) int {
		aLower := strings.ToLower(a)
		bLower := strings.ToLower(b)
		if aLower < bLower {
			return -1
		} else if aLower > bLower {
			return 1
		}
		return 0
	})(FromSlice(data)))

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := []string{"apple", "Banana", "cherry", "Date"}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, v)
		}
	}
}

func TestSortByDesc(t *testing.T) {
	data := []Record{
		NewRecord().String("name", "Alice").Int("age", 30).Build(),
		NewRecord().String("name", "Bob").Int("age", 25).Build(),
		NewRecord().String("name", "Charlie").Int("age", 35).Build(),
	}

	result, err := Collect(SortByDesc("age")(FromSlice(data)))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should be sorted by age descending: Charlie(35), Alice(30), Bob(25)
	expected := []string{"Charlie", "Alice", "Bob"}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, record := range result {
		name := GetOr(record, "name", "")
		if name != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, name)
		}
	}
}

func TestSortEmptyStream(t *testing.T) {
	data := []int{}

	result, err := Collect(SortAsc[int]()(FromSlice(data)))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d elements", len(result))
	}
}

func TestSortSingleElement(t *testing.T) {
	data := []int{42}

	result, err := Collect(SortDesc[int]()(FromSlice(data)))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 1 || result[0] != 42 {
		t.Errorf("Expected [42], got %v", result)
	}
}

func TestSortByMissingFields(t *testing.T) {
	data := []Record{
		NewRecord().String("name", "Alice").Int("age", 30).Build(),
		NewRecord().String("name", "Bob").Build(), // Missing age
		NewRecord().Int("age", 25).Build(),        // Missing name
	}

	result, err := Collect(SortBy("name", "age")(FromSlice(data)))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Records with missing fields should be sorted appropriately
	if len(result) != 3 {
		t.Errorf("Expected 3 records, got %d", len(result))
	}
}

func TestBottomK(t *testing.T) {
	data := []int{5, 2, 8, 1, 9, 3, 7, 4, 6}

	result, err := Collect(BottomK(3, func(a, b int) int {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	})(FromSlice(data)))

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should get bottom 3 elements in ascending order: [1, 2, 3]
	expected := []int{1, 2, 3}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

func TestSortErrorHandling(t *testing.T) {
	// Create a stream that returns an error
	errorStream := func() (int, error) {
		return 0, fmt.Errorf("test error")
	}

	result, err := Collect(SortAsc[int]()(errorStream))

	// Should return the error, not EOS
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result on error, got %v", result)
	}
}

func TestTopKWithFewerElements(t *testing.T) {
	data := []int{5, 2}

	result, err := Collect(TopK(5, func(a, b int) int {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	})(FromSlice(data)))

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should return all elements when K > stream length
	expected := []int{5, 2}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
}

func TestSortTimeWindow(t *testing.T) {
	// Simple test for SortTimeWindow functionality
	data := []int{5, 2, 8, 1}

	result, err := Collect(SortTimeWindow(100*time.Millisecond, func(a, b int) int {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	})(FromSlice(data)))

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Since we're using a simple slice (no real timing), this should sort the whole batch
	// The exact windowing behavior depends on timing, but we can at least verify the function works
	if len(result) == 0 {
		t.Errorf("Expected some results, got empty slice")
	}

	// Verify function doesn't panic and produces reasonable output
	t.Logf("SortTimeWindow result: %v", result)
}