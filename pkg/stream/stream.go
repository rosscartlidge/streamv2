package stream

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"
)

// ============================================================================
// STREAMV2 - GENERICS-FIRST STREAM PROCESSING LIBRARY
// ============================================================================

// EOS signals end of stream
var EOS = errors.New("end of stream")

// Stream represents a generic data stream - the heart of V2
type Stream[T any] func() (T, error)

// Filter transforms one stream into another with full type flexibility
type Filter[T, U any] func(Stream[T]) Stream[U]

// Common stream type aliases for convenience
type RecordStream = Stream[Record]
type IntStream = Stream[int64]
type FloatStream = Stream[float64]
type StringStream = Stream[string]
type BoolStream = Stream[bool]

// Record represents structured data with native Go types
type Record map[string]any

// ============================================================================
// SMART RECORD SYSTEM - NATIVE GO TYPES
// ============================================================================

// R creates records from key-value pairs with automatic type handling
func R(pairs ...any) Record {
	if len(pairs)%2 != 0 {
		panic("R() requires even number of arguments (key-value pairs)")
	}

	r := make(Record)
	for i := 0; i < len(pairs); i += 2 {
		key := pairs[i].(string)
		r[key] = pairs[i+1] // Direct storage - no wrapping!
	}
	return r
}

// RecordFrom creates a Record from map[string]any
func RecordFrom(m map[string]any) Record {
	return Record(m) // Direct conversion - zero overhead!
}

// RecordsFrom creates Records from slice of maps
func RecordsFrom(maps []map[string]any) []Record {
	records := make([]Record, len(maps))
	for i, m := range maps {
		records[i] = Record(m)
	}
	return records
}

// ============================================================================
// TYPE-SAFE RECORD ACCESS WITH AUTOMATIC CONVERSION
// ============================================================================

// Get retrieves a typed value from a record with automatic conversion
func Get[T any](r Record, field string) (T, bool) {
	val, exists := r[field]
	if !exists {
		var zero T
		return zero, false
	}

	// Direct type assertion first (fast path)
	if typed, ok := val.(T); ok {
		return typed, true
	}

	// Smart type conversion (slower path)
	if converted, ok := convertTo[T](val); ok {
		return converted, true
	}

	var zero T
	return zero, false
}

// GetOr retrieves a typed value with a default fallback
func GetOr[T any](r Record, field string, defaultVal T) T {
	if val, ok := Get[T](r, field); ok {
		return val
	}
	return defaultVal
}

// Set assigns a value to a record field
func (r Record) Set(field string, value any) Record {
	r[field] = value
	return r
}

// Has checks if a field exists
func (r Record) Has(field string) bool {
	_, exists := r[field]
	return exists
}

// Keys returns all field names
func (r Record) Keys() []string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	return keys
}

// ============================================================================
// SMART TYPE CONVERSION SYSTEM
// ============================================================================

func convertTo[T any](val any) (T, bool) {
	var zero T
	targetType := reflect.TypeOf(zero)

	// Handle nil
	if val == nil {
		return zero, false
	}

	sourceVal := reflect.ValueOf(val)

	// Try direct conversion for basic types
	if sourceVal.Type().ConvertibleTo(targetType) {
		converted := sourceVal.Convert(targetType)
		return converted.Interface().(T), true
	}

	// Custom conversions for common cases
	switch target := any(zero).(type) {
	case int64:
		if converted, ok := convertToInt64(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	case float64:
		if converted, ok := convertToFloat64(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	case string:
		if converted, ok := convertToString(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	case bool:
		if converted, ok := convertToBool(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	case time.Time:
		if converted, ok := convertToTime(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	default:
		_ = target
		return zero, false
	}
}

func convertToInt64(val any) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int16:
		return int64(v), true
	case int8:
		return int64(v), true
	case uint64:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint8:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	case string:
		// Could add string parsing here
		return 0, false
	default:
		return 0, false
	}
}

func convertToFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int16:
		return float64(v), true
	case int8:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint8:
		return float64(v), true
	default:
		return 0, false
	}
}

func convertToString(val any) (string, bool) {
	switch v := val.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	default:
		return fmt.Sprintf("%v", val), true
	}
}

func convertToBool(val any) (bool, bool) {
	switch v := val.(type) {
	case bool:
		return v, true
	case int64:
		return v != 0, true
	case int:
		return v != 0, true
	case float64:
		return v != 0, true
	case string:
		return v != "", true
	default:
		return false, false
	}
}

func convertToTime(val any) (time.Time, bool) {
	switch v := val.(type) {
	case time.Time:
		return v, true
	case string:
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, true
		}
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t, true
		}
		return time.Time{}, false
	case int64:
		return time.Unix(v, 0), true
	default:
		return time.Time{}, false
	}
}

