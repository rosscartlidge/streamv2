package stream

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// STREAM TRANSFORMATION FILTERS
// ============================================================================

// Filter transforms one stream into another with full type flexibility
type Filter[T, U any] func(Stream[T]) Stream[U]

// ============================================================================
// FUNCTIONAL OPERATIONS - TYPE SAFE AND COMPOSABLE
// ============================================================================

// Map transforms each element in a stream with automatic parallelization for large datasets
func Map[T, U any](fn func(T) U) Filter[T, U] {
	return func(input Stream[T]) Stream[U] {
		// Try to estimate dataset size by sampling
		complexity := estimateFunctionComplexity(fn)
		
		// For auto-parallel decision, we need to peek at the stream
		// If we can't estimate size, use heuristics based on function complexity
		if complexity >= 5 { // Complex functions benefit more from parallelization
			return autoParallelMap(fn, input, complexity)
		}
		
		// For simple operations, check if we should parallelize based on data volume
		return adaptiveMap(fn, input, complexity)
	}
}

// autoParallelMap automatically uses parallel processing for complex operations
func autoParallelMap[T, U any](fn func(T) U, input Stream[T], complexity int) Stream[U] {
	// Use parallel processing with worker count based on complexity and available CPUs
	workers := calculateOptimalWorkers(complexity)
	return Parallel(workers, fn)(input)
}

// adaptiveMap decides between sequential and parallel based on data characteristics
func adaptiveMap[T, U any](fn func(T) U, input Stream[T], complexity int) Stream[U] {
	// For now, use a simple heuristic: parallel for moderate+ complexity
	// In future, this could sample the stream to estimate size
	if complexity >= 3 {
		workers := calculateOptimalWorkers(complexity)
		return Parallel(workers, fn)(input)
	}
	
	// Sequential implementation for simple operations
	return func() (U, error) {
		item, err := input()
		if err != nil {
			var zero U
			return zero, err
		}
		return fn(item), nil
	}
}

// calculateOptimalWorkers determines optimal worker count based on operation characteristics
func calculateOptimalWorkers(complexity int) int {
	// Base workers on complexity and system capabilities
	baseWorkers := 4 // Conservative default
	
	// More workers for more complex operations
	if complexity >= 10 {
		baseWorkers = 16
	} else if complexity >= 7 {
		baseWorkers = 8
	} else if complexity >= 5 {
		baseWorkers = 6
	}
	
	// Don't exceed reasonable bounds
	if baseWorkers > 16 {
		baseWorkers = 16
	}
	
	return baseWorkers
}

// Where keeps only elements matching a predicate with automatic parallelization
func Where[T any](predicate func(T) bool) Filter[T, T] {
	return func(input Stream[T]) Stream[T] {
		complexity := estimateFunctionComplexity(predicate)
		
		// Auto-parallelize complex predicates
		if complexity >= 4 {
			return autoParallelFilter(predicate, input, complexity)
		}
		
		// Sequential implementation for simple predicates
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

// autoParallelFilter implements parallel filtering for complex predicates
func autoParallelFilter[T any](predicate func(T) bool, input Stream[T], complexity int) Stream[T] {
	// Create a filter function that returns the item or nil
	filterFn := func(item T) *T {
		if predicate(item) {
			return &item
		}
		return nil
	}
	
	workers := calculateOptimalWorkers(complexity)
	
	// Use parallel processing, then filter out nils
	parallelStream := Parallel(workers, filterFn)(input)
	
	return func() (T, error) {
		for {
			result, err := parallelStream()
			if err != nil {
				var zero T
				return zero, err
			}
			if result != nil {
				return *result, nil
			}
			// Continue if result was nil (filtered out)
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
				result[field] = val
			}
		}
		return result
	})
}

