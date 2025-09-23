package stream

import (
	"context"
	"testing"
	"time"
)

// TestStreamingSum tests the StreamingSum filter
func TestStreamingSum(t *testing.T) {
	t.Run("IntStreamingSum", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		runningSum := StreamingSum[int64]()(stream)
		results, err := Collect(runningSum)
		if err != nil {
			t.Fatalf("Failed to collect streaming sum: %v", err)
		}
		
		expected := []int64{1, 3, 6, 10, 15} // Running sum
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, result := range results {
			if result != expected[i] {
				t.Errorf("Expected %v at position %d, got %v", expected[i], i, result)
			}
		}
	})
	
	t.Run("FloatStreamingSum", func(t *testing.T) {
		input := []float64{1.5, 2.5, 1.0}
		stream := FromSlice(input)
		
		runningSum := StreamingSum[float64]()(stream)
		results, err := Collect(runningSum)
		if err != nil {
			t.Fatalf("Failed to collect streaming sum: %v", err)
		}
		
		expected := []float64{1.5, 4.0, 5.0}
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

// TestStreamingCount tests the StreamingCount filter
func TestStreamingCount(t *testing.T) {
	t.Run("CountProgression", func(t *testing.T) {
		input := []string{"a", "b", "c", "d"}
		stream := FromSlice(input)
		
		runningCount := StreamingCount[string]()(stream)
		results, err := Collect(runningCount)
		if err != nil {
			t.Fatalf("Failed to collect streaming count: %v", err)
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
}

// TestStreamingAvg tests the StreamingAvg filter
func TestStreamingAvg(t *testing.T) {
	t.Run("RunningAverage", func(t *testing.T) {
		input := []int64{2, 4, 6}
		stream := FromSlice(input)
		
		runningAvg := StreamingAvg[int64]()(stream)
		results, err := Collect(runningAvg)
		if err != nil {
			t.Fatalf("Failed to collect streaming average: %v", err)
		}
		
		expected := []float64{2.0, 3.0, 4.0} // 2/1, 6/2, 12/3
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

// TestStreamingMax tests the StreamingMax filter
func TestStreamingMax(t *testing.T) {
	t.Run("RunningMaximum", func(t *testing.T) {
		input := []int64{3, 1, 4, 1, 5}
		stream := FromSlice(input)
		
		runningMax := StreamingMax[int64]()(stream)
		results, err := Collect(runningMax)
		if err != nil {
			t.Fatalf("Failed to collect streaming max: %v", err)
		}
		
		expected := []int64{3, 3, 4, 4, 5}
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

// TestStreamingMin tests the StreamingMin filter
func TestStreamingMin(t *testing.T) {
	t.Run("RunningMinimum", func(t *testing.T) {
		input := []int64{3, 1, 4, 1, 5}
		stream := FromSlice(input)
		
		runningMin := StreamingMin[int64]()(stream)
		results, err := Collect(runningMin)
		if err != nil {
			t.Fatalf("Failed to collect streaming min: %v", err)
		}
		
		expected := []int64{3, 1, 1, 1, 1}
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

// TestStreamingStats tests the StreamingStats filter
func TestStreamingStats(t *testing.T) {
	t.Run("RunningStatistics", func(t *testing.T) {
		input := []int64{1, 2, 3}
		stream := FromSlice(input)
		
		runningStats := StreamingStats[int64]()(stream)
		results, err := Collect(runningStats)
		if err != nil {
			t.Fatalf("Failed to collect streaming stats: %v", err)
		}
		
		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}
		
		// Check first record (just 1)
		first := results[0]
		if GetOr(first, "count", int64(0)) != 1 {
			t.Errorf("Expected count=1, got %v", first["count"])
		}
		if GetOr(first, "sum", int64(0)) != 1 {
			t.Errorf("Expected sum=1, got %v", first["sum"])
		}
		if GetOr(first, "avg", 0.0) != 1.0 {
			t.Errorf("Expected avg=1.0, got %v", first["avg"])
		}
		if GetOr(first, "min", int64(0)) != 1 {
			t.Errorf("Expected min=1, got %v", first["min"])
		}
		if GetOr(first, "max", int64(0)) != 1 {
			t.Errorf("Expected max=1, got %v", first["max"])
		}
		
		// Check third record (1, 2, 3)
		third := results[2]
		if GetOr(third, "count", int64(0)) != 3 {
			t.Errorf("Expected count=3, got %v", third["count"])
		}
		if GetOr(third, "sum", int64(0)) != 6 {
			t.Errorf("Expected sum=6, got %v", third["sum"])
		}
		if GetOr(third, "avg", 0.0) != 2.0 {
			t.Errorf("Expected avg=2.0, got %v", third["avg"])
		}
		if GetOr(third, "min", int64(0)) != 1 {
			t.Errorf("Expected min=1, got %v", third["min"])
		}
		if GetOr(third, "max", int64(0)) != 3 {
			t.Errorf("Expected max=3, got %v", third["max"])
		}
	})
}

// TestCountWindow tests the CountWindow filter
func TestCountWindow(t *testing.T) {
	t.Run("WindowsOfSize3", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5, 6, 7}
		stream := FromSlice(input)
		
		windowed := CountWindow[int64](3)(stream)
		results := make([][]int64, 0)
		
		for {
			window, err := windowed()
			if err != nil {
				break
			}
			
			windowData, err := Collect(window)
			if err != nil {
				t.Fatalf("Failed to collect window: %v", err)
			}
			results = append(results, windowData)
		}
		
		// Should have windows: [1,2,3], [4,5,6], [7]
		expected := [][]int64{
			{1, 2, 3},
			{4, 5, 6},
			{7},
		}
		
		if len(results) != len(expected) {
			t.Fatalf("Expected %d windows, got %d", len(expected), len(results))
		}
		
		for i, window := range results {
			if len(window) != len(expected[i]) {
				t.Errorf("Window %d: expected %d items, got %d", i, len(expected[i]), len(window))
			}
			for j, item := range window {
				if item != expected[i][j] {
					t.Errorf("Window %d, item %d: expected %v, got %v", i, j, expected[i][j], item)
				}
			}
		}
	})
}

// TestTimeWindow tests the TimeWindow filter
func TestTimeWindow(t *testing.T) {
	t.Run("TimeBasedWindows", func(t *testing.T) {
		// Create a stream with time delays
		stream := Generate(func() (int64, error) {
			time.Sleep(10 * time.Millisecond)
			return 1, nil
		})
		
		// Limit to only a few items to avoid infinite stream
		limitedStream := Limit[int64](5)(stream)
		
		windowed := TimeWindow[int64](50 * time.Millisecond)(limitedStream)
		
		windowCount := 0
		for {
			window, err := windowed()
			if err != nil {
				break
			}
			
			windowData, err := Collect(window)
			if err != nil {
				t.Fatalf("Failed to collect window: %v", err)
			}
			
			if len(windowData) > 0 {
				windowCount++
			}
		}
		
		// Should have at least 1 window (exact count depends on timing)
		if windowCount == 0 {
			t.Errorf("Expected at least 1 window, got %d", windowCount)
		}
	})
}

// TestSlidingCountWindow tests the SlidingCountWindow filter
func TestSlidingCountWindow(t *testing.T) {
	t.Run("SlidingWindows", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		// Window size 3, step size 2
		windowed := SlidingCountWindow[int64](3, 2)(stream)
		results := make([][]int64, 0)
		
		for {
			window, err := windowed()
			if err != nil {
				break
			}
			
			windowData, err := Collect(window)
			if err != nil {
				t.Fatalf("Failed to collect window: %v", err)
			}
			results = append(results, windowData)
		}
		
		// Should have windows: [1,2,3], [3,4,5]
		expected := [][]int64{
			{1, 2, 3},
			{3, 4, 5},
		}
		
		if len(results) != len(expected) {
			t.Fatalf("Expected %d windows, got %d", len(expected), len(results))
		}
		
		for i, window := range results {
			if len(window) != len(expected[i]) {
				t.Errorf("Window %d: expected %d items, got %d", i, len(expected[i]), len(window))
			}
			for j, item := range window {
				if item != expected[i][j] {
					t.Errorf("Window %d, item %d: expected %v, got %v", i, j, expected[i][j], item)
				}
			}
		}
	})
}

// TestNewCountTrigger tests the NewCountTrigger function
func TestNewCountTrigger(t *testing.T) {
	t.Run("TriggerEvery3", func(t *testing.T) {
		trigger := NewCountTrigger[int64](3)
		
		// Should not trigger for first 2 items
		if trigger.ShouldFire(1, nil) {
			t.Errorf("Expected trigger not to fire on first item")
		}
		if trigger.ShouldFire(2, nil) {
			t.Errorf("Expected trigger not to fire on second item")
		}
		
		// Should trigger on third item
		if !trigger.ShouldFire(3, nil) {
			t.Errorf("Expected trigger to fire on third item")
		}
		
		// Reset and test again
		state := trigger.ResetState()
		if trigger.ShouldFire(1, state) {
			t.Errorf("Expected trigger not to fire after reset")
		}
	})
}

// TestNewValueChangeTrigger tests the NewValueChangeTrigger function
func TestNewValueChangeTrigger(t *testing.T) {
	t.Run("TriggerOnChange", func(t *testing.T) {
		// Trigger when string length changes
		trigger := NewValueChangeTrigger(func(s string) any {
			return len(s)
		})
		
		// First item should trigger (no previous value)
		if !trigger.ShouldFire("hello", nil) {
			t.Errorf("Expected trigger to fire on first item")
		}
		
		// Same length should not trigger
		state := trigger.ResetState()
		if trigger.ShouldFire("world", state) {
			t.Errorf("Expected trigger not to fire on same length")
		}
		
		// Different length should trigger
		if !trigger.ShouldFire("hi", state) {
			t.Errorf("Expected trigger to fire on different length")
		}
	})
}

// TestTriggeredWindow tests the TriggeredWindow function
func TestTriggeredWindow(t *testing.T) {
	t.Run("CountTriggeredWindows", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5, 6}
		stream := FromSlice(input)
		
		trigger := NewCountTrigger[int64](2)
		windowed := TriggeredWindow(trigger)(stream)
		
		results := make([][]int64, 0)
		for {
			window, err := windowed()
			if err != nil {
				break
			}
			
			windowData, err := Collect(window)
			if err != nil {
				t.Fatalf("Failed to collect window: %v", err)
			}
			results = append(results, windowData)
		}
		
		// Should have windows triggered every 2 items: [1,2], [3,4], [5,6]
		expected := [][]int64{
			{1, 2},
			{3, 4},
			{5, 6},
		}
		
		if len(results) != len(expected) {
			t.Fatalf("Expected %d windows, got %d", len(expected), len(results))
		}
		
		for i, window := range results {
			if len(window) != len(expected[i]) {
				t.Errorf("Window %d: expected %d items, got %d", i, len(expected[i]), len(window))
			}
			for j, item := range window {
				if item != expected[i][j] {
					t.Errorf("Window %d, item %d: expected %v, got %v", i, j, expected[i][j], item)
				}
			}
		}
	})
}

// TestStreamingGroupBy tests the StreamingGroupBy function
func TestStreamingGroupBy(t *testing.T) {
	t.Run("GroupByCategory", func(t *testing.T) {
		records := []Record{
			NewRecord().String("category", "A").Int("value", 1).Build(),
			NewRecord().String("category", "B").Int("value", 2).Build(),
			NewRecord().String("category", "A").Int("value", 3).Build(),
			NewRecord().String("category", "B").Int("value", 4).Build(),
		}
		stream := FromRecordsUnsafe(records)
		
		grouped := StreamingGroupBy([]string{"category"}, 2)(stream)
		results, err := Collect(grouped)
		if err != nil {
			t.Fatalf("Failed to collect grouped results: %v", err)
		}
		
		// Should have summary records for groups
		if len(results) == 0 {
			t.Fatalf("Expected grouped results, got none")
		}
		
		// Check that we have group information
		found := false
		for _, result := range results {
			if _, exists := result["largest_group_key"]; exists {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Expected to find largest_group_key in results")
		}
	})
}

// TestParallel tests the Parallel filter  
func TestParallel(t *testing.T) {
	t.Run("ParallelProcessing", func(t *testing.T) {
		input := []int64{1, 2, 3, 4, 5}
		stream := FromSlice(input)
		
		// Square numbers in parallel
		squared := Parallel(2, func(x int64) int64 {
			return x * x
		})(stream)
		
		results, err := Collect(squared)
		if err != nil {
			t.Fatalf("Failed to collect parallel results: %v", err)
		}
		
		if len(results) != len(input) {
			t.Fatalf("Expected %d results, got %d", len(input), len(results))
		}
		
		// Results may be out of order due to parallel processing
		// So we'll check that we have the right values
		expectedMap := make(map[int64]bool)
		for _, x := range input {
			expectedMap[x*x] = true
		}
		
		for _, result := range results {
			if !expectedMap[result] {
				t.Errorf("Unexpected result: %v", result)
			}
		}
	})
}

// TestSplit tests the Split filter
func TestSplit(t *testing.T) {
	t.Run("SplitByKey", func(t *testing.T) {
		records := []Record{
			NewRecord().String("department", "eng").String("name", "Alice").Build(),
			NewRecord().String("department", "sales").String("name", "Bob").Build(),
			NewRecord().String("department", "eng").String("name", "Charlie").Build(),
		}
		stream := FromRecordsUnsafe(records)
		
		splitStream := Split([]string{"department"})(stream)
		
		substreams := make([]Stream[Record], 0)
		
		// Collect first few substreams
		for i := 0; i < 2; i++ {
			substream, err := splitStream()
			if err != nil {
				break
			}
			substreams = append(substreams, substream)
		}
		
		if len(substreams) == 0 {
			t.Fatalf("Expected at least 1 substream, got %d", len(substreams))
		}
		
		// Each substream should contain records
		for i, substream := range substreams {
			results, err := Collect(substream)
			if err != nil {
				t.Errorf("Failed to collect substream %d: %v", i, err)
				continue
			}
			
			if len(results) == 0 {
				t.Errorf("Expected non-empty substream %d", i)
			}
		}
	})
}

// Test timeout scenario for context cancellation
func TestContextTimeout(t *testing.T) {
	t.Run("ContextCancellation", func(t *testing.T) {
		// Create a slow stream
		slowStream := Generate(func() (int64, error) {
			time.Sleep(10 * time.Millisecond)
			return 1, nil
		})
		
		// Create context with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		
		contextualStream := WithContext(ctx, slowStream)
		
		// Try to collect - should be interrupted by context cancellation
		start := time.Now()
		_, _ = Collect(contextualStream) // Ignore error, just check timing
		elapsed := time.Since(start)
		
		// Should finish quickly due to context cancellation
		if elapsed > 200*time.Millisecond {
			t.Errorf("Expected quick termination due to context, took %v", elapsed)
		}
	})
}