package stream

import (
	"testing"
	"time"
)

// TestFromSlice tests the FromSlice function
func TestFromSlice(t *testing.T) {
	t.Run("IntSlice", func(t *testing.T) {
		data := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(data)
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != len(data) {
			t.Fatalf("Expected %d results, got %d", len(data), len(results))
		}
		
		for i, result := range results {
			if result != data[i] {
				t.Errorf("Expected %v at position %d, got %v", data[i], i, result)
			}
		}
	})
	
	t.Run("StringSlice", func(t *testing.T) {
		data := []string{"a", "b", "c"}
		stream := FromSlice(data)
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != len(data) {
			t.Fatalf("Expected %d results, got %d", len(data), len(results))
		}
		
		for i, result := range results {
			if result != data[i] {
				t.Errorf("Expected %v at position %d, got %v", data[i], i, result)
			}
		}
	})
	
	t.Run("EmptySlice", func(t *testing.T) {
		data := []int64{}
		stream := FromSlice(data)
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 0 {
			t.Fatalf("Expected 0 results, got %d", len(results))
		}
	})
}

// TestFromChannel tests the FromChannel function
func TestFromChannel(t *testing.T) {
	t.Run("IntChannel", func(t *testing.T) {
		ch := make(chan int64, 3)
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)
		
		stream := FromChannel(ch)
		results, err := Collect(stream)
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
	
	t.Run("ClosedChannel", func(t *testing.T) {
		ch := make(chan string)
		close(ch)
		
		stream := FromChannel(ch)
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 0 {
			t.Fatalf("Expected 0 results from closed channel, got %d", len(results))
		}
	})
}

// TestFrom tests the From variadic function
func TestFrom(t *testing.T) {
	t.Run("MultipleValues", func(t *testing.T) {
		stream := From(int64(1), int64(2), int64(3))
		
		results, err := Collect(stream)
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
	
	t.Run("NoValues", func(t *testing.T) {
		stream := From[int64]()
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 0 {
			t.Fatalf("Expected 0 results, got %d", len(results))
		}
	})
}

// TestFromSliceAny tests the FromSliceAny function
func TestFromSliceAny(t *testing.T) {
	t.Run("MixedTypes", func(t *testing.T) {
		data := []any{1, "hello", 3.14, true}
		stream := FromSliceAny(data)
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != len(data) {
			t.Fatalf("Expected %d results, got %d", len(data), len(results))
		}
		
		for i, result := range results {
			if result != data[i] {
				t.Errorf("Expected %v at position %d, got %v", data[i], i, result)
			}
		}
	})
	
	t.Run("CustomTypes", func(t *testing.T) {
		type CustomStruct struct {
			ID   int
			Name string
		}
		
		data := []CustomStruct{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
		}
		
		stream := FromSliceAny(data)
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != len(data) {
			t.Fatalf("Expected %d results, got %d", len(data), len(results))
		}
		
		for i, result := range results {
			if result != data[i] {
				t.Errorf("Expected %v at position %d, got %v", data[i], i, result)
			}
		}
	})
}

// TestGenerate tests the Generate function
func TestGenerate(t *testing.T) {
	t.Run("Counter", func(t *testing.T) {
		counter := int64(0)
		stream := Generate(func() (int64, error) {
			if counter >= 5 {
				return 0, EOS
			}
			counter++
			return counter, nil
		})
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		expected := []int64{1, 2, 3, 4, 5}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("EmptyGenerator", func(t *testing.T) {
		stream := Generate(func() (string, error) {
			return "", EOS
		})
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 0 {
			t.Fatalf("Expected 0 results, got %d", len(results))
		}
	})
}

// TestRange tests the Range function
func TestRange(t *testing.T) {
	t.Run("PositiveStep", func(t *testing.T) {
		stream := Range(1, 5, 1)
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		expected := []int64{1, 2, 3, 4}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("StepOfTwo", func(t *testing.T) {
		stream := Range(0, 10, 2)
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		expected := []int64{0, 2, 4, 6, 8}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("NegativeStep", func(t *testing.T) {
		stream := Range(5, 1, -1)
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		expected := []int64{5, 4, 3, 2}
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("EmptyRange", func(t *testing.T) {
		stream := Range(5, 5, 1)
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 0 {
			t.Fatalf("Expected 0 results, got %d", len(results))
		}
	})
}

// TestOnce tests the Once function
func TestOnce(t *testing.T) {
	t.Run("SingleValue", func(t *testing.T) {
		stream := Once(int64(42))
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		
		if results[0] != 42 {
			t.Errorf("Expected 42, got %v", results[0])
		}
	})
	
	t.Run("StringValue", func(t *testing.T) {
		stream := Once("hello")
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		
		if results[0] != "hello" {
			t.Errorf("Expected 'hello', got %v", results[0])
		}
	})
}

// TestField tests the Field function
func TestField(t *testing.T) {
	t.Run("StringField", func(t *testing.T) {
		record := Field("name", "Alice")
		
		if len(record) != 1 {
			t.Fatalf("Expected 1 field, got %d", len(record))
		}
		
		value, exists := record["name"]
		if !exists {
			t.Fatalf("Expected 'name' field to exist")
		}
		
		if value != "Alice" {
			t.Errorf("Expected 'Alice', got %v", value)
		}
	})
	
	t.Run("IntField", func(t *testing.T) {
		record := Field("age", int64(30))
		
		value, exists := record["age"]
		if !exists {
			t.Fatalf("Expected 'age' field to exist")
		}
		
		if value != int64(30) {
			t.Errorf("Expected 30, got %v", value)
		}
	})
}

// TestGet tests the Get function
func TestGet(t *testing.T) {
	record := NewRecord().
		String("name", "Alice").
		Int("age", 30).
		Build()
	
	t.Run("ExistingField", func(t *testing.T) {
		name, exists := Get[string](record, "name")
		if !exists {
			t.Fatalf("Expected 'name' field to exist")
		}
		
		if name != "Alice" {
			t.Errorf("Expected 'Alice', got %v", name)
		}
	})
	
	t.Run("NonExistingField", func(t *testing.T) {
		_, exists := Get[string](record, "email")
		if exists {
			t.Fatalf("Expected 'email' field to not exist")
		}
	})
	
	t.Run("TypeMismatch", func(t *testing.T) {
		// This should not panic but return zero value and false
		_, exists := Get[int64](record, "name")
		if exists {
			t.Errorf("Expected type mismatch to return false")
		}
	})
}

// TestGetOr tests the GetOr function
func TestGetOr(t *testing.T) {
	record := NewRecord().
		String("name", "Alice").
		Int("age", 30).
		Build()
	
	t.Run("ExistingField", func(t *testing.T) {
		name := GetOr(record, "name", "Unknown")
		if name != "Alice" {
			t.Errorf("Expected 'Alice', got %v", name)
		}
	})
	
	t.Run("NonExistingField", func(t *testing.T) {
		email := GetOr(record, "email", "default@example.com")
		if email != "default@example.com" {
			t.Errorf("Expected 'default@example.com', got %v", email)
		}
	})
	
	t.Run("TypeMismatch", func(t *testing.T) {
		// Should return default value for type mismatch
		age := GetOr[int64](record, "name", 0)
		if age != 0 {
			t.Errorf("Expected 0 (default), got %v", age)
		}
	})
}

// TestSetField tests the SetField function
func TestSetField(t *testing.T) {
	t.Run("NewField", func(t *testing.T) {
		record := make(Record)
		updated := SetField(record, "name", "Alice")
		
		if len(updated) != 1 {
			t.Fatalf("Expected 1 field, got %d", len(updated))
		}
		
		value := GetOr(updated, "name", "")
		if value != "Alice" {
			t.Errorf("Expected 'Alice', got %v", value)
		}
	})
	
	t.Run("UpdateExisting", func(t *testing.T) {
		record := NewRecord().String("name", "Alice").Build()
		updated := SetField(record, "name", "Bob")
		
		value := GetOr(updated, "name", "")
		if value != "Bob" {
			t.Errorf("Expected 'Bob', got %v", value)
		}
	})
}

// TestFromChannelAny tests the FromChannelAny function
func TestFromChannelAny(t *testing.T) {
	t.Run("CustomStruct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		
		ch := make(chan Person, 2)
		ch <- Person{Name: "Alice", Age: 30}
		ch <- Person{Name: "Bob", Age: 25}
		close(ch)
		
		stream := FromChannelAny(ch)
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
		
		if results[0].Name != "Alice" || results[0].Age != 30 {
			t.Errorf("Expected Alice/30, got %+v", results[0])
		}
		
		if results[1].Name != "Bob" || results[1].Age != 25 {
			t.Errorf("Expected Bob/25, got %+v", results[1])
		}
	})
}

// TestGenerateAny tests the GenerateAny function
func TestGenerateAny(t *testing.T) {
	t.Run("TimestampGenerator", func(t *testing.T) {
		count := 0
		stream := GenerateAny(func() (time.Time, error) {
			if count >= 3 {
				return time.Time{}, EOS
			}
			count++
			return time.Unix(int64(count), 0), nil
		})
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}
		
		for i, result := range results {
			expected := time.Unix(int64(i+1), 0)
			if !result.Equal(expected) {
				t.Errorf("Expected %v at position %d, got %v", expected, i, result)
			}
		}
	})
}

// TestOnceAny tests the OnceAny function
func TestOnceAny(t *testing.T) {
	t.Run("CustomStruct", func(t *testing.T) {
		type Config struct {
			Debug bool
			Port  int
		}
		
		config := Config{Debug: true, Port: 8080}
		stream := OnceAny(config)
		
		results, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		
		if results[0].Debug != true || results[0].Port != 8080 {
			t.Errorf("Expected debug=true/port=8080, got %+v", results[0])
		}
	})
}