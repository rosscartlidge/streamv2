package main

import (
	"fmt"
	"os"
	"time"
	"math/rand"
	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Println("🌊 Infinite Stream + CSV Integration")
	fmt.Println("====================================")

	// Test 1: CSV to Infinite Processing Pipeline
	fmt.Println("\n📊 Test 1: CSV Input → Infinite Processing")
	fmt.Println("------------------------------------------")
	
	// Create initial sales data
	initialSales := []stream.Record{
		stream.NewRecord().Int("id", 1).Float("amount", 150.0).String("region", "US").String("product", "laptop").Build(),
		stream.NewRecord().Int("id", 2).Float("amount", 200.0).String("region", "EU").String("product", "mouse").Build(),
		stream.NewRecord().Int("id", 3).Float("amount", 175.0).String("region", "US").String("product", "keyboard").Build(),
		stream.NewRecord().Int("id", 4).Float("amount", 300.0).String("region", "ASIA").String("product", "laptop").Build(),
		stream.NewRecord().Int("id", 5).Float("amount", 125.0).String("region", "EU").String("product", "mouse").Build(),
		stream.NewRecord().Int("id", 6).Float("amount", 250.0).String("region", "ASIA").String("product", "keyboard").Build(),
	}
	
	// Write to CSV
	salesFile := "/tmp/initial_sales.csv"
	err := stream.StreamToCSVFile(stream.FromSlice(initialSales), salesFile)
	if err != nil {
		fmt.Printf("Error writing sales CSV: %v\n", err)
		return
	}
	
	// Read CSV and create processing pipeline
	salesStream, err := stream.CSVToStreamFromFile(salesFile)
	if err != nil {
		fmt.Printf("Error reading sales CSV: %v\n", err)
		return
	}
	
	fmt.Println("CSV Data Loaded:")
	csvRecords, _ := stream.Collect(salesStream)
	for _, record := range csvRecords {
		id := stream.GetOr(record, "id", int64(0))
		amount := stream.GetOr(record, "amount", 0.0)
		region := stream.GetOr(record, "region", "")
		product := stream.GetOr(record, "product", "")
		fmt.Printf("  ID=%d: $%.0f %s in %s\n", id, amount, product, region)
	}
	
	// Process with windowing and streaming aggregators
	fmt.Println("\nProcessing with CountWindow(3) + StreamingGroupBy:")
	
	// Reload stream for processing
	salesStream2, _ := stream.CSVToStreamFromFile(salesFile)
	
	// Window processing
	windows := stream.CountWindow[stream.Record](3)(salesStream2)
	
	windowNum := 0
	for {
		window, err := windows()
		if err == stream.EOS {
			break
		}
		if err != nil {
			break
		}
		
		windowNum++
		
		// Group by region within window
		groupedResults, _ := stream.Collect(
			stream.GroupBy([]string{"region"}, 
				stream.SumField[float64]("total_amount", "amount"),
				stream.CountField("count", "region"),
			)(window),
		)
		
		fmt.Printf("  Window %d results:\n", windowNum)
		for _, result := range groupedResults {
			region := stream.GetOr(result, "region", "")
			total := stream.GetOr(result, "total_amount", 0.0)
			count := stream.GetOr(result, "count", int64(0))
			fmt.Printf("    %s: %d sales, total=$%.0f\n", region, count, total)
		}
	}

	// Test 2: Infinite Stream → Streaming CSV Output
	fmt.Println("\n🔄 Test 2: Infinite Stream → Streaming CSV Output")
	fmt.Println("--------------------------------------------------")
	
	// Create infinite order stream
	counter := 0
	infiniteOrderStream := func() (stream.Record, error) {
		time.Sleep(200 * time.Millisecond) // Simulate real-time orders
		counter++
		
		products := []string{"laptop", "mouse", "keyboard", "monitor"}
		regions := []string{"US", "EU", "ASIA"}
		
		return stream.NewRecord().
			Int("order_id", int64(counter)).
			Int("timestamp", time.Now().Unix()).
			String("product", products[rand.Intn(len(products))]).
			String("region", regions[rand.Intn(len(regions))]).
			Int("amount", int64(rand.Intn(400) + 100)). // $100-$500
			Int("customer_id", int64(1000 + rand.Intn(100))).
			Build(), nil
	}
	
	// Set up streaming CSV output
	streamingOrdersFile, err := os.Create("/tmp/streaming_orders.csv")
	if err != nil {
		fmt.Printf("Error creating streaming file: %v\n", err)
		return
	}
	defer streamingOrdersFile.Close()
	
	headers := []string{"order_id", "timestamp", "product", "region", "amount", "customer_id"}
	csvWriter := stream.NewStreamingCSVWriter(streamingOrdersFile, headers)
	
	fmt.Println("📝 Writing infinite order stream to CSV (10 orders):")
	
	// Process orders with streaming aggregation + CSV output
	orderTee := stream.Tee(infiniteOrderStream, 2)
	
	// Stream 1: Running totals
	runningTotal := stream.StreamingSum[int]()(
		stream.ExtractField[int]("amount")(orderTee[0]))
	
	// Stream 2: CSV output
	csvOutputStream := orderTee[1]
	
	// Process 10 orders
	for i := 0; i < 10; i++ {
		// Get running total
		total, err := runningTotal()
		if err != nil {
			break
		}
		
		// Get order for CSV
		order, err := csvOutputStream()
		if err != nil {
			break
		}
		
		// Write to CSV
		err = csvWriter.WriteRecord(order)
		if err != nil {
			fmt.Printf("Error writing to CSV: %v\n", err)
			break
		}
		
		orderId := stream.GetOr(order, "order_id", 0)
		amount := stream.GetOr(order, "amount", 0)
		product := stream.GetOr(order, "product", "")
		
		fmt.Printf("  Order %d: %s ($%d) - Running total: $%d\n", 
			orderId, product, amount, total)
	}
	
	csvWriter.Close()
	fmt.Println("✅ Streaming CSV output complete!")

	// Test 3: CSV-to-CSV Processing Pipeline
	fmt.Println("\n🔗 Test 3: CSV-to-CSV Processing Pipeline")
	fmt.Println("-----------------------------------------")
	
	// Read the streaming orders and process them
	streamingOrdersStream, err := stream.CSVToStreamFromFile("/tmp/streaming_orders.csv")
	if err != nil {
		fmt.Printf("Error reading streaming orders: %v\n", err)
		return
	}
	
	// Complex processing pipeline
	processedOrdersStream := stream.Map(func(order stream.Record) stream.Record {
		amount := stream.GetOr(order, "amount", 0)
		region := stream.GetOr(order, "region", "")
		product := stream.GetOr(order, "product", "")
		
		// Calculate fees and taxes based on region
		var tax, shipping float64
		switch region {
		case "US":
			tax = float64(amount) * 0.08  // 8% tax
			shipping = 10.0
		case "EU":
			tax = float64(amount) * 0.20  // 20% VAT
			shipping = 15.0
		case "ASIA":
			tax = float64(amount) * 0.05  // 5% tax
			shipping = 25.0
		}
		
		total := float64(amount) + tax + shipping
		
		// Determine order priority
		var priority string
		if amount >= 300 {
			priority = "high"
		} else if amount >= 200 {
			priority = "medium"
		} else {
			priority = "low"
		}
		
		return stream.NewRecord().
			Int("order_id", stream.GetOr(order, "order_id", int64(0))).
			String("product", product).
			String("region", region).
			Int("amount", int64(amount)).
			Float("tax", tax).
			Float("shipping", shipping).
			Float("total", total).
			String("priority", priority).
			String("processed_at", time.Now().Format("2006-01-02 15:04:05")).
			Build()
	})(streamingOrdersStream)
	
	// Write processed orders to new CSV
	processedFile := "/tmp/processed_orders.csv"
	err = stream.StreamToCSVFile(processedOrdersStream, processedFile)
	if err != nil {
		fmt.Printf("Error writing processed orders: %v\n", err)
		return
	}
	
	fmt.Println("📋 Processed Orders Summary:")
	
	// Read and analyze processed orders
	processedStream, _ := stream.CSVToStreamFromFile(processedFile)
	processedRecords, _ := stream.Collect(processedStream)
	
	for i, record := range processedRecords {
		orderId := stream.GetOr(record, "order_id", int64(0))
		total := stream.GetOr(record, "total", 0.0)
		priority := stream.GetOr(record, "priority", "")
		region := stream.GetOr(record, "region", "")
		
		fmt.Printf("  %d: Order #%d (%s) - Total=$%.2f (%s priority)\n", 
			i+1, orderId, region, total, priority)
	}

	// Test 4: Windowed Analytics with CSV Output
	fmt.Println("\n📊 Test 4: Windowed Analytics with CSV Output")
	fmt.Println("---------------------------------------------")
	
	// Use processed orders for windowed analytics
	processedStream2, _ := stream.CSVToStreamFromFile(processedFile)
	
	// Window by 3 orders and calculate analytics
	windows2 := stream.CountWindow[stream.Record](3)(processedStream2)
	
	analyticsResults := []stream.Record{}
	windowNum2 := 0
	
	for {
		window, err := windows2()
		if err == stream.EOS {
			break
		}
		if err != nil {
			break
		}
		
		windowNum2++
		
		// Calculate window analytics
		windowRecords, _ := stream.Collect(window)
		
		var totalAmount, totalTax, totalShipping float64
		regionCount := make(map[string]int)
		priorityCount := make(map[string]int)
		
		for _, record := range windowRecords {
			totalAmount += stream.GetOr(record, "amount", 0.0)
			totalTax += stream.GetOr(record, "tax", 0.0)
			totalShipping += stream.GetOr(record, "shipping", 0.0)
			
			region := stream.GetOr(record, "region", "")
			priority := stream.GetOr(record, "priority", "")
			
			regionCount[region]++
			priorityCount[priority]++
		}
		
		// Find dominant region and priority
		var dominantRegion, dominantPriority string
		var maxRegionCount, maxPriorityCount int
		
		for region, count := range regionCount {
			if count > maxRegionCount {
				maxRegionCount = count
				dominantRegion = region
			}
		}
		
		for priority, count := range priorityCount {
			if count > maxPriorityCount {
				maxPriorityCount = count
				dominantPriority = priority
			}
		}
		
		analyticsResult := stream.NewRecord().
			Int("window_id", int64(windowNum2)).
			Int("order_count", int64(len(windowRecords))).
			Float("total_amount", totalAmount).
			Float("total_tax", totalTax).
			Float("total_shipping", totalShipping).
			Float("avg_order_value", totalAmount/float64(len(windowRecords))).
			String("dominant_region", dominantRegion).
			String("dominant_priority", dominantPriority).
			Int("analysis_timestamp", time.Now().Unix()).
			Build()
		
		analyticsResults = append(analyticsResults, analyticsResult)
		
		fmt.Printf("  Window %d: %d orders, avg=$%.2f, dominant=%s/%s\n", 
			windowNum2, len(windowRecords), totalAmount/float64(len(windowRecords)), 
			dominantRegion, dominantPriority)
	}
	
	// Write analytics to CSV
	analyticsFile := "/tmp/order_analytics.csv"
	err = stream.StreamToCSVFile(stream.FromSlice(analyticsResults), analyticsFile)
	if err != nil {
		fmt.Printf("Error writing analytics: %v\n", err)
		return
	}
	
	fmt.Printf("📈 Analytics saved to: %s\n", analyticsFile)

	// Test 5: Real-time Dashboard Simulation
	fmt.Println("\n📺 Test 5: Real-time Dashboard Simulation")
	fmt.Println("-----------------------------------------")
	
	fmt.Println("Simulating real-time dashboard with CSV logging...")
	
	// Create dashboard log file
	dashboardLogFile, err := os.Create("/tmp/dashboard_log.csv")
	if err != nil {
		fmt.Printf("Error creating dashboard log: %v\n", err)
		return
	}
	defer dashboardLogFile.Close()
	
	dashboardHeaders := []string{"timestamp", "metric", "value", "trend"}
	dashboardWriter := stream.NewStreamingCSVWriter(dashboardLogFile, dashboardHeaders)
	
	// Simulate real-time metrics
	metrics := []string{"total_sales", "avg_order_value", "orders_per_minute"}
	baseValues := []float64{10000, 250, 15}
	
	fmt.Println("📊 Live Dashboard (5 updates):")
	for i := 0; i < 5; i++ {
		for j, metric := range metrics {
			// Simulate metric changes
			change := (rand.Float64() - 0.5) * 0.2 // ±10% change
			value := baseValues[j] * (1 + change)
			
			var trend string
			if change > 0.05 {
				trend = "up"
			} else if change < -0.05 {
				trend = "down"
			} else {
				trend = "stable"
			}
			
			dashboardRecord := stream.NewRecord().
				Int("timestamp", time.Now().Unix()).
				String("metric", metric).
				Float("value", value).
				String("trend", trend).
				Build()
			
			err = dashboardWriter.WriteRecord(dashboardRecord)
			if err != nil {
				fmt.Printf("Error writing dashboard log: %v\n", err)
				continue
			}
			
			fmt.Printf("  %s: %.2f (%s)\n", metric, value, trend)
		}
		fmt.Println("  ---")
		time.Sleep(500 * time.Millisecond)
	}
	
	dashboardWriter.Close()
	fmt.Println("📈 Dashboard simulation complete!")

	// Cleanup
	fmt.Println("\n🧹 Cleanup")
	fmt.Println("----------")
	
	files := []string{
		salesFile, "/tmp/streaming_orders.csv", processedFile,
		analyticsFile, "/tmp/dashboard_log.csv",
	}
	
	for _, file := range files {
		if err := os.Remove(file); err == nil {
			fmt.Printf("Removed: %s\n", file)
		}
	}
	
	fmt.Println("\n🎉 Infinite Stream + CSV Integration: COMPLETE!")
	fmt.Println("✨ Seamless integration between infinite streams and CSV I/O!")
}

func init() {
	rand.Seed(time.Now().UnixNano())
}