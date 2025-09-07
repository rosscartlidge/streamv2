package stream

import (
	"context"
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

// ============================================================================
// STREAM UTILITIES
// ============================================================================

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