package stream

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
	
	"golang.org/x/sync/errgroup"
)

// ============================================================================
// STREAM TRANSFORMATION FILTERS
// ============================================================================

// Filter transforms one stream into another with full type flexibility
type Filter[T, U any] func(Stream[T]) Stream[U]

// ============================================================================
// FUNCTIONAL OPERATIONS - TYPE SAFE AND COMPOSABLE
// ============================================================================

// Map transforms each element in a stream
func Map[T, U any](fn func(T) U) Filter[T, U] {
	return func(input Stream[T]) Stream[U] {
		return func() (U, error) {
			item, err := input()
			if err != nil {
				var zero U
				return zero, err
			}
			return fn(item), nil
		}
	}
}

// Where keeps only elements matching a predicate
func Where[T any](predicate func(T) bool) Filter[T, T] {
	return func(input Stream[T]) Stream[T] {
		return func() (T, error) {
			for {
				item, err := input()
				if err != nil {
					var zero T
					return zero, err
				}
				if predicate(item) {
					return item, nil
				}
			}
		}
	}
}

// Take limits stream to first N elements
func Take[T any](n int) Filter[T, T] {
	return func(input Stream[T]) Stream[T] {
		count := 0
		return func() (T, error) {
			if count >= n {
				var zero T
				return zero, EOS
			}
			count++
			return input()
		}
	}
}

// Skip skips first N elements
func Skip[T any](n int) Filter[T, T] {
	return func(input Stream[T]) Stream[T] {
		skipped := 0
		return func() (T, error) {
			for skipped < n {
				if _, err := input(); err != nil {
					var zero T
					return zero, err
				}
				skipped++
			}
			return input()
		}
	}
}

// ============================================================================
// STREAM COMPOSITION - BEAUTIFUL CHAINING
// ============================================================================

// Pipe composes two filters
func Pipe[T, U, V any](f1 Filter[T, U], f2 Filter[U, V]) Filter[T, V] {
	return func(input Stream[T]) Stream[V] {
		return f2(f1(input))
	}
}

// Pipe3 composes three filters
func Pipe3[T, U, V, W any](f1 Filter[T, U], f2 Filter[U, V], f3 Filter[V, W]) Filter[T, W] {
	return func(input Stream[T]) Stream[W] {
		return f3(f2(f1(input)))
	}
}

// Chain applies multiple filters of the same type
func Chain[T any](filters ...Filter[T, T]) Filter[T, T] {
	return func(input Stream[T]) Stream[T] {
		result := input
		for _, filter := range filters {
			result = filter(result)
		}
		return result
	}
}

// ============================================================================
// RECORD-SPECIFIC OPERATIONS - SQL-LIKE POWER
// ============================================================================

// Select extracts specific fields from records
func Select(fields ...string) Filter[Record, Record] {
	return Map(func(r Record) Record {
		result := make(Record)
		for _, field := range fields {
			if val, exists := r[field]; exists {
				// Since we're copying from a valid Record, the fields should be valid
				// But let's validate to be safe
				if !isValidFieldType(val) {
					panic(fmt.Sprintf("Selected field '%s' has invalid type %s", field, getFieldTypeName(reflect.TypeOf(val))))
				}
				result[field] = val
			}
		}
		return result
	})
}

// Update modifies records
func Update(fn func(Record) Record) Filter[Record, Record] {
	return Map(func(r Record) Record {
		result := fn(r)
		// Validate the result to ensure user function returned valid Record
		if err := ValidateRecord(result); err != nil {
			panic(fmt.Sprintf("Update function returned invalid Record: %v", err))
		}
		return result
	})
}

// ExtractField gets a typed field from records
func ExtractField[T any](field string) Filter[Record, T] {
	return Map(func(r Record) T {
		val, _ := Get[T](r, field)
		return val
	})
}

// ============================================================================
// CONCURRENT PROCESSING
// ============================================================================

