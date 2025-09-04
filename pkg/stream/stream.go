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
