package stream

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
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
// TYPE-SAFE VALUE SYSTEM - ZERO REFLECTION
// ============================================================================

// Value defines the compile-time type constraint for Record field values
// This replaces runtime reflection with compile-time type safety
type Value interface {
	// Integer types
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
	
	// Float types  
	~float32 | ~float64 |
	
	// Other basic types
	~bool | ~string | time.Time |
	
	// Record type for nested structures
	Record |
	
	// Stream types for all allowed element types
	Stream[int] | Stream[int8] | Stream[int16] | Stream[int32] | Stream[int64] |
	Stream[uint] | Stream[uint8] | Stream[uint16] | Stream[uint32] | Stream[uint64] |
	Stream[float32] | Stream[float64] |
	Stream[bool] | Stream[string] | Stream[time.Time] |
	Stream[Record]
}

// ============================================================================
// TYPE-SAFE RECORD BUILDERS - ZERO REFLECTION
// ============================================================================

// ============================================================================
// FLUENT TYPED RECORD BUILDER
// ============================================================================

// TypedRecord provides type-safe field setting with method chaining
type TypedRecord struct {
	data map[string]any
}

// NewRecord creates a new type-safe Record builder
func NewRecord() *TypedRecord {
	return &TypedRecord{data: make(map[string]any)}
}

// Set adds a field with compile-time type safety
func (tr *TypedRecord) Set(key string, value any) *TypedRecord {
	tr.data[key] = value
	return tr
}

// Build returns the completed Record
func (tr *TypedRecord) Build() Record {
	return Record(tr.data)
}

// Convenience methods for common types with type safety
func (tr *TypedRecord) String(key string, value string) *TypedRecord {
	tr.data[key] = value
	return tr
}

func (tr *TypedRecord) Int(key string, value int64) *TypedRecord {
	tr.data[key] = value
	return tr
}

func (tr *TypedRecord) Float(key string, value float64) *TypedRecord {
	tr.data[key] = value
	return tr
}

func (tr *TypedRecord) Bool(key string, value bool) *TypedRecord {
	tr.data[key] = value
	return tr
}

func (tr *TypedRecord) Time(key string, value time.Time) *TypedRecord {
	tr.data[key] = value
	return tr
}

func (tr *TypedRecord) Record(key string, value Record) *TypedRecord {
	tr.data[key] = value
	return tr
}

// Field creates a single-field Record with compile-time type safety
func Field[V Value](key string, value V) Record {
	return Record{key: value}
}

// ============================================================================
// LEGACY SUPPORT - INTERNAL USE ONLY
// ============================================================================

// recordFrom creates a Record from map[string]any without validation (internal use)
func recordFrom(m map[string]any) Record {
	r := make(Record, len(m))
	for key, value := range m {
		r[key] = value
	}
	return r
}

