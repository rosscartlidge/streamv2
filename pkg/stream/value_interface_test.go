package stream

import (
	"testing"
	"time"
)

// TestValueInterfaceConstraints verifies compile-time type safety
func TestValueInterfaceConstraints(t *testing.T) {
	t.Run("TypeSafeRecordBuilder", func(t *testing.T) {
		// All these should compile without errors
		user := NewRecord().
			String("name", "Alice").                 // string - type safe
			Int("age", 30).                         // int64 - type safe
			Float("salary", 75000.50).              // float64 - type safe
			Bool("active", true).                   // bool - type safe
			Time("created", time.Now()).            // time.Time - type safe
			Record("tags", NewRecord().             // nested Record - type safe
				String("level", "senior").
				String("department", "engineering").
				Build()).
			Build()

		// Verify values are stored correctly
		if user["name"] != "Alice" {
			t.Errorf("Expected name=Alice, got %v", user["name"])
		}
		if user["age"] != int64(30) {
			t.Errorf("Expected age=30, got %v", user["age"])
		}
		if user["salary"] != 75000.50 {
			t.Errorf("Expected salary=75000.50, got %v", user["salary"])
		}
		if user["active"] != true {
			t.Errorf("Expected active=true, got %v", user["active"])
		}

		// Check nested record
		tags, ok := user["tags"].(Record)
		if !ok {
			t.Error("Expected tags to be a Record")
		}
		if tags["level"] != "senior" {
			t.Errorf("Expected level=senior, got %v", tags["level"])
		}
	})

	t.Run("TypeSafeFieldConstructor", func(t *testing.T) {
		// Single field creation with type safety
		nameField := Field("name", "Bob")
		ageField := Field("age", int64(25))
		salaryField := Field("salary", 50000.0)

		if nameField["name"] != "Bob" {
			t.Errorf("Expected name=Bob, got %v", nameField["name"])
		}
		if ageField["age"] != int64(25) {
			t.Errorf("Expected age=25, got %v", ageField["age"])
		}
		if salaryField["salary"] != 50000.0 {
			t.Errorf("Expected salary=50000.0, got %v", salaryField["salary"])
		}
	})

	t.Run("TypeSafeRecordSet", func(t *testing.T) {
		// Create record with type-safe methods
		record := NewRecord().
			String("title", "Manager").
			Int("experience", 5).
			Float("rating", 4.8).
			Bool("certified", true).
			Build()

		if record["title"] != "Manager" {
			t.Errorf("Expected title=Manager, got %v", record["title"])
		}
		if record["experience"] != int64(5) {
			t.Errorf("Expected experience=5, got %v", record["experience"])
		}
		if record["rating"] != 4.8 {
			t.Errorf("Expected rating=4.8, got %v", record["rating"])
		}
		if record["certified"] != true {
			t.Errorf("Expected certified=true, got %v", record["certified"])
		}
	})

	t.Run("StreamTypesSupported", func(t *testing.T) {
		// Test that Stream types are supported via Set method
		intStream := FromSlice([]int{1, 2, 3})
		stringStream := FromSlice([]string{"a", "b", "c"})
		
		record := NewRecord().
			Set("numbers", intStream).               // Stream[int]
			Set("letters", stringStream).            // Stream[string]
			Build()

		// Verify streams are stored
		if _, ok := record["numbers"].(Stream[int]); !ok {
			t.Error("Expected numbers to be Stream[int]")
		}
		if _, ok := record["letters"].(Stream[string]); !ok {
			t.Error("Expected letters to be Stream[string]")
		}
	})

	t.Run("AllIntegerTypes", func(t *testing.T) {
		// Test all integer types are supported via Set method
		record := NewRecord().
			Set("int", int(1)).
			Set("int8", int8(2)).
			Set("int16", int16(3)).
			Set("int32", int32(4)).
			Set("int64", int64(5)).
			Set("uint", uint(6)).
			Set("uint8", uint8(7)).
			Set("uint16", uint16(8)).
			Set("uint32", uint32(9)).
			Set("uint64", uint64(10)).
			Build()

		if record["int"] != int(1) {
			t.Errorf("Expected int=1, got %v", record["int"])
		}
		if record["uint64"] != uint64(10) {
			t.Errorf("Expected uint64=10, got %v", record["uint64"])
		}
	})

	t.Run("AllFloatTypes", func(t *testing.T) {
		// Test all float types are supported via Set method
		record := NewRecord().
			Set("float32", float32(3.14)).
			Set("float64", float64(2.718)).
			Build()

		if record["float32"] != float32(3.14) {
			t.Errorf("Expected float32=3.14, got %v", record["float32"])
		}
		if record["float64"] != float64(2.718) {
			t.Errorf("Expected float64=2.718, got %v", record["float64"])
		}
	})
}

