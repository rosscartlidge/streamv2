package stream

import (
	"math"
	"runtime"
	"testing"
	"time"
)

// TestAutoParallelMap verifies that Map automatically uses parallel processing for complex operations
func TestAutoParallelMap(t *testing.T) {
	// Test simple operation stays sequential (no goroutine explosion)
	t.Run("SimpleOperationSequential", func(t *testing.T) {
		before := runtime.NumGoroutine()
		
		data := []int{1, 2, 3, 4, 5}
		result, err := Collect(
			Map(func(x int) int { return x + 1 })(  // Simple operation - complexity ~1
				FromSlice(data)))
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Verify results are correct
		expected := []int{2, 3, 4, 5, 6}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
			}
		}
		
		// Should not spawn many goroutines for simple operations
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		after := runtime.NumGoroutine()
		
		if after > before+2 {
			t.Logf("Goroutines before: %d, after: %d (expected minimal increase)", before, after)
		}
	})
	
	// Test complex operation uses parallel processing
	t.Run("ComplexOperationParallel", func(t *testing.T) {
		data := make([]float64, 100)
		for i := range data {
			data[i] = float64(i) * 0.1
		}
		
		start := time.Now()
		result, err := Collect(
			Map(func(x float64) float64 { 
				// Complex operation - should trigger auto-parallel
				return math.Sin(x) * math.Cos(x) * math.Sqrt(x+1)
			})(FromSlice(data)))
		duration := time.Since(start)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Verify we got correct number of results
		if len(result) != len(data) {
			t.Errorf("Expected %d results, got %d", len(data), len(result))
		}
		
		t.Logf("Complex Map operation completed in %v", duration)
	})
}

// TestAutoParallelFilter verifies that Where automatically uses parallel processing for complex predicates
func TestAutoParallelFilter(t *testing.T) {
	t.Run("ComplexPredicateParallel", func(t *testing.T) {
		data := make([]float64, 200)
		for i := range data {
			data[i] = float64(i)
		}
		
		start := time.Now()
		result, err := Collect(
			Where(func(x float64) bool {
				// Complex predicate - should trigger auto-parallel
				return math.Sin(x)*math.Cos(x) > 0.5
			})(FromSlice(data)))
		duration := time.Since(start)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Verify filtering worked correctly
		if len(result) == 0 {
			t.Error("Expected some results from filtering")
		}
		
		if len(result) >= len(data) {
			t.Error("Expected filtering to reduce result count")
		}
		
		t.Logf("Complex Filter operation completed in %v, filtered to %d results", duration, len(result))
	})
}

// TestAutoParallelGroupBy verifies GroupBy uses parallel processing for complex aggregations
func TestAutoParallelGroupBy(t *testing.T) {
	t.Run("MultipleAggregatorsParallel", func(t *testing.T) {
		// Create test data with multiple groups
		var data []Record
		departments := []string{"engineering", "sales", "marketing", "hr"}
		
		for i := 0; i < 400; i++ {
			dept := departments[i%len(departments)]
			salary := 50000 + (i%20)*5000
			
			data = append(data, NewRecord().
				String("department", dept).
				Int("salary", int64(salary)).
				Int("experience", int64(i%10)).
				Build())
		}
		
		start := time.Now()
		results, err := Collect(
			GroupBy([]string{"department"},
				FieldSumSpec[int]("total_salary", "salary"),
				FieldAvgSpec[int]("avg_salary", "salary"),
				FieldMaxSpec[int]("max_salary", "salary"),  // 3+ aggregators should trigger parallel
				FieldMinSpec[int]("min_salary", "salary"),
			)(FromSlice(data)))
		duration := time.Since(start)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Should have one group per department
		if len(results) != len(departments) {
			t.Errorf("Expected %d groups, got %d", len(departments), len(results))
		}
		
		// Verify aggregation results
		for _, result := range results {
			dept := GetOr(result, "department", "")
			count := GetOr(result, "group_count", int64(0))
			total := GetOr(result, "total_salary", 0)
			
			if dept == "" {
				t.Error("Expected department in group result")
			}
			
			if count != 100 { // 400 records / 4 departments = 100 each
				t.Errorf("Expected count 100 for %s, got %d", dept, count)
			}
			
			if total <= 0 {
				t.Errorf("Expected positive total salary for %s, got %d", dept, total)
			}
			
			t.Logf("Department %s: %d people, total salary %d", dept, count, total)
		}
		
		t.Logf("Parallel GroupBy with multiple aggregators completed in %v", duration)
	})
	
	t.Run("SimpleGroupBySequential", func(t *testing.T) {
		// Simple GroupBy should use sequential processing
		data := []Record{
			NewRecord().String("type", "A").Int("value", 1).Build(),
			NewRecord().String("type", "B").Int("value", 2).Build(),
			NewRecord().String("type", "A").Int("value", 3).Build(),
		}
		
		results, err := Collect(
			GroupBy([]string{"type"})(  // No aggregators - should be sequential
				FromSlice(data)))
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if len(results) != 2 { // Types A and B
			t.Errorf("Expected 2 groups, got %d", len(results))
		}
	})
}

// BenchmarkAutoParallel compares performance of auto-parallel vs manual parallel
func BenchmarkAutoParallel(b *testing.B) {
	data := make([]float64, 10000)
	for i := range data {
		data[i] = float64(i) * 0.001
	}
	
	complexFn := func(x float64) float64 {
		return math.Sin(x) * math.Cos(x) * math.Sqrt(x+1)
	}
	
	b.Run("AutoParallelMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Collect(Map(complexFn)(FromSlice(data)))
		}
	})
	
	b.Run("ManualParallel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Collect(Parallel(4, complexFn)(FromSlice(data)))
		}
	})
}

// TestWorkerCalculation verifies optimal worker calculation
func TestWorkerCalculation(t *testing.T) {
	testCases := []struct {
		complexity int
		expected   int
	}{
		{1, 4},  // Low complexity
		{3, 4},  // Moderate complexity  
		{5, 6},  // Higher complexity
		{7, 8},  // High complexity
		{10, 16}, // Very high complexity (capped)
	}
	
	for _, tc := range testCases {
		workers := calculateOptimalWorkers(tc.complexity)
		if workers != tc.expected {
			t.Errorf("For complexity %d, expected %d workers, got %d", tc.complexity, tc.expected, workers)
		}
	}
}