// Update modifies records
func Update(fn func(Record) Record) Filter[Record, Record] {
	return Map(func(r Record) Record {
		return fn(r)
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

// Parallel processes elements concurrently using simple goroutines
func Parallel[T, U any](workers int, fn func(T) U) Filter[T, U] {
	return func(input Stream[T]) Stream[U] {
		inputCh := make(chan T, workers)
		outputCh := make(chan U, workers)
		workerDone := make(chan struct{}, workers)

		// Start workers
		for i := 0; i < workers; i++ {
			go func() {
				for item := range inputCh {
					result := fn(item)
					outputCh <- result
				}
				workerDone <- struct{}{}
			}()
		}

		// Feed input and manage cleanup
		go func() {
			defer close(inputCh)
			for {
				item, err := input()
				if err != nil {
					break // Input stream ended
				}
				inputCh <- item
			}
		}()

		// Cleanup coordinator
		go func() {
			// Wait for all workers to signal completion
			for i := 0; i < workers; i++ {
				<-workerDone
			}
			close(outputCh)
		}()

		// Return simple stream
		return func() (U, error) {
			item, ok := <-outputCh
			if !ok {
				var zero U
				return zero, EOS
			}
			return item, nil
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
					case <-ctx.Done():
						mu.RUnlock()
						return
					default:
						// Channel full - consumer too slow, mark as abandoned to prevent leak
						mu.RUnlock()
						mu.Lock()
						abandoned[i] = true
						close(ch) // Close abandoned channel
						mu.Unlock()
						mu.RLock()
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

// DotFlatten flattens nested records using dot notation for field names.
// Nested records become prefixed fields: {"user": {"name": "Alice"}} → {"user.name": "Alice"}
// Stream fields are not flattened (use CrossFlatten for those).
func DotFlatten(separator string, fields ...string) Filter[Record, Record] {
	if separator == "" {
		separator = "."
	}
	
	return func(input Stream[Record]) Stream[Record] {
		return func() (Record, error) {
			record, err := input()
			if err != nil {
				return nil, err
			}
			
			return dotFlattenRecord(record, "", separator, fields...), nil
		}
	}
}

// dotFlattenRecord recursively flattens a record using dot notation
// If fields are specified, only flattens those top-level fields
func dotFlattenRecord(record Record, prefix, separator string, fields ...string) Record {
	result := make(Record)
	
	// Create a set of fields to flatten for quick lookup
	fieldsToFlatten := make(map[string]bool)
	if len(fields) > 0 {
		for _, field := range fields {
			fieldsToFlatten[field] = true
		}
	}
	
	for key, value := range record {
		newKey := key
		if prefix != "" {
			newKey = prefix + separator + key
		}
		
		// Check if this field should be flattened (only applies to top-level fields)
		shouldFlatten := len(fields) == 0 || prefix != "" || fieldsToFlatten[key]
		
		// If the value is a nested record, flatten it recursively
		if nestedRecord, ok := value.(Record); ok && shouldFlatten {
			flattened := dotFlattenRecord(nestedRecord, newKey, separator)
			for flatKey, flatValue := range flattened {
				result[flatKey] = flatValue
			}
		} else {
			// For non-record values (including streams), or fields not to be flattened, keep as-is
			result[newKey] = value
		}
	}
	
	return result
}

// CrossFlatten creates cartesian product of stream fields within records.
// Uses the original flatten algorithm that can handle infinite streams.
// Example: {"id": 1, "tags": Stream["a", "b"]} → [{"id": 1, "tag": "a"}, {"id": 1, "tag": "b"}]
func CrossFlatten(separator string, fields ...string) Filter[Record, Record] {
	if separator == "" {
		separator = "."
	}
	
	return func(input Stream[Record]) Stream[Record] {
		var expandedStream Stream[Record]
		
		return func() (Record, error) {
			for {
				// If we have an expanded stream, try to get next item from it
				if expandedStream != nil {
					record, err := expandedStream()
					if err == nil {
						return record, nil
					}
					// Expanded stream is exhausted, clear it
					expandedStream = nil
				}
				
				// Get next input record
				record, err := input()
				if err != nil {
					return nil, err
				}
				
				// Use field-specific flatten algorithm 
				flattened := crossFlattenRecord(record, separator, fields...)
				if len(flattened) > 1 {
					// Multiple records produced, create stream from them
					expandedStream = FromSlice(flattened)
				} else if len(flattened) == 1 {
					// Single record produced, return it directly
					return flattened[0], nil
				} else {
					// No records produced (shouldn't happen), return original
					return record, nil
				}
			}
		}
	}
}

// cross performs cartesian product of record slices
func cross(columns [][]Record) []Record {
	if len(columns) == 0 {
		return nil
	}
	if len(columns) == 1 {
		return columns[0]
	}
	var rs []Record
	for _, lr := range cross(columns[1:]) {
		for _, rr := range columns[0] {
			r := make(Record)
			for f := range rr {
				r[f] = rr[f]
			}
			for f := range lr {
				r[f] = lr[f]
			}
			rs = append(rs, r)
		}
	}
	return rs
}

// flattenRecord implements the original flatten algorithm that handles both
// dot notation for nested records and cross products for streams
func flattenRecord(r Record, sep string) []Record {
	var columns [][]Record
	var fs []string
	
	for f := range r {
		if s, ok := r[f].(Stream[interface{}]); ok {
			var rs []Record
			for {
				if record, err := s(); err == nil {
					if f != "" {
						er := make(Record)
						if recordMap, ok := record.(Record); ok {
							// Stream contains records - add field prefix
							for fi := range recordMap {
								ef := []string{f}
								if fi != "" {
									ef = append(ef, fi)
								}
								er[strings.Join(ef, sep)] = recordMap[fi]
							}
						} else {
							// Stream contains scalar values - use field name directly
							er[f] = record
						}
						rs = append(rs, flattenRecord(er, sep)...)
					} else {
						if recordMap, ok := record.(Record); ok {
							rs = append(rs, flattenRecord(recordMap, sep)...)
						} else {
							// Scalar in unnamed field
							rs = append(rs, Record{"": record})
						}
					}
				} else {
					break
				}
			}
			columns = append(columns, rs)
		} else {
			fs = append(fs, f)
		}
	}
	
	if len(columns) == 0 {
		return []Record{r}
	}
	
	crs := cross(columns)
	if len(fs) == 0 {
		return crs
	}
	
	for _, cr := range crs {
		for _, f := range fs {
			cr[f] = r[f]
		}
	}
	
	return crs
}

// crossFlattenRecord expands specified stream fields using cartesian product
// If no fields specified, expands all stream fields
func crossFlattenRecord(r Record, sep string, fields ...string) []Record {
	var columns [][]Record
	var nonStreamFields []string
	
	// Create a set of fields to expand for quick lookup
	fieldsToExpand := make(map[string]bool)
	if len(fields) > 0 {
		for _, field := range fields {
			fieldsToExpand[field] = true
		}
	}
	
	for f := range r {
		if s, ok := r[f].(Stream[interface{}]); ok {
			// Check if this field should be expanded
			shouldExpand := len(fields) == 0 || fieldsToExpand[f]
			
			if shouldExpand {
				var rs []Record
				for {
					if record, err := s(); err == nil {
						// Create a record with this stream value
						newRecord := Record{f: record}
						rs = append(rs, newRecord)
					} else {
						break
					}
				}
				if len(rs) > 0 {
					columns = append(columns, rs)
				}
			} else {
				// Keep stream field as-is (don't expand)
				nonStreamFields = append(nonStreamFields, f)
			}
		} else {
			// Non-stream field
			nonStreamFields = append(nonStreamFields, f)
		}
	}
	
	// If no stream fields to expand, return original record
	if len(columns) == 0 {
		return []Record{r}
	}
	
	// Create cartesian product of expanded fields
	crs := cross(columns)
	
	// Add non-stream fields to each result record
	for _, cr := range crs {
		for _, f := range nonStreamFields {
			cr[f] = r[f]
		}
	}
	
	return crs
}

// JoinOption configures join behavior
type JoinOption func(*joinConfig)

// joinConfig holds join configuration
type joinConfig struct {
	leftPrefix  string
	rightPrefix string
}

// WithPrefixes sets custom prefixes for field name conflicts
// Default is "left." and "right."
func WithPrefixes(leftPrefix, rightPrefix string) JoinOption {
	return func(config *joinConfig) {
		config.leftPrefix = leftPrefix
		config.rightPrefix = rightPrefix
	}
}

// InnerJoin performs an inner join between left stream and right stream.
// Only records with matching keys in both streams are returned.
// WARNING: Right stream is collected into memory - must be finite and reasonably sized.
func InnerJoin(rightStream Stream[Record], leftKey, rightKey string, options ...JoinOption) Filter[Record, Record] {
	return createJoin(rightStream, leftKey, rightKey, innerJoinType, options...)
}

// LeftJoin performs a left join between left stream and right stream.
// All records from left stream are returned, with matching right records when available.
// WARNING: Right stream is collected into memory - must be finite and reasonably sized.
func LeftJoin(rightStream Stream[Record], leftKey, rightKey string, options ...JoinOption) Filter[Record, Record] {
	return createJoin(rightStream, leftKey, rightKey, leftJoinType, options...)
}

// RightJoin performs a right join between left stream and right stream.
// All records from right stream are returned, with matching left records when available.
// WARNING: Right stream is collected into memory - must be finite and reasonably sized.
func RightJoin(rightStream Stream[Record], leftKey, rightKey string, options ...JoinOption) Filter[Record, Record] {
	return createJoin(rightStream, leftKey, rightKey, rightJoinType, options...)
}

// FullJoin performs a full outer join between left stream and right stream.
// All records from both streams are returned, with matching when available.
// WARNING: Right stream is collected into memory - must be finite and reasonably sized.
func FullJoin(rightStream Stream[Record], leftKey, rightKey string, options ...JoinOption) Filter[Record, Record] {
	return createJoin(rightStream, leftKey, rightKey, fullJoinType, options...)
}

type joinType int

const (
	innerJoinType joinType = iota
	leftJoinType
	rightJoinType
	fullJoinType
)

// createJoin implements the hash join algorithm for all join types
func createJoin(rightStream Stream[Record], leftKey, rightKey string, jType joinType, options ...JoinOption) Filter[Record, Record] {
	// Apply configuration options
	config := &joinConfig{
		leftPrefix:  "left.",
		rightPrefix: "right.",
	}
	for _, option := range options {
		option(config)
	}

	return func(leftStream Stream[Record]) Stream[Record] {
		// Build hash table from right stream (WARNING: collects entire right stream into memory)
		rightMap := make(map[string][]Record)
		rightKeysUsed := make(map[string]bool) // Track which right keys were matched (for full join)
		
		// Collect right stream into hash map
		for {
			rightRecord, err := rightStream()
			if err != nil {
				break // End of right stream
			}
			
			// Get the join key value from right record
			rightKeyValue := getJoinKeyValue(rightRecord, rightKey)
			if rightKeyValue != "" {
				rightMap[rightKeyValue] = append(rightMap[rightKeyValue], rightRecord)
			}
		}

		// For right and full joins, we need to track unmatched right records
		var unmatchedRightRecords []Record
		if jType == rightJoinType || jType == fullJoinType {
			// Prepare list of all right records for later processing
			for key, records := range rightMap {
				for _, record := range records {
					unmatchedRightRecords = append(unmatchedRightRecords, record)
				}
				rightKeysUsed[key] = false
			}
		}

		var pendingResults []Record
		pendingIndex := 0
		leftFinished := false

		return func() (Record, error) {
			// Return pending results first
			if pendingIndex < len(pendingResults) {
				result := pendingResults[pendingIndex]
				pendingIndex++
				return result, nil
			}

			// Reset pending results
			pendingResults = nil
			pendingIndex = 0

			// Process left stream
			if !leftFinished {
				leftRecord, err := leftStream()
				if err != nil {
					leftFinished = true
					// Handle right/full join unmatched right records
					if jType == rightJoinType || jType == fullJoinType {
						for key, used := range rightKeysUsed {
							if !used {
								for _, rightRecord := range rightMap[key] {
									merged := mergeRecords(nil, rightRecord, config.leftPrefix, config.rightPrefix)
									pendingResults = append(pendingResults, merged)
								}
							}
						}
						if len(pendingResults) > 0 {
							result := pendingResults[0]
							pendingIndex = 1
							return result, nil
						}
					}
					return nil, EOS
				}

				// Get the join key value from left record
				leftKeyValue := getJoinKeyValue(leftRecord, leftKey)
				
				// Look up matching right records
				if matchingRightRecords, exists := rightMap[leftKeyValue]; exists && leftKeyValue != "" {
					// Mark this right key as used
					rightKeysUsed[leftKeyValue] = true
					
					// Create joined records for each match
					for _, rightRecord := range matchingRightRecords {
						merged := mergeRecords(leftRecord, rightRecord, config.leftPrefix, config.rightPrefix)
						pendingResults = append(pendingResults, merged)
					}
				} else {
					// No match found
					if jType == leftJoinType || jType == fullJoinType {
						// Left/Full join: include left record with nil right
						merged := mergeRecords(leftRecord, nil, config.leftPrefix, config.rightPrefix)
						pendingResults = append(pendingResults, merged)
					}
					// Inner/Right join: skip this left record
				}

				// Return first result if any
				if len(pendingResults) > 0 {
					result := pendingResults[0]
					pendingIndex = 1
					return result, nil
				}

				// No results, continue to next left record
				return createJoin(rightStream, leftKey, rightKey, jType, options...)(leftStream)()
			}

			return nil, EOS
		}
	}
}

// getJoinKeyValue extracts the join key value from a record
func getJoinKeyValue(record Record, keyField string) string {
	if value, exists := record[keyField]; exists {
		return fmt.Sprintf("%v", value)
	}
	return ""
}

// mergeRecords combines left and right records, handling field name conflicts
func mergeRecords(leftRecord, rightRecord Record, leftPrefix, rightPrefix string) Record {
	result := make(Record)

	// Add left record fields
	if leftRecord != nil {
		for key, value := range leftRecord {
			result[key] = value
		}
	}

	// Add right record fields, handling conflicts
	if rightRecord != nil {
		for key, value := range rightRecord {
			if _, exists := result[key]; exists && leftPrefix != "" && rightPrefix != "" {
				// Conflict: add prefixed versions
				if leftRecord != nil {
					result[leftPrefix+key] = result[key]
					delete(result, key)
				}
				result[rightPrefix+key] = value
			} else {
				// No conflict or no prefixes
				result[key] = value
			}
		}
	}

	return result
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
						
						// Emit new substream 
						select {
						case newSubstreams <- substream:
						case <-ctx.Done():
							return
						default:
							// Consumer too slow, abandon this group
							mu.Lock()
							abandonedGroups[key] = true
							close(groupChan)
							delete(groupChannels, key)
							mu.Unlock()
							continue
						}
					}
					
					// Send record to group channel with non-blocking fallback
					select {
					case groupChannels[key] <- record:
					case <-ctx.Done():
						return
					default:
						// Group consumer too slow - abandon it
						mu.Lock()
						abandonedGroups[key] = true
						close(groupChannels[key])
						delete(groupChannels, key)
						mu.Unlock()
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
			return FromSliceAny(batch), nil
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
			timer := time.NewTimer(duration)
			defer timer.Stop()
			
			for {
				// Try to get next item with deadline-based timeout
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
						return FromSliceAny(batch), nil
					}
					batch = append(batch, item)
					
					// Check if deadline passed after processing item
					if time.Now().After(deadline) {
						break
					}
					
				case <-timer.C:
					// Window expired - return what we have
					break
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
			
			return FromSliceAny(batch), nil
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
					// Stream ended - only return window if we have full size
					if len(buffer) < windowSize {
						return nil, err
					}
					break
				}
				buffer = append(buffer, item)
			}
			
			// Only create window if we have full size
			if len(buffer) < windowSize {
				return nil, EOS
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
			
			return FromSliceAny(window), nil
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
				return NewRecord().
					Int("count", count).
					Set("sum", sum).
					Set("avg", float64(sum)/float64(count)).
					Set("min", currentMin).
					Set("max", currentMax).
					Build(), err
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
			return NewRecord().
				Int("count", count).
				Set("sum", sum).
				Set("avg", float64(sum)/float64(count)).
				Set("min", currentMin).
				Set("max", currentMax).
				Build(), nil
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
		return true // Fire on first element (no previous value to compare)
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
					return FromSliceAny(batch), nil
				}
				
				batch = append(batch, item)
				
				if trigger.ShouldFire(item, triggerState) {
					// Fire trigger - emit current batch
					return FromSliceAny(batch), nil
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
	summary := NewRecord().
		Int("total_processed", int64(totalProcessed)).
		Int("active_groups", int64(len(groupStats))).
		Int("timestamp", time.Now().Unix()).
		Build()
	
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
		summary["largest_group_key"] = largestKey
		summary["largest_group_count"] = largestGroup.count
	}
	
	return summary
}

// Note: convertToFloat64 function is defined in stream.go