// ============================================================================
// STREAM CREATION - GENERICS MAKE THIS BEAUTIFUL
// ============================================================================

// FromSlice creates a stream from a slice of any type
func FromSlice[T any](items []T) Stream[T] {
	index := 0
	return func() (T, error) {
		if index >= len(items) {
			var zero T
			return zero, EOS
		}
		item := items[index]
		index++
		return item, nil
	}
}

// FromChannel creates a stream from a channel
func FromChannel[T any](ch <-chan T) Stream[T] {
	return func() (T, error) {
		item, ok := <-ch
		if !ok {
			var zero T
			return zero, EOS
		}
		return item, nil
	}
}

// FromRecords creates a RecordStream from Records
func FromRecords(records []Record) RecordStream {
	return FromSlice(records)
}

// FromMaps creates a RecordStream from maps
func FromMaps(maps []map[string]any) RecordStream {
	return FromSlice(RecordsFrom(maps))
}

// Generate creates a stream using a generator function
func Generate[T any](generator func() (T, error)) Stream[T] {
	return generator
}

// Range creates a numeric stream
func Range(start, end, step int64) IntStream {
	current := start
	return func() (int64, error) {
		if (step > 0 && current >= end) || (step < 0 && current <= end) {
			return 0, EOS
		}
		value := current
		current += step
		return value, nil
	}
}

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

// Filter keeps only elements matching a predicate
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
				result[field] = val
			}
		}
		return result
	})
}

// Update modifies records
func Update(fn func(Record) Record) Filter[Record, Record] {
	return Map(fn)
}

// ExtractField gets a typed field from records
func ExtractField[T any](field string) Filter[Record, T] {
	return Map(func(r Record) T {
		val, _ := Get[T](r, field)
		return val
	})
}

// ============================================================================
// AGGREGATION FUNCTIONS - TYPE SAFE
// ============================================================================

// Sum aggregates numeric values
func Sum[T Numeric](stream Stream[T]) (T, error) {
	var total T
	for {
		val, err := stream()
		if err != nil {
			if errors.Is(err, EOS) {
				return total, nil
			}
			return total, err
		}
		total += val
	}
}

// Count counts elements
func Count[T any](stream Stream[T]) (int64, error) {
	var count int64
	for {
		_, err := stream()
		if err != nil {
			if errors.Is(err, EOS) {
				return count, nil
			}
			return count, err
		}
		count++
	}
}

// Max finds maximum value
func Max[T Comparable](stream Stream[T]) (T, error) {
	var max T
	first := true

	for {
		val, err := stream()
		if err != nil {
			if errors.Is(err, EOS) {
				if first {
					return max, errors.New("empty stream")
				}
				return max, nil
			}
			return max, err
		}

		if first || val > max {
			max = val
			first = false
		}
	}
}

// Min finds minimum value
func Min[T Comparable](stream Stream[T]) (T, error) {
	var min T
	first := true

	for {
		val, err := stream()
		if err != nil {
			if errors.Is(err, EOS) {
				if first {
					return min, errors.New("empty stream")
				}
				return min, nil
			}
			return min, err
		}

		if first || val < min {
			min = val
			first = false
		}
	}
}

// Collect gathers all stream elements into a slice
func Collect[T any](stream Stream[T]) ([]T, error) {
	var result []T
	for {
		item, err := stream()
		if err != nil {
			if errors.Is(err, EOS) {
				return result, nil
			}
			return result, err
		}
		result = append(result, item)
	}
}

// ForEach executes a function for each element
func ForEach[T any](fn func(T)) func(Stream[T]) error {
	return func(stream Stream[T]) error {
		for {
			item, err := stream()
			if err != nil {
				if errors.Is(err, EOS) {
					return nil
				}
				return err
			}
			fn(item)
		}
	}
}

// ============================================================================
// TYPE CONSTRAINTS
// ============================================================================

// Numeric constraint for mathematical operations
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Comparable constraint for ordering operations
type Comparable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

// ============================================================================
// CONCURRENT PROCESSING
// ============================================================================

// Parallel processes elements concurrently
func Parallel[T, U any](workers int, fn func(T) U) Filter[T, U] {
	return func(input Stream[T]) Stream[U] {
		inputCh := make(chan T, workers)
		outputCh := make(chan U, workers)

		// Start workers
		for i := 0; i < workers; i++ {
			go func() {
				for item := range inputCh {
					outputCh <- fn(item)
				}
			}()
		}

		// Feed input
		go func() {
			defer close(inputCh)
			defer func() {
				// Give workers time to finish, then close output
				go func() {
					time.Sleep(100 * time.Millisecond)
					close(outputCh)
				}()
			}()

			for {
				item, err := input()
				if err != nil {
					return
				}
				inputCh <- item
			}
		}()

		return FromChannel(outputCh)
	}
}

