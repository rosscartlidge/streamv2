package stream

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestMap tests the Map filter
func TestMap(t *testing.T) {
	t.Run("IntToString", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		mapped := Map(func(x int64) string {
			return fmt.Sprintf("num_%d", x)
		})(stream)
		
		results, err := Collect(mapped)
		if err != nil {
			t.Fatalf("Failed to collect mapped stream: %v", err)
		}
		
		expected := []string{"num_1", "num_2", "num_3", "num_4", "num_5"}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("DoubleInts", func(t *testing.T) {
		input := []int64{1, 2, 3}
		stream := FromSlice(input)
		
		doubled := Map(func(x int64) int64 {
			return x * 2
		})(stream)
		
		results, err := Collect(doubled)
		if err != nil {
			t.Fatalf("Failed to collect doubled stream: %v", err)
		}
		
		expected := []int64{2, 4, 6}
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
}

// TestWhere tests the Where filter
func TestWhere(t *testing.T) {
	t.Run("FilterEvenNumbers", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5, 6}
		stream := FromSlice(input)
		
		evens := Where(func(x int64) bool {
			return x%2 == 0
		})(stream)
		
		results, err := Collect(evens)
		if err != nil {
			t.Fatalf("Failed to collect filtered stream: %v", err)
		}
		
		expected := []int64{2, 4, 6}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("FilterStrings", func(t *testing.T) {
		input := []string{"apple", "banana", "cherry", "date"}
		stream := FromSlice(input)
		
		longWords := Where(func(s string) bool {
			return len(s) >= 6
		})(stream)
		
		results, err := Collect(longWords)
		if err != nil {
			t.Fatalf("Failed to collect filtered stream: %v", err)
		}
		
		expected := []string{"banana", "cherry"}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("NoMatches", func(t *testing.T) {
		input := []int64{1, 3, 5}
		stream := FromSlice(input)
		
		evens := Where(func(x int64) bool {
			return x%2 == 0
		})(stream)
		
		results, err := Collect(evens)
		if err != nil {
			t.Fatalf("Failed to collect filtered stream: %v", err)
		}
		
		if len(results) != 0 {
			t.Fatalf("Expected 0 results, got %d", len(results))
		}
	})
}

// TestTake tests the Take filter
func TestTake(t *testing.T) {
	t.Run("TakeFirst3", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		first3 := Take[int64](3)(stream)
		results, err := Collect(first3)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		expected := []int64{1, 2, 3}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("TakeZero", func(t *testing.T) {
		input := []int64{1, 2, 3}
		stream := FromSlice(input)
		
		none := Take[int64](0)(stream)
		results, err := Collect(none)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 0 {
			t.Fatalf("Expected 0 results, got %d", len(results))
		}
	})
	
	t.Run("TakeMoreThanAvailable", func(t *testing.T) {
		input := []int64{1, 2}
		stream := FromSlice(input)
		
		first10 := Take[int64](10)(stream)
		results, err := Collect(first10)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
		
		for i, result := range results {
			if result != input[i] {
				t.Errorf("Expected %v at position %d, got %v", input[i], i, result)
			}
		}
	})
}

// TestSkip tests the Skip filter
func TestSkip(t *testing.T) {
	t.Run("SkipFirst2", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		skip2 := Skip[int64](2)(stream)
		results, err := Collect(skip2)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		expected := []int64{3, 4, 5}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("SkipAll", func(t *testing.T) {
		input := []int64{1, 2, 3}
		stream := FromSlice(input)
		
		skipAll := Skip[int64](5)(stream)
		results, err := Collect(skipAll)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 0 {
			t.Fatalf("Expected 0 results, got %d", len(results))
		}
	})
	
	t.Run("SkipZero", func(t *testing.T) {
		input := []int64{1, 2, 3}
		stream := FromSlice(input)
		
		skipNone := Skip[int64](0)(stream)
		results, err := Collect(skipNone)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != len(input) {
			t.Fatalf("Expected %d results, got %d", len(input), len(results))
		}
		
		for i, result := range results {
			if result != input[i] {
				t.Errorf("Expected %v at position %d, got %v", input[i], i, result)
			}
		}
	})
}

// TestPipe tests the Pipe function
func TestPipe(t *testing.T) {
	t.Run("MapThenFilter", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		// Double then filter evens
		pipeline := Pipe(
			Map(func(x int64) int64 { return x * 2 }),
			Where(func(x int64) bool { return x > 4 }),
		)(stream)
		
		results, err := Collect(pipeline)
		if err != nil {
			t.Fatalf("Failed to collect pipeline: %v", err)
		}
		
		expected := []int64{6, 8, 10} // 2*3, 2*4, 2*5 where > 4
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("FilterThenTake", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5, 6, 7, 8}
		stream := FromSlice(input)
		
		// Filter evens then take first 2
		pipeline := Pipe(
			Where(func(x int64) bool { return x%2 == 0 }),
			Take[int64](2),
		)(stream)
		
		results, err := Collect(pipeline)
		if err != nil {
			t.Fatalf("Failed to collect pipeline: %v", err)
		}
		
		expected := []int64{2, 4}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
}

// TestPipe3 tests the Pipe3 function
func TestPipe3(t *testing.T) {
	t.Run("MapFilterTake", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		// Triple, filter > 5, take first 2
		pipeline := Pipe3(
			Map(func(x int64) int64 { return x * 3 }),
			Where(func(x int64) bool { return x > 5 }),
			Take[int64](2),
		)(stream)
		
		results, err := Collect(pipeline)
		if err != nil {
			t.Fatalf("Failed to collect pipeline: %v", err)
		}
		
		expected := []int64{6, 9} // 2*3, 3*3 where > 5, take 2
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
}

// TestChain tests the Chain function
func TestChain(t *testing.T) {
	t.Run("MultipleFilters", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		stream := FromSlice(input)
		
		chained := Chain(
			Where(func(x int64) bool { return x%2 == 0 }), // evens
			Where(func(x int64) bool { return x > 4 }),    // > 4
			Take[int64](2),                               // first 2
		)(stream)
		
		results, err := Collect(chained)
		if err != nil {
			t.Fatalf("Failed to collect chained filters: %v", err)
		}
		
		expected := []int64{6, 8}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
}

// TestSelect tests the Select filter
func TestSelect(t *testing.T) {
	t.Run("SelectSpecificFields", func(t *testing.T) {
		records := []Record{
			NewRecord().String("name", "Alice").Int("age", 30).String("city", "NYC").Build(),
			NewRecord().String("name", "Bob").Int("age", 25).String("city", "SF").Build(),
		}
		stream := FromRecordsUnsafe(records)
		
		selected := Select("name", "age")(stream)
		results, err := Collect(selected)
		if err != nil {
			t.Fatalf("Failed to collect selected fields: %v", err)
		}
		
		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
		
		// Check first record
		if GetOr(results[0], "name", "") != "Alice" {
			t.Errorf("Expected Alice, got %v", results[0]["name"])
		}
		if GetOr(results[0], "age", int64(0)) != 30 {
			t.Errorf("Expected 30, got %v", results[0]["age"])
		}
		if _, exists := results[0]["city"]; exists {
			t.Errorf("Expected city field to be removed")
		}
		
		// Check second record
		if GetOr(results[1], "name", "") != "Bob" {
			t.Errorf("Expected Bob, got %v", results[1]["name"])
		}
		if GetOr(results[1], "age", int64(0)) != 25 {
			t.Errorf("Expected 25, got %v", results[1]["age"])
		}
		if _, exists := results[1]["city"]; exists {
			t.Errorf("Expected city field to be removed")
		}
	})
	
	t.Run("SelectNonExistentField", func(t *testing.T) {
		records := []Record{
			NewRecord().String("name", "Alice").Build(),
		}
		stream := FromRecordsUnsafe(records)
		
		selected := Select("name", "email")(stream)
		results, err := Collect(selected)
		if err != nil {
			t.Fatalf("Failed to collect selected fields: %v", err)
		}
		
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		
		if GetOr(results[0], "name", "") != "Alice" {
			t.Errorf("Expected Alice, got %v", results[0]["name"])
		}
		
		// Non-existent field should not be in result
		if _, exists := results[0]["email"]; exists {
			t.Errorf("Expected email field to not exist")
		}
	})
}

// TestUpdate tests the Update filter
func TestUpdate(t *testing.T) {
	t.Run("AddField", func(t *testing.T) {
		records := []Record{
			NewRecord().String("name", "Alice").Int("age", 30).Build(),
			NewRecord().String("name", "Bob").Int("age", 25).Build(),
		}
		stream := FromRecordsUnsafe(records)
		
		updated := Update(func(r Record) Record {
			age := GetOr(r, "age", int64(0))
			category := "adult"
			if age < 18 {
				category = "minor"
			}
			return SetField(r, "category", category)
		})(stream)
		
		results, err := Collect(updated)
		if err != nil {
			t.Fatalf("Failed to collect updated records: %v", err)
		}
		
		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
		
		for _, result := range results {
			category := GetOr(result, "category", "")
			if category != "adult" {
				t.Errorf("Expected 'adult' category, got %v", category)
			}
		}
	})
	
	t.Run("ModifyField", func(t *testing.T) {
		records := []Record{
			NewRecord().String("name", "alice").Build(),
			NewRecord().String("name", "bob").Build(),
		}
		stream := FromRecordsUnsafe(records)
		
		updated := Update(func(r Record) Record {
			name := GetOr(r, "name", "")
			// Capitalize first letter
			if len(name) > 0 {
				name = string(name[0]-32) + name[1:]
			}
			return SetField(r, "name", name)
		})(stream)
		
		results, err := Collect(updated)
		if err != nil {
			t.Fatalf("Failed to collect updated records: %v", err)
		}
		
		expected := []string{"Alice", "Bob"}
		for i, result := range results {
			name := GetOr(result, "name", "")
			if name != expected[i] {
				t.Errorf("Expected %v, got %v", expected[i], name)
			}
		}
	})
}

// TestExtractField tests the ExtractField function
func TestExtractField(t *testing.T) {
	t.Run("ExtractString", func(t *testing.T) {
		records := []Record{
			NewRecord().String("name", "Alice").Int("age", 30).Build(),
			NewRecord().String("name", "Bob").Int("age", 25).Build(),
		}
		stream := FromRecordsUnsafe(records)
		
		names := ExtractField[string]("name")(stream)
		results, err := Collect(names)
		if err != nil {
			t.Fatalf("Failed to collect extracted field: %v", err)
		}
		
		expected := []string{"Alice", "Bob"}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("ExtractInt", func(t *testing.T) {
		records := []Record{
			NewRecord().String("name", "Alice").Int("age", 30).Build(),
			NewRecord().String("name", "Bob").Int("age", 25).Build(),
		}
		stream := FromRecordsUnsafe(records)
		
		ages := ExtractField[int64]("age")(stream)
		results, err := Collect(ages)
		if err != nil {
			t.Fatalf("Failed to collect extracted field: %v", err)
		}
		
		expected := []int64{30, 25}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
}

// TestWithContext tests the WithContext function
func TestWithContext(t *testing.T) {
	t.Run("ContextCancellation", func(t *testing.T) {
		input := Range(1, 1000000, 1) // Large range
		
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		
		contextualStream := WithContext(ctx, input)
		
		// This should terminate early due to context cancellation
		results, err := Collect(contextualStream)
		
		// We expect either an error due to context cancellation 
		// or fewer results than the full range
		if err == nil && len(results) >= 999999 {
			t.Errorf("Expected context cancellation to limit results, got %d", len(results))
		}
	})
}

// TestTee tests the Tee function
func TestTee(t *testing.T) {
	t.Run("TeeInto2", func(t *testing.T) {
		input := FromSlice([]int64{1, 2, 3, 4, 5})
		
		streams := Tee(input, 2)
		if len(streams) != 2 {
			t.Fatalf("Expected 2 streams, got %d", len(streams))
		}
		
		// Collect from first stream
		results1, err1 := Collect(streams[0])
		if err1 != nil {
			t.Fatalf("Failed to collect from first stream: %v", err1)
		}
		
		// Collect from second stream
		results2, err2 := Collect(streams[1])
		if err2 != nil {
			t.Fatalf("Failed to collect from second stream: %v", err2)
		}
		
		expected := []int64{1, 2, 3, 4, 5}
		
		// Both streams should have same data
		if len(results1) != len(expected) || len(results2) != len(expected) {
			t.Fatalf("Expected %d results from each stream, got %d and %d", 
				len(expected), len(results1), len(results2))
		}
		
		for i := range expected {
			if results1[i] != expected[i] {
				t.Errorf("Stream 1: expected %v at position %d, got %v", expected[i], i, results1[i])
			}
			if results2[i] != expected[i] {
				t.Errorf("Stream 2: expected %v at position %d, got %v", expected[i], i, results2[i])
			}
		}
	})
}

// TestFlatMap tests the FlatMap function
func TestFlatMap(t *testing.T) {
	t.Run("IntToRange", func(t *testing.T) {
		input := FromSlice([]int64{2, 3})
		
		// Each int becomes a range from 1 to that number
		flattened := FlatMap(func(x int64) Stream[int64] {
			return Range(1, x+1, 1)
		})(input)
		
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened stream: %v", err)
		}
		
		expected := []int64{1, 2, 1, 2, 3} // Range(1,3) + Range(1,4)
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("StringToChars", func(t *testing.T) {
		input := FromSlice([]string{"hi", "go"})
		
		// Each string becomes a stream of its characters
		flattened := FlatMap(func(s string) Stream[string] {
			chars := make([]string, len(s))
			for i, r := range s {
				chars[i] = string(r)
			}
			return FromSlice(chars)
		})(input)
		
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened stream: %v", err)
		}
		
		expected := []string{"h", "i", "g", "o"}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
}