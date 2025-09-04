package stream

import (
	"fmt"
)

// ============================================================================
// NESTED STREAM SUPPORT - STREAMS AS RECORD FIELDS
// ============================================================================

// StreamValue wraps a stream to be stored in records
type StreamValue[T any] struct {
	stream Stream[T]
	cached []T // Cache for multiple iterations
	index  int // Current position in cache
}

// NewStreamValue creates a StreamValue from a stream
func NewStreamValue[T any](stream Stream[T]) *StreamValue[T] {
	return &StreamValue[T]{
		stream: stream,
		cached: nil,
		index:  0,
	}
}

// Stream returns the underlying stream
func (sv *StreamValue[T]) Stream() Stream[T] {
	if sv.cached != nil {
		// Return cached stream
		return sv.cachedStream()
	}
	return sv.stream
}

// cachedStream creates a stream from cached data
func (sv *StreamValue[T]) cachedStream() Stream[T] {
	index := 0
	return func() (T, error) {
		if index >= len(sv.cached) {
			var zero T
			return zero, EOS
		}
		val := sv.cached[index]
		index++
		return val, nil
	}
}

// Cache forces the stream to be materialized for multiple access
func (sv *StreamValue[T]) Cache() error {
	if sv.cached != nil {
		return nil // Already cached
	}

	var items []T
	for {
		item, err := sv.stream()
		if err != nil {
			if err == EOS {
				break
			}
			return err
		}
		items = append(items, item)
	}

	sv.cached = items
	sv.index = 0
	return nil
}

// String provides string representation
func (sv *StreamValue[T]) String() string {
	if sv.cached != nil {
		return fmt.Sprintf("StreamValue[cached:%d items]", len(sv.cached))
	}
	return fmt.Sprintf("StreamValue[uncached]")
}

// ============================================================================
// ENHANCED RECORD OPERATIONS FOR STREAMS
// ============================================================================

// GetStream retrieves a stream from a record with type safety
func GetStream[T any](r Record, field string) (Stream[T], bool) {
	val, exists := r[field]
	if !exists {
		return nil, false
	}

	// Check if it's a StreamValue
	if sv, ok := val.(*StreamValue[T]); ok {
		return sv.Stream(), true
	}

	// Check if it's a direct stream
	if stream, ok := val.(Stream[T]); ok {
		return stream, true
	}

	// Check if it's a slice that can be converted to stream
	if slice, ok := val.([]T); ok {
		return FromSlice(slice), true
	}

	return nil, false
}

// SetStream stores a stream in a record
func (r Record) SetStream(field string, stream Stream[any]) Record {
	r[field] = NewStreamValue(stream)
	return r
}

// SetTypedStream stores a typed stream in a record
func SetTypedStream[T any](r Record, field string, stream Stream[T]) Record {
	r[field] = NewStreamValue(stream)
	return r
}

// ============================================================================
// STREAM RECORD CONSTRUCTORS
// ============================================================================

// RWithStream creates a record with stream fields
func RWithStreams(pairs ...any) Record {
	if len(pairs)%2 != 0 {
		panic("RWithStreams() requires even number of arguments")
	}

	r := make(Record)
	for i := 0; i < len(pairs); i += 2 {
		key := pairs[i].(string)
		value := pairs[i+1]

		// Wrap streams automatically
		if stream, ok := value.(Stream[any]); ok {
			r[key] = NewStreamValue(stream)
		} else {
			r[key] = value
		}
	}
	return r
}

// ============================================================================
// NESTED STREAM OPERATIONS
// ============================================================================

// FlatMap flattens nested streams (stream of streams → single stream)
func FlatMap[T, U any](fn func(T) Stream[U]) Filter[T, U] {
	return func(input Stream[T]) Stream[U] {
		var currentSubStream Stream[U]

		return func() (U, error) {
			for {
				// Try to get from current sub-stream
				if currentSubStream != nil {
					val, err := currentSubStream()
					if err == nil {
						return val, nil
					}
					if err != EOS {
						var zero U
						return zero, err
					}
					// Current sub-stream is exhausted
					currentSubStream = nil
				}

				// Get next item from main stream
				item, err := input()
				if err != nil {
					var zero U
					return zero, err
				}

				// Create new sub-stream
				currentSubStream = fn(item)
			}
		}
	}
}

