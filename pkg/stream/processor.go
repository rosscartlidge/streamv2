package stream

import (
	"runtime"
	"time"
)

// ============================================================================
// PROCESSOR INTERFACE - UNIFIED CPU/GPU ABSTRACTION
// ============================================================================

// Processor provides high-performance analytics with automatic backend selection
type Processor interface {
	// Network analytics
	TopTalkers(topN int) func([]NetFlow) []TopTalker
	PortScanDetection(timeWindow time.Duration, threshold int) func([]NetFlow) []PortScanAlert
	DDoSDetection(baseline, threshold float64) func([]NetFlow) []DDoSAlert
	TrafficMatrix() func([]NetFlow) TrafficMatrix

	// Statistical operations
	StatsInt64(percentiles []float64) func([]int64) StatsSummary
	StatsFloat64(percentiles []float64) func([]float64) StatsSummary
	HistogramInt64(buckets int) func([]int64) []HistogramBucketInt64
	HistogramFloat64(buckets int) func([]float64) []HistogramBucketFloat64

	// System info
	Capabilities() ProcessorInfo
	Metrics() ProcessorMetrics
	Close() error
}

// ProcessorInfo describes processor capabilities
type ProcessorInfo struct {
	Backend      string    // "CPU", "GPU", or "Hybrid"
	CPUCores     int       // Available CPU cores
	GPUDevices   []GPUInfo // Available GPU devices
	MaxBatchSize int       // Optimal batch size
	Memory       int64     // Available memory (bytes)
}

// GPUInfo describes a GPU device
type GPUInfo struct {
	ID          int    // GPU device ID
	Name        string // GPU name
	Memory      int64  // GPU memory (bytes)
	Cores       int    // CUDA cores
	ComputeCaps string // Compute capability
}

// ProcessorMetrics tracks performance
type ProcessorMetrics struct {
	RecordsProcessed int64         // Total records processed
	AverageLatency   time.Duration // Average operation latency
	ThroughputRPS    float64       // Records per second
	BackendUsage     BackendUsage  // Backend utilization
}

// BackendUsage tracks CPU/GPU utilization
type BackendUsage struct {
	CPUUsage float64 // CPU utilization (0-100%)
	GPUUsage float64 // GPU utilization (0-100%)
	Mode     string  // "CPU", "GPU", or "Adaptive"
}

// ============================================================================
// PROCESSOR FACTORY - AUTOMATIC BACKEND SELECTION
// ============================================================================

