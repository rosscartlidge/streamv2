package stream

import (
	"context"
	"sync"
	"time"
)

// ============================================================================
// ADVANCED WINDOWING SYSTEM
// ============================================================================

// WindowBuilder provides a fluent API for configuring advanced windows
type WindowBuilder[T any] struct {
	triggers     []AdvancedTrigger[T]
	lateness     time.Duration
	accumMode    AccumulationMode
	allowedLateness bool
}

// AccumulationMode defines how late data is handled
type AccumulationMode int

const (
	DiscardingMode AccumulationMode = iota // Discard late data
	AccumulateMode                         // Accumulate late data into existing windows
)

// AdvancedTriggerType defines different trigger conditions for advanced windowing
type AdvancedTriggerType int

const (
	AdvancedCountTriggerType AdvancedTriggerType = iota
	AdvancedTimeTriggerType
	AdvancedProcessingTimeTriggerType
	AdvancedCustomTriggerType
)

// AdvancedTrigger defines when a window should fire/emit results in advanced windowing
type AdvancedTrigger[T any] interface {
	// ShouldFire returns true if the window should fire based on current state
	ShouldFire(window *WindowState[T], element T, processingTime time.Time) bool
	
	// Reset resets the trigger state (called after firing)
	Reset()
	
	// GetType returns the trigger type for optimization
	GetType() AdvancedTriggerType
}

// WindowState holds the current state of a window
type WindowState[T any] struct {
	ID           string
	StartTime    time.Time
	EndTime      time.Time
	Elements     []T
	ElementCount int
	LastActivity time.Time
	LastFired    time.Time
	Metadata     map[string]interface{}
	mu           sync.RWMutex
}

// ActivityDetector determines if an element represents activity for session windows
type ActivityDetector[T any] func(T) bool

// ============================================================================
// WINDOW BUILDER - FLUENT API
// ============================================================================

// Window creates a new window builder for advanced windowing configuration
func Window[T any]() *WindowBuilder[T] {
	return &WindowBuilder[T]{
		triggers:        make([]AdvancedTrigger[T], 0),
		lateness:        0,
		accumMode:       DiscardingMode,
		allowedLateness: false,
	}
}

// TriggerOnCount adds a count-based trigger to the window
func (wb *WindowBuilder[T]) TriggerOnCount(count int) *WindowBuilder[T] {
	wb.triggers = append(wb.triggers, NewAdvancedCountTrigger[T](count))
	return wb
}

// TriggerOnTime adds a time-based trigger to the window
func (wb *WindowBuilder[T]) TriggerOnTime(duration time.Duration) *WindowBuilder[T] {
	wb.triggers = append(wb.triggers, NewAdvancedTimeTrigger[T](duration))
	return wb
}

// TriggerOnProcessingTime adds a processing time trigger
func (wb *WindowBuilder[T]) TriggerOnProcessingTime(interval time.Duration) *WindowBuilder[T] {
	wb.triggers = append(wb.triggers, NewAdvancedProcessingTimeTrigger[T](interval))
	return wb
}

// TriggerOn adds a custom trigger to the window
func (wb *WindowBuilder[T]) TriggerOn(trigger AdvancedTrigger[T]) *WindowBuilder[T] {
	wb.triggers = append(wb.triggers, trigger)
	return wb
}

// AllowLateness configures how late data should be handled
func (wb *WindowBuilder[T]) AllowLateness(lateness time.Duration) *WindowBuilder[T] {
	wb.lateness = lateness
	wb.allowedLateness = true
	return wb
}

// AccumulationMode sets the accumulation mode for late data
func (wb *WindowBuilder[T]) AccumulationMode() *WindowBuilder[T] {
	wb.accumMode = AccumulateMode
	return wb
}

// DiscardingMode sets the discarding mode for late data (default)
func (wb *WindowBuilder[T]) DiscardingMode() *WindowBuilder[T] {
	wb.accumMode = DiscardingMode
	return wb
}

// Apply creates the windowing filter with the configured settings
func (wb *WindowBuilder[T]) Apply() Filter[T, Stream[T]] {
	return func(input Stream[T]) Stream[Stream[T]] {
		return wb.createAdvancedWindow(input)
	}
}