// Parallel processes elements concurrently using errgroup for proper lifecycle management
func Parallel[T, U any](workers int, fn func(T) U) Filter[T, U] {
	return func(input Stream[T]) Stream[U] {
		ctx, cancel := context.WithCancel(context.Background())
		g, gctx := errgroup.WithContext(ctx)
		
		inputCh := make(chan T, workers)
		outputCh := make(chan U, workers)

		// Start workers using errgroup
		for i := 0; i < workers; i++ {
			g.Go(func() error {
				for {
					select {
					case <-gctx.Done():
						return gctx.Err()
					case item, ok := <-inputCh:
						if !ok {
							return nil // Input closed normally
						}
						// Process item with cancellation check
						result := fn(item)
						select {
						case outputCh <- result:
						case <-gctx.Done():
							return gctx.Err()
						}
					}
				}
			})
		}

		// Feed input with errgroup
		g.Go(func() error {
			defer close(inputCh)
			
			for {
				select {
				case <-gctx.Done():
					return gctx.Err()
				default:
					item, err := input()
					if err != nil {
						return nil // Input stream ended normally
					}
					select {
					case inputCh <- item:
					case <-gctx.Done():
						return gctx.Err()
					}
				}
			}
		})

		// Cleanup goroutine 
		go func() {
			g.Wait() // Wait for all goroutines to finish
			close(outputCh)
		}()

		// Return cancellable stream
		return func() (U, error) {
			select {
			case <-gctx.Done():
				cancel() // Ensure cleanup
				var zero U
				return zero, gctx.Err()
			case item, ok := <-outputCh:
				if !ok {
					cancel() // Cleanup when stream ends
					var zero U
					return zero, EOS
				}
				return item, nil
			}
		}
	}
}

// ============================================================================
// CONTEXT SUPPORT
// ============================================================================

// WithContext adds context support to a stream
func WithContext[T any](ctx context.Context, stream Stream[T]) Stream[T] {
	return func() (T, error) {
		select {
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		default:
			return stream()
		}
	}
}

// ============================================================================
// STREAM UTILITIES
// ============================================================================

