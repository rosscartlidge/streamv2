package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// INFINITE STREAM SPLIT EXAMPLE - REAL-TIME STREAM PARTITIONING
// ============================================================================

func main() {
	fmt.Println("♾️  StreamV2 Infinite Split Example")
	fmt.Println("===================================")

	demonstrateRealtimeLogProcessing()
}

// ============================================================================
// REAL-TIME LOG PROCESSING WITH SPLIT
// ============================================================================

func demonstrateRealtimeLogProcessing() {
	fmt.Println("\n📊 Real-time Log Processing by Service")
	fmt.Println("--------------------------------------")

	// Create infinite stream of simulated log entries
	logStream := createInfiniteLogStream()

	// Split by service - emit new substreams as we discover services
	serviceStreams := stream.Split([]string{"service"})(logStream)

	// Process first few service streams for demo
	fmt.Println("Waiting for service streams to appear...")
	
	processedServices := make(map[string]bool)
	startTime := time.Now()
	
	// Process service streams as they appear
	for time.Since(startTime) < 10*time.Second { // Run for 10 seconds
		serviceStream, err := serviceStreams()
		if err == stream.EOS {
			fmt.Println("No more service streams")
			break
		}
		
		// Get first log entry to identify the service
		firstLog, err := serviceStream()
		if err != nil {
			continue
		}
		
		serviceName := stream.GetOr(firstLog, "service", "unknown")
		
		if !processedServices[serviceName] {
			processedServices[serviceName] = true
			fmt.Printf("📡 Discovered service: %s\n", serviceName)
			
			// Start processing this service's logs in background
			go func(svc string, logStream stream.Stream[stream.Record], firstLog stream.Record) {
				processServiceLogs(svc, logStream, firstLog)
			}(serviceName, serviceStream, firstLog)
		}
		
		// Small delay to prevent tight loop
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n✨ All service streams are now processing in background!")
	fmt.Println("In a real system, these would run indefinitely...")
	
	// Give background goroutines time to process some logs
	time.Sleep(3 * time.Second)
}

// ============================================================================
// INFINITE LOG STREAM SIMULATOR
// ============================================================================

func createInfiniteLogStream() stream.Stream[stream.Record] {
	services := []string{"auth", "api", "database", "cache", "frontend", "STOP"}
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	
	return func() (stream.Record, error) {
		// Simulate network delay
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		
		// Generate random log entry
		service := services[rand.Intn(len(services))]
		// When STOP selected "infinite" stream ends for demo purposes.
		if service == "STOP" {
			var zero stream.Record
			return zero, stream.EOS
		}
		level := levels[rand.Intn(len(levels))]
		timestamp := time.Now().Format("15:04:05")
		
		// Simulate different message types
		var message string
		switch level {
		case "ERROR":
			message = fmt.Sprintf("Connection failed to %s", service)
		case "WARN":
			message = fmt.Sprintf("High memory usage in %s", service)
		case "INFO":
			message = fmt.Sprintf("Request processed by %s", service)
		case "DEBUG":
			message = fmt.Sprintf("Debug trace in %s", service)
		}
		
		return stream.NewRecord().
			String("timestamp", timestamp).
			String("service", service).
			String("level", level).
			String("message", message).
			String("request_id", fmt.Sprintf("req-%d", rand.Intn(10000))).
			Build(), nil
	}
}

// ============================================================================
// SERVICE LOG PROCESSOR
// ============================================================================

func processServiceLogs(serviceName string, logStream stream.Stream[stream.Record], firstLog stream.Record) {
	fmt.Printf("🔄 Started processing logs for service: %s\n", serviceName)
	
	logCount := 1 // Count the first log
	errorCount := 0
	warnCount := 0
	
	// Process first log
	if stream.GetOr(firstLog, "level", "") == "ERROR" {
		errorCount++
	} else if stream.GetOr(firstLog, "level", "") == "WARN" {
		warnCount++
	}
	
	// Create a timeout for demo purposes
	timeout := time.After(5 * time.Second)
	
	// Process logs from this service
	for {
		select {
		case <-timeout:
			// Demo timeout - in real system this wouldn't exist
			fmt.Printf("📊 Service %s processed %d logs (%d errors, %d warnings) [DEMO ENDED]\n", 
				serviceName, logCount, errorCount, warnCount)
			return
			
		default:
			// Try to get next log entry
			logEntry, err := logStream()
			if err == stream.EOS {
				fmt.Printf("📊 Service %s stream ended after %d logs\n", serviceName, logCount)
				return
			}
			
			logCount++
			level := stream.GetOr(logEntry, "level", "")
			timestamp := stream.GetOr(logEntry, "timestamp", "")
			message := stream.GetOr(logEntry, "message", "")
			
			// Count by level
			switch level {
			case "ERROR":
				errorCount++
				fmt.Printf("🚨 [%s] %s ERROR: %s\n", serviceName, timestamp, message)
			case "WARN":
				warnCount++
				if rand.Float32() < 0.3 { // Show some warnings
					fmt.Printf("⚠️  [%s] %s WARN: %s\n", serviceName, timestamp, message)
				}
			default:
				if rand.Float32() < 0.1 { // Show some info logs
					fmt.Printf("ℹ️  [%s] %s %s: %s\n", serviceName, timestamp, level, message)
				}
			}
			
			// Report stats periodically
			if logCount%10 == 0 {
				fmt.Printf("📈 [%s] Processed %d logs (%d errors, %d warnings)\n", 
					serviceName, logCount, errorCount, warnCount)
			}
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
