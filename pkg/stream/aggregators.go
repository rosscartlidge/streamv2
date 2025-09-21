package stream

import (
	"errors"
	"fmt"
)

// ============================================================================
// TYPE CONSTRAINTS FOR AGGREGATIONS
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
// BASIC AGGREGATION FUNCTIONS - TYPE SAFE
// ============================================================================

// Sum aggregates numeric values
func Sum[T Numeric](stream Stream[T]) (T, error) {
	return RunAggregator(stream, SumAggregator[T, T](func(val T) T { return val }))
}

// Count counts elements
func Count[T any](stream Stream[T]) (int64, error) {
	return RunAggregator(stream, CountAggregator[T]())
}

// Max finds maximum value
func Max[T Comparable](stream Stream[T]) (T, error) {
	return RunAggregator(stream, MaxAggregator[T, T](func(val T) T { return val }))
}

// Min finds minimum value
func Min[T Comparable](stream Stream[T]) (T, error) {
	return RunAggregator(stream, MinAggregator[T, T](func(val T) T { return val }))
}

// Avg calculates average of numeric values
func Avg[T Numeric](stream Stream[T]) (float64, error) {
	return RunAggregator(stream, AvgAggregator[T, T](func(val T) T { return val }))
}

// ============================================================================
// STREAM COLLECTION AND PROCESSING
// ============================================================================

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