// Tee splits a stream into multiple identical streams for parallel consumption.
// Works with both finite and infinite streams using a broadcasting dispatcher with proper cleanup.
func Tee[T any](stream Stream[T], n int) []Stream[T] {
	if n <= 0 {
		return nil
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	channels := make([]chan T, n)
	abandoned := make([]bool, n) // Track abandoned streams
	var mu sync.RWMutex
	
	for i := 0; i < n; i++ {
		channels[i] = make(chan T, 100)
	}
	
	// Start broadcaster goroutine with cancellation
	go func() {
		defer func() {
			for _, ch := range channels {
				close(ch)
			}
		}()
		
		for {
			select {
			case <-ctx.Done():
				return
			default:
				item, err := stream()
				if err != nil {
					return // Source stream ended
				}
				
				// Broadcast to non-abandoned channels only
				mu.RLock()
				activeCount := 0
				for i, ch := range channels {
					if abandoned[i] {
						continue
					}
					activeCount++
					
					select {
					case ch <- item:
						// Successfully sent
					case <-time.After(5 * time.Second):
						// Consumer too slow - mark as abandoned to prevent leak
						mu.RUnlock()
						mu.Lock()
						abandoned[i] = true
						close(ch) // Close abandoned channel
						mu.Unlock()
						mu.RLock()
					case <-ctx.Done():
						mu.RUnlock()
						return
					}
				}
				mu.RUnlock()
				
				// If all streams abandoned, terminate
				if activeCount == 0 {
					cancel()
					return
				}
			}
		}
	}()
	
	// Create output streams with cleanup tracking
	streams := make([]Stream[T], n)
	for i := 0; i < n; i++ {
		ch := channels[i]
		streams[i] = func() (T, error) {
			select {
			case <-ctx.Done():
				var zero T
				return zero, ctx.Err()
			case item, ok := <-ch:
				if !ok {
					// Mark this stream as done
					cancel()
					var zero T
					return zero, EOS
				}
				return item, nil
			}
		}
	}
	
	return streams
}

// FlatMap transforms elements and flattens the resulting streams
func FlatMap[T, U any](fn func(T) Stream[U]) Filter[T, U] {
	return func(input Stream[T]) Stream[U] {
		var currentStream Stream[U]
		
		return func() (U, error) {
			for {
				// If we have a current stream, try to get next item from it
				if currentStream != nil {
					item, err := currentStream()
					if err == nil {
						return item, nil
					}
					if err != EOS {
						var zero U
						return zero, err
					}
					// Current stream is exhausted, set to nil to get next
					currentStream = nil
				}
				
				// Get next item from input to create new stream
				inputItem, err := input()
				if err != nil {
					var zero U
					return zero, err
				}
				
				// Create new current stream
				currentStream = fn(inputItem)
			}
		}
	}
}

// Split splits a stream of records into substreams based on key fields.
// Each substream contains all records that share the same key values.
// Works with both finite and infinite streams using a central dispatcher.
func Split(keyFields []string) Filter[Record, Stream[Record]] {
	return func(input Stream[Record]) Stream[Stream[Record]] {
		ctx, cancel := context.WithCancel(context.Background())
		newSubstreams := make(chan Stream[Record], 10)
		groupChannels := make(map[string]chan Record)
		abandonedGroups := make(map[string]bool)
		var mu sync.RWMutex
		
		// Start dispatcher goroutine with cancellation
		go func() {
			defer func() {
				for _, ch := range groupChannels {
					close(ch)
				}
				close(newSubstreams)
			}()
			
			for {
				select {
				case <-ctx.Done():
					return
				default:
					record, err := input()
					if err != nil {
						return // Input stream ended
					}
					
					key := buildGroupKey(record, keyFields)
					
					// Check if group was abandoned
					mu.RLock()
					isAbandoned := abandonedGroups[key]
					mu.RUnlock()
					
					if isAbandoned {
						continue // Skip abandoned groups
					}
					
					// Create new substream if needed
					if _, exists := groupChannels[key]; !exists {
						groupChan := make(chan Record, 100)
						groupChannels[key] = groupChan
						
						// Create substream with cancellation
						substream := func() (Record, error) {
							select {
							case <-ctx.Done():
								return nil, ctx.Err()
							case record, ok := <-groupChan:
								if !ok {
									return nil, EOS
								}
								return record, nil
							}
						}
						
						// Emit new substream with timeout
						select {
						case newSubstreams <- substream:
						case <-time.After(1 * time.Second):
							// Consumer too slow, abandon this group
							mu.Lock()
							abandonedGroups[key] = true
							close(groupChan)
							delete(groupChannels, key)
							mu.Unlock()
							continue
						case <-ctx.Done():
							return
						}
					}
					
					// Send record to group channel with timeout
					select {
					case groupChannels[key] <- record:
					case <-time.After(5 * time.Second):
						// Group consumer too slow - abandon it
						mu.Lock()
						abandonedGroups[key] = true
						close(groupChannels[key])
						delete(groupChannels, key)
						mu.Unlock()
					case <-ctx.Done():
						return
					}
				}
			}
		}()
		
		// Return stream with cancellation support
		return func() (Stream[Record], error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case substream, ok := <-newSubstreams:
				if !ok {
					cancel() // Cleanup when done
					return nil, EOS
				}
				return substream, nil
			}
		}
	}
}

// ============================================================================
// WINDOWING FUNCTIONS FOR INFINITE STREAMS
// ============================================================================

