package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// STREAM FACTORY PROTOTYPE IMPLEMENTATION
// ============================================================================

// StreamFactory creates fresh streams on demand
type StreamFactory[T any] func() stream.Stream[T]

// ============================================================================
// FACTORY CONSTRUCTORS
// ============================================================================

// NewSliceFactory creates a factory from an immutable slice
func NewSliceFactory[T any](data []T) StreamFactory[T] {
	// Capture the slice data (safe since we copy it)
	dataCopy := make([]T, len(data))
	copy(dataCopy, data)
	
	return func() stream.Stream[T] {
		return stream.FromSlice(dataCopy)
	}
}

// NewGeneratorFactory creates a factory with deterministic generator
func NewGeneratorFactory[T any](generator func() T, count int) StreamFactory[T] {
	return func() stream.Stream[T] {
		return stream.Generate(generator).Take(count)
	}
}

// NewSeededRandomFactory creates reproducible random data
func NewSeededRandomFactory(seed int64, count int) StreamFactory[float64] {
	return func() stream.Stream[float64] {
		rng := rand.New(rand.NewSource(seed))
		return stream.Generate(func() float64 {
			return rng.Float64()
		}).Take(count)
	}
}

// ============================================================================
// CACHED FACTORY FOR EXPENSIVE OPERATIONS
// ============================================================================

type CachedFactory[T any] struct {
	computer func() []T
	cached   []T
	once     sync.Once
}

func NewCachedFactory[T any](computer func() []T) StreamFactory[T] {
	cf := &CachedFactory[T]{computer: computer}
	
	return func() stream.Stream[T] {
		cf.once.Do(func() {
			cf.cached = cf.computer()
		})
		return stream.FromSlice(cf.cached)
	}
}

// ============================================================================
// DUAL FORMAT FACTORY FOR CPU/GPU FLEXIBILITY
// ============================================================================

type DualFormatFactory[T any, N int] struct {
	dataSource func() []T
	cached     []T
	once       sync.Once
}

func NewDualFormatFactory[T any, N int](dataSource func() []T) *DualFormatFactory[T, N] {
	return &DualFormatFactory[T, N]{dataSource: dataSource}
}

func (df *DualFormatFactory[T, N]) Stream() stream.Stream[T] {
	df.once.Do(func() {
		df.cached = df.dataSource()
	})
	return stream.FromSlice(df.cached)
}

func (df *DualFormatFactory[T, N]) Array() ([N]T, int) {
	df.once.Do(func() {
		df.cached = df.dataSource()
	})
	
	var arr [N]T
	actualSize := min(len(df.cached), N)
	copy(arr[:actualSize], df.cached)
	return arr, actualSize
}

// ============================================================================
// RECORD HELPER FUNCTIONS
// ============================================================================

// GetStreamFactory safely extracts and calls a stream factory
func GetStreamFactory[T any](record stream.Record, key string) (stream.Stream[T], bool) {
	if val, exists := record[key]; exists {
		if factory, ok := val.(StreamFactory[T]); ok {
			return factory(), true
		}
	}
	return nil, false
}

// GetStreamOr returns stream from factory or empty stream
func GetStreamOr[T any](record stream.Record, key string) stream.Stream[T] {
	if s, exists := GetStreamFactory[T](record, key); exists {
		return s
	}
	return stream.Empty[T]()
}

// GetFactory extracts the factory without calling it
func GetFactory[T any](record stream.Record, key string) (StreamFactory[T], bool) {
	if val, exists := record[key]; exists {
		if factory, ok := val.(StreamFactory[T]); ok {
			return factory, true
		}
	}
	return nil, false
}

// ============================================================================
// FLUENT RECORD BUILDER
// ============================================================================

type RecordBuilder struct {
	data stream.Record
}

func NewRecord() *RecordBuilder {
	return &RecordBuilder{data: make(stream.Record)}
}

func (rb *RecordBuilder) Int(key string, value int64) *RecordBuilder {
	rb.data[key] = value
	return rb
}

func (rb *RecordBuilder) Float(key string, value float64) *RecordBuilder {
	rb.data[key] = value
	return rb
}

