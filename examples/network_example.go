package main

import (
	"fmt"
	"log"
	"time"

	"github.com/rosscartlidge/streamv2/pkg/stream"
)

// ============================================================================
// NETWORK ANALYTICS EXAMPLE WITH STREAMV2
// ============================================================================

func main() {
	fmt.Println("🌐 StreamV2 Network Analytics Example")
	fmt.Println("=====================================")

	// Initialize processor with automatic backend selection
	processor := stream.NewProcessor(
		stream.WithGPUPreferred(), // Try GPU first, fallback to CPU
		stream.WithBatchSize(50000),
	)
	defer processor.Close()

	fmt.Printf("✅ Initialized %s processor\n", processor.Capabilities().Backend)

	demonstrateNetworkFlows()
	demonstrateNetworkAnalytics(processor)
	demonstrateStreamProcessing()
}

// ============================================================================
// NETWORK FLOW PROCESSING
// ============================================================================

func demonstrateNetworkFlows() {
	fmt.Println("\n📡 Network Flow Processing")
	fmt.Println("--------------------------")

	// Generate sample network flows
	flows := stream.GenerateTestFlows(1000)
	fmt.Printf("Generated %d network flows\n", len(flows))

	// Show a few sample flows
	fmt.Println("\nSample flows:")
	for i := 0; i < 3 && i < len(flows); i++ {
		fmt.Printf("%d: %s\n", i+1, flows[i].String())
	}

	// Convert flows to stream
	flowStream := stream.FromNetFlows(flows)

	// Filter TCP flows on port 80
	httpFlows := stream.Chain(
		stream.FilterByProtocol(6), // TCP
		stream.FilterByPort(80),    // HTTP port
	)(flowStream)

	httpFlowList, _ := stream.Collect(httpFlows)
	fmt.Printf("\nTCP flows on port 80: %d\n", len(httpFlowList))

	// Convert flows to records for generic processing
	flowRecords := stream.NetFlowToRecords()(stream.FromNetFlows(flows))
	recordList, _ := stream.Collect(flowRecords)
	fmt.Printf("Flows as records: %d\n", len(recordList))

	if len(recordList) > 0 {
		fmt.Printf("Sample record: %v\n", recordList[0])
	}
}

// ============================================================================
// HIGH-PERFORMANCE NETWORK ANALYTICS
// ============================================================================

func demonstrateNetworkAnalytics(processor stream.Processor) {
	fmt.Println("\n🔍 Network Analytics")
	fmt.Println("--------------------")

	// Generate realistic network traffic
	flows := stream.GenerateTestFlows(50000)
	fmt.Printf("Analyzing %d network flows\n", len(flows))

	start := time.Now()

	// High-performance parallel analytics
	topTalkers := processor.TopTalkers(10)(flows)
	portScans := processor.PortScanDetection(time.Minute*5, 50)(flows)
	ddosAlerts := processor.DDoSDetection(1000.0, 3.0)(flows)

	elapsed := time.Since(start)

	fmt.Printf("\n⚡ Analysis completed in %v\n", elapsed)
	fmt.Printf("• Top Talkers: %d identified\n", len(topTalkers))
	fmt.Printf("• Port Scans: %d detected\n", len(portScans))
	fmt.Printf("• DDoS Alerts: %d generated\n", len(ddosAlerts))

	// Show top talkers
	if len(topTalkers) > 0 {
		fmt.Println("\n🔥 Top Traffic Sources:")
		for i, talker := range topTalkers {
			if i >= 5 {
				break
			} // Show top 5
			fmt.Printf("%d. %s\n", i+1, talker.String())
		}
	}

	// Show performance metrics
	metrics := processor.Metrics()
	fmt.Printf("\n📊 Performance Metrics:\n")
	fmt.Printf("• Records Processed: %d\n", metrics.RecordsProcessed)
	fmt.Printf("• Backend Mode: %s\n", metrics.BackendUsage.Mode)

	capabilities := processor.Capabilities()
	fmt.Printf("• CPU Cores: %d\n", capabilities.CPUCores)
	fmt.Printf("• Max Batch Size: %d\n", capabilities.MaxBatchSize)
}

// ============================================================================
// STREAM-BASED NETWORK PROCESSING
// ============================================================================

func demonstrateStreamProcessing() {
	fmt.Println("\n🔄 Stream-Based Network Processing")
	fmt.Println("----------------------------------")

	// Create a stream of network flows
	flows := stream.GenerateTestFlows(10000)
	flowStream := stream.FromNetFlows(flows)

	// Process flows in batches for better performance
	batchProcessor := stream.Batched(1000, func(batch []stream.NetFlow) []stream.Record {
		// Analyze each batch
		var results []stream.Record

		// Count flows by protocol
		protocolCount := make(map[uint8]int)
		var totalBytes uint64

		for _, flow := range batch {
			protocolCount[flow.Protocol]++
			totalBytes += flow.Bytes
		}

		// Create summary record
		summary := stream.R(
			"batch_size", len(batch),
			"total_bytes", totalBytes,
			"tcp_flows", protocolCount[6],
			"udp_flows", protocolCount[17],
			"other_flows", len(batch)-protocolCount[6]-protocolCount[17],
		)

		results = append(results, summary)
		return results
	})

	// Process the stream
	summaryStream := batchProcessor(flowStream)
	summaries, _ := stream.Collect(summaryStream)

	fmt.Printf("Processed %d batches\n", len(summaries))
	if len(summaries) > 0 {
		fmt.Printf("Sample batch summary: %v\n", summaries[0])
	}

	// Demonstrate flow grouping by endpoints
	smallFlowSet := stream.GenerateTestFlows(100)
	groupedFlows := stream.GroupByEndpoints()(stream.FromNetFlows(smallFlowSet))
	groups, _ := stream.Collect(groupedFlows)

	fmt.Printf("\nGrouped %d flows into %d endpoint pairs\n", len(smallFlowSet), len(groups))
	if len(groups) > 0 {
		fmt.Printf("Sample group: %v\n", groups[0])
	}

	// Demonstrate parallel processing
	fmt.Println("\n⚡ Parallel Flow Processing")

	start := time.Now()

	// Process flows in parallel - each worker enriches flow data
	enrichedFlows := stream.Parallel(4, func(flow stream.NetFlow) stream.NetFlow {
		// Simulate expensive processing (e.g., GeoIP lookup)
		time.Sleep(time.Microsecond * 10)

		// Add computed field (simplified)
		flow.ASN = uint16((flow.SrcIP >> 16) & 0xFFFF) // Fake ASN based on IP
		return flow
	})(stream.FromNetFlows(stream.GenerateTestFlows(1000)))

	enrichedList, _ := stream.Collect(enrichedFlows)
	elapsed := time.Since(start)

	fmt.Printf("Enriched %d flows in %v (parallel processing)\n", len(enrichedList), elapsed)
	if len(enrichedList) > 0 {
		fmt.Printf("Sample enriched flow: %s (ASN: %d)\n",
			enrichedList[0].String(), enrichedList[0].ASN)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