// TestNewRecordBuilder verifies new record builder works correctly
func TestNewRecordBuilder(t *testing.T) {
	t.Run("NewRecordFunction", func(t *testing.T) {
		// New NewRecord() builder function should work
		user := NewRecord().String("name", "Charlie").Int("age", 40).Bool("active", true).Build()

		if user["name"] != "Charlie" {
			t.Errorf("Expected name=Charlie, got %v", user["name"])
		}
		if user["age"] != int64(40) {
			t.Errorf("Expected age=40, got %v", user["age"])
		}
		if user["active"] != true {
			t.Errorf("Expected active=true, got %v", user["active"])
		}
	})

	t.Run("RecordFromMap", func(t *testing.T) {
		m := map[string]any{
			"name":   "David",
			"score":  95.5,
			"passed": true,
		}

		record := recordFrom(m)

		if record["name"] != "David" {
			t.Errorf("Expected name=David, got %v", record["name"])
		}
		if record["score"] != 95.5 {
			t.Errorf("Expected score=95.5, got %v", record["score"])
		}
		if record["passed"] != true {
			t.Errorf("Expected passed=true, got %v", record["passed"])
		}
	})
}

// BenchmarkConstructorPerformance compares old vs new approaches
func BenchmarkConstructorPerformance(b *testing.B) {
	b.Run("NewTypeSafeBuilder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewRecord().
				String("name", "Alice").
				Int("age", 30).
				Float("salary", 75000.50).
				Bool("active", true).
				Build()
		}
	})

	b.Run("NewRecordBuilder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewRecord().String("name", "Alice").Int("age", 30).Float("salary", 75000.50).Bool("active", true).Build()
		}
	})

	b.Run("TypeSafeField", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Field("name", "Alice")
		}
	})
}
// TestStringParsing tests the string parsing functionality in convertToInt64
func TestStringParsing(t *testing.T) {
	t.Run("ValidStringNumbers", func(t *testing.T) {
		record := NewRecord().
			String("string_positive", "42").
			String("string_negative", "-123").
			String("string_zero", "0").
			Build()
		
		// Test positive number parsing
		parsed := GetOr(record, "string_positive", int64(999))
		if parsed != 42 {
			t.Errorf("Expected 42, got %d", parsed)
		}
		
		// Test negative number parsing
		parsed = GetOr(record, "string_negative", int64(999))
		if parsed != -123 {
			t.Errorf("Expected -123, got %d", parsed)
		}
		
		// Test zero parsing
		parsed = GetOr(record, "string_zero", int64(999))
		if parsed != 0 {
			t.Errorf("Expected 0, got %d", parsed)
		}
	})
	
	t.Run("InvalidStringNumbers", func(t *testing.T) {
		record := NewRecord().
			String("invalid_letters", "abc").
			String("invalid_mixed", "123abc").
			String("invalid_empty", "").
			Build()
		
		// Test invalid letter string - should return default
		parsed := GetOr(record, "invalid_letters", int64(999))
		if parsed != 999 {
			t.Errorf("Expected default 999, got %d", parsed)
		}
		
		// Test invalid mixed string - should return default
		parsed = GetOr(record, "invalid_mixed", int64(888))
		if parsed != 888 {
			t.Errorf("Expected default 888, got %d", parsed)
		}
		
		// Test empty string - should return default
		parsed = GetOr(record, "invalid_empty", int64(777))
		if parsed != 777 {
			t.Errorf("Expected default 777, got %d", parsed)
		}
	})
}
