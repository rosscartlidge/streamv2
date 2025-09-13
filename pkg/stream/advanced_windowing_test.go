package stream

import (
	"fmt"
	"testing"
	"time"
)

// TestSessionWindow verifies session-based windowing functionality
func TestSessionWindow(t *testing.T) {
	t.Run("BasicSessionWindow", func(t *testing.T) {
		// Create test data with activity patterns
		events := []string{
			"login",     // Activity - starts session 1
			"click",     // Non-activity - extends session 1
			"view",      // Non-activity - extends session 1
			"purchase",  // Activity - starts session 2 or extends session 1
			"logout",    // Activity - might start session 3
		}
		
		// Activity detector: login, purchase, logout are activities
		activityDetector := func(event string) bool {
			return event == "login" || event == "purchase" || event == "logout"
		}
		
		timeout := 100 * time.Millisecond
		
		sessionWindows, err := Collect(
			SessionWindow(timeout, activityDetector)(
				FromSlice(events)))
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Should have at least one session window
		if len(sessionWindows) == 0 {
			t.Error("Expected at least one session window")
		}
		
		// Verify each session contains events
		for i, session := range sessionWindows {
			sessionEvents, err := Collect(session)
			if err != nil {
				t.Errorf("Error collecting session %d: %v", i, err)
				continue
			}
			
			if len(sessionEvents) == 0 {
				t.Errorf("Session %d should contain events", i)
			}
			
			t.Logf("Session %d: %v", i, sessionEvents)
		}
	})
	
	t.Run("SessionWindowTimeout", func(t *testing.T) {
		// Test that sessions properly timeout
		events := []string{"start", "event1"}
		
		activityDetector := func(event string) bool {
			return event == "start"
		}
		
		timeout := 50 * time.Millisecond
		
		start := time.Now()
		sessionWindows, err := Collect(
			SessionWindow(timeout, activityDetector)(
				FromSlice(events)))
		duration := time.Since(start)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Should complete in reasonable time (not hang)
		if duration > 200*time.Millisecond {
			t.Errorf("Session window took too long: %v", duration)
		}
		
		t.Logf("Session window completed in %v with %d sessions", duration, len(sessionWindows))
	})
}

// TestWindowBuilder verifies the fluent window builder API
func TestWindowBuilder(t *testing.T) {
	t.Run("CountTrigger", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		
		windows, err := Collect(
			Window[int]().
				TriggerOnCount(3).  // Fire every 3 elements
				Apply()(
				FromSlice(data)))
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Should have multiple windows
		if len(windows) < 2 {
			t.Errorf("Expected at least 2 windows, got %d", len(windows))
		}
		
		// Verify window contents
		totalElements := 0
		for i, window := range windows {
			elements, err := Collect(window)
			if err != nil {
				t.Errorf("Error collecting window %d: %v", i, err)
				continue
			}
			
			totalElements += len(elements)
			t.Logf("Window %d: %v (size: %d)", i, elements, len(elements))
		}
		
		// All elements should be accounted for
		if totalElements != len(data) {
			t.Errorf("Expected %d total elements, got %d", len(data), totalElements)
		}
	})
	
	t.Run("TimeTrigger", func(t *testing.T) {
		// Create a stream that produces elements with delays
		data := []int{1, 2, 3}
		
		windows, err := Collect(
			Window[int]().
				TriggerOnTime(50 * time.Millisecond).  // Fire every 50ms
				Apply()(
				FromSlice(data)))
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Should have at least one window
		if len(windows) == 0 {
			t.Error("Expected at least one window")
		}
		
		for i, window := range windows {
			elements, _ := Collect(window)
			t.Logf("Time window %d: %v", i, elements)
		}
	})
	
	t.Run("MultipleTriggers", func(t *testing.T) {
		data := make([]int, 50) // Large dataset
		for i := range data {
			data[i] = i + 1
		}
		
		windows, err := Collect(
			Window[int]().
				TriggerOnCount(10).  // Fire every 10 elements
				TriggerOnTime(100 * time.Millisecond).  // Or every 100ms
				Apply()(
				FromSlice(data)))
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Should have multiple windows
		if len(windows) < 3 {
			t.Errorf("Expected at least 3 windows, got %d", len(windows))
		}
		
		t.Logf("Created %d windows with multiple triggers", len(windows))
	})
	
	t.Run("LatenessHandling", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		
		windows, err := Collect(
			Window[int]().
				TriggerOnCount(3).
				AllowLateness(1 * time.Second).
				AccumulationMode().  // Late data accumulated
				Apply()(
				FromSlice(data)))
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		// Should handle configuration without errors
		if len(windows) == 0 {
			t.Error("Expected at least one window")
		}
		
		t.Logf("Lateness handling created %d windows", len(windows))
	})
}

