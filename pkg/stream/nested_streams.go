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

// detectStreamFields automatically finds all fields that contain streams
func detectStreamFields(record Record) []string {
	var streamFields []string
	
	for fieldName, value := range record {
		// Check if this field contains a StreamValue of any type
		isStream := false
		
		// Check common StreamValue types
		if _, ok := value.(*StreamValue[int64]); ok {
			isStream = true
		} else if _, ok := value.(*StreamValue[string]); ok {
			isStream = true
		} else if _, ok := value.(*StreamValue[float64]); ok {
			isStream = true
		} else if _, ok := value.(*StreamValue[any]); ok {
			isStream = true
		} else if _, ok := value.(*StreamValue[int]); ok {
			isStream = true
		} else if _, ok := value.(*StreamValue[uint64]); ok {
			isStream = true
		} else if _, ok := value.(*StreamValue[bool]); ok {
			isStream = true
		}
		// Add more types as needed
		
		if isStream {
			streamFields = append(streamFields, fieldName)
		}
	}
	
	return streamFields
}

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

// ExpandStreamsDot expands stream fields using dot product (element-wise pairing)
// For streams of different lengths, shorter streams are exhausted first
// If no streamFields are specified, all stream fields in the record are expanded
func ExpandStreamsDot(streamFields ...string) Filter[Record, Record] {
	return func(input Stream[Record]) Stream[Record] {
		// Collect all records first for easier processing
		records, err := Collect(input)
		if err != nil {
			return func() (Record, error) { return nil, err }
		}

		var expandedRecords []Record
		
		// Process each record
		for _, record := range records {
			// Auto-detect stream fields if none specified
			fieldsToExpand := streamFields
			if len(fieldsToExpand) == 0 {
				fieldsToExpand = detectStreamFields(record)
			}
			
			// Cache all stream values first to ensure multiple access
			for _, fieldName := range fieldsToExpand {
				if val, exists := record[fieldName]; exists {
					if sv, ok := val.(*StreamValue[int64]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[string]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[float64]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[any]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[int]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[uint64]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[bool]); ok {
						sv.Cache()
					}
				}
			}
			
			// Extract stream values for the specified fields
			streamValues := make(map[string][]any)
			baseRecord := make(Record)
			maxLength := 0

			// Process each field in the record
			for k, v := range record {
				isStreamField := false
				for _, sf := range fieldsToExpand {
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
					} else if sv, ok := v.(*StreamValue[int]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, any(val))
						}
					} else if sv, ok := v.(*StreamValue[uint64]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, any(val))
						}
					} else if sv, ok := v.(*StreamValue[bool]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, any(val))
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

// ExpandStreamsCross expands stream fields using cross product (Cartesian product)
// Every value from each stream is combined with every value from other streams
// If no streamFields are specified, all stream fields in the record are expanded
func ExpandStreamsCross(streamFields ...string) Filter[Record, Record] {
	return func(input Stream[Record]) Stream[Record] {
		// Collect all records first for easier processing
		records, err := Collect(input)
		if err != nil {
			return func() (Record, error) { return nil, err }
		}

		var expandedRecords []Record
		
		// Process each record
		for _, record := range records {
			// Auto-detect stream fields if none specified
			fieldsToExpand := streamFields
			if len(fieldsToExpand) == 0 {
				fieldsToExpand = detectStreamFields(record)
			}
			
			// Cache all stream values first to ensure multiple access
			for _, fieldName := range fieldsToExpand {
				if val, exists := record[fieldName]; exists {
					if sv, ok := val.(*StreamValue[int64]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[string]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[float64]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[any]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[int]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[uint64]); ok {
						sv.Cache()
					} else if sv, ok := val.(*StreamValue[bool]); ok {
						sv.Cache()
					}
				}
			}
			
			// Extract stream values for the specified fields
			streamValues := make(map[string][]any)
			baseRecord := make(Record)

			// Process each field in the record
			for k, v := range record {
				isStreamField := false
				for _, sf := range fieldsToExpand {
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
					} else if sv, ok := v.(*StreamValue[int]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, any(val))
						}
					} else if sv, ok := v.(*StreamValue[uint64]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, any(val))
						}
					} else if sv, ok := v.(*StreamValue[bool]); ok {
						vals, _ := Collect(sv.Stream())
						for _, val := range vals {
							values = append(values, any(val))
						}
					}
					
					if len(values) > 0 {
						streamValues[k] = values
					}
				} else {
					// Regular field - copy to base record
					baseRecord[k] = v
				}
			}

			// Generate all combinations using recursive cross product
			combinations := generateCrossProduct(fieldsToExpand, streamValues)
			
			// Create expanded records for each combination
			for _, combo := range combinations {
				expandedRecord := make(Record)
				
				// Copy base fields
				for k, v := range baseRecord {
					expandedRecord[k] = v
				}
				
				// Add combination values
				for field, value := range combo {
					expandedRecord[field] = value
				}
				
				expandedRecords = append(expandedRecords, expandedRecord)
			}
		}

		return FromSlice(expandedRecords)
	}
}

// generateCrossProduct creates all combinations of stream values
func generateCrossProduct(fields []string, values map[string][]any) []map[string]any {
	if len(fields) == 0 {
		return []map[string]any{{}}
	}
	
	field := fields[0]
	fieldValues, exists := values[field]
	if !exists || len(fieldValues) == 0 {
		// If this field has no values, continue with remaining fields
		return generateCrossProduct(fields[1:], values)
	}
	
	// Get combinations for remaining fields
	restCombinations := generateCrossProduct(fields[1:], values)
	
	var allCombinations []map[string]any
	
	// For each value in this field, combine with all rest combinations
	for _, value := range fieldValues {
		for _, restCombo := range restCombinations {
			combo := make(map[string]any)
			combo[field] = value
			for k, v := range restCombo {
				combo[k] = v
			}
			allCombinations = append(allCombinations, combo)
		}
	}
	
	return allCombinations
}

// ExpandStreams is an alias for ExpandStreamsDot for backward compatibility
func ExpandStreams(streamFields ...string) Filter[Record, Record] {
	return ExpandStreamsDot(streamFields...)
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