// CountWindow groups elements into batches of N elements.
// Each batch is emitted as a finite stream, enabling aggregations on infinite streams.
// Perfect for processing infinite streams in manageable chunks.
func CountWindow[T any](windowSize int) Filter[T, Stream[T]] {
	if windowSize <= 0 {
		panic("CountWindow size must be positive")
	}
	
	return func(input Stream[T]) Stream[Stream[T]] {
		return func() (Stream[T], error) {
			// Collect windowSize elements into a batch
			batch := make([]T, 0, windowSize)
			
			for len(batch) < windowSize {
				item, err := input()
				if err != nil {
					// If we hit EOS or error before filling window
					if len(batch) == 0 {
						// No elements collected, propagate error
						return nil, err
					}
					// Partial batch - emit what we have
					break
				}
				batch = append(batch, item)
			}
			
			// Convert batch to stream
			return FromSlice(batch), nil
		}
	}
}

// TimeWindow groups elements into time-based windows.
// Collects elements for the specified duration, then emits as a finite stream.
func TimeWindow[T any](duration time.Duration) Filter[T, Stream[T]] {
	if duration <= 0 {
		panic("TimeWindow duration must be positive")
	}
	
	return func(input Stream[T]) Stream[Stream[T]] {
		return func() (Stream[T], error) {
			batch := make([]T, 0)
			deadline := time.Now().Add(duration)
			
			for time.Now().Before(deadline) {
				// Try to get next item with a small timeout
				done := make(chan bool, 1)
				var item T
				var err error
				
				go func() {
					item, err = input()
					done <- true
				}()
				
				select {
				case <-done:
					if err != nil {
						// Stream ended or error
						if len(batch) == 0 {
							return nil, err
						}
						// Return partial batch
						return FromSlice(batch), nil
					}
					batch = append(batch, item)
					
				case <-time.After(100 * time.Millisecond):
					// Timeout - check if window expired
					if time.Now().After(deadline) {
						break
					}
					// Continue waiting
				}
			}
			
			if len(batch) == 0 {
				// No items collected in time window, try once more
				item, err := input()
				if err != nil {
					return nil, err
				}
				batch = append(batch, item)
			}
			
			return FromSlice(batch), nil
		}
	}
}

// SlidingCountWindow creates overlapping windows of size windowSize with step stepSize.
// Each window slides by stepSize elements, creating overlapping batches.
func SlidingCountWindow[T any](windowSize, stepSize int) Filter[T, Stream[T]] {
	if windowSize <= 0 || stepSize <= 0 {
		panic("SlidingCountWindow size and step must be positive")
	}
	if stepSize > windowSize {
		panic("SlidingCountWindow step cannot be larger than window size")
	}
	
	return func(input Stream[T]) Stream[Stream[T]] {
		buffer := make([]T, 0, windowSize)
		
		return func() (Stream[T], error) {
			// Fill initial buffer if needed
			for len(buffer) < windowSize {
				item, err := input()
				if err != nil {
					if len(buffer) == 0 {
						return nil, err
					}
					// Partial window
					break
				}
				buffer = append(buffer, item)
			}
			
			// Create current window
			window := make([]T, len(buffer))
			copy(window, buffer)
			
			// Slide the buffer by stepSize
			if stepSize >= len(buffer) {
				// Step is larger than buffer, clear it
				buffer = buffer[:0]
			} else {
				// Slide buffer
				copy(buffer, buffer[stepSize:])
				buffer = buffer[:len(buffer)-stepSize]
			}
			
			return FromSlice(window), nil
		}
	}
}

// ============================================================================
// STREAMING AGGREGATORS - REAL-TIME RUNNING TOTALS
// ============================================================================

// StreamingSum emits running sum continuously as each element arrives.
// Perfect for real-time dashboards and monitoring.
func StreamingSum[T Numeric]() Filter[T, T] {
	return func(input Stream[T]) Stream[T] {
		var runningSum T
		
		return func() (T, error) {
			value, err := input()
			if err != nil {
				return runningSum, err
			}
			
			runningSum += value
			return runningSum, nil
		}
	}
}

// StreamingCount emits running count as each element arrives.
func StreamingCount[T any]() Filter[T, int64] {
	return func(input Stream[T]) Stream[int64] {
		var count int64 = 0
		
		return func() (int64, error) {
			_, err := input()
			if err != nil {
				return count, err
			}
			
			count++
			return count, nil
		}
	}
}