// ============================================================================
// GENERALIZED AGGREGATION SUPPORT
// ============================================================================

// Aggregator defines a composable aggregation operation
type Aggregator[T, A, R any] struct {
	Initial    func() A           // Create initial accumulator
	Accumulate func(A, T) A      // Process each element
	Finalize   func(A) R         // Produce final result
}

// AggregateWith runs a single custom aggregator on a stream
func AggregateWith[T, A, R any](stream Stream[T], agg Aggregator[T, A, R]) (R, error) {
	acc := agg.Initial()
	
	for {
		val, err := stream()
		if err != nil {
			if errors.Is(err, EOS) {
				return agg.Finalize(acc), nil
			}
			var zero R
			return zero, err
		}
		
		acc = agg.Accumulate(acc, val)
	}
}

// AggregateMultiple runs multiple aggregators in parallel using Tee
func AggregateMultiple[T any, A1, A2, R1, R2 any](
	stream Stream[T],
	agg1 Aggregator[T, A1, R1],
	agg2 Aggregator[T, A2, R2],
) (R1, R2, error) {
	streams := Tee(stream, 2)
	
	result1, err1 := AggregateWith(streams[0], agg1)
	result2, err2 := AggregateWith(streams[1], agg2)
	
	if err1 != nil {
		var zero1 R1
		var zero2 R2
		return zero1, zero2, err1
	}
	if err2 != nil {
		var zero1 R1
		var zero2 R2
		return zero1, zero2, err2
	}
	
	return result1, result2, nil
}

// AggregateTriple runs three aggregators in parallel
func AggregateTriple[T any, A1, A2, A3, R1, R2, R3 any](
	stream Stream[T],
	agg1 Aggregator[T, A1, R1],
	agg2 Aggregator[T, A2, R2],
	agg3 Aggregator[T, A3, R3],
) (R1, R2, R3, error) {
	streams := Tee(stream, 3)
	
	result1, err1 := AggregateWith(streams[0], agg1)
	result2, err2 := AggregateWith(streams[1], agg2)
	result3, err3 := AggregateWith(streams[2], agg3)
	
	if err1 != nil {
		var zero1 R1
		var zero2 R2
		var zero3 R3
		return zero1, zero2, zero3, err1
	}
	if err2 != nil {
		var zero1 R1
		var zero2 R2
		var zero3 R3
		return zero1, zero2, zero3, err2
	}
	if err3 != nil {
		var zero1 R1
		var zero2 R2
		var zero3 R3
		return zero1, zero2, zero3, err3
	}
	
	return result1, result2, result3, nil
}

// ============================================================================
// BUILT-IN AGGREGATORS
// ============================================================================

// SumAgg creates a sum aggregator
func SumAgg[T Numeric]() Aggregator[T, T, T] {
	return Aggregator[T, T, T]{
		Initial:    func() T { var zero T; return zero },
		Accumulate: func(acc T, val T) T { return acc + val },
		Finalize:   func(acc T) T { return acc },
	}
}

// CountAgg creates a count aggregator  
func CountAgg[T any]() Aggregator[T, int64, int64] {
	return Aggregator[T, int64, int64]{
		Initial:    func() int64 { return 0 },
		Accumulate: func(acc int64, val T) int64 { return acc + 1 },
		Finalize:   func(acc int64) int64 { return acc },
	}
}

// MinAgg creates a min aggregator
func MinAgg[T Comparable]() Aggregator[T, *T, T] {
	return Aggregator[T, *T, T]{
		Initial: func() *T { return nil },
		Accumulate: func(acc *T, val T) *T {
			if acc == nil || val < *acc {
				return &val
			}
			return acc
		},
		Finalize: func(acc *T) T {
			if acc == nil {
				var zero T
				return zero
			}
			return *acc
		},
	}
}

// MaxAgg creates a max aggregator
func MaxAgg[T Comparable]() Aggregator[T, *T, T] {
	return Aggregator[T, *T, T]{
		Initial: func() *T { return nil },
		Accumulate: func(acc *T, val T) *T {
			if acc == nil || val > *acc {
				return &val
			}
			return acc
		},
		Finalize: func(acc *T) T {
			if acc == nil {
				var zero T
				return zero
			}
			return *acc
		},
	}
}

// AvgAgg creates an average aggregator
func AvgAgg[T Numeric]() Aggregator[T, [2]float64, float64] {
	return Aggregator[T, [2]float64, float64]{
		Initial: func() [2]float64 { return [2]float64{0, 0} }, // [sum, count]
		Accumulate: func(acc [2]float64, val T) [2]float64 {
			return [2]float64{acc[0] + float64(val), acc[1] + 1}
		},
		Finalize: func(acc [2]float64) float64 {
			if acc[1] == 0 {
				return 0
			}
			return acc[0] / acc[1]
		},
	}
}

