package stream

import (
	"runtime"
	"testing"
	"time"
)

// TestNoGoroutineLeaks verifies that our fixed functions don't leak goroutines
func TestNoGoroutineLeaks(t *testing.T) {
	// Test Parallel function
	t.Run("Parallel", func(t *testing.T) {
		before := runtime.NumGoroutine()
		
		// Create a parallel stream
		data := []int{1, 2, 3, 4, 5}
		stream := Parallel(2, func(x int) int { return x * 2 })(FromSlice(data))
		
		// Get first item only, then abandon
		_, err := stream()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Don't consume rest - this should trigger cleanup
		
		// Force GC and wait for cleanup
		runtime.GC()
		time.Sleep(200 * time.Millisecond)
		
		after := runtime.NumGoroutine()
		if after > before+3 { // Allow for some tolerance
			t.Errorf("Potential goroutine leak in Parallel: %d -> %d", before, after)
		}
	})
	
	// Test Tee function
	t.Run("Tee", func(t *testing.T) {
		before := runtime.NumGoroutine()
		
		// Create tee streams
		data := []int{1, 2, 3, 4, 5}
		streams := Tee(FromSlice(data), 3)
		
		// Read first item from first stream only
		_, err := streams[0]()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Abandon all streams - should trigger cleanup
		
		// Force cleanup
		runtime.GC()
		time.Sleep(200 * time.Millisecond)
		
		after := runtime.NumGoroutine()
		if after > before+2 { // Allow tolerance
			t.Errorf("Potential goroutine leak in Tee: %d -> %d", before, after)
		}
	})
	
	// Test Split function
	t.Run("Split", func(t *testing.T) {
		before := runtime.NumGoroutine()
		
		// Create split stream
		records := []Record{
			R("key", "A", "value", 1),
			R("key", "B", "value", 2),
			R("key", "A", "value", 3),
		}
		
		splitStream := Split([]string{"key"})(FromSlice(records))
		
		// Get first substream only
		substream, err := splitStream()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Read one item from substream, then abandon
		if substream != nil {
			_, err := substream()
			if err != nil && err != EOS {
				t.Errorf("Expected no error or EOS, got %v", err)
			}
		}
		
		// Abandon everything - should trigger cleanup
		
		// Force cleanup
		runtime.GC()
		time.Sleep(200 * time.Millisecond)
		
		after := runtime.NumGoroutine()
		if after > before+2 { // Allow tolerance
			t.Errorf("Potential goroutine leak in Split: %d -> %d", before, after)
		}
	})
}

// TestChannelOperationsSafety tests that channel operations don't block indefinitely
func TestChannelOperationsSafety(t *testing.T) {
	t.Run("Parallel with abandoned consumer", func(t *testing.T) {
		done := make(chan bool, 1)
		
		go func() {
			// This should complete quickly without hanging
			data := make([]int, 1000) // Large dataset
			for i := range data {
				data[i] = i
			}
			
			stream := Parallel(4, func(x int) int { 
				time.Sleep(1 * time.Millisecond) // Simulate work
				return x * 2 
			})(FromSlice(data))
			
			// Read a few items then abandon - should not hang
			for i := 0; i < 5; i++ {
				_, err := stream()
				if err != nil {
					break
				}
			}
			// Don't read rest - this triggers the leak scenario
			
			done <- true
		}()
		
		select {
		case <-done:
			// Success - completed without hanging
		case <-time.After(3 * time.Second):
			t.Error("Parallel operation hung - potential goroutine leak")
		}
	})
	
	t.Run("Tee with slow consumers", func(t *testing.T) {
		done := make(chan bool, 1)
		
		go func() {
			data := make([]int, 100)
			for i := range data {
				data[i] = i
			}
			
			streams := Tee(FromSlice(data), 3)
			
			// Read from first stream only - others should timeout and cleanup
			for i := 0; i < 10; i++ {
				_, err := streams[0]()
				if err != nil {
					break
				}
			}
			// Don't read from streams[1] and streams[2] - they should be abandoned
			
			done <- true
		}()
		
		select {
		case <-done:
			// Success
		case <-time.After(10 * time.Second):
			t.Error("Tee operation hung - potential goroutine leak")
		}
	})
}