// StreamingAvg emits running average as each element arrives.
func StreamingAvg[T Numeric]() Filter[T, float64] {
	return func(input Stream[T]) Stream[float64] {
		var sum T
		var count int64 = 0
		
		return func() (float64, error) {
			value, err := input()
			if err != nil {
				if count == 0 {
					return 0, err
				}
				return float64(sum) / float64(count), err
			}
			
			sum += value
			count++
			return float64(sum) / float64(count), nil
		}
	}
}

// StreamingMax emits running maximum as each element arrives.
func StreamingMax[T Comparable]() Filter[T, T] {
	return func(input Stream[T]) Stream[T] {
		var currentMax T
		var hasValue bool = false
		
		return func() (T, error) {
			value, err := input()
			if err != nil {
				return currentMax, err
			}
			
			if !hasValue || value > currentMax {
				currentMax = value
				hasValue = true
			}
			
			return currentMax, nil
		}
	}
}

// StreamingMin emits running minimum as each element arrives.
func StreamingMin[T Comparable]() Filter[T, T] {
	return func(input Stream[T]) Stream[T] {
		var currentMin T
		var hasValue bool = false
		
		return func() (T, error) {
			value, err := input()
			if err != nil {
				return currentMin, err
			}
			
			if !hasValue || value < currentMin {
				currentMin = value
				hasValue = true
			}
			
			return currentMin, nil
		}
	}
}

// StreamingStats emits comprehensive running statistics for each element.
// Returns Record with count, sum, avg, min, max.
func StreamingStats[T Numeric]() Filter[T, Record] {
	return func(input Stream[T]) Stream[Record] {
		var sum T
		var count int64 = 0
		var currentMin, currentMax T
		var hasValue bool = false
		
		return func() (Record, error) {
			value, err := input()
			if err != nil {
				if count == 0 {
					return nil, err
				}
				// Return final stats
				return R(
					"count", count,
					"sum", sum,
					"avg", float64(sum)/float64(count),
					"min", currentMin,
					"max", currentMax,
				), err
			}
			
			// Update statistics
			sum += value
			count++
			
			if !hasValue {
				currentMin = value
				currentMax = value
				hasValue = true
			} else {
				if value < currentMin {
					currentMin = value
				}
				if value > currentMax {
					currentMax = value
				}
			}
			
			// Return current stats
			return R(
				"count", count,
				"sum", sum,
				"avg", float64(sum)/float64(count),
				"min", currentMin,
				"max", currentMax,
			), nil
		}
	}
}

// ============================================================================
// TRIGGER-BASED AGGREGATION SYSTEM
// ============================================================================

// Trigger interface defines when to emit aggregation results
type Trigger[T any] interface {
	ShouldFire(element T, state any) bool
	ResetState() any
}

// CountTrigger fires every N elements
type CountTrigger[T any] struct {
	N     int
	count int
}

func NewCountTrigger[T any](n int) *CountTrigger[T] {
	return &CountTrigger[T]{N: n, count: 0}
}

func (ct *CountTrigger[T]) ShouldFire(element T, state any) bool {
	ct.count++
	return ct.count >= ct.N
}

func (ct *CountTrigger[T]) ResetState() any {
	ct.count = 0
	return nil
}

// ValueChangeTrigger fires when a specific field value changes
type ValueChangeTrigger[T any] struct {
	ExtractFunc func(T) any
	lastValue   any
	initialized bool
}

func NewValueChangeTrigger[T any](extractFunc func(T) any) *ValueChangeTrigger[T] {
	return &ValueChangeTrigger[T]{
		ExtractFunc: extractFunc,
		initialized: false,
	}
}

