package stream

import (
	"testing"
)

// TestSum tests the Sum aggregator
func TestSum(t *testing.T) {
	t.Run("IntSum", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		result, err := Sum(stream)
		if err != nil {
			t.Fatalf("Failed to sum stream: %v", err)
		}
		
		expected := int64(15)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("FloatSum", func(t *testing.T) {
		input := []float64{1.5, 2.5, 3.0}
		stream := FromSlice(input)
		
		result, err := Sum(stream)
		if err != nil {
			t.Fatalf("Failed to sum stream: %v", err)
		}
		
		expected := 7.0
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("EmptySum", func(t *testing.T) {
		stream := FromSlice([]int64{})
		
		result, err := Sum(stream)
		if err != nil {
			t.Fatalf("Failed to sum empty stream: %v", err)
		}
		
		if result != 0 {
			t.Errorf("Expected 0 for empty stream, got %v", result)
		}
	})
}

// TestCount tests the Count aggregator  
func TestCount(t *testing.T) {
	t.Run("CountInts", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		result, err := Count(stream)
		if err != nil {
			t.Fatalf("Failed to count stream: %v", err)
		}
		
		expected := int64(5)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("CountStrings", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		stream := FromSlice(input)
		
		result, err := Count(stream)
		if err != nil {
			t.Fatalf("Failed to count stream: %v", err)
		}
		
		expected := int64(3)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("CountEmpty", func(t *testing.T) {
		stream := FromSlice([]string{})
		
		result, err := Count(stream)
		if err != nil {
			t.Fatalf("Failed to count empty stream: %v", err)
		}
		
		if result != 0 {
			t.Errorf("Expected 0 for empty stream, got %v", result)
		}
	})
}

// TestMax tests the Max aggregator
func TestMax(t *testing.T) {
	t.Run("MaxInts", func(t *testing.T) {
		input := []int64{3, 1, 4, 1, 5, 9, 2}
		stream := FromSlice(input)
		
		result, err := Max(stream)
		if err != nil {
			t.Fatalf("Failed to find max: %v", err)
		}
		
		expected := int64(9)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("MaxFloats", func(t *testing.T) {
		input := []float64{3.14, 2.71, 1.41, 1.73}
		stream := FromSlice(input)
		
		result, err := Max(stream)
		if err != nil {
			t.Fatalf("Failed to find max: %v", err)
		}
		
		expected := 3.14
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("MaxStrings", func(t *testing.T) {
		input := []string{"apple", "zebra", "banana"}
		stream := FromSlice(input)
		
		result, err := Max(stream)
		if err != nil {
			t.Fatalf("Failed to find max: %v", err)
		}
		
		expected := "zebra"
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("MaxSingle", func(t *testing.T) {
		input := []int64{42}
		stream := FromSlice(input)
		
		result, err := Max(stream)
		if err != nil {
			t.Fatalf("Failed to find max: %v", err)
		}
		
		expected := int64(42)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// TestMin tests the Min aggregator
func TestMin(t *testing.T) {
	t.Run("MinInts", func(t *testing.T) {
		input := []int64{3, 1, 4, 1, 5, 9, 2}
		stream := FromSlice(input)
		
		result, err := Min(stream)
		if err != nil {
			t.Fatalf("Failed to find min: %v", err)
		}
		
		expected := int64(1)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("MinFloats", func(t *testing.T) {
		input := []float64{3.14, 2.71, 1.41, 1.73}
		stream := FromSlice(input)
		
		result, err := Min(stream)
		if err != nil {
			t.Fatalf("Failed to find min: %v", err)
		}
		
		expected := 1.41
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("MinStrings", func(t *testing.T) {
		input := []string{"zebra", "apple", "banana"}
		stream := FromSlice(input)
		
		result, err := Min(stream)
		if err != nil {
			t.Fatalf("Failed to find min: %v", err)
		}
		
		expected := "apple"
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// TestAvg tests the Avg aggregator
func TestAvg(t *testing.T) {
	t.Run("AvgInts", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		result, err := Avg(stream)
		if err != nil {
			t.Fatalf("Failed to calculate average: %v", err)
		}
		
		expected := 3.0
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("AvgFloats", func(t *testing.T) {
		input := []float64{2.0, 4.0, 6.0}
		stream := FromSlice(input)
		
		result, err := Avg(stream)
		if err != nil {
			t.Fatalf("Failed to calculate average: %v", err)
		}
		
		expected := 4.0
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("AvgSingle", func(t *testing.T) {
		input := []int64{42}
		stream := FromSlice(input)
		
		result, err := Avg(stream)
		if err != nil {
			t.Fatalf("Failed to calculate average: %v", err)
		}
		
		expected := 42.0
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// TestCollect tests the Collect function (already used above, but explicit test)
func TestCollect(t *testing.T) {
	t.Run("CollectInts", func(t *testing.T) {
		input := []int64{1, 2, 3}
		stream := FromSlice(input)
		
		result, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect stream: %v", err)
		}
		
		if len(result) != len(input) {
			t.Fatalf("Expected %d items, got %d", len(input), len(result))
		}
		
		for i, item := range result {
			if item != input[i] {
				t.Errorf("Expected %v at position %d, got %v", input[i], i, item)
			}
		}
	})
	
	t.Run("CollectEmpty", func(t *testing.T) {
		stream := FromSlice([]string{})
		
		result, err := Collect(stream)
		if err != nil {
			t.Fatalf("Failed to collect empty stream: %v", err)
		}
		
		if len(result) != 0 {
			t.Fatalf("Expected empty slice, got %d items", len(result))
		}
	})
}

// TestForEach tests the ForEach function
func TestForEach(t *testing.T) {
	t.Run("SideEffects", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		sum := int64(0)
		err := ForEach(func(x int64) {
			sum += x
		})(stream)
		
		if err != nil {
			t.Fatalf("Failed to iterate stream: %v", err)
		}
		
		expected := int64(15)
		if sum != expected {
			t.Errorf("Expected sum %v, got %v", expected, sum)
		}
	})
	
	t.Run("StringProcessing", func(t *testing.T) {
		input := []string{"hello", "world"}
		stream := FromSlice(input)
		
		var concatenated string
		err := ForEach(func(s string) {
			concatenated += s + " "
		})(stream)
		
		if err != nil {
			t.Fatalf("Failed to iterate stream: %v", err)
		}
		
		expected := "hello world "
		if concatenated != expected {
			t.Errorf("Expected '%v', got '%v'", expected, concatenated)
		}
	})
}

// TestSumAggregator tests the SumAggregator function
func TestSumAggregator(t *testing.T) {
	t.Run("IntExtraction", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int64
		}
		
		people := []Person{
			{"Alice", 30},
			{"Bob", 25},
			{"Charlie", 35},
		}
		stream := FromSliceAny(people)
		
		agg := SumAggregator(func(p Person) int64 { return p.Age })
		result, err := RunAggregator(stream, agg)
		if err != nil {
			t.Fatalf("Failed to run aggregator: %v", err)
		}
		
		expected := int64(90)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// TestMinAggregator tests the MinAggregator function
func TestMinAggregator(t *testing.T) {
	t.Run("ScoreExtraction", func(t *testing.T) {
		type Game struct {
			Player string
			Score  int64
		}
		
		games := []Game{
			{"Alice", 100},
			{"Bob", 85},
			{"Charlie", 120},
		}
		stream := FromSliceAny(games)
		
		agg := MinAggregator(func(g Game) int64 { return g.Score })
		result, err := RunAggregator(stream, agg)
		if err != nil {
			t.Fatalf("Failed to run aggregator: %v", err)
		}
		
		expected := int64(85)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// TestMaxAggregator tests the MaxAggregator function
func TestMaxAggregator(t *testing.T) {
	t.Run("ScoreExtraction", func(t *testing.T) {
		type Game struct {
			Player string
			Score  int64
		}
		
		games := []Game{
			{"Alice", 100},
			{"Bob", 85},
			{"Charlie", 120},
		}
		stream := FromSliceAny(games)
		
		agg := MaxAggregator(func(g Game) int64 { return g.Score })
		result, err := RunAggregator(stream, agg)
		if err != nil {
			t.Fatalf("Failed to run aggregator: %v", err)
		}
		
		expected := int64(120)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// TestAvgAggregator tests the AvgAggregator function
func TestAvgAggregator(t *testing.T) {
	t.Run("ScoreAverage", func(t *testing.T) {
		type Student struct {
			Name  string
			Score float64
		}
		
		students := []Student{
			{"Alice", 90.0},
			{"Bob", 80.0},
			{"Charlie", 85.0},
		}
		stream := FromSliceAny(students)
		
		agg := AvgAggregator(func(s Student) float64 { return s.Score })
		result, err := RunAggregator(stream, agg)
		if err != nil {
			t.Fatalf("Failed to run aggregator: %v", err)
		}
		
		expected := 85.0
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// TestCountAggregator tests the CountAggregator function
func TestCountAggregator(t *testing.T) {
	t.Run("CountItems", func(t *testing.T) {
		type Item struct {
			ID   int
			Name string
		}
		
		items := []Item{
			{1, "Item1"},
			{2, "Item2"},
			{3, "Item3"},
		}
		stream := FromSliceAny(items)
		
		agg := CountAggregator[Item]()
		result, err := RunAggregator(stream, agg)
		if err != nil {
			t.Fatalf("Failed to run aggregator: %v", err)
		}
		
		expected := int64(3)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// TestAggregateMultiple tests the AggregateMultiple function
func TestAggregateMultiple(t *testing.T) {
	t.Run("SumAndCount", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		sumAgg := SumAggregator(func(x int64) int64 { return x })
		countAgg := CountAggregator[int64]()
		
		sum, count, err := AggregateMultiple(stream, sumAgg, countAgg)
		if err != nil {
			t.Fatalf("Failed to run multiple aggregators: %v", err)
		}
		
		expectedSum := int64(15)
		expectedCount := int64(5)
		
		if sum != expectedSum {
			t.Errorf("Expected sum %v, got %v", expectedSum, sum)
		}
		if count != expectedCount {
			t.Errorf("Expected count %v, got %v", expectedCount, count)
		}
	})
}

// TestAggregates tests the Aggregates function
func TestAggregates(t *testing.T) {
	t.Run("MultipleSpecs", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		result, err := Aggregates(stream,
			SumSpec[int64]("total"),
			CountSpec[int64]("count"),
			MinSpec[int64]("minimum"),
			MaxSpec[int64]("maximum"),
			AvgSpec[int64]("average"),
		)
		
		if err != nil {
			t.Fatalf("Failed to run aggregates: %v", err)
		}
		
		if GetOr(result, "total", int64(0)) != 15 {
			t.Errorf("Expected total=15, got %v", result["total"])
		}
		if GetOr(result, "count", int64(0)) != 5 {
			t.Errorf("Expected count=5, got %v", result["count"])
		}
		if GetOr(result, "minimum", int64(0)) != 1 {
			t.Errorf("Expected minimum=1, got %v", result["minimum"])
		}
		if GetOr(result, "maximum", int64(0)) != 5 {
			t.Errorf("Expected maximum=5, got %v", result["maximum"])
		}
		if GetOr(result, "average", 0.0) != 3.0 {
			t.Errorf("Expected average=3.0, got %v", result["average"])
		}
	})
}

// TestFieldCountSpec tests the FieldCountSpec function
func TestFieldCountSpec(t *testing.T) {
	t.Run("RecordCount", func(t *testing.T) {
		records := []Record{
			{"id": 1, "name": "Alice", "score": 95},
			{"id": 2, "name": "Bob", "score": 87},
			{"id": 3, "name": "Charlie", "score": 92},
		}
		stream := FromSlice(records)
		
		// Test FieldCountSpec in GroupBy
		result := GroupBy([]string{}, 
			FieldCountSpec("custom_count", "id"),
		)(stream)
		
		results, err := Collect(result)
		if err != nil {
			t.Fatalf("Failed to collect results: %v", err)
		}
		
		if len(results) != 1 {
			t.Fatalf("Expected 1 group, got %d", len(results))
		}
		
		// Should have both group_count (built-in) and custom_count (our spec)
		groupCount := GetOr(results[0], "group_count", int64(0))
		customCount := GetOr(results[0], "custom_count", int64(0))
		
		if groupCount != 3 {
			t.Errorf("Expected group_count=3, got %v", groupCount)
		}
		if customCount != 3 {
			t.Errorf("Expected custom_count=3, got %v", customCount)
		}
	})
}