// RunAggregator runs a single aggregator on a stream and returns the result
func RunAggregator[T, A, R any](stream Stream[T], agg Aggregator[T, A, R]) (R, error) {
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

// ============================================================================
// GENERIC AGGREGATOR FACTORIES
// ============================================================================

// SumAggregator creates a sum aggregator with custom value extraction
func SumAggregator[I any, T Numeric](extract func(I) T) Aggregator[I, T, T] {
	return Aggregator[I, T, T]{
		Initial:    func() T { var zero T; return zero },
		Accumulate: func(acc T, input I) T { return acc + extract(input) },
		Finalize:   func(acc T) T { return acc },
	}
}

// MinAggregator creates a min aggregator with custom value extraction  
func MinAggregator[I any, T Comparable](extract func(I) T) Aggregator[I, *T, T] {
	return Aggregator[I, *T, T]{
		Initial: func() *T { return nil },
		Accumulate: func(acc *T, input I) *T {
			val := extract(input)
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

// MaxAggregator creates a max aggregator with custom value extraction
func MaxAggregator[I any, T Comparable](extract func(I) T) Aggregator[I, *T, T] {
	return Aggregator[I, *T, T]{
		Initial: func() *T { return nil },
		Accumulate: func(acc *T, input I) *T {
			val := extract(input)
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

// AvgAggregator creates an average aggregator with custom value extraction
func AvgAggregator[I any, T Numeric](extract func(I) T) Aggregator[I, [2]float64, float64] {
	return Aggregator[I, [2]float64, float64]{
		Initial: func() [2]float64 { return [2]float64{0, 0} }, // [sum, count]
		Accumulate: func(acc [2]float64, input I) [2]float64 {
			val := extract(input)
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

// CountAggregator creates a count aggregator (doesn't need value extraction)
func CountAggregator[I any]() Aggregator[I, int64, int64] {
	return Aggregator[I, int64, int64]{
		Initial:    func() int64 { return 0 },
		Accumulate: func(acc int64, _ I) int64 { return acc + 1 },
		Finalize:   func(acc int64) int64 { return acc },
	}
}

// ============================================================================
// BUILT-IN AGGREGATORS - using the generic factories
// ============================================================================






// ============================================================================
// FIELD AGGREGATORS - operate on Record streams by extracting fields
// ============================================================================

// SumAggregatorField creates an aggregator that sums a numeric field in records
func SumAggregatorField[T Numeric](fieldName string) Aggregator[Record, T, T] {
	return SumAggregator[Record, T](func(r Record) T {
		var zero T
		return GetOr(r, fieldName, zero)
	})
}

// AvgAggregatorField creates an aggregator that averages a numeric field in records
func AvgAggregatorField[T Numeric](fieldName string) Aggregator[Record, [2]float64, float64] {
	return AvgAggregator[Record, T](func(r Record) T {
		var zero T
		return GetOr(r, fieldName, zero)
	})
}

// MinAggregatorField creates an aggregator that finds the minimum of a field in records
func MinAggregatorField[T Comparable](fieldName string) Aggregator[Record, *T, T] {
	return MinAggregator[Record, T](func(r Record) T {
		var zero T
		return GetOr(r, fieldName, zero)
	})
}

// MaxAggregatorField creates an aggregator that finds the maximum of a field in records
func MaxAggregatorField[T Comparable](fieldName string) Aggregator[Record, *T, T] {
	return MaxAggregator[Record, T](func(r Record) T {
		var zero T
		return GetOr(r, fieldName, zero)
	})
}

// CountAggregatorField creates an aggregator that counts records (field name is ignored but maintained for consistency)
func CountAggregatorField(fieldName string) Aggregator[Record, int64, int64] {
	return CountAggregator[Record]()
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
func SumStream[T Numeric](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: SumAggregator[T, T](func(val T) T { return val })}
}

func CountStream[T any](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: CountAggregator[T]()}
}

func MinStream[T Comparable](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: MinAggregator[T, T](func(val T) T { return val })}
}

func MaxStream[T Comparable](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: MaxAggregator[T, T](func(val T) T { return val })}
}

func AvgStream[T Numeric](name string) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: AvgAggregator[T, T](func(val T) T { return val })}
}

// CustomSpec creates a spec for any custom aggregator
func CustomSpec[T, A, R any](name string, agg Aggregator[T, A, R]) AggregatorSpec[T] {
	return AggregatorSpec[T]{Name: name, Agg: agg}
}


// ============================================================================
// GROUPBY OPERATIONS - SQL-LIKE GROUPING
// ============================================================================

// GroupBy groups records by the specified fields and applies aggregations.
//
// Returns one record per group containing:
//   - The key fields from the group
//   - Any additional aggregations specified
//
// Example:
//   grouped := stream.GroupBy([]string{"department"})(users)
//   // Each result contains: department
//
//   groupedWithStats := stream.GroupBy([]string{"department"}, 
//       stream.AvgField[float64]("avg_salary", "salary"),
//       stream.CountField("count", "name"))(users)
//   // Each result contains: department, avg_salary, count
// GroupBy groups records and applies custom aggregations to each group
func GroupBy(keyFields []string, aggregators ...AggregatorSpec[Record]) Filter[Record, Record] {
	return func(input Stream[Record]) Stream[Record] {
		// Collect all records
		records, err := Collect(input)
		if err != nil {
			return func() (Record, error) { return nil, err }
		}

		// Group records by key fields
		groups := make(map[string][]Record)
		for _, record := range records {
			key := buildGroupKey(record, keyFields)
			groups[key] = append(groups[key], record)
		}

		// Process each group
		var results []Record
		for _, groupRecords := range groups {
			if len(groupRecords) == 0 {
				continue
			}

			// Create result record with key fields
			result := make(Record)
			for _, field := range keyFields {
				if val, exists := groupRecords[0][field]; exists {
					result[field] = val
				}
			}


			// Apply custom aggregations to this group
			if len(aggregators) > 0 {
				groupStream := FromSlice(groupRecords)
				
				// Run each aggregator directly on the group
				for _, spec := range aggregators {
					var value interface{}
					var err error
					
					// Type-assert the specific Record aggregator types
					switch agg := spec.Agg.(type) {
					case Aggregator[Record, int64, int64]:
						value, err = AggregateWith(groupStream, agg)
						groupStream = FromSlice(groupRecords) // Reset for next aggregator
					case Aggregator[Record, [2]float64, float64]:
						value, err = AggregateWith(groupStream, agg)
						groupStream = FromSlice(groupRecords) // Reset for next aggregator
					case Aggregator[Record, *int64, int64]:
						value, err = AggregateWith(groupStream, agg)
						groupStream = FromSlice(groupRecords) // Reset for next aggregator
					default:
						err = fmt.Errorf("unsupported aggregator type for '%s'", spec.Name)
					}
					
					if err == nil {
						result[spec.Name] = value
					}
				}
			}

			results = append(results, result)
		}

		return FromSlice(results)
	}
}

// buildGroupKey creates a composite key from the specified fields
func buildGroupKey(record Record, keyFields []string) string {
	key := ""
	for i, field := range keyFields {
		if i > 0 {
			key += "|" // Use pipe separator to avoid collisions
		}
		if val, exists := record[field]; exists {
			key += fmt.Sprintf("%v", val)
		} else {
			key += "<nil>"
		}
	}
	return key
}

// ============================================================================
// RECORD FIELD AGGREGATIONS - FOR GROUPBY
// ============================================================================

// SumField creates an aggregator that sums a numeric field in records
func SumField[T Numeric](name, fieldName string) AggregatorSpec[Record] {
	return AggregatorSpec[Record]{Name: name, Agg: SumAggregatorField[T](fieldName)}
}

// AvgField creates an aggregator that averages a numeric field in records
func AvgField[T Numeric](name, fieldName string) AggregatorSpec[Record] {
	return AggregatorSpec[Record]{Name: name, Agg: AvgAggregatorField[T](fieldName)}
}

// MinField creates an aggregator that finds the minimum of a field in records
func MinField[T Comparable](name, fieldName string) AggregatorSpec[Record] {
	return AggregatorSpec[Record]{Name: name, Agg: MinAggregatorField[T](fieldName)}
}

// MaxField creates an aggregator that finds the maximum of a field in records
func MaxField[T Comparable](name, fieldName string) AggregatorSpec[Record] {
	return AggregatorSpec[Record]{Name: name, Agg: MaxAggregatorField[T](fieldName)}
}

// CountField creates an aggregator that counts records (field name is ignored but maintained for consistency)
func CountField(name, fieldName string) AggregatorSpec[Record] {
	return AggregatorSpec[Record]{Name: name, Agg: CountAggregatorField(fieldName)}
}