package stream

import (
	"context"
	"reflect"
)

// ============================================================================
// EXECUTOR ARCHITECTURE FOR TRANSPARENT ACCELERATION
// ============================================================================

// OpType represents the type of stream operation
type OpType int

const (
	OpMap OpType = iota
	OpFilter
	OpReduce
	OpGroupBy
	OpWindow
	OpSort
	OpParallel
)

// Operation describes a stream operation with metadata for executor selection
type Operation struct {
	Type         OpType
	DataType     reflect.Type
	InputSize    int64        // Estimated size (-1 if unknown/infinite)
	IsVectorizable bool       // Can benefit from SIMD/GPU vectorization
	MemoryUsage  int64        // Estimated memory usage in bytes
	Complexity   int          // Computational complexity (1=simple, 10=heavy)
}

// ExecutionContext provides runtime information for operation execution
type ExecutionContext struct {
	Ctx           context.Context
	MaxMemory     int64  // Available memory limit
	MaxGoroutines int    // CPU parallelism limit
	GPUMemory     int64  // Available GPU memory (0 if no GPU)
	BatchSize     int    // Preferred batch size for vectorization
}

// Executor interface abstracts different execution backends (CPU, GPU, distributed)
type Executor interface {
	// CanHandle returns true if this executor can efficiently handle the operation
	CanHandle(op Operation, ctx ExecutionContext) bool
	
	// GetScore returns a score (0-100) indicating how well this executor handles the operation
	// Higher scores indicate better suitability
	GetScore(op Operation, ctx ExecutionContext) int
	
	// ExecuteMap performs map operation with this executor
	ExecuteMap(fn interface{}, input interface{}, ctx ExecutionContext) interface{}
	
	// ExecuteFilter performs filter operation with this executor
	ExecuteFilter(predicate interface{}, input interface{}, ctx ExecutionContext) interface{}
	
	// ExecuteReduce performs reduce operation with this executor
	ExecuteReduce(fn interface{}, initial interface{}, input interface{}, ctx ExecutionContext) interface{}
	
	// ExecuteParallel performs parallel processing with this executor
	ExecuteParallel(workers int, fn interface{}, input interface{}, ctx ExecutionContext) interface{}
	
	// Name returns the executor name for debugging/telemetry
	Name() string
}

// ============================================================================
// EXECUTOR MANAGER - SELECTS BEST EXECUTOR FOR EACH OPERATION
// ============================================================================

// ExecutorManager coordinates multiple executors and selects the best one
type ExecutorManager struct {
	executors []Executor
	fallback  Executor // Default CPU executor
}

// NewExecutorManager creates a new executor manager with automatic backend detection
func NewExecutorManager() *ExecutorManager {
	em := &ExecutorManager{
		executors: make([]Executor, 0, 3), // CPU, GPU, potentially distributed
		fallback:  NewCPUExecutor(),       // Always have CPU fallback
	}
	
	// Add CPU executor
	em.AddExecutor(em.fallback)
	
	// Auto-detect and add GPU executor if available
	if gpuExec := NewGPUExecutor(); gpuExec != nil {
		em.AddExecutor(gpuExec)
	}
	
	return em
}

// AddExecutor adds an executor to the manager
func (em *ExecutorManager) AddExecutor(executor Executor) {
	em.executors = append(em.executors, executor)
}

// SelectBest chooses the best executor for the given operation
func (em *ExecutorManager) SelectBest(op Operation, ctx ExecutionContext) Executor {
	var bestExecutor Executor
	bestScore := -1
	
	for _, executor := range em.executors {
		if executor.CanHandle(op, ctx) {
			score := executor.GetScore(op, ctx)
			if score > bestScore {
				bestScore = score
				bestExecutor = executor
			}
		}
	}
	
	// Always fall back to CPU executor if no better option
	if bestExecutor == nil {
		bestExecutor = em.fallback
	}
	
	return bestExecutor
}

// ============================================================================
// GLOBAL EXECUTOR MANAGER - TRANSPARENT TO USERS
// ============================================================================

var globalExecutorManager *ExecutorManager

func init() {
	globalExecutorManager = NewExecutorManager()
}

// getExecutor returns the best executor for the given operation (internal use only)
func getExecutor(op Operation, ctx ExecutionContext) Executor {
	return globalExecutorManager.SelectBest(op, ctx)
}

// ============================================================================
// OPERATION BUILDERS - HELP CREATE OPERATION METADATA
// ============================================================================

// NewMapOperation creates metadata for a Map operation
func NewMapOperation(dataType reflect.Type, size int64, complexity int) Operation {
	return Operation{
		Type:           OpMap,
		DataType:       dataType,
		InputSize:      size,
		IsVectorizable: isVectorizable(dataType),
		MemoryUsage:    estimateMapMemory(dataType, size),
		Complexity:     complexity,
	}
}

// NewFilterOperation creates metadata for a Filter operation  
func NewFilterOperation(dataType reflect.Type, size int64, complexity int) Operation {
	return Operation{
		Type:           OpFilter,
		DataType:       dataType,
		InputSize:      size,
		IsVectorizable: isVectorizable(dataType),
		MemoryUsage:    estimateFilterMemory(dataType, size),
		Complexity:     complexity,
	}
}

// NewParallelOperation creates metadata for a Parallel operation
func NewParallelOperation(dataType reflect.Type, size int64, workers int, complexity int) Operation {
	return Operation{
		Type:           OpParallel,
		DataType:       dataType,
		InputSize:      size,
		IsVectorizable: isVectorizable(dataType),
		MemoryUsage:    estimateParallelMemory(dataType, size, workers),
		Complexity:     complexity,
	}
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// isVectorizable returns true if the data type can benefit from SIMD/GPU vectorization
func isVectorizable(dataType reflect.Type) bool {
	switch dataType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Bool:
		return true
	default:
		return false
	}
}

// Memory estimation functions (simple heuristics for now)
func estimateMapMemory(dataType reflect.Type, size int64) int64 {
	if size < 0 {
		return 1024 * 1024 // 1MB default for unknown size
	}
	return size * int64(dataType.Size()) * 2 // Input + output
}

func estimateFilterMemory(dataType reflect.Type, size int64) int64 {
	if size < 0 {
		return 1024 * 1024
	}
	return size * int64(dataType.Size()) * 2 // Worst case: no filtering
}

func estimateParallelMemory(dataType reflect.Type, size int64, workers int) int64 {
	base := estimateMapMemory(dataType, size)
	return base + int64(workers*1024*1024) // Add buffer memory per worker
}

// estimateFunctionComplexity provides a rough estimate of function computational complexity
// This is a heuristic for executor selection - more sophisticated analysis could be added later
func estimateFunctionComplexity(fn interface{}) int {
	// For now, return a conservative default to avoid over-parallelization
	// In the future, this could analyze the function:
	// - Simple arithmetic: 1-2
	// - Math functions (sin, cos, sqrt): 4-6  
	// - Complex business logic: 7-10
	return 2 // Conservative default - most operations stay sequential
}