// ExpandStreams expands stream fields in records  
func ExpandStreams(streamFields ...string) Filter[Record, Record] {
	return func(input Stream[Record]) Stream[Record] {
		// Collect all records first for easier processing
		records, err := Collect(input)
		if err != nil {
			return func() (Record, error) { return nil, err }
		}

		var expandedRecords []Record
		
		// Process each record
		for _, record := range records {
			// Extract stream values for the specified fields
			streamValues := make(map[string][]any)
			baseRecord := make(Record)
			maxLength := 0

			// Process each field in the record
			for k, v := range record {
				isStreamField := false
				for _, sf := range streamFields {
					if k == sf {
						isStreamField = true
						break
					}
				}

				if isStreamField {
					// Extract values from this stream field
					var values []any
					if sv, ok := v.(*StreamValue[int64]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, any(val))
						}
					} else if sv, ok := v.(*StreamValue[string]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, any(val))
						}
					} else if sv, ok := v.(*StreamValue[float64]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, any(val))
						}
					} else if sv, ok := v.(*StreamValue[any]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, val)
						}
					}
					
					if len(values) > 0 {
						streamValues[k] = values
						if len(values) > maxLength {
							maxLength = len(values)
						}
					}
				} else {
					// Regular field - copy to base record
					baseRecord[k] = v
				}
			}

			// Generate expanded records by iterating through stream values
			for i := 0; i < maxLength; i++ {
				expandedRecord := make(Record)
				
				// Copy base fields
				for k, v := range baseRecord {
					expandedRecord[k] = v
				}
				
				// Add stream values at current index
				for field, values := range streamValues {
					if i < len(values) {
						expandedRecord[field] = values[i]
					}
					// Note: if stream is shorter, field is omitted from this record
				}
				
				expandedRecords = append(expandedRecords, expandedRecord)
			}
		}

		return FromSlice(expandedRecords)
	}
}

// ============================================================================
// EXAMPLES AND USAGE PATTERNS
// ============================================================================

// CreateNestedRecord demonstrates creating records with nested streams
func CreateNestedRecord() Record {
	// Create some sample streams
	numbersStream := FromSlice([]int64{1, 2, 3, 4, 5})
	stringsStream := FromSlice([]string{"a", "b", "c"})

	// Create record with nested streams
	record := R(
		"id", 123,
		"name", "example",
	)

	// Add streams to record
	SetTypedStream(record, "numbers", numbersStream)
	SetTypedStream(record, "letters", stringsStream)

	return record
}

// ProcessNestedStreams demonstrates processing records with streams
func ProcessNestedStreams(record Record) {
	// Extract and process nested streams
	if numbersStream, ok := GetStream[int64](record, "numbers"); ok {
		// Process the nested stream
		sum, _ := Sum(numbersStream)
		fmt.Printf("Sum of nested numbers: %d\n", sum)
	}

	if lettersStream, ok := GetStream[string](record, "letters"); ok {
		// Collect nested stream
		letters, _ := Collect(lettersStream)
		fmt.Printf("Nested letters: %v\n", letters)
	}
}

// GroupByWithNestedStreams groups records and creates streams of grouped data
func GroupByWithNestedStreams(keyFields []string) Filter[Record, Record] {
	return func(input Stream[Record]) Stream[Record] {
		// Collect all records
		records, err := Collect(input)
		if err != nil {
			return func() (Record, error) { return nil, err }
		}

		// Group records
		groups := make(map[string][]Record)
		for _, record := range records {
			key := ""
			for _, field := range keyFields {
				if val, exists := record[field]; exists {
					key += fmt.Sprintf("%v,", val)
				}
			}
			groups[key] = append(groups[key], record)
		}

		// Create results with grouped data as streams
		var results []Record
		for _, groupRecords := range groups {
			result := make(Record)

			// Add key fields from first record
			if len(groupRecords) > 0 {
				for _, field := range keyFields {
					if val, exists := groupRecords[0][field]; exists {
						result[field] = val
					}
				}
			}

			// Add grouped records as a stream
			result["grouped_records"] = NewStreamValue(FromSlice(groupRecords))
			result["group_count"] = len(groupRecords)

			results = append(results, result)
		}

		return FromSlice(results)
	}
}
