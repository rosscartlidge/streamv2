package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Println("🔧 Flexible I/O with io.Reader/Writer")
	fmt.Println("=====================================")

	// Test 1: CSV from string buffer (no files!)
	fmt.Println("\n📝 Test 1: CSV from string buffer")
	fmt.Println("----------------------------------")
	
	csvData := `name,age,salary,active
Alice,30,75000.50,true
Bob,25,65000,false
Charlie,35,85000.25,true`

	// Read CSV directly from string
	csvStream := stream.CSVToStream(strings.NewReader(csvData))
	records, _ := stream.Collect(csvStream)
	
	fmt.Printf("Read %d records from string buffer:\n", len(records))
	for i, record := range records {
		name := stream.GetOr(record, "name", "")
		age := stream.GetOr(record, "age", int64(0))
		salary := stream.GetOr(record, "salary", 0.0)
		fmt.Printf("  %d: %s (age=%d, salary=$%.2f)\n", i+1, name, age, salary)
	}

	// Test 2: CSV to memory buffer (no files!)
	fmt.Println("\n💾 Test 2: CSV to memory buffer") 
	fmt.Println("-------------------------------")
	
	salesData := []stream.Record{
		stream.NewRecord().Int("id", 1).Float("amount", 150.75).String("customer", "Alice").Build(),
		stream.NewRecord().Int("id", 2).Float("amount", 200.50).String("customer", "Bob").Build(),
		stream.NewRecord().Int("id", 3).Float("amount", 175.25).String("customer", "Charlie").Build(),
	}
	
	// Write CSV to memory buffer
	var csvBuffer bytes.Buffer
	err := stream.StreamToCSV(stream.FromSlice(salesData), &csvBuffer)
	if err != nil {
		fmt.Printf("Error writing to buffer: %v\n", err)
		return
	}
	
	fmt.Println("CSV written to memory buffer:")
	fmt.Println(csvBuffer.String())

	// Test 3: Round-trip through memory (CSV → Process → CSV)
	fmt.Println("\n🔄 Test 3: Memory round-trip processing")
	fmt.Println("--------------------------------------")
	
	// Read back from buffer
	readStream := stream.CSVToStream(&csvBuffer)
	
	// Process: add tax calculation
	processedStream := stream.Map(func(record stream.Record) stream.Record {
		amount := stream.GetOr(record, "amount", 0.0)
		customer := stream.GetOr(record, "customer", "")
		
		tax := amount * 0.1
		total := amount + tax
		
		return stream.NewRecord().
			String("customer", customer).
			Float("original_amount", amount).
			Float("tax", tax).
			Float("total", total).
			String("processed_at", time.Now().Format("15:04:05")).
			Build()
	})(readStream)
	
	// Write processed results to new buffer
	var processedBuffer bytes.Buffer
	err = stream.StreamToCSV(processedStream, &processedBuffer)
	if err != nil {
		fmt.Printf("Error writing processed data: %v\n", err)
		return
	}
	
	fmt.Println("Processed CSV in memory:")
	fmt.Println(processedBuffer.String())

	// Test 4: TSV to stdout
	fmt.Println("\n📤 Test 4: TSV to stdout")
	fmt.Println("------------------------")
	
	inventoryData := []stream.Record{
		stream.NewRecord().String("sku", "LAP001").String("name", "Gaming Laptop").Float("price", 1299.99).Build(),
		stream.NewRecord().String("sku", "MOU001").String("name", "Wireless Mouse").Float("price", 29.99).Build(),
		stream.NewRecord().String("sku", "KEY001").String("name", "Mechanical Keyboard").Float("price", 89.99).Build(),
	}
	
	fmt.Println("Writing TSV directly to stdout:")
	err = stream.StreamToTSV(stream.FromSlice(inventoryData), os.Stdout)
	if err != nil {
		fmt.Printf("Error writing to stdout: %v\n", err)
	}

	// Test 5: Pipe CSV through multiple transformations in memory
	fmt.Println("\n🏭 Test 5: Multi-stage memory pipeline")
	fmt.Println("-------------------------------------")
	
	// Stage 1: Create sales data
	salesStream := stream.FromSlice([]stream.Record{
		stream.NewRecord().String("region", "US").String("product", "laptop").Int("amount", 1200).Build(),
		stream.NewRecord().String("region", "EU").String("product", "mouse").Int("amount", 25).Build(),
		stream.NewRecord().String("region", "ASIA").String("product", "keyboard").Int("amount", 75).Build(),
		stream.NewRecord().String("region", "US").String("product", "monitor").Int("amount", 300).Build(),
		stream.NewRecord().String("region", "EU").String("product", "laptop").Int("amount", 1100).Build(),
	})
	
	// Stage 1: CSV buffer 1
	var stage1Buffer bytes.Buffer
	stream.StreamToCSV(salesStream, &stage1Buffer)
	fmt.Println("Stage 1 - Raw sales data:")
	fmt.Print(stage1Buffer.String())
	
	// Stage 2: Read and add regional processing
	stage1Stream := stream.CSVToStream(&stage1Buffer)
	stage2Stream := stream.Map(func(record stream.Record) stream.Record {
		amount := stream.GetOr(record, "amount", 0)
		region := stream.GetOr(record, "region", "")
		product := stream.GetOr(record, "product", "")
		
		// Regional markup
		var markup float64
		switch region {
		case "US":
			markup = 1.08 // 8% tax
		case "EU": 
			markup = 1.20 // 20% VAT
		case "ASIA":
			markup = 1.05 // 5% tax
		}
		
		finalPrice := float64(amount) * markup
		
		return stream.NewRecord().
			String("region", region).
			String("product", product).
			Int("base_amount", amount).
			Float("markup", markup).
			Float("final_price", finalPrice).
			Build()
	})(stage1Stream)
	
	// Stage 2: CSV buffer 2  
	var stage2Buffer bytes.Buffer
	stream.StreamToCSV(stage2Stream, &stage2Buffer)
	fmt.Println("\nStage 2 - With regional pricing:")
	fmt.Print(stage2Buffer.String())
	
	// Stage 3: Read and summarize by region
	stage2ReadStream := stream.CSVToStream(&stage2Buffer)
	results, _ := stream.Collect(
		stream.GroupBy([]string{"region"}, 
			stream.FieldSumSpec[float64]("total_sales", "final_price"),
			stream.FieldAvgSpec[float64]("avg_price", "final_price"),
		)(stage2ReadStream),
	)
	
	// Stage 3: Final results
	var stage3Buffer bytes.Buffer
	stream.StreamToCSV(stream.FromSlice(results), &stage3Buffer)
	fmt.Println("\nStage 3 - Regional summary:")
	fmt.Print(stage3Buffer.String())

	// Test 6: Show the power - HTTP-like usage (simulated)
	fmt.Println("\n🌐 Test 6: HTTP-like usage simulation")
	fmt.Println("------------------------------------")
	
	// Simulate receiving CSV from HTTP request body
	httpBody := `user_id,action,timestamp
1001,login,2023-01-15T10:30:00Z
1002,purchase,2023-01-15T10:31:00Z
1001,logout,2023-01-15T10:45:00Z
1003,login,2023-01-15T10:46:00Z`
	
	fmt.Println("Simulating CSV from HTTP request body:")
	
	// Process directly from "HTTP body"
	httpStream := stream.CSVToStream(strings.NewReader(httpBody))
	
	// Add processing timestamp and user session info
	processedHTTPStream := stream.Map(func(record stream.Record) stream.Record {
		userID := stream.GetOr(record, "user_id", "")
		action := stream.GetOr(record, "action", "")
		timestamp := stream.GetOr(record, "timestamp", "")
		
		return stream.NewRecord().
			String("user_id", userID).
			String("action", action).
			String("original_timestamp", timestamp).
			Int("processed_at", time.Now().Unix()).
			String("session_id", fmt.Sprintf("sess_%s_%d", userID, time.Now().Unix()%1000)).
			Build()
	})(httpStream)
	
	// "Send" processed CSV back as HTTP response (to stdout)
	fmt.Println("Processed CSV response:")
	stream.StreamToCSV(processedHTTPStream, os.Stdout)
	
	fmt.Println("\n🎉 Flexible I/O Examples Complete!")
	fmt.Println("✨ No files required - pure stream processing!")
}