func (vct *ValueChangeTrigger[T]) ShouldFire(element T, state any) bool {
	currentValue := vct.ExtractFunc(element)
	
	if !vct.initialized {
		vct.lastValue = currentValue
		vct.initialized = true
		return false // Don't fire on first element
	}
	
	if currentValue != vct.lastValue {
		vct.lastValue = currentValue
		return true
	}
	
	return false
}

func (vct *ValueChangeTrigger[T]) ResetState() any {
	// Keep the last value for next comparison
	return nil
}

// TriggeredWindow creates windows based on trigger conditions
func TriggeredWindow[T any](trigger Trigger[T]) Filter[T, Stream[T]] {
	return func(input Stream[T]) Stream[Stream[T]] {
		return func() (Stream[T], error) {
			batch := make([]T, 0)
			triggerState := trigger.ResetState()
			
			for {
				item, err := input()
				if err != nil {
					// Stream ended
					if len(batch) == 0 {
						return nil, err
					}
					// Return partial batch
					return FromSlice(batch), nil
				}
				
				batch = append(batch, item)
				
				if trigger.ShouldFire(item, triggerState) {
					// Fire trigger - emit current batch
					return FromSlice(batch), nil
				}
			}
		}
	}
}

// ============================================================================
// STREAMING GROUPBY - CONTINUOUS GROUP UPDATES
// ============================================================================

// StreamingGroupBy maintains running group statistics and emits updates.
// Unlike regular GroupBy, this works with infinite streams by emitting
// updated group totals as new records arrive.
func StreamingGroupBy(keyFields []string, updateInterval int) Filter[Record, Record] {
	return func(input Stream[Record]) Stream[Record] {
		groupStats := make(map[string]*groupAccumulator)
		processedCount := 0
		
		return func() (Record, error) {
			// Process updateInterval records before emitting
			for i := 0; i < updateInterval; i++ {
				record, err := input()
				if err != nil {
					// Stream ended or error
					if len(groupStats) == 0 {
						return nil, err
					}
					// Emit final group summary
					return emitGroupSummary(groupStats, processedCount), err
				}
				
				key := buildGroupKey(record, keyFields)
				
				// Update or create group stats
				if stats, exists := groupStats[key]; exists {
					stats.update(record)
				} else {
					stats := newGroupAccumulator(record, keyFields)
					stats.update(record)
					groupStats[key] = stats
				}
				
				processedCount++
			}
			
			// Emit current group summary
			return emitGroupSummary(groupStats, processedCount), nil
		}
	}
}

// groupAccumulator maintains running statistics for a group
type groupAccumulator struct {
	keyValues map[string]any
	count     int64
	numericSums map[string]float64
}

func newGroupAccumulator(firstRecord Record, keyFields []string) *groupAccumulator {
	acc := &groupAccumulator{
		keyValues:   make(map[string]any),
		count:       0,
		numericSums: make(map[string]float64),
	}
	
	// Store key field values
	for _, field := range keyFields {
		if val, exists := firstRecord[field]; exists {
			acc.keyValues[field] = val
		}
	}
	
	return acc
}

func (acc *groupAccumulator) update(record Record) {
	acc.count++
	
	// Update numeric field sums
	for field, value := range record {
		if numValue, ok := convertToFloat64(value); ok {
			acc.numericSums[field] += numValue
		}
	}
}

func emitGroupSummary(groupStats map[string]*groupAccumulator, totalProcessed int) Record {
	summary := R(
		"total_processed", totalProcessed,
		"active_groups", len(groupStats),
		"timestamp", time.Now().Unix(),
	)
	
	// Add details about largest group
	var largestGroup *groupAccumulator
	var largestKey string
	for key, stats := range groupStats {
		if largestGroup == nil || stats.count > largestGroup.count {
			largestGroup = stats
			largestKey = key
		}
	}
	
	if largestGroup != nil {
		summary.Set("largest_group_key", largestKey)
		summary.Set("largest_group_count", largestGroup.count)
	}
	
	return summary
}

// Note: convertToFloat64 function is defined in stream.go