func (rb *RecordBuilder) String(key string, value string) *RecordBuilder {
	rb.data[key] = value
	return rb
}

func (rb *RecordBuilder) StreamFactory[T any](key string, factory StreamFactory[T]) *RecordBuilder {
	rb.data[key] = factory
	return rb
}

func (rb *RecordBuilder) SliceFactory[T any](key string, data []T) *RecordBuilder {
	rb.data[key] = NewSliceFactory(data)
	return rb
}

func (rb *RecordBuilder) CachedFactory[T any](key string, computer func() []T) *RecordBuilder {
	rb.data[key] = NewCachedFactory(computer)
	return rb
}

func (rb *RecordBuilder) Build() stream.Record {
	return rb.data
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Simulate GPU array processing
func gpuMapArray(fn func(float64) float64, arr [256]float64) [256]float64 {
	var result [256]float64
	for i, val := range arr {
		result[i] = fn(val)
	}
	return result
}

// Convert slice to fixed array
func sliceTo256Array(data []float64) [256]float64 {
	var arr [256]float64
	copy(arr[:], data)
	return arr
}

// Sum array elements
func sumArray256(arr [256]float64) float64 {
	sum := 0.0
	for _, val := range arr {
		sum += val
	}
	return sum
}

// ============================================================================
// EXAMPLE USAGE
// ============================================================================

func main() {
	fmt.Println("Stream Factory Prototype Demo")
	fmt.Println("=" * 40)
	
	basicUsageExample()
	fmt.Println()
	
	copyabilityExample()
	fmt.Println()
	
	multipleConsumerExample()
	fmt.Println()
	
	cachedFactoryExample()
	fmt.Println()
	
	dualFormatExample()
	fmt.Println()
	
	groupByExample()
}

// Basic stream factory usage
func basicUsageExample() {
	fmt.Println("1. Basic Stream Factory Usage")
	fmt.Println("-" * 30)
	
	// Create record with stream factory
	record := NewRecord().
		Int("id", 1).
		String("name", "Alice").
		SliceFactory("scores", []float64{85.5, 90.0, 87.5, 92.0}).
		SliceFactory("grades", []string{"A", "A-", "B+", "A"}).
		Build()
	
	// Use stream factory multiple times
	if scoreStream, exists := GetStreamFactory[float64](record, "scores"); exists {
		avg, _ := stream.Avg(scoreStream)
		fmt.Printf("Average score: %.2f\n", avg)
	}
	
	if gradeStream, exists := GetStreamFactory[string](record, "grades"); exists {
		count, _ := stream.Count(gradeStream)
		fmt.Printf("Number of grades: %d\n", count)
	}
}

// Demonstrate record copyability
func copyabilityExample() {
	fmt.Println("2. Record Copyability")
	fmt.Println("-" * 20)
	
	original := NewRecord().
		String("name", "Original").
		SliceFactory("data", []int64{1, 2, 3, 4, 5}).
		Build()
	
	// Safe copy
	copy := original
	copy["name"] = "Copy" // Modify copy
	
	fmt.Printf("Original name: %v\n", original["name"])
	fmt.Printf("Copy name: %v\n", copy["name"])
	
	// Both have independent access to the same factory
	if origStream, exists := GetStreamFactory[int64](original, "data"); exists {
		sum, _ := stream.Sum(origStream)
		fmt.Printf("Original sum: %d\n", sum)
	}
	
	if copyStream, exists := GetStreamFactory[int64](copy, "data"); exists {
		max, _ := stream.Max(copyStream)
		fmt.Printf("Copy max: %d\n", max)
	}
}

// Multiple consumers from same factory
func multipleConsumerExample() {
	fmt.Println("3. Multiple Stream Consumers")
	fmt.Println("-" * 29)
	
	record := NewRecord().
		SliceFactory("numbers", []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
		Build()
	
	// Consumer 1: Calculate sum
	if stream1, exists := GetStreamFactory[int64](record, "numbers"); exists {
		sum, _ := stream.Sum(stream1)
		fmt.Printf("Sum: %d\n", sum)
	}
	
	// Consumer 2: Calculate average (independent stream)
	if stream2, exists := GetStreamFactory[int64](record, "numbers"); exists {
		avg, _ := stream.Avg(stream2)
		fmt.Printf("Average: %.2f\n", avg)
	}
	
	// Consumer 3: Find maximum (another independent stream)
	if stream3, exists := GetStreamFactory[int64](record, "numbers"); exists {
		max, _ := stream.Max(stream3)
		fmt.Printf("Maximum: %d\n", max)
	}
}

// Cached factory for expensive operations
func cachedFactoryExample() {
	fmt.Println("4. Cached Factory Example")
	fmt.Println("-" * 25)
	
	// Create expensive computation factory
	expensiveFactory := NewCachedFactory(func() []float64 {
		fmt.Println("   [Computing expensive data...]")
		time.Sleep(100 * time.Millisecond) // Simulate expensive operation
		
		data := make([]float64, 1000)
		for i := range data {
			data[i] = rand.Float64() * 100
		}
		return data
	})
	
	record := NewRecord().
		CachedFactory("expensive_data", expensiveFactory).
		Build()
	
	// First access - triggers computation
	if stream1, exists := GetStreamFactory[float64](record, "expensive_data"); exists {
		count, _ := stream.Count(stream1)
		fmt.Printf("First access - count: %d\n", count)
	}
	
	// Second access - uses cached data
	if stream2, exists := GetStreamFactory[float64](record, "expensive_data"); exists {
		avg, _ := stream.Avg(stream2)
		fmt.Printf("Second access - average: %.4f\n", avg)
	}
}

// Dual format factory for CPU/GPU processing
func dualFormatExample() {
	fmt.Println("5. Dual Format CPU/GPU Example")
	fmt.Println("-" * 33)
	
	// Create dual format factory
	dualFactory := NewDualFormatFactory[float64, 256](func() []float64 {
		data := make([]float64, 1000)
		for i := range data {
			data[i] = rand.Float64() * math.Pi
		}
		return data
	})
	
	// CPU processing - use as stream
	cpuStream := dualFactory.Stream()
	cpuProcessed := stream.Map(func(x float64) float64 { 
		return math.Sin(x) 
	})(cpuStream)
	
	cpuResult, _ := stream.Take[float64](10)(cpuProcessed)
	cpuValues, _ := stream.Collect(cpuResult)
	fmt.Printf("CPU processing (first 10): %v\n", cpuValues[:5]) // Show first 5
	
	// GPU processing - use as array
	gpuArray, size := dualFactory.Array()
	gpuResult := gpuMapArray(func(x float64) float64 { 
		return math.Sin(x) 
	}, gpuArray)
	
	gpuAvg := sumArray256(gpuResult) / float64(size)
	fmt.Printf("GPU processing average: %.6f\n", gpuAvg)
}

// GroupBy with stream factories
func groupByExample() {
	fmt.Println("6. GroupBy with Stream Factories")
	fmt.Println("-" * 32)
	
	// Create records with stream factories
	records := []stream.Record{
		NewRecord().
			String("department", "Engineering").
			SliceFactory("salaries", []int64{75000, 85000, 95000, 105000}).
			Build(),
		
		NewRecord().
			String("department", "Sales").
			SliceFactory("salaries", []int64{65000, 70000, 80000, 90000}).
			Build(),
		
		NewRecord().
			String("department", "Marketing").
			SliceFactory("salaries", []int64{60000, 65000, 75000, 85000}).
			Build(),
	}
	
	// GroupBy safely copies records with factories
	grouped := stream.GroupBy([]string{"department"})(stream.FromSlice(records))
	
	// Process each group
	groupResults, _ := stream.Collect(grouped)
	for _, group := range groupResults {
		dept := stream.GetOr(group, "department", "Unknown")
		
		// Each group processes salaries independently
		if salaryStream, exists := GetStreamFactory[int64](group, "salaries"); exists {
			avgSalary, _ := stream.Avg(salaryStream)
			fmt.Printf("%s avg salary: $%.0f\n", dept, avgSalary)
		}
	}
}

func (s string) mul(count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}