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
// SIMPLIFIED MULTI-AGGREGATION
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

// ============================================================================
// GROUPBY OPERATIONS - SQL-LIKE GROUPING
// ============================================================================

// GroupBy groups records by the specified fields and applies aggregations.
//
// Returns one record per group containing:
//   - The key fields from the group
//   - "group_count": number of records in the group
//   - Any additional aggregations specified
//
// Example:
//   grouped := stream.GroupBy([]string{"department"})(users)
//   // Each result contains: department, group_count
//
//   groupedWithAvg := stream.GroupByWith([]string{"department"}, 
//       stream.AvgSpec[float64]("avg_salary"))(users)
//   // Each result contains: department, group_count, avg_salary
func GroupBy(keyFields []string) Filter[Record, Record] {
	return GroupByWith(keyFields)
}

// GroupByWith groups records and applies custom aggregations to each group
func GroupByWith(keyFields []string, aggregators ...AggregatorSpec[Record]) Filter[Record, Record] {
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

			// Add group count
			result["group_count"] = int64(len(groupRecords))

			// Apply custom aggregations to this group
			if len(aggregators) > 0 {
				groupStream := FromSlice(groupRecords)
				aggregateResults, err := Aggregates(groupStream, aggregators...)
				if err == nil {
					// Merge aggregation results into the result record
					for key, value := range aggregateResults {
						result[key] = value
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

// FieldSumSpec creates an aggregator that sums a numeric field in records
func FieldSumSpec[T Numeric](name, fieldName string) AggregatorSpec[Record] {
	agg := Aggregator[Record, T, T]{
		Initial: func() T { var zero T; return zero },
		Accumulate: func(acc T, r Record) T {
			val := GetOr(r, fieldName, T(0))
			return acc + val
		},
		Finalize: func(acc T) T { return acc },
	}
	return AggregatorSpec[Record]{Name: name, Agg: agg}
}

// FieldAvgSpec creates an aggregator that averages a numeric field in records
func FieldAvgSpec[T Numeric](name, fieldName string) AggregatorSpec[Record] {
	agg := Aggregator[Record, [2]float64, float64]{
		Initial: func() [2]float64 { return [2]float64{0, 0} }, // [sum, count]
		Accumulate: func(acc [2]float64, r Record) [2]float64 {
			val := GetOr(r, fieldName, T(0))
			return [2]float64{acc[0] + float64(val), acc[1] + 1}
		},
		Finalize: func(acc [2]float64) float64 {
			if acc[1] == 0 {
				return 0
			}
			return acc[0] / acc[1]
		},
	}
	return AggregatorSpec[Record]{Name: name, Agg: agg}
}

// FieldMinSpec creates an aggregator that finds the minimum of a field in records
func FieldMinSpec[T Comparable](name, fieldName string) AggregatorSpec[Record] {
	agg := Aggregator[Record, *T, T]{
		Initial: func() *T { return nil },
		Accumulate: func(acc *T, r Record) *T {
			val := GetOr(r, fieldName, T(0))
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
	return AggregatorSpec[Record]{Name: name, Agg: agg}
}

// FieldMaxSpec creates an aggregator that finds the maximum of a field in records
func FieldMaxSpec[T Comparable](name, fieldName string) AggregatorSpec[Record] {
	agg := Aggregator[Record, *T, T]{
		Initial: func() *T { return nil },
		Accumulate: func(acc *T, r Record) *T {
			val := GetOr(r, fieldName, T(0))
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
	return AggregatorSpec[Record]{Name: name, Agg: agg}
}