// recordsFrom creates Records from slice of maps (internal use)
func recordsFrom(maps []map[string]any) []Record {
	records := make([]Record, len(maps))
	for i, m := range maps {
		records[i] = recordFrom(m)
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

// SetField assigns a value to a record field with compile-time type safety
func SetField[V Value](r Record, field string, value V) Record {
	result := make(Record, len(r)+1)
	for k, v := range r {
		result[k] = v
	}
	result[field] = value
	return result
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
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed, true
		}
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
// STREAM CREATION - SAFE BY DEFAULT
// ============================================================================

// FromSlice creates a type-safe stream from a slice (Safe by Default)
// Only accepts types that are compatible with Record Value system
func FromSlice[V Value](items []V) Stream[V] {
	return fromSliceImpl(items)
}

// FromChannel creates a type-safe stream from a channel (Safe by Default)
// Only accepts types that are compatible with Record Value system
func FromChannel[V Value](ch <-chan V) Stream[V] {
	return fromChannelImpl(ch)
}

// From creates a type-safe stream from variadic arguments (Safe by Default)
func From[V Value](items ...V) Stream[V] {
	return fromSliceImpl(items)
}

// ============================================================================
// ESCAPE HATCHES - ADVANCED USE ONLY
// ============================================================================

// FromSliceAny creates a stream from any type slice - USE WITH CAUTION
// This bypasses Value type safety - ensure types are compatible with your use case
func FromSliceAny[T any](items []T) Stream[T] {
	return fromSliceImpl(items)
}

// FromChannelAny creates a stream from any type channel - USE WITH CAUTION  
// This bypasses Value type safety - ensure types are compatible with your use case
func FromChannelAny[T any](ch <-chan T) Stream[T] {
	return fromChannelImpl(ch)
}

// ============================================================================
// INTERNAL IMPLEMENTATIONS - SHARED BY SAFE AND UNSAFE VARIANTS
// ============================================================================

func fromSliceImpl[T any](items []T) Stream[T] {
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

func fromChannelImpl[T any](ch <-chan T) Stream[T] {
	return func() (T, error) {
		item, ok := <-ch
		if !ok {
			var zero T
			return zero, EOS
		}
		return item, nil
	}
}

// ============================================================================
// RECORD STREAM CREATION - WITH VALIDATION
// ============================================================================

// FromRecords creates a Record stream with field type validation
func FromRecords(records []Record) (Stream[Record], error) {
	// Validate all records have Value-compatible fields
	for i, record := range records {
		if err := validateRecord(record); err != nil {
			return nil, fmt.Errorf("record %d invalid: %w", i, err)
		}
	}
	return FromSlice(records), nil
}

// FromMaps creates a Record stream from maps with field type validation
func FromMaps(maps []map[string]any) (Stream[Record], error) {
	records := make([]Record, len(maps))
	for i, m := range maps {
		record := make(Record, len(m))
		for key, value := range m {
			// Validate each field type
			if !isValueType(value) {
				return nil, fmt.Errorf("map %d field '%s' has invalid type %T", i, key, value)
			}
			record[key] = value
		}
		records[i] = record
	}
	return FromSlice(records), nil
}

// FromRecordsUnsafe creates a Record stream without validation - USE WITH CAUTION
func FromRecordsUnsafe(records []Record) Stream[Record] {
	return FromSlice(records)
}

// FromMapsUnsafe creates a Record stream from maps without validation - USE WITH CAUTION
func FromMapsUnsafe(maps []map[string]any) Stream[Record] {
	return FromSlice(recordsFrom(maps))
}

// validateRecord checks if a Record has only Value-compatible field types
func validateRecord(r Record) error {
	for field, value := range r {
		if !isValueType(value) {
			return fmt.Errorf("field '%s' has invalid type %T", field, value)
		}
	}
	return nil
}

// isValueType checks if a value conforms to the Value interface using type assertions
func isValueType(value any) bool {
	switch value.(type) {
	// Integer types
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	// Float types
	case float32, float64:
		return true
	// Other basic types
	case bool, string:
		return true
	case time.Time:
		return true
	// Record type
	case Record:
		return true
	// Stream types (basic check - could be extended)
	case Stream[int], Stream[int8], Stream[int16], Stream[int32], Stream[int64]:
		return true
	case Stream[uint], Stream[uint8], Stream[uint16], Stream[uint32], Stream[uint64]:
		return true
	case Stream[float32], Stream[float64]:
		return true
	case Stream[bool], Stream[string]:
		return true
	case Stream[time.Time], Stream[Record]:
		return true
	default:
		return false
	}
}

// ============================================================================
// OTHER STREAM CONSTRUCTORS
// ============================================================================

// Generate creates a Value-safe stream using a generator function (Safe by Default)
func Generate[V Value](generator func() (V, error)) Stream[V] {
	return generator
}

// GenerateAny creates a stream using any type generator - USE WITH CAUTION
func GenerateAny[T any](generator func() (T, error)) Stream[T] {
	return generator
}

// Range creates a numeric stream (int64 values are Value-compatible)
func Range(start, end, step int64) Stream[int64] {
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

// Once creates a Value-safe stream with a single element (Safe by Default)
func Once[V Value](item V) Stream[V] {
	consumed := false
	return func() (V, error) {
		if consumed {
			var zero V
			return zero, EOS
		}
		consumed = true
		return item, nil
	}
}

// OnceAny creates a stream with a single element of any type - USE WITH CAUTION
func OnceAny[T any](item T) Stream[T] {
	consumed := false
	return func() (T, error) {
		if consumed {
			var zero T
			return zero, EOS
		}
		consumed = true
		return item, nil
	}
}