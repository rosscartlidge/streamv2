package stream

import (
	"reflect"
	"testing"
)

func TestExecutorArchitecture(t *testing.T) {
	// Test executor manager initialization
	t.Run("ExecutorManager", func(t *testing.T) {
		em := NewExecutorManager()
		
		// Should have at least CPU executor
		if len(em.executors) < 1 {
			t.Error("Expected at least one executor (CPU)")
		}
		
		// Should have CPU fallback
		if em.fallback == nil {
			t.Error("Expected CPU fallback executor")
		}
		
		if em.fallback.Name() != "CPU" {
			t.Errorf("Expected CPU fallback, got %s", em.fallback.Name())
		}
	})
	
	// Test operation metadata creation
	t.Run("OperationMetadata", func(t *testing.T) {
		op := NewMapOperation(reflect.TypeOf(float64(0)), 1000, 5)
		
		if op.Type != OpMap {
			t.Errorf("Expected OpMap, got %v", op.Type)
		}
		
		if op.InputSize != 1000 {
			t.Errorf("Expected size 1000, got %d", op.InputSize)
		}
		
		if op.Complexity != 5 {
			t.Errorf("Expected complexity 5, got %d", op.Complexity)
		}
		
		if !op.IsVectorizable {
			t.Error("Expected float64 to be vectorizable")
		}
	})
	
	// Test CPU executor scoring
	t.Run("CPUExecutorScoring", func(t *testing.T) {
		cpu := NewCPUExecutor()
		
		// CPU should handle any operation
		op := NewMapOperation(reflect.TypeOf(int(0)), 100, 3)
		ctx := ExecutionContext{MaxGoroutines: 4}
		
		if !cpu.CanHandle(op, ctx) {
			t.Error("CPU should be able to handle any operation")
		}
		
		score := cpu.GetScore(op, ctx)
		if score <= 0 {
			t.Errorf("Expected positive score, got %d", score)
		}
	})
	
	// Test GPU executor (should be nil since no CUDA hardware)
	t.Run("GPUExecutorUnavailable", func(t *testing.T) {
		gpu := NewGPUExecutor()
		
		// Should be nil since no CUDA hardware available
		if gpu != nil {
			t.Error("Expected GPU executor to be nil when no CUDA hardware")
		}
	})
	
	// Test executor selection
	t.Run("ExecutorSelection", func(t *testing.T) {
		em := NewExecutorManager()
		
		// Small, non-vectorizable operation should prefer CPU
		op := NewMapOperation(reflect.TypeOf(""), 10, 8) // String operation, complex
		ctx := ExecutionContext{MaxGoroutines: 4}
		
		executor := em.SelectBest(op, ctx)
		if executor.Name() != "CPU" {
			t.Errorf("Expected CPU for small string operation, got %s", executor.Name())
		}
		
		// Large, vectorizable operation would prefer GPU (but falls back to CPU)
		op2 := NewMapOperation(reflect.TypeOf(float64(0)), 100000, 2) // Large float operation
		executor2 := em.SelectBest(op2, ctx)
		
		// Should still be CPU since no GPU available
		if executor2.Name() != "CPU" {
			t.Errorf("Expected CPU fallback, got %s", executor2.Name())
		}
	})
	
	// Test function complexity estimation
	t.Run("ComplexityEstimation", func(t *testing.T) {
		complexity := estimateFunctionComplexity(func(x int) int { return x * 2 })
		
		if complexity <= 0 || complexity > 10 {
			t.Errorf("Expected reasonable complexity (1-10), got %d", complexity)
		}
	})
}

// TestTransparentExecution verifies that the executor architecture produces correct results
func TestTransparentExecution(t *testing.T) {
	// Simple Map should work correctly (may use sequential or parallel based on complexity)
	data := []int{1, 2, 3, 4, 5}
	result, err := Collect(
		Map(func(x int) int { return x * x })(
			FromSlice(data)))
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if len(result) != len(data) {
		t.Errorf("Expected %d results, got %d", len(data), len(result))
	}
	
	// For parallel processing, order may not be preserved, so check content instead
	expected := map[int]bool{1: true, 4: true, 9: true, 16: true, 25: true}
	for _, v := range result {
		if !expected[v] {
			t.Errorf("Unexpected result value: %d", v)
		}
	}
}

// BenchmarkExecutorSelection measures overhead of executor selection
func BenchmarkExecutorSelection(b *testing.B) {
	em := NewExecutorManager()
	op := NewMapOperation(reflect.TypeOf(float64(0)), 1000, 3)
	ctx := ExecutionContext{MaxGoroutines: 4}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = em.SelectBest(op, ctx)
	}
}