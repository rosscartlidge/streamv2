package benchmarks_test

import (
	"testing"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// STREAMV2 vs GO-STREAMS PERFORMANCE BENCHMARKS
// ============================================================================
//
// This file contains benchmarks comparing StreamV2 against go-streams
// equivalent operations to demonstrate performance differences.
//
// Note: go-streams benchmarks are simulated/estimated based on typical
// interface{} overhead patterns since we don't import go-streams here.
// ============================================================================

const (
	benchmarkSize = 100000
)

// generateTestData creates test data for benchmarks
func generateTestData(size int) []int64 {
	data := make([]int64, size)
	for i := 0; i < size; i++ {
		data[i] = int64(i + 1)
	}
	return data
}

// ============================================================================
// MAP OPERATION BENCHMARKS
// ============================================================================

func BenchmarkStreamV2_Map(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dataStream := stream.FromSlice(data)
		mapped := stream.Map(func(x int64) int64 { return x * 2 })(dataStream)
		
		// Consume the stream
		_, _ = stream.Collect(mapped)
	}
}

func BenchmarkGoStreams_Map_Simulated(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate go-streams overhead:
		// 1. Box each int64 to interface{}
		// 2. Type assert back to int64  
		// 3. Box result back to interface{}
		result := make([]interface{}, len(data))
		for j, val := range data {
			// This simulates the interface{} boxing/unboxing overhead
			var boxed interface{} = val
			unboxed := boxed.(int64)
			result[j] = interface{}(unboxed * 2)
		}
		
		// Simulate consumption
		for _, val := range result {
			_ = val.(int64) // Type assertion on consumption
		}
	}
}

// ============================================================================
// FILTER OPERATION BENCHMARKS
// ============================================================================

func BenchmarkStreamV2_Filter(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dataStream := stream.FromSlice(data)
		filtered := stream.Where(func(x int64) bool { return x%2 == 0 })(dataStream)
		
		// Consume the stream
		_, _ = stream.Collect(filtered)
	}
}

func BenchmarkGoStreams_Filter_Simulated(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate go-streams overhead with filtering
		result := make([]interface{}, 0, len(data)/2)
		for _, val := range data {
			var boxed interface{} = val
			unboxed := boxed.(int64)
			if unboxed%2 == 0 {
				result = append(result, interface{}(unboxed))
			}
		}
		
		// Consume results
		for _, val := range result {
			_ = val.(int64)
		}
	}
}

// ============================================================================
// CHAINED OPERATIONS BENCHMARKS
// ============================================================================

func BenchmarkStreamV2_ChainedOps(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dataStream := stream.FromSlice(data)
		processed := stream.Chain(
			stream.Where(func(x int64) bool { return x%2 == 0 }),
			stream.Map(func(x int64) int64 { return x * x }),
			stream.Take[int64](1000),
		)(dataStream)
		
		// Consume the stream
		_, _ = stream.Collect(processed)
	}
}

func BenchmarkGoStreams_ChainedOps_Simulated(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate chained go-streams operations
		intermediate := make([]interface{}, 0, len(data)/2)
		
		// Filter step
		for _, val := range data {
			var boxed interface{} = val
			unboxed := boxed.(int64)
			if unboxed%2 == 0 {
				intermediate = append(intermediate, interface{}(unboxed))
			}
		}
		
		// Map step
		mapped := make([]interface{}, len(intermediate))
		for j, val := range intermediate {
			unboxed := val.(int64)
			mapped[j] = interface{}(unboxed * unboxed)
		}
		
		// Take step
		var result []interface{}
		limit := 1000
		if len(mapped) < limit {
			limit = len(mapped)
		}
		result = mapped[:limit]
		
		// Consume
		for _, val := range result {
			_ = val.(int64)
		}
	}
}

// ============================================================================
// AGGREGATION BENCHMARKS
// ============================================================================

func BenchmarkStreamV2_Sum(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dataStream := stream.FromSlice(data)
		_, _ = stream.Sum(dataStream)
	}
}

func BenchmarkGoStreams_Sum_Simulated(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate go-streams sum with interface{} overhead
		var sum int64
		for _, val := range data {
			var boxed interface{} = val
			unboxed := boxed.(int64)
			sum += unboxed
		}
		_ = sum
	}
}