// TestTriggerImplementations verifies individual trigger behavior
func TestTriggerImplementations(t *testing.T) {
	t.Run("CountTrigger", func(t *testing.T) {
		trigger := NewAdvancedCountTrigger[int](3)
		window := &WindowState[int]{
			StartTime: time.Now(),
			Elements:  []int{},
		}
		
		// Should not fire until threshold reached
		if trigger.ShouldFire(window, 1, time.Now()) {
			t.Error("Should not fire on first element")
		}
		
		if trigger.ShouldFire(window, 2, time.Now()) {
			t.Error("Should not fire on second element")
		}
		
		if !trigger.ShouldFire(window, 3, time.Now()) {
			t.Error("Should fire on third element")
		}
		
		// Reset should allow firing again
		trigger.Reset()
		if trigger.ShouldFire(window, 4, time.Now()) {
			t.Error("Should not fire immediately after reset")
		}
	})
	
	t.Run("TimeTrigger", func(t *testing.T) {
		duration := 100 * time.Millisecond
		trigger := NewAdvancedTimeTrigger[int](duration)
		
		startTime := time.Now()
		window := &WindowState[int]{
			StartTime: startTime,
			Elements:  []int{},
		}
		
		// Should not fire immediately
		if trigger.ShouldFire(window, 1, startTime) {
			t.Error("Should not fire immediately")
		}
		
		// Should fire after duration
		time.Sleep(duration + 10*time.Millisecond)
		if !trigger.ShouldFire(window, 2, time.Now()) {
			t.Error("Should fire after duration")
		}
	})
	
	t.Run("ProcessingTimeTrigger", func(t *testing.T) {
		interval := 50 * time.Millisecond
		trigger := NewAdvancedProcessingTimeTrigger[int](interval)
		window := &WindowState[int]{StartTime: time.Now()}
		
		// Should not fire immediately
		if trigger.ShouldFire(window, 1, time.Now()) {
			t.Error("Should not fire immediately")
		}
		
		// Should fire after interval
		time.Sleep(interval + 10*time.Millisecond)
		if !trigger.ShouldFire(window, 2, time.Now()) {
			t.Error("Should fire after processing time interval")
		}
	})
}

// TestAdvancedWindowingScenarios tests complex real-world scenarios
func TestAdvancedWindowingScenarios(t *testing.T) {
	t.Run("UserSessionAnalysis", func(t *testing.T) {
		// Simulate user activity events
		type UserEvent struct {
			UserID    string
			EventType string
			Timestamp time.Time
		}
		
		events := []UserEvent{
			{"user1", "login", time.Now()},
			{"user1", "page_view", time.Now().Add(1 * time.Second)},
			{"user1", "click", time.Now().Add(2 * time.Second)},
			{"user1", "purchase", time.Now().Add(3 * time.Second)},
			{"user1", "logout", time.Now().Add(10 * time.Second)},
		}
		
		// Activity detector for user sessions
		activityDetector := func(event UserEvent) bool {
			return event.EventType == "login" || event.EventType == "purchase" || event.EventType == "logout"
		}
		
		sessionTimeout := 5 * time.Second
		
		sessions, err := Collect(
			SessionWindow(sessionTimeout, activityDetector)(
				FromSliceAny(events)))
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		t.Logf("User session analysis created %d sessions", len(sessions))
		
		for i, session := range sessions {
			sessionEvents, _ := Collect(session)
			t.Logf("Session %d: %d events", i, len(sessionEvents))
		}
	})
	
	t.Run("MetricsCollection", func(t *testing.T) {
		// Simulate metrics data
		type Metric struct {
			Name  string
			Value float64
			Time  time.Time
		}
		
		metrics := make([]Metric, 100)
		for i := range metrics {
			metrics[i] = Metric{
				Name:  fmt.Sprintf("cpu_usage_%d", i%10),
				Value: float64(i % 100),
				Time:  time.Now().Add(time.Duration(i) * time.Millisecond),
			}
		}
		
		// Collect metrics every 10 elements or every 50ms
		windows, err := Collect(
			Window[Metric]().
				TriggerOnCount(10).
				TriggerOnTime(50 * time.Millisecond).
				Apply()(
				FromSliceAny(metrics)))
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		t.Logf("Metrics collection created %d windows", len(windows))
		
		// Verify windows contain metrics
		totalMetrics := 0
		for _, window := range windows {
			windowMetrics, _ := Collect(window)
			totalMetrics += len(windowMetrics)
		}
		
		if totalMetrics != len(metrics) {
			t.Errorf("Expected %d total metrics, got %d", len(metrics), totalMetrics)
		}
	})
}

// BenchmarkAdvancedWindowing measures performance of advanced windowing
func BenchmarkAdvancedWindowing(b *testing.B) {
	data := make([]int, 10000)
	for i := range data {
		data[i] = i
	}
	
	b.Run("SessionWindow", func(b *testing.B) {
		activityDetector := func(n int) bool {
			return n%100 == 0 // Every 100th element is activity
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Collect(
				SessionWindow(100*time.Millisecond, activityDetector)(
					FromSlice(data)))
		}
	})
	
	b.Run("WindowBuilder", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Collect(
				Window[int]().
					TriggerOnCount(100).
					TriggerOnTime(50 * time.Millisecond).
					Apply()(
					FromSlice(data)))
		}
	})
}