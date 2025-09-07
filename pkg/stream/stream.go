package stream

import (
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

// Common stream type aliases
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

// FromRecords creates a Record stream from Records
func FromRecords(records []Record) Stream[Record] {
	return FromSlice(records)
}

// FromMaps creates a Record stream from maps
func FromMaps(maps []map[string]any) Stream[Record] {
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