// ============================================================================
// GENERALIZED AGGREGATES FUNCTION
// ============================================================================

// AggregatorSpec represents a named aggregator specification
type AggregatorSpec[T any] struct {
	Name string
	Agg  interface{} // Type-erased aggregator
}

// Aggregates runs multiple named aggregators and returns results in a Record
func Aggregates[T any](stream Stream[T], specs ...AggregatorSpec[T]) (Record, error) {
	if len(specs) == 0 {
		return Record{}, nil
	}

	// Split stream into multiple streams using Tee
	streams := Tee(stream, len(specs))
	
	// Create result record
	result := Record{}
	
	// Run each aggregator on its own stream
	for i, spec := range specs {
		var err error
		var value interface{}
		
		// Type-assert the aggregator and run it
		switch agg := spec.Agg.(type) {
		case Aggregator[T, T, T]:
			value, err = AggregateWith(streams[i], agg)
		case Aggregator[T, int64, int64]:
			value, err = AggregateWith(streams[i], agg)
		case Aggregator[T, *T, T]:
			value, err = AggregateWith(streams[i], agg)
		case Aggregator[T, [2]float64, float64]:
			value, err = AggregateWith(streams[i], agg)
		default:
			// For custom aggregators, try to use reflection or provide a generic interface
			return result, fmt.Errorf("unsupported aggregator type for '%s'", spec.Name)
		}
		
		if err != nil {
			return result, err
		}
		
		result[spec.Name] = value
	}
	
	return result, nil
}

// Helper functions to create aggregator specs
func SumSpec[T Numeric](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: SumAgg[T]()}
}

func CountSpec[T any](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: CountAgg[T]()}
}

func MinSpec[T Comparable](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: MinAgg[T]()}
}

func MaxSpec[T Comparable](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: MaxAgg[T]()}
}

func AvgSpec[T Numeric](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: AvgAgg[T]()}
}

// CustomSpec creates a spec for any custom aggregator
func CustomSpec[T, A, R any](name string, agg Aggregator[T, A, R]) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: agg}
}

// ============================================================================
// SIMPLIFIED MULTI-AGGREGATION (keeping for backward compatibility)
// ============================================================================

// Stats holds multiple aggregation results computed in a single pass
type Stats[T Numeric] struct {
	Count int64
	Sum   T
	Min   T
	Max   T
	Avg   float64
}

// MultiAggregate computes multiple statistics in a single pass through the stream
func MultiAggregate[T Numeric](stream Stream[T]) (Stats[T], error) {
	var stats Stats[T]
	first := true
	
	for {
		val, err := stream()
		if err != nil {
			if errors.Is(err, EOS) {
				if stats.Count > 0 {
					stats.Avg = float64(stats.Sum) / float64(stats.Count)
				}
				return stats, nil
			}
			return stats, err
		}
		
		stats.Count++
		stats.Sum += val
		
		if first {
			stats.Min = val
			stats.Max = val
			first = false
		} else {
			if val < stats.Min {
				stats.Min = val
			}
			if val > stats.Max {
				stats.Max = val
			}
		}
	}
}

// SumAndCount computes both sum and count in a single pass
func SumAndCount[T Numeric](stream Stream[T]) (sum T, count int64, err error) {
	var zero T
	for {
		val, streamErr := stream()
		if streamErr != nil {
			if errors.Is(streamErr, EOS) {
				return sum, count, nil
			}
			return zero, 0, streamErr
		}
		sum += val
		count++
	}
}

// Average computes the average in a single pass (combines sum and count)
func Average[T Numeric](stream Stream[T]) (float64, error) {
	sum, count, err := SumAndCount(stream)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, errors.New("empty stream")
	}
	return float64(sum) / float64(count), nil
}

// Tee splits a stream into multiple identical streams for parallel consumption
func Tee[T any](stream Stream[T], n int) []Stream[T] {
	if n <= 0 {
		return nil
	}
	
	// Collect all values first
	values, err := Collect(stream)
	if err != nil {
		// Return n error streams
		errorStreams := make([]Stream[T], n)
		for i := range errorStreams {
			errorStreams[i] = func() (T, error) {
				var zero T
				return zero, err
			}
		}
		return errorStreams
	}
	
	// Create n independent streams from the collected values
	streams := make([]Stream[T], n)
	for i := 0; i < n; i++ {
		streams[i] = FromSlice(values) // Each stream gets its own copy
	}
	
	return streams
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