// ============================================================================
// SESSION WINDOWS
// ============================================================================

// SessionWindow creates activity-based windows that group related events
// Windows extend as long as activity is detected within the timeout period
func SessionWindow[T any](timeout time.Duration, activityDetector ActivityDetector[T]) Filter[T, Stream[T]] {
	return func(input Stream[T]) Stream[Stream[T]] {
		ctx, cancel := context.WithCancel(context.Background())
		windowCh := make(chan Stream[T], 10)
		
		go func() {
			defer close(windowCh)
			defer cancel()
			
			sessions := make(map[string]*SessionState[T])
			var mu sync.RWMutex
			
			// Cleanup goroutine for expired sessions
			go func() {
				ticker := time.NewTicker(timeout / 2)
				defer ticker.Stop()
				
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						mu.Lock()
						now := time.Now()
						for id, session := range sessions {
							if now.Sub(session.LastActivity) > timeout {
								// Session expired - emit window
								if len(session.Elements) > 0 {
									windowCh <- FromSlice(session.Elements)
								}
								delete(sessions, id)
							}
						}
						mu.Unlock()
					}
				}
			}()
			
			// Process input stream
			for {
				element, err := input()
				if err != nil {
					// End of stream - emit all remaining sessions
					mu.Lock()
					for _, session := range sessions {
						if len(session.Elements) > 0 {
							windowCh <- FromSlice(session.Elements)
						}
					}
					mu.Unlock()
					return
				}
				
				now := time.Now()
				isActivity := activityDetector(element)
				
				if isActivity {
					mu.Lock()
					// Find or create session for this activity
					sessionID := generateSessionID(element, now)
					
					if session, exists := sessions[sessionID]; exists {
						// Extend existing session
						session.Elements = append(session.Elements, element)
						session.LastActivity = now
					} else {
						// Create new session
						sessions[sessionID] = &SessionState[T]{
							ID:           sessionID,
							StartTime:    now,
							LastActivity: now,
							Elements:     []T{element},
						}
					}
					mu.Unlock()
				} else {
					// Non-activity elements extend the most recent session if within timeout
					mu.Lock()
					var mostRecentSession *SessionState[T]
					var mostRecentTime time.Time
					
					for _, session := range sessions {
						if session.LastActivity.After(mostRecentTime) {
							mostRecentTime = session.LastActivity
							mostRecentSession = session
						}
					}
					
					if mostRecentSession != nil && now.Sub(mostRecentSession.LastActivity) <= timeout {
						mostRecentSession.Elements = append(mostRecentSession.Elements, element)
						mostRecentSession.LastActivity = now
					}
					mu.Unlock()
				}
			}
		}()
		
		// Return stream of windows
		return func() (Stream[T], error) {
			window, ok := <-windowCh
			if !ok {
				return nil, EOS
			}
			return window, nil
		}
	}
}

// SessionState holds the state of a session window
type SessionState[T any] struct {
	ID           string
	StartTime    time.Time
	LastActivity time.Time
	Elements     []T
}

// generateSessionID creates a unique session identifier
func generateSessionID[T any](element T, timestamp time.Time) string {
	// Simple session ID generation - in practice this might be more sophisticated
	return timestamp.Format("2006-01-02T15:04:05.000Z")
}

// ============================================================================
// TRIGGER IMPLEMENTATIONS
// ============================================================================

// AdvancedCountTrigger fires when element count reaches threshold
type AdvancedCountTrigger[T any] struct {
	threshold int
	current   int
}

func NewAdvancedCountTrigger[T any](threshold int) *AdvancedCountTrigger[T] {
	return &AdvancedCountTrigger[T]{threshold: threshold}
}

func (t *AdvancedCountTrigger[T]) ShouldFire(window *WindowState[T], element T, processingTime time.Time) bool {
	t.current++
	return t.current >= t.threshold
}

func (t *AdvancedCountTrigger[T]) Reset() {
	t.current = 0
}

func (t *AdvancedCountTrigger[T]) GetType() AdvancedTriggerType {
	return AdvancedCountTriggerType
}

// AdvancedTimeTrigger fires after a duration since window start
type AdvancedTimeTrigger[T any] struct {
	duration time.Duration
}

func NewAdvancedTimeTrigger[T any](duration time.Duration) *AdvancedTimeTrigger[T] {
	return &AdvancedTimeTrigger[T]{duration: duration}
}

func (t *AdvancedTimeTrigger[T]) ShouldFire(window *WindowState[T], element T, processingTime time.Time) bool {
	return processingTime.Sub(window.StartTime) >= t.duration
}

func (t *AdvancedTimeTrigger[T]) Reset() {
	// Time triggers don't need reset
}

func (t *AdvancedTimeTrigger[T]) GetType() AdvancedTriggerType {
	return AdvancedTimeTriggerType
}

// AdvancedProcessingTimeTrigger fires at regular processing time intervals
type AdvancedProcessingTimeTrigger[T any] struct {
	interval time.Duration
	lastFire time.Time
}

func NewAdvancedProcessingTimeTrigger[T any](interval time.Duration) *AdvancedProcessingTimeTrigger[T] {
	return &AdvancedProcessingTimeTrigger[T]{
		interval: interval,
		lastFire: time.Now(),
	}
}

func (t *AdvancedProcessingTimeTrigger[T]) ShouldFire(window *WindowState[T], element T, processingTime time.Time) bool {
	return processingTime.Sub(t.lastFire) >= t.interval
}

func (t *AdvancedProcessingTimeTrigger[T]) Reset() {
	t.lastFire = time.Now()
}

func (t *AdvancedProcessingTimeTrigger[T]) GetType() AdvancedTriggerType {
	return AdvancedProcessingTimeTriggerType
}

// ============================================================================
// ADVANCED WINDOW IMPLEMENTATION
// ============================================================================

func (wb *WindowBuilder[T]) createAdvancedWindow(input Stream[T]) Stream[Stream[T]] {
	ctx, cancel := context.WithCancel(context.Background())
	windowCh := make(chan Stream[T], 10)
	
	go func() {
		defer close(windowCh)
		defer cancel()
		
		window := &WindowState[T]{
			ID:           generateWindowID(),
			StartTime:    time.Now(),
			Elements:     make([]T, 0),
			ElementCount: 0,
			Metadata:     make(map[string]interface{}),
		}
		
		for {
			select {
			case <-ctx.Done():
				return
			default:
				element, err := input()
				if err != nil {
					// End of stream - emit final window if it has elements
					if len(window.Elements) > 0 {
						windowCh <- FromSlice(window.Elements)
					}
					return
				}
				
				now := time.Now()
				
				// Check if element is late
				if wb.allowedLateness && wb.isLateData(element, now) {
					if wb.accumMode == DiscardingMode {
						continue // Discard late data
					}
					// AccumulateMode: add to existing window
				}
				
				// Add element to window
				window.mu.Lock()
				window.Elements = append(window.Elements, element)
				window.ElementCount++
				window.LastActivity = now
				window.mu.Unlock()
				
				// Check all triggers
				shouldFire := false
				for _, trigger := range wb.triggers {
					if trigger.ShouldFire(window, element, now) {
						shouldFire = true
						trigger.Reset()
					}
				}
				
				// Fire window if any trigger activated
				if shouldFire {
					windowCh <- FromSlice(window.Elements)
					
					// Reset window for next batch
					window = &WindowState[T]{
						ID:           generateWindowID(),
						StartTime:    now,
						Elements:     make([]T, 0),
						ElementCount: 0,
						Metadata:     make(map[string]interface{}),
					}
				}
			}
		}
	}()
	
	// Return stream of windows
	return func() (Stream[T], error) {
		window, ok := <-windowCh
		if !ok {
			return nil, EOS
		}
		return window, nil
	}
}

// isLateData determines if an element is late based on configured lateness
func (wb *WindowBuilder[T]) isLateData(element T, processingTime time.Time) bool {
	// This is a simplified implementation
	// In practice, you might extract timestamp from the element
	// and compare with processing time minus lateness allowance
	return false
}

// generateWindowID creates a unique window identifier
func generateWindowID() string {
	return time.Now().Format("window-2006-01-02T15:04:05.000Z")
}