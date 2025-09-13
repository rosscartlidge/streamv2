package stream

import (
	"testing"
	"time"
)

// TestSafeConstructors verifies the new Value-constrained constructors
func TestSafeConstructors(t *testing.T) {
	t.Run("SafeFromSlice", func(t *testing.T) {
		// These should all compile and work
		intStream := FromSlice([]int64{1, 2, 3, 4, 5})
		stringStream := FromSlice([]string{"a", "b", "c"})
		boolStream := FromSlice([]bool{true, false, true})
		floatStream := FromSlice([]float64{1.1, 2.2, 3.3})
		
		// Verify they work correctly
		val, err := intStream()
		if err != nil || val != 1 {
			t.Errorf("Expected first int to be 1, got %v, err %v", val, err)
		}
		
		str, err := stringStream()
		if err != nil || str != "a" {
			t.Errorf("Expected first string to be 'a', got %v, err %v", str, err)
		}
		
		bl, err := boolStream()
		if err != nil || bl != true {
			t.Errorf("Expected first bool to be true, got %v, err %v", bl, err)
		}
		
		fl, err := floatStream()
		if err != nil || fl != 1.1 {
			t.Errorf("Expected first float to be 1.1, got %v, err %v", fl, err)
		}
	})

	t.Run("SafeFrom", func(t *testing.T) {
		// Variadic constructor with type safety
		numbers := From(int64(1), int64(2), int64(3))
		words := From("hello", "world", "test")
		flags := From(true, false, true, false)
		
		// Collect all values
		numSlice, err := Collect(numbers)
		if err != nil {
			t.Errorf("Error collecting numbers: %v", err)
		}
		if len(numSlice) != 3 || numSlice[0] != 1 || numSlice[2] != 3 {
			t.Errorf("Expected [1,2,3], got %v", numSlice)
		}
		
		wordSlice, err := Collect(words)
		if err != nil {
			t.Errorf("Error collecting words: %v", err)
		}
		if len(wordSlice) != 3 || wordSlice[0] != "hello" {
			t.Errorf("Expected [hello,world,test], got %v", wordSlice)
		}
		
		flagSlice, err := Collect(flags)
		if err != nil {
			t.Errorf("Error collecting flags: %v", err)
		}
		if len(flagSlice) != 4 || flagSlice[0] != true {
			t.Errorf("Expected [true,false,true,false], got %v", flagSlice)
		}
	})

	t.Run("SafeOnce", func(t *testing.T) {
		// Single value streams with type safety
		singleInt := Once(int64(42))
		singleString := Once("hello")
		singleTime := Once(time.Now())
		
		val, err := singleInt()
		if err != nil || val != 42 {
			t.Errorf("Expected 42, got %v, err %v", val, err)
		}
		
		// Second call should return EOS
		_, err = singleInt()
		if err != EOS {
			t.Errorf("Expected EOS on second call, got %v", err)
		}
		
		str, err := singleString()
		if err != nil || str != "hello" {
			t.Errorf("Expected 'hello', got %v, err %v", str, err)
		}
		
		_, err = singleTime()
		if err != nil {
			t.Errorf("Expected time value, got err %v", err)
		}
	})

	t.Run("SafeFromRecords", func(t *testing.T) {
		// Test safe record stream creation
		records := []Record{
			NewRecord().String("name", "Alice").Int("age", 30).Build(),
			NewRecord().String("name", "Bob").Int("age", 25).Build(),
		}
		
		recordStream, err := FromRecords(records)
		if err != nil {
			t.Errorf("Expected no error from valid records, got %v", err)
		}
		
		// Verify we can consume the stream
		first, err := recordStream()
		if err != nil {
			t.Errorf("Error reading first record: %v", err)
		}
		if first["name"] != "Alice" {
			t.Errorf("Expected name=Alice, got %v", first["name"])
		}
	})

	t.Run("SafeFromMaps", func(t *testing.T) {
		// Test safe map to record conversion
		maps := []map[string]any{
			{"name": "Charlie", "age": int64(35), "active": true},
			{"name": "Diana", "age": int64(28), "active": false},
		}
		
		recordStream, err := FromMaps(maps)
		if err != nil {
			t.Errorf("Expected no error from valid maps, got %v", err)
		}
		
		records, err := Collect(recordStream)
		if err != nil {
			t.Errorf("Error collecting records: %v", err)
		}
		
		if len(records) != 2 {
			t.Errorf("Expected 2 records, got %d", len(records))
		}
		if records[0]["name"] != "Charlie" {
			t.Errorf("Expected name=Charlie, got %v", records[0]["name"])
		}
	})
}

