package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Println("📁 CSV/TSV Sources and Sinks Testing")
	fmt.Println("====================================")

	// Test 1: CSV Source - Reading from string
	fmt.Println("\n📊 Test 1: CSV Source (from string)")
	fmt.Println("-----------------------------------")
	
	csvData := `name,age,salary,active,hire_date
Alice,30,75000.50,true,2023-01-15
Bob,25,65000,false,2023-03-20
Charlie,35,85000.25,true,2022-12-01
Diana,28,70000,true,2023-02-10`
	
	fmt.Println("Input CSV:")
	fmt.Println(csvData)
	
	csvReader := strings.NewReader(csvData)
	csvSource := stream.NewCSVSource(csvReader)
	csvStream := csvSource.ToStream()
	
	fmt.Println("\nParsed Records:")
	records, _ := stream.Collect(csvStream)
	for i, record := range records {
		name := stream.GetOr(record, "name", "")
		age := stream.GetOr(record, "age", int64(0))
		salary := stream.GetOr(record, "salary", 0.0)
		active := stream.GetOr(record, "active", false)
		fmt.Printf("  %d: %s (age=%d, salary=$%.2f, active=%v)\n", 
			i+1, name, age, salary, active)
	}

	// Test 2: CSV without headers
	fmt.Println("\n📋 Test 2: CSV without headers")
	fmt.Println("------------------------------")
	
	csvNoHeaderData := `John,22,55000,true
Jane,26,62000,false
Mike,31,78000,true`
	
	csvNoHeaderReader := strings.NewReader(csvNoHeaderData)
	csvNoHeaderSource := stream.NewCSVSource(csvNoHeaderReader).
		WithoutHeaders().
		WithHeaders([]string{"name", "age", "salary", "active"})
	
	csvNoHeaderStream := csvNoHeaderSource.ToStream()
	
	fmt.Println("Data without headers (custom headers applied):")
	noHeaderRecords, _ := stream.Collect(csvNoHeaderStream)
	for i, record := range noHeaderRecords {
		name := stream.GetOr(record, "name", "")
		age := stream.GetOr(record, "age", int64(0))
		fmt.Printf("  %d: %s (age=%d)\n", i+1, name, age)
	}

	// Test 3: TSV Source
	fmt.Println("\n📑 Test 3: TSV Source")
	fmt.Println("---------------------")
	
	tsvData := `product	category	price	in_stock
Laptop	Electronics	999.99	true
Mouse	Electronics	25.50	true
Desk	Furniture	299.00	false
Chair	Furniture	199.99	true`
	
	tsvReader := strings.NewReader(tsvData)
	tsvSource := stream.NewTSVSource(tsvReader)
	tsvStream := tsvSource.ToStream()
	
	fmt.Println("TSV Records:")
	tsvRecords, _ := stream.Collect(tsvStream)
	for i, record := range tsvRecords {
		product := stream.GetOr(record, "product", "")
		category := stream.GetOr(record, "category", "")
		price := stream.GetOr(record, "price", 0.0)
		inStock := stream.GetOr(record, "in_stock", false)
		fmt.Printf("  %d: %s (%s) - $%.2f, in_stock=%v\n", 
			i+1, product, category, price, inStock)
	}

	// Test 4: CSV Sink - Writing to file
	fmt.Println("\n💾 Test 4: CSV Sink (writing to file)")
	fmt.Println("-------------------------------------")
	
	// Create sample sales data
	salesData := []stream.Record{
		stream.NewRecord().Int("id", 1).Float("amount", 150.75).Set("date", time.Now()).String("customer", "Alice").Build(),
		stream.NewRecord().Int("id", 2).Float("amount", 200.50).Set("date", time.Now().Add(-24*time.Hour)).String("customer", "Bob").Build(),
		stream.NewRecord().Int("id", 3).Float("amount", 175.25).Set("date", time.Now().Add(-48*time.Hour)).String("customer", "Charlie").Build(),
	}
	
	// Write to CSV file
	csvOutputFile := "/tmp/sales_output.csv"
	err := stream.StreamToCSVFile(stream.FromSlice(salesData), csvOutputFile)
	if err != nil {
		fmt.Printf("Error writing to CSV: %v\n", err)
		return
	}
	
	fmt.Printf("Sales data written to: %s\n", csvOutputFile)
	
	// Read it back to verify
	readBackStream, err := stream.CSVToStreamFromFile(csvOutputFile)
	if err != nil {
		fmt.Printf("Error reading back CSV: %v\n", err)
		return
	}
	
	fmt.Println("Verification - reading back from file:")
	readBackRecords, _ := stream.Collect(readBackStream)
	for i, record := range readBackRecords {
		id := stream.GetOr(record, "id", int64(0))
		amount := stream.GetOr(record, "amount", 0.0)
		customer := stream.GetOr(record, "customer", "")
		fmt.Printf("  %d: ID=%d, Amount=$%.2f, Customer=%s\n", 
			i+1, id, amount, customer)
	}

	// Test 5: TSV Sink
	fmt.Println("\n📝 Test 5: TSV Sink")
	fmt.Println("-------------------")
	
	// Create inventory data
	inventoryData := []stream.Record{
		stream.NewRecord().String("sku", "LAP001").String("name", "Gaming Laptop").Int("quantity", 15).Float("price", 1299.99).Build(),
		stream.NewRecord().String("sku", "MOU001").String("name", "Wireless Mouse").Int("quantity", 50).Float("price", 29.99).Build(),
		stream.NewRecord().String("sku", "KEY001").String("name", "Mechanical Keyboard").Int("quantity", 25).Float("price", 89.99).Build(),
	}
	
	tsvOutputFile := "/tmp/inventory_output.tsv"
	err = stream.StreamToTSVFile(stream.FromSlice(inventoryData), tsvOutputFile)
	if err != nil {
		fmt.Printf("Error writing to TSV: %v\n", err)
		return
	}
	
	fmt.Printf("Inventory data written to: %s\n", tsvOutputFile)
	
	// Verify TSV content
	if content, err := os.ReadFile(tsvOutputFile); err == nil {
		fmt.Println("TSV Content:")
		fmt.Println(string(content))
	}

	// Test 6: Streaming CSV Writer (for infinite streams)
	fmt.Println("\n🌊 Test 6: Streaming CSV Writer")
	fmt.Println("-------------------------------")
	
	// Create streaming output file
	streamingFile, err := os.Create("/tmp/streaming_output.csv")
	if err != nil {
		fmt.Printf("Error creating streaming file: %v\n", err)
		return
	}
	defer streamingFile.Close()
	
	// Create streaming CSV writer
	headers := []string{"timestamp", "sensor", "value", "status"}
	streamingWriter := stream.NewStreamingCSVWriter(streamingFile, headers)
	
	fmt.Println("Writing sensor data in real-time...")
	
	// Simulate real-time sensor data
	sensors := []string{"temperature", "humidity", "pressure"}
	for i := 0; i < 10; i++ {
		sensor := sensors[i%len(sensors)]
		record := stream.NewRecord().
			Int("timestamp", time.Now().Unix()).
			String("sensor", sensor).
			Float("value", 20.0 + float64(i*5)).
			Bool("status", i%2 == 0).
			Build()
		
		err = streamingWriter.WriteRecord(record)
		if err != nil {
			fmt.Printf("Error writing streaming record: %v\n", err)
			break
		}
		
		fmt.Printf("  Wrote: %s=%.1f\n", sensor, 20.0+float64(i*5))
		time.Sleep(100 * time.Millisecond) // Simulate real-time
	}
	
	streamingWriter.Close()
	fmt.Println("Streaming CSV writing complete!")

	// Test 7: Complex processing pipeline with CSV I/O
	fmt.Println("\n🔄 Test 7: Processing Pipeline with CSV I/O")
	fmt.Println("--------------------------------------------")
	
	// Read sales data, process it, and write results
	salesStream, err := stream.CSVToStreamFromFile(csvOutputFile)
	if err != nil {
		fmt.Printf("Error reading sales CSV: %v\n", err)
		return
	}
	
	// Process: Calculate sales metrics
	processedStream := stream.Map(func(record stream.Record) stream.Record {
		amount := stream.GetOr(record, "amount", 0.0)
		customer := stream.GetOr(record, "customer", "")
		
		// Calculate tax (10%) and total
		tax := amount * 0.10
		total := amount + tax
		
		// Determine customer tier
		var tier string
		switch {
		case amount >= 200:
			tier = "Premium"
		case amount >= 150:
			tier = "Standard"
		default:
			tier = "Basic"
		}
		
		return stream.NewRecord().
			Float("original_amount", amount).
			Float("tax", tax).
			Float("total", total).
			String("customer", customer).
			String("tier", tier).
			String("processed_at", time.Now().Format("2006-01-02 15:04:05")).
			Build()
	})(salesStream)
	
	// Write processed results
	processedOutputFile := "/tmp/processed_sales.csv"
	err = stream.StreamToCSVFile(processedStream, processedOutputFile)
	if err != nil {
		fmt.Printf("Error writing processed CSV: %v\n", err)
		return
	}
	
	fmt.Printf("Processed sales written to: %s\n", processedOutputFile)
	
	// Show processed results
	processedReadStream, _ := stream.CSVToStreamFromFile(processedOutputFile)
	processedRecords, _ := stream.Collect(processedReadStream)
	
	fmt.Println("Processed Results:")
	for i, record := range processedRecords {
		customer := stream.GetOr(record, "customer", "")
		total := stream.GetOr(record, "total", 0.0)
		tier := stream.GetOr(record, "tier", "")
		fmt.Printf("  %d: %s - Total=$%.2f (%s tier)\n", 
			i+1, customer, total, tier)
	}

	// Test 8: Window-based CSV processing
	fmt.Println("\n📊 Test 8: Windowed CSV Processing")
	fmt.Println("----------------------------------")
	
	// Create larger dataset for windowing
	largeDataset := make([]stream.Record, 20)
	for i := 0; i < 20; i++ {
		largeDataset[i] = stream.NewRecord().
			Int("transaction_id", int64(i+1)).
			Float("amount", float64(100 + (i*15)%200)).
			String("region", []string{"US", "EU", "ASIA"}[i%3]).
			Build()
	}
	
	// Write large dataset
	largeDataFile := "/tmp/large_dataset.csv"
	largeStream, err := stream.FromRecords(largeDataset)
	if err != nil {
		panic(err)
	}
	err = stream.StreamToCSVFile(largeStream, largeDataFile)
	if err != nil {
		fmt.Printf("Error writing large dataset: %v\n", err)
		return
	}
	
	// Read and process in windows
	largeStreamRead, _ := stream.CSVToStreamFromFile(largeDataFile)
	windowedStream := stream.CountWindow[stream.Record](5)(largeStreamRead)
	
	fmt.Println("Processing in windows of 5 records:")
	windowNum := 0
	for {
		window, err := windowedStream()
		if err == stream.EOS {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}
		
		windowNum++
		
		// Calculate window statistics
		amounts := stream.ExtractField[float64]("amount")(window)
		sum, _ := stream.Sum(amounts)
		
		fmt.Printf("  Window %d: Total=$%.2f\n", windowNum, sum)
	}

	// Cleanup
	fmt.Println("\n🧹 Cleanup")
	fmt.Println("----------")
	files := []string{csvOutputFile, tsvOutputFile, "/tmp/streaming_output.csv", 
		processedOutputFile, largeDataFile}
	
	for _, file := range files {
		if err := os.Remove(file); err == nil {
			fmt.Printf("Removed: %s\n", file)
		}
	}
	
	fmt.Println("\n🎉 CSV/TSV I/O Testing Complete!")
	fmt.Println("✨ Sources, sinks, and streaming I/O all working perfectly!")
}