func BenchmarkStreamV2_MultiAggregate(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dataStream := stream.FromSlice(data)
		_, _ = stream.MultiAggregate(dataStream)
	}
}

func BenchmarkGoStreams_MultiAggregate_Simulated(b *testing.B) {
	data := generateTestData(benchmarkSize)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate multiple passes required by go-streams
		// Pass 1: Count
		var count int64
		for _, val := range data {
			var boxed interface{} = val
			_ = boxed.(int64) // Type assertion
			count++
		}
		
		// Pass 2: Sum  
		var sum int64
		for _, val := range data {
			var boxed interface{} = val
			unboxed := boxed.(int64)
			sum += unboxed
		}
		
		// Pass 3: Min
		var min int64 = data[0]
		for _, val := range data[1:] {
			var boxed interface{} = val
			unboxed := boxed.(int64)
			if unboxed < min {
				min = unboxed
			}
		}
		
		// Pass 4: Max
		var max int64 = data[0]
		for _, val := range data[1:] {
			var boxed interface{} = val
			unboxed := boxed.(int64)
			if unboxed > max {
				max = unboxed
			}
		}
		
		// Compute average
		avg := float64(sum) / float64(count)
		_, _, _, _, _ = count, sum, min, max, avg
	}
}

// ============================================================================
// RECORD PROCESSING BENCHMARKS
// ============================================================================

func BenchmarkStreamV2_RecordProcessing(b *testing.B) {
	// Create test records
	records := make([]stream.Record, benchmarkSize)
	for i := 0; i < benchmarkSize; i++ {
		records[i] = stream.R(
			"id", int64(i),
			"value", float64(i)*1.5,
			"active", i%2 == 0,
		)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recordStream := stream.FromRecords(records)
		processed := stream.Chain(
			stream.Where(func(r stream.Record) bool {
				return stream.GetOr(r, "active", false)
			}),
			stream.Update(func(r stream.Record) stream.Record {
				value := stream.GetOr(r, "value", 0.0)
				return r.Set("doubled", value*2)
			}),
		)(recordStream)
		
		// Consume stream
		_, _ = stream.Collect(processed)
	}
}

func BenchmarkGoStreams_RecordProcessing_Simulated(b *testing.B) {
	// Create test records as interface{} maps
	records := make([]interface{}, benchmarkSize)
	for i := 0; i < benchmarkSize; i++ {
		records[i] = map[string]interface{}{
			"id":     int64(i),
			"value":  float64(i) * 1.5,
			"active": i%2 == 0,
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := make([]interface{}, 0, len(records)/2)
		
		for _, record := range records {
			// Simulate go-streams record processing
			recordMap := record.(map[string]interface{})
			
			// Filter step
			active, ok := recordMap["active"].(bool)
			if !ok || !active {
				continue
			}
			
			// Transform step
			value, ok := recordMap["value"].(float64)
			if !ok {
				value = 0.0
			}
			
			// Create new record
			newRecord := make(map[string]interface{})
			for k, v := range recordMap {
				newRecord[k] = v
			}
			newRecord["doubled"] = value * 2
			
			result = append(result, newRecord)
		}
		
		// Consume results
		for _, record := range result {
			_ = record.(map[string]interface{})
		}
	}
}

// ============================================================================
// MEMORY ALLOCATION BENCHMARKS
// ============================================================================

func BenchmarkStreamV2_MemoryEfficiency(b *testing.B) {
	data := generateTestData(10000) // Smaller dataset for memory testing
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// StreamV2 uses direct types, minimal allocations
		dataStream := stream.FromSlice(data)
		mapped := stream.Map(func(x int64) int64 { return x * 2 })(dataStream)
		results, _ := stream.Collect(mapped)
		_ = results
	}
}

func BenchmarkGoStreams_MemoryEfficiency_Simulated(b *testing.B) {
	data := generateTestData(10000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate go-streams memory overhead
		// Box all data to interface{}
		boxed := make([]interface{}, len(data))
		for j, val := range data {
			boxed[j] = interface{}(val)
		}
		
		// Process with interface{} overhead
		results := make([]interface{}, len(boxed))
		for j, val := range boxed {
			unboxed := val.(int64)
			results[j] = interface{}(unboxed * 2)
		}
		
		// Unbox for final consumption
		final := make([]int64, len(results))
		for j, val := range results {
			final[j] = val.(int64)
		}
		_ = final
	}
}