// TestUnsafeConstructors verifies the escape hatch constructors
func TestUnsafeConstructors(t *testing.T) {
	t.Run("UnsafeFromSliceAny", func(t *testing.T) {
		// These work with any type, including non-Value types
		type CustomStruct struct {
			Field string
		}
		
		customStream := FromSliceAny([]CustomStruct{
			{Field: "test1"},
			{Field: "test2"},
		})
		
		val, err := customStream()
		if err != nil {
			t.Errorf("Error reading custom struct: %v", err)
		}
		if val.Field != "test1" {
			t.Errorf("Expected field=test1, got %v", val.Field)
		}
	})

	t.Run("UnsafeFromRecords", func(t *testing.T) {
		// Create records with invalid types (bypassing safety)
		badRecords := []Record{
			{"valid": "string", "invalid": make(chan int)}, // Channel is not a Value type
		}
		
		// Unsafe version should work without validation
		recordStream := FromRecordsUnsafe(badRecords)
		record, err := recordStream()
		if err != nil {
			t.Errorf("Unsafe constructor should work: %v", err)
		}
		if record["valid"] != "string" {
			t.Errorf("Expected valid field, got %v", record["valid"])
		}
	})
}

// TestConstructorValidation verifies that validation catches invalid types
func TestConstructorValidation(t *testing.T) {
	t.Run("InvalidRecordTypes", func(t *testing.T) {
		// Test that invalid Record fields are caught
		badRecords := []Record{
			{"name": "Alice", "invalid": make(chan int)}, // Channel not allowed
		}
		
		_, err := FromRecords(badRecords)
		if err == nil {
			t.Error("Expected error for invalid record type, got nil")
		}
		if err != nil && err.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("InvalidMapTypes", func(t *testing.T) {
		// Test that invalid map fields are caught
		badMaps := []map[string]any{
			{"name": "Alice", "invalid": func() {}}, // Function not allowed
		}
		
		_, err := FromMaps(badMaps)
		if err == nil {
			t.Error("Expected error for invalid map type, got nil")
		}
	})

	t.Run("ValidTypes", func(t *testing.T) {
		// Test that all valid Value types are accepted
		validMaps := []map[string]any{
			{
				"int":     int(1),
				"int8":    int8(2),
				"int16":   int16(3),
				"int32":   int32(4),
				"int64":   int64(5),
				"uint":    uint(6),
				"uint8":   uint8(7),
				"uint16":  uint16(8),
				"uint32":  uint32(9),
				"uint64":  uint64(10),
				"float32": float32(1.1),
				"float64": float64(2.2),
				"bool":    true,
				"string":  "test",
				"time":    time.Now(),
				"record":  Record{"nested": "value"},
			},
		}
		
		_, err := FromMaps(validMaps)
		if err != nil {
			t.Errorf("Expected no error for valid types, got %v", err)
		}
	})
}

// BenchmarkSafeVsUnsafe compares performance of safe vs unsafe constructors
func BenchmarkSafeVsUnsafe(b *testing.B) {
	data := make([]int64, 1000)
	for i := range data {
		data[i] = int64(i)
	}
	
	b.Run("SafeFromSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stream := FromSlice(data)
			// Consume a few elements to ensure it's working
			stream()
			stream()
			stream()
		}
	})

	b.Run("UnsafeFromSliceAny", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stream := FromSliceAny(data)
			// Consume a few elements to ensure it's working
			stream()
			stream()
			stream()
		}
	})

	records := []Record{
		NewRecord().String("name", "Alice").Int("age", 30).Build(),
		NewRecord().String("name", "Bob").Int("age", 25).Build(),
	}
	
	b.Run("SafeFromRecords", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stream, _ := FromRecords(records)
			stream()
			stream()
		}
	})

	b.Run("UnsafeFromRecords", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stream := FromRecordsUnsafe(records)
			stream()
			stream()
		}
	})
}

// TestStreamCompatibility verifies streams work with existing pipeline operations
func TestStreamCompatibility(t *testing.T) {
	t.Run("PipelineIntegration", func(t *testing.T) {
		// Create safe stream and use with existing filters
		numbers := FromSlice([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		
		// Use with existing pipeline operations
		evens := Where(func(x int64) bool { return x%2 == 0 })(numbers)
		doubled := Map(func(x int64) int64 { return x * 2 })(evens)
		
		result, err := Collect(doubled)
		if err != nil {
			t.Errorf("Pipeline error: %v", err)
		}
		
		expected := []int64{4, 8, 12, 16, 20} // 2*2, 4*2, 6*2, 8*2, 10*2
		if len(result) != len(expected) {
			t.Errorf("Expected %d results, got %d", len(expected), len(result))
		}
		for i, val := range result {
			if val != expected[i] {
				t.Errorf("Expected %d at index %d, got %d", expected[i], i, val)
			}
		}
	})
}