// NewProcessor creates the best processor for the current system
func NewProcessor(opts ...ProcessorOption) Processor {
	config := &ProcessorConfig{
		PreferGPU:     true,
		FallbackToCPU: true,
		Adaptive:      true,
		MaxWorkers:    runtime.NumCPU(),
		BatchSize:     100000,
		MemoryLimit:   1024 * 1024 * 1024, // 1GB default
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Try GPU first if preferred
	if config.PreferGPU {
		if gpu, err := newGPUProcessor(config); err == nil {
			return &adaptiveProcessor{
				gpu:    gpu,
				cpu:    newCPUProcessor(config),
				config: config,
			}
		}
	}

	// Fallback to CPU or fail
	if config.FallbackToCPU {
		return newCPUProcessor(config)
	}

	panic("No suitable processor available")
}

// ProcessorConfig configures processor behavior
type ProcessorConfig struct {
	PreferGPU     bool
	FallbackToCPU bool
	Adaptive      bool
	MaxWorkers    int
	BatchSize     int
	MemoryLimit   int64
	GPUDevice     int
}

// ProcessorOption configures a processor
type ProcessorOption func(*ProcessorConfig)

// WithCPUOnly forces CPU-only processing
func WithCPUOnly() ProcessorOption {
	return func(c *ProcessorConfig) {
		c.PreferGPU = false
		c.FallbackToCPU = true
	}
}

// WithGPUPreferred tries GPU first, falls back to CPU
func WithGPUPreferred() ProcessorOption {
	return func(c *ProcessorConfig) {
		c.PreferGPU = true
		c.FallbackToCPU = true
	}
}

// WithBatchSize sets optimal batch size
func WithBatchSize(size int) ProcessorOption {
	return func(c *ProcessorConfig) {
		c.BatchSize = size
	}
}

// WithWorkers sets number of CPU workers
func WithWorkers(workers int) ProcessorOption {
	return func(c *ProcessorConfig) {
		c.MaxWorkers = workers
	}
}

// WithMemoryLimit sets memory usage limit
func WithMemoryLimit(bytes int64) ProcessorOption {
	return func(c *ProcessorConfig) {
		c.MemoryLimit = bytes
	}
}

// ============================================================================
// STATISTICAL TYPES
// ============================================================================

// StatsSummary contains statistical summary
type StatsSummary struct {
	Count       int64               // Number of values
	Sum         float64             // Sum of values
	Mean        float64             // Average value
	Min         float64             // Minimum value
	Max         float64             // Maximum value
	Variance    float64             // Variance
	StdDev      float64             // Standard deviation
	Percentiles map[float64]float64 // Requested percentiles
}

// HistogramBucketInt64 represents a histogram bucket for int64 values
type HistogramBucketInt64 struct {
	Min   int64 // Bucket minimum (inclusive)
	Max   int64 // Bucket maximum (exclusive)
	Count int64 // Number of values in bucket
}

// HistogramBucketFloat64 represents a histogram bucket for float64 values
type HistogramBucketFloat64 struct {
	Min   float64 // Bucket minimum (inclusive)
	Max   float64 // Bucket maximum (exclusive)
	Count int64   // Number of values in bucket
}

// ============================================================================
// BATCHED PROCESSING - EFFICIENT BULK OPERATIONS
// ============================================================================

// Batched processes stream elements in batches for better performance
func Batched[T, U any](batchSize int, processor func([]T) []U) Filter[T, U] {
	return func(input Stream[T]) Stream[U] {
		var batch []T
		var results []U
		var resultIndex int

		return func() (U, error) {
			// Return cached results first
			if resultIndex < len(results) {
				result := results[resultIndex]
				resultIndex++
				return result, nil
			}

			// Collect new batch
			batch = batch[:0] // Reset slice but keep capacity
			for len(batch) < batchSize {
				item, err := input()
				if err != nil {
					if err == EOS && len(batch) > 0 {
						break // Process partial batch
					}
					var zero U
					return zero, err
				}
				batch = append(batch, item)
			}

			if len(batch) == 0 {
				var zero U
				return zero, EOS
			}

			// Process batch
			results = processor(batch)
			resultIndex = 0

			// Return first result
			if len(results) > 0 {
				result := results[0]
				resultIndex = 1
				return result, nil
			}

			// Empty results, try again
			var zero U
			return zero, EOS
		}
	}
}

// ============================================================================
// PLACEHOLDER IMPLEMENTATIONS (TO BE REPLACED WITH REAL IMPLEMENTATIONS)
// ============================================================================

// Placeholder processor types
type adaptiveProcessor struct {
	cpu    *cpuProcessor
	gpu    *gpuProcessor
	config *ProcessorConfig
}

type cpuProcessor struct {
	config *ProcessorConfig
}

type gpuProcessor struct {
	config *ProcessorConfig
}

// Factory functions
func newCPUProcessor(config *ProcessorConfig) *cpuProcessor {
	return &cpuProcessor{config: config}
}

func newGPUProcessor(config *ProcessorConfig) (*gpuProcessor, error) {
	// Return error for now - GPU implementation would go here
	return nil, EOS // Using EOS as generic error for now
}

// Placeholder implementations - these would be replaced with real algorithms
func (ap *adaptiveProcessor) TopTalkers(topN int) func([]NetFlow) []TopTalker {
	return ap.cpu.TopTalkers(topN)
}

func (ap *adaptiveProcessor) PortScanDetection(timeWindow time.Duration, threshold int) func([]NetFlow) []PortScanAlert {
	return ap.cpu.PortScanDetection(timeWindow, threshold)
}

func (ap *adaptiveProcessor) DDoSDetection(baseline, threshold float64) func([]NetFlow) []DDoSAlert {
	return ap.cpu.DDoSDetection(baseline, threshold)
}

func (ap *adaptiveProcessor) TrafficMatrix() func([]NetFlow) TrafficMatrix {
	return ap.cpu.TrafficMatrix()
}

func (ap *adaptiveProcessor) StatsInt64(percentiles []float64) func([]int64) StatsSummary {
	return ap.cpu.StatsInt64(percentiles)
}

func (ap *adaptiveProcessor) StatsFloat64(percentiles []float64) func([]float64) StatsSummary {
	return ap.cpu.StatsFloat64(percentiles)
}

func (ap *adaptiveProcessor) HistogramInt64(buckets int) func([]int64) []HistogramBucketInt64 {
	return ap.cpu.HistogramInt64(buckets)
}

func (ap *adaptiveProcessor) HistogramFloat64(buckets int) func([]float64) []HistogramBucketFloat64 {
	return ap.cpu.HistogramFloat64(buckets)
}

func (ap *adaptiveProcessor) Capabilities() ProcessorInfo {
	return ProcessorInfo{
		Backend:      "Hybrid",
		CPUCores:     runtime.NumCPU(),
		MaxBatchSize: ap.config.BatchSize,
		Memory:       ap.config.MemoryLimit,
	}
}

func (ap *adaptiveProcessor) Metrics() ProcessorMetrics {
	return ProcessorMetrics{
		BackendUsage: BackendUsage{Mode: "Adaptive"},
	}
}

func (ap *adaptiveProcessor) Close() error {
	return nil
}

// CPU processor placeholder implementations
func (cpu *cpuProcessor) TopTalkers(topN int) func([]NetFlow) []TopTalker {
	return func(flows []NetFlow) []TopTalker {
		// Placeholder - real implementation would go here
		talkers := make(map[uint32]uint64)
		for _, flow := range flows {
			talkers[flow.SrcIP] += flow.Bytes
			talkers[flow.DstIP] += flow.Bytes
		}

		// Convert to slice and return top N (simplified)
		var result []TopTalker
		for ip, bytes := range talkers {
			result = append(result, TopTalker{IP: ip, Bytes: bytes})
			if len(result) >= topN {
				break
			}
		}
		return result
	}
}

func (cpu *cpuProcessor) PortScanDetection(timeWindow time.Duration, threshold int) func([]NetFlow) []PortScanAlert {
	return func(flows []NetFlow) []PortScanAlert {
		// Placeholder implementation
		return []PortScanAlert{}
	}
}

func (cpu *cpuProcessor) DDoSDetection(baseline, threshold float64) func([]NetFlow) []DDoSAlert {
	return func(flows []NetFlow) []DDoSAlert {
		// Placeholder implementation
		return []DDoSAlert{}
	}
}

func (cpu *cpuProcessor) TrafficMatrix() func([]NetFlow) TrafficMatrix {
	return func(flows []NetFlow) TrafficMatrix {
		// Placeholder implementation
		return TrafficMatrix{}
	}
}

func (cpu *cpuProcessor) StatsInt64(percentiles []float64) func([]int64) StatsSummary {
	return func(values []int64) StatsSummary {
		// Placeholder implementation
		return StatsSummary{Count: int64(len(values))}
	}
}

func (cpu *cpuProcessor) StatsFloat64(percentiles []float64) func([]float64) StatsSummary {
	return func(values []float64) StatsSummary {
		// Placeholder implementation
		return StatsSummary{Count: int64(len(values))}
	}
}

func (cpu *cpuProcessor) HistogramInt64(buckets int) func([]int64) []HistogramBucketInt64 {
	return func(values []int64) []HistogramBucketInt64 {
		// Placeholder implementation
		return make([]HistogramBucketInt64, buckets)
	}
}

func (cpu *cpuProcessor) HistogramFloat64(buckets int) func([]float64) []HistogramBucketFloat64 {
	return func(values []float64) []HistogramBucketFloat64 {
		// Placeholder implementation
		return make([]HistogramBucketFloat64, buckets)
	}
}

func (cpu *cpuProcessor) Capabilities() ProcessorInfo {
	return ProcessorInfo{
		Backend:      "CPU",
		CPUCores:     runtime.NumCPU(),
		MaxBatchSize: cpu.config.BatchSize,
		Memory:       cpu.config.MemoryLimit,
	}
}

func (cpu *cpuProcessor) Metrics() ProcessorMetrics {
	return ProcessorMetrics{
		BackendUsage: BackendUsage{Mode: "CPU"},
	}
}

func (cpu *cpuProcessor) Close() error {
	return nil
}
