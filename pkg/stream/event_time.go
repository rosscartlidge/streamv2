package stream

import (
	"sort"
	"sync"
	"time"
)

// RecordTimestampExtractor extracts event time from Record elements
type RecordTimestampExtractor func(Record) time.Time

// WatermarkGenerator generates watermarks based on observed event times
type WatermarkGenerator func(maxEventTime time.Time) time.Time

// NewRecordTimestampExtractor creates a timestamp extractor for Record types
func NewRecordTimestampExtractor(fieldName string) RecordTimestampExtractor {
	return func(record Record) time.Time {
		if timestampValue, exists := record[fieldName]; exists {
			switch ts := timestampValue.(type) {
			case time.Time:
				return ts
			case string:
				if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
					return parsed
				}
			}
		}
		return time.Now()
	}
}

// BoundedOutOfOrdernessWatermark generates watermarks allowing specified lateness
func BoundedOutOfOrdernessWatermark(maxLateness time.Duration) WatermarkGenerator {
	return func(maxEventTime time.Time) time.Time {
		return maxEventTime.Add(-maxLateness)
	}
}

// PeriodicWatermarkGenerator creates a watermark generator that updates periodically
func PeriodicWatermarkGenerator(interval time.Duration, baseGenerator WatermarkGenerator) WatermarkGenerator {
	var lastUpdate time.Time
	var cachedWatermark time.Time
	var initialized bool

	return func(maxEventTime time.Time) time.Time {
		now := time.Now()

		// Initialize or update if interval has passed
		if !initialized || now.Sub(lastUpdate) >= interval {
			cachedWatermark = baseGenerator(maxEventTime)
			lastUpdate = now
			initialized = true
		}

		return cachedWatermark
	}
}

// ============================================================================
// WATERMARK TRACKING
// ============================================================================

// WatermarkTracker manages watermark progression for event-time processing
type WatermarkTracker struct {
	mu               sync.RWMutex
	currentWatermark time.Time
	maxEventTime     time.Time
	generator        WatermarkGenerator
	initialized      bool
}

// NewWatermarkTracker creates a new watermark tracker
func NewWatermarkTracker(generator WatermarkGenerator) *WatermarkTracker {
	return &WatermarkTracker{
		generator: generator,
	}
}

// UpdateWatermark processes a new event time and potentially advances the watermark
func (wt *WatermarkTracker) UpdateWatermark(eventTime time.Time) time.Time {
	wt.mu.Lock()
	defer wt.mu.Unlock()

	// Track the maximum event time seen
	if !wt.initialized || eventTime.After(wt.maxEventTime) {
		wt.maxEventTime = eventTime
		wt.initialized = true

		// Generate new watermark
		newWatermark := wt.generator(wt.maxEventTime)

		// Watermarks can only advance (monotonic property)
		if newWatermark.After(wt.currentWatermark) {
			wt.currentWatermark = newWatermark
		}
	}

	return wt.currentWatermark
}

// GetWatermark returns the current watermark
func (wt *WatermarkTracker) GetWatermark() time.Time {
	wt.mu.RLock()
	defer wt.mu.RUnlock()
	return wt.currentWatermark
}

// ============================================================================
// EVENT-TIME WINDOW CONFIGURATION
// ============================================================================

// LateDataPolicy defines how to handle data that arrives after window firing
type LateDataPolicy int

const (
	DropLateData   LateDataPolicy = iota // Ignore late-arriving data
	UpdateWindow                         // Update window results with late data
	SideOutputLate                       // Send late data to separate stream
)

// EventTimeWindowConfig holds configuration for event-time windows on Records
type EventTimeWindowConfig struct {
	TimestampExtractor RecordTimestampExtractor
	WatermarkGenerator WatermarkGenerator
	LateDataPolicy     LateDataPolicy
	AllowedLateness    time.Duration
}

// EventTimeWindowOption configures event-time windows
type EventTimeWindowOption func(*EventTimeWindowConfig)

// WithTimestampExtractor sets the timestamp extraction function
func WithTimestampExtractor(extractor RecordTimestampExtractor) EventTimeWindowOption {
	return func(config *EventTimeWindowConfig) {
		config.TimestampExtractor = extractor
	}
}

// WithWatermarkGenerator sets the watermark generation strategy
func WithWatermarkGenerator(generator WatermarkGenerator) EventTimeWindowOption {
	return func(config *EventTimeWindowConfig) {
		config.WatermarkGenerator = generator
	}
}

// WithLateDataPolicy sets how late data should be handled
func WithLateDataPolicy(policy LateDataPolicy) EventTimeWindowOption {
	return func(config *EventTimeWindowConfig) {
		config.LateDataPolicy = policy
	}
}

// WithAllowedLateness sets the maximum allowed lateness (convenience for watermark generation)
func WithAllowedLateness(lateness time.Duration) EventTimeWindowOption {
	return func(config *EventTimeWindowConfig) {
		config.AllowedLateness = lateness
		if config.WatermarkGenerator == nil {
			config.WatermarkGenerator = BoundedOutOfOrdernessWatermark(lateness)
		}
	}
}

// ============================================================================
// EVENT-TIME WINDOW STATE
// ============================================================================

// TimestampedRecord pairs a Record with its event timestamp
type TimestampedRecord struct {
	Record    Record
	Timestamp time.Time
}

// EventTimeWindowState tracks the state of an event-time window for Records
type EventTimeWindowState struct {
	mu           sync.RWMutex
	elements     []TimestampedRecord
	windowStart  time.Time
	windowEnd    time.Time
	fired        bool
	latePolicy   LateDataPolicy
	lateElements []TimestampedRecord // Elements that arrived after firing
}

// NewEventTimeWindowState creates a new event-time window state for Records
func NewEventTimeWindowState(start, end time.Time, policy LateDataPolicy) *EventTimeWindowState {
	return &EventTimeWindowState{
		windowStart:  start,
		windowEnd:    end,
		latePolicy:   policy,
		elements:     make([]TimestampedRecord, 0),
		lateElements: make([]TimestampedRecord, 0),
	}
}

// AddElement adds an element to the window if it belongs to this window
func (ws *EventTimeWindowState) AddElement(element Record, eventTime time.Time) bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Check if element belongs to this window
	if eventTime.Before(ws.windowStart) || !eventTime.Before(ws.windowEnd) {
		return false // Element doesn't belong to this window
	}

	timestampedElement := TimestampedRecord{
		Record:    element,
		Timestamp: eventTime,
	}

	if ws.fired {
		// Window already fired - handle as late data
		switch ws.latePolicy {
		case DropLateData:
			return true // Drop silently
		case UpdateWindow, SideOutputLate:
			ws.lateElements = append(ws.lateElements, timestampedElement)
		}
	} else {
		// Normal case - add to window
		ws.elements = append(ws.elements, timestampedElement)
	}

	return true
}

// ShouldFire determines if the window should fire based on watermark
func (ws *EventTimeWindowState) ShouldFire(watermark time.Time) bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return !ws.fired && !watermark.Before(ws.windowEnd)
}

// Fire triggers the window to emit results
func (ws *EventTimeWindowState) Fire() []Record {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.fired {
		return nil // Already fired
	}

	ws.fired = true

	// Sort elements by timestamp for consistent ordering
	sort.Slice(ws.elements, func(i, j int) bool {
		return ws.elements[i].Timestamp.Before(ws.elements[j].Timestamp)
	})

	// Extract just the elements (not timestamps)
	result := make([]Record, len(ws.elements))
	for i, elem := range ws.elements {
		result[i] = elem.Record
	}

	return result
}

// HasFired returns true if the window has already fired
func (ws *EventTimeWindowState) HasFired() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.fired
}

// ============================================================================
// SIMPLE EVENT-TIME TUMBLING WINDOW
// ============================================================================

// EventTimeTumblingWindow creates tumbling windows based on event time for Records
func EventTimeTumblingWindow(
	windowSize time.Duration,
	options ...EventTimeWindowOption,
) func(Stream[Record]) Stream[Stream[Record]] {

	// Apply default configuration
	config := &EventTimeWindowConfig{
		LateDataPolicy:     DropLateData,
		WatermarkGenerator: BoundedOutOfOrdernessWatermark(30 * time.Second), // Default 30s lateness
	}

	for _, option := range options {
		option(config)
	}

	if config.TimestampExtractor == nil {
		panic("EventTimeTumblingWindow requires a timestamp extractor")
	}

	return func(input Stream[Record]) Stream[Stream[Record]] {
		watermarkTracker := NewWatermarkTracker(config.WatermarkGenerator)
		windowsMap := make(map[time.Time]*EventTimeWindowState)
		var mu sync.RWMutex

		return func() (Stream[Record], error) {
			for {
				// Get next element from input stream
				element, err := input()
				if err == EOS {
					// Handle end of stream - fire all remaining windows
					mu.Lock()

					// Find any windows that should fire
					currentWatermark := watermarkTracker.GetWatermark()
					var readyWindows []*EventTimeWindowState

					for _, window := range windowsMap {
						if window.ShouldFire(currentWatermark) || err == EOS {
							readyWindows = append(readyWindows, window)
						}
					}

					if len(readyWindows) > 0 {
						// Fire the earliest window
						sort.Slice(readyWindows, func(i, j int) bool {
							return readyWindows[i].windowStart.Before(readyWindows[j].windowStart)
						})

						window := readyWindows[0]
						result := window.Fire()

						// Remove fired window from map
						delete(windowsMap, window.windowStart)
						mu.Unlock()

						if len(result) > 0 {
							return FromSlice(result), nil
						}
						// Continue to check for more windows
						continue
					}

					mu.Unlock()
					return nil, EOS
				}

				if err != nil {
					return nil, err
				}

				// Extract event time
				eventTime := config.TimestampExtractor(element)

				// Update watermark
				watermark := watermarkTracker.UpdateWatermark(eventTime)

				mu.Lock()

				// Determine which window this element belongs to
				windowStart := eventTime.Truncate(windowSize)

				// Get or create window
				window, exists := windowsMap[windowStart]
				if !exists {
					windowEnd := windowStart.Add(windowSize)
					window = NewEventTimeWindowState(windowStart, windowEnd, config.LateDataPolicy)
					windowsMap[windowStart] = window
				}

				// Add element to window
				window.AddElement(element, eventTime)

				// Check if any windows should fire
				var readyWindows []*EventTimeWindowState
				for _, w := range windowsMap {
					if w.ShouldFire(watermark) {
						readyWindows = append(readyWindows, w)
					}
				}

				if len(readyWindows) > 0 {
					// Fire the earliest ready window
					sort.Slice(readyWindows, func(i, j int) bool {
						return readyWindows[i].windowStart.Before(readyWindows[j].windowStart)
					})

					windowToFire := readyWindows[0]
					result := windowToFire.Fire()

					// Remove fired window from map
					delete(windowsMap, windowToFire.windowStart)
					mu.Unlock()

					if len(result) > 0 {
						return FromSlice(result), nil
					}
					// Continue processing if window was empty
					continue
				}

				mu.Unlock()
				// Continue processing more elements
			}
		}
	}
}

// ============================================================================
// EVENT-TIME SLIDING WINDOW
// ============================================================================

// EventTimeSlidingWindow creates sliding windows based on event time for Records
func EventTimeSlidingWindow(
	windowSize time.Duration,
	slideInterval time.Duration,
	options ...EventTimeWindowOption,
) func(Stream[Record]) Stream[Stream[Record]] {

	// Apply default configuration
	config := &EventTimeWindowConfig{
		LateDataPolicy:     DropLateData,
		WatermarkGenerator: BoundedOutOfOrdernessWatermark(30 * time.Second), // Default 30s lateness
	}

	for _, option := range options {
		option(config)
	}

	if config.TimestampExtractor == nil {
		panic("EventTimeSlidingWindow requires a timestamp extractor")
	}

	return func(input Stream[Record]) Stream[Stream[Record]] {
		watermarkTracker := NewWatermarkTracker(config.WatermarkGenerator)
		windowsMap := make(map[time.Time]*EventTimeWindowState)
		var mu sync.RWMutex

		return func() (Stream[Record], error) {
			for {
				// Get next element from input stream
				element, err := input()
				if err == EOS {
					// Handle end of stream - fire all remaining windows
					mu.Lock()

					// Find any windows that should fire
					currentWatermark := watermarkTracker.GetWatermark()
					var readyWindows []*EventTimeWindowState

					for _, window := range windowsMap {
						if window.ShouldFire(currentWatermark) || err == EOS {
							readyWindows = append(readyWindows, window)
						}
					}

					if len(readyWindows) > 0 {
						// Fire the earliest window
						sort.Slice(readyWindows, func(i, j int) bool {
							return readyWindows[i].windowStart.Before(readyWindows[j].windowStart)
						})

						window := readyWindows[0]
						result := window.Fire()

						// Remove fired window from map
						delete(windowsMap, window.windowStart)
						mu.Unlock()

						if len(result) > 0 {
							return FromSlice(result), nil
						}
						// Continue to check for more windows
						continue
					}

					mu.Unlock()
					return nil, EOS
				}

				if err != nil {
					return nil, err
				}

				// Extract event time
				eventTime := config.TimestampExtractor(element)

				// Update watermark
				watermark := watermarkTracker.UpdateWatermark(eventTime)

				mu.Lock()

				// For sliding windows, create multiple windows for each slide
				// Calculate which windows this element should belong to
				windowStart := eventTime.Truncate(slideInterval)
				for {
					// Check if this element falls within a window starting at windowStart
					windowEnd := windowStart.Add(windowSize)
					if eventTime.Before(windowStart) || !eventTime.Before(windowEnd) {
						// Element doesn't belong to this window, try previous window
						windowStart = windowStart.Add(-slideInterval)
						if windowStart.Add(windowSize).Before(eventTime) {
							// Element is too far ahead, stop looking backwards
							break
						}
						continue
					}

					// Element belongs to this window - get or create it
					window, exists := windowsMap[windowStart]
					if !exists {
						window = NewEventTimeWindowState(windowStart, windowEnd, config.LateDataPolicy)
						windowsMap[windowStart] = window
					}

					// Add element to window
					window.AddElement(element, eventTime)

					// Move to next potential window
					windowStart = windowStart.Add(-slideInterval)
					if windowStart.Add(windowSize).Before(eventTime) {
						// Element is too far ahead, stop looking backwards
						break
					}
				}

				// Check if any windows should fire
				var readyWindows []*EventTimeWindowState
				for _, w := range windowsMap {
					if w.ShouldFire(watermark) {
						readyWindows = append(readyWindows, w)
					}
				}

				if len(readyWindows) > 0 {
					// Fire the earliest ready window
					sort.Slice(readyWindows, func(i, j int) bool {
						return readyWindows[i].windowStart.Before(readyWindows[j].windowStart)
					})

					windowToFire := readyWindows[0]
					result := windowToFire.Fire()

					// Remove fired window from map
					delete(windowsMap, windowToFire.windowStart)
					mu.Unlock()

					if len(result) > 0 {
						return FromSlice(result), nil
					}
					// Continue processing if window was empty
					continue
				}

				mu.Unlock()
				// Continue processing more elements
			}
		}
	}
}

// ============================================================================
// EVENT-TIME SESSION WINDOW
// ============================================================================

// EventTimeSessionState tracks the state of an event-time session window
type EventTimeSessionState struct {
	mu           sync.RWMutex
	elements     []TimestampedRecord
	sessionStart time.Time
	sessionEnd   time.Time
	lastActivity time.Time
	fired        bool
	latePolicy   LateDataPolicy
	lateElements []TimestampedRecord
}

// NewEventTimeSessionState creates a new session state
func NewEventTimeSessionState(policy LateDataPolicy) *EventTimeSessionState {
	return &EventTimeSessionState{
		latePolicy:   policy,
		elements:     make([]TimestampedRecord, 0),
		lateElements: make([]TimestampedRecord, 0),
	}
}

// AddElement adds an element to the session
func (ss *EventTimeSessionState) AddElement(element Record, eventTime time.Time) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	timestampedElement := TimestampedRecord{
		Record:    element,
		Timestamp: eventTime,
	}

	if ss.fired {
		// Session already fired - handle as late data
		switch ss.latePolicy {
		case DropLateData:
			return // Drop silently
		case UpdateWindow, SideOutputLate:
			ss.lateElements = append(ss.lateElements, timestampedElement)
		}
	} else {
		// Normal case - add to session
		if len(ss.elements) == 0 {
			// First element in session
			ss.sessionStart = eventTime
			ss.sessionEnd = eventTime
		} else {
			// Update session bounds
			if eventTime.Before(ss.sessionStart) {
				ss.sessionStart = eventTime
			}
			if eventTime.After(ss.sessionEnd) {
				ss.sessionEnd = eventTime
			}
		}

		ss.elements = append(ss.elements, timestampedElement)
		ss.lastActivity = eventTime
	}
}

// ShouldFire determines if the session should fire based on timeout and watermark
func (ss *EventTimeSessionState) ShouldFire(watermark time.Time, sessionTimeout time.Duration) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if ss.fired || len(ss.elements) == 0 {
		return false
	}

	// Session fires when watermark passes the last activity + timeout
	sessionEndTime := ss.lastActivity.Add(sessionTimeout)
	return !watermark.Before(sessionEndTime)
}

// Fire triggers the session to emit results
func (ss *EventTimeSessionState) Fire() []Record {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.fired {
		return nil // Already fired
	}

	ss.fired = true

	// Sort elements by timestamp for consistent ordering
	sort.Slice(ss.elements, func(i, j int) bool {
		return ss.elements[i].Timestamp.Before(ss.elements[j].Timestamp)
	})

	// Extract just the elements (not timestamps)
	result := make([]Record, len(ss.elements))
	for i, elem := range ss.elements {
		result[i] = elem.Record
	}

	return result
}

// HasFired returns true if the session has already fired
func (ss *EventTimeSessionState) HasFired() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.fired
}

// EventTimeSessionWindow creates session windows based on event time for Records
func EventTimeSessionWindow(
	sessionTimeout time.Duration,
	options ...EventTimeWindowOption,
) func(Stream[Record]) Stream[Stream[Record]] {

	// Apply default configuration
	config := &EventTimeWindowConfig{
		LateDataPolicy:     DropLateData,
		WatermarkGenerator: BoundedOutOfOrdernessWatermark(30 * time.Second), // Default 30s lateness
	}

	for _, option := range options {
		option(config)
	}

	if config.TimestampExtractor == nil {
		panic("EventTimeSessionWindow requires a timestamp extractor")
	}

	return func(input Stream[Record]) Stream[Stream[Record]] {
		watermarkTracker := NewWatermarkTracker(config.WatermarkGenerator)
		sessionsMap := make(map[string]*EventTimeSessionState) // Using string key for session ID
		var mu sync.RWMutex

		return func() (Stream[Record], error) {
			for {
				// Get next element from input stream
				element, err := input()
				if err == EOS {
					// Handle end of stream - fire all remaining sessions
					mu.Lock()

					// Find any sessions that should fire
					currentWatermark := watermarkTracker.GetWatermark()
					var readySessions []*EventTimeSessionState

					for _, session := range sessionsMap {
						if session.ShouldFire(currentWatermark, sessionTimeout) || err == EOS {
							readySessions = append(readySessions, session)
						}
					}

					if len(readySessions) > 0 {
						// Fire the earliest session (by last activity)
						sort.Slice(readySessions, func(i, j int) bool {
							return readySessions[i].lastActivity.Before(readySessions[j].lastActivity)
						})

						session := readySessions[0]
						result := session.Fire()

						// Remove fired session from map
						for id, s := range sessionsMap {
							if s == session {
								delete(sessionsMap, id)
								break
							}
						}
						mu.Unlock()

						if len(result) > 0 {
							return FromSlice(result), nil
						}
						// Continue to check for more sessions
						continue
					}

					mu.Unlock()
					return nil, EOS
				}

				if err != nil {
					return nil, err
				}

				// Extract event time
				eventTime := config.TimestampExtractor(element)

				// Update watermark
				watermark := watermarkTracker.UpdateWatermark(eventTime)

				mu.Lock()

				// Determine session ID (for now, use a global session - could be extended to per-user sessions)
				sessionID := "global" // In practice, this could be extracted from the record

				// Get or create session
				session, exists := sessionsMap[sessionID]
				if !exists {
					session = NewEventTimeSessionState(config.LateDataPolicy)
					sessionsMap[sessionID] = session
				}

				// Check if this element extends the session or if the session has timed out
				if len(session.elements) > 0 && eventTime.Sub(session.lastActivity) > sessionTimeout {
					// Session has timed out - fire current session and start new one
					if session.ShouldFire(watermark, sessionTimeout) {
						result := session.Fire()
						delete(sessionsMap, sessionID)

						// Create new session for this element
						newSession := NewEventTimeSessionState(config.LateDataPolicy)
						newSession.AddElement(element, eventTime)
						sessionsMap[sessionID] = newSession

						mu.Unlock()

						if len(result) > 0 {
							return FromSlice(result), nil
						}
						continue
					}
				}

				// Add element to session
				session.AddElement(element, eventTime)

				// Check if any sessions should fire
				var readySessions []*EventTimeSessionState
				for _, s := range sessionsMap {
					if s.ShouldFire(watermark, sessionTimeout) {
						readySessions = append(readySessions, s)
					}
				}

				if len(readySessions) > 0 {
					// Fire the earliest ready session
					sort.Slice(readySessions, func(i, j int) bool {
						return readySessions[i].lastActivity.Before(readySessions[j].lastActivity)
					})

					sessionToFire := readySessions[0]
					result := sessionToFire.Fire()

					// Remove fired session from map
					for id, s := range sessionsMap {
						if s == sessionToFire {
							delete(sessionsMap, id)
							break
						}
					}
					mu.Unlock()

					if len(result) > 0 {
						return FromSlice(result), nil
					}
					// Continue processing if session was empty
					continue
				}

				mu.Unlock()
				// Continue processing more elements
			}
		}
	}
}
