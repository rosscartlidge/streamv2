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
	fmt.Println("🗃️ JSON I/O with Structured Data Support")
	fmt.Println("==========================================")

	// Test 1: JSON Lines from string buffer
	fmt.Println("\n📝 Test 1: JSON Lines (JSONL) format")
	fmt.Println("------------------------------------")
	
	jsonlData := `{"name": "Alice", "age": 30, "salary": 75000.50, "active": true}
{"name": "Bob", "age": 25, "salary": 65000, "active": false}
{"name": "Charlie", "age": 35, "salary": 85000.25, "active": true}`

	fmt.Println("Input JSON Lines:")
	fmt.Println(jsonlData)
	
	// Read JSON Lines directly from string
	jsonlStream := stream.JSONToStream(strings.NewReader(jsonlData))
	records, _ := stream.Collect(jsonlStream)
	
	fmt.Printf("\nParsed %d records from JSON Lines:\n", len(records))
	for i, record := range records {
		name := stream.GetOr(record, "name", "")
		age := stream.GetOr(record, "age", int64(0))
		salary := stream.GetOr(record, "salary", 0.0)
		active := stream.GetOr(record, "active", false)
		fmt.Printf("  %d: %s (age=%d, salary=$%.2f, active=%v)\n", 
			i+1, name, age, salary, active)
	}

	// Test 2: JSON Array format
	fmt.Println("\n📋 Test 2: JSON Array format")
	fmt.Println("----------------------------")
	
	jsonArrayData := `[
		{"product": "laptop", "category": "electronics", "price": 999.99, "in_stock": true},
		{"product": "mouse", "category": "electronics", "price": 25.50, "in_stock": true},
		{"product": "desk", "category": "furniture", "price": 299.00, "in_stock": false}
	]`
	
	// Read JSON Array
	jsonArrayStream := stream.NewJSONSource(strings.NewReader(jsonArrayData)).
		WithFormat(stream.JSONArray).
		ToStream()
	
	arrayRecords, _ := stream.Collect(jsonArrayStream)
	fmt.Printf("Read %d products from JSON array:\n", len(arrayRecords))
	for i, record := range arrayRecords {
		product := stream.GetOr(record, "product", "")
		price := stream.GetOr(record, "price", 0.0)
		inStock := stream.GetOr(record, "in_stock", false)
		fmt.Printf("  %d: %s - $%.2f (in_stock=%v)\n", i+1, product, price, inStock)
	}

	// Test 3: Structured data with nested objects and arrays
	fmt.Println("\n🏗️ Test 3: Structured data with nesting")
	fmt.Println("---------------------------------------")
	
	structuredData := `{"user_id": 1001, "name": "Alice", "profile": {"email": "alice@example.com", "preferences": {"theme": "dark", "notifications": true}}, "orders": [{"id": 1, "amount": 150.75}, {"id": 2, "amount": 200.50}], "created_at": "2023-01-15T10:30:00Z"}
{"user_id": 1002, "name": "Bob", "profile": {"email": "bob@example.com", "preferences": {"theme": "light", "notifications": false}}, "orders": [{"id": 3, "amount": 75.25}], "created_at": "2023-01-16T14:20:00Z"}`
	
	structuredStream := stream.JSONToStream(strings.NewReader(structuredData))
	structuredRecords, _ := stream.Collect(structuredStream)
	
	fmt.Printf("Processing %d users with structured data:\n", len(structuredRecords))
	for i, record := range structuredRecords {
		userID := stream.GetOr(record, "user_id", int64(0))
		name := stream.GetOr(record, "name", "")
		
		// Access nested profile data
		if profileRecord, ok := stream.Get[stream.Record](record, "profile"); ok {
			email := stream.GetOr(profileRecord, "email", "")
			fmt.Printf("  %d: User %d (%s) - %s\n", i+1, userID, name, email)
			
			// Access nested preferences
			if prefs, ok := stream.Get[stream.Record](profileRecord, "preferences"); ok {
				theme := stream.GetOr(prefs, "theme", "")
				notifications := stream.GetOr(prefs, "notifications", false)
				fmt.Printf("      Preferences: theme=%s, notifications=%v\n", theme, notifications)
			}
		}
		
		// Access orders array (converted to Stream[any])
		if ordersStream, ok := stream.Get[stream.Stream[any]](record, "orders"); ok {
			orders, _ := stream.Collect(ordersStream)
			fmt.Printf("      Orders: %d items\n", len(orders))
			for j, order := range orders {
				if orderMap, ok := order.(stream.Record); ok {
					orderID := stream.GetOr(orderMap, "id", int64(0))
					amount := stream.GetOr(orderMap, "amount", 0.0)
					fmt.Printf("        %d: Order #%d - $%.2f\n", j+1, orderID, amount)
				}
			}
		}
	}

	// Test 4: JSON output to memory buffer
	fmt.Println("\n💾 Test 4: Writing JSON Lines to memory")
	fmt.Println("---------------------------------------")
	
	// Create sample e-commerce data with nested structures
	salesData := []stream.Record{
		stream.R(
			"order_id", 1001,
			"customer", stream.R("name", "Alice", "email", "alice@example.com"),
			"items", stream.FromSlice([]any{
				stream.R("product", "laptop", "quantity", 1, "price", 1299.99),
				stream.R("product", "mouse", "quantity", 2, "price", 29.99),
			}),
			"total", 1359.97,
			"timestamp", time.Now(),
		),
		stream.R(
			"order_id", 1002, 
			"customer", stream.R("name", "Bob", "email", "bob@example.com"),
			"items", stream.FromSlice([]any{
				stream.R("product", "keyboard", "quantity", 1, "price", 89.99),
			}),
			"total", 89.99,
			"timestamp", time.Now().Add(-time.Hour),
		),
	}
	
	// Write as JSON Lines to memory buffer
	var jsonlBuffer bytes.Buffer
	err := stream.StreamToJSON(stream.FromSlice(salesData), &jsonlBuffer)
	if err != nil {
		fmt.Printf("Error writing JSON Lines: %v\n", err)
		return
	}
	
	fmt.Println("Generated JSON Lines with structured data:")
	fmt.Println(jsonlBuffer.String())

	// Test 5: JSON Array with pretty printing
	fmt.Println("\n🎨 Test 5: Pretty-printed JSON Array")
	fmt.Println("------------------------------------")
	
	// Write as pretty-printed JSON Array
	var prettyBuffer bytes.Buffer
	sink := stream.NewJSONSink(&prettyBuffer).
		WithFormat(stream.JSONArray).
		WithPrettyPrint()
	
	err = sink.WriteRecords(salesData)
	if err != nil {
		fmt.Printf("Error writing pretty JSON: %v\n", err)
		return
	}
	
	fmt.Println("Pretty-printed JSON Array:")
	fmt.Println(prettyBuffer.String())

	// Test 6: Round-trip processing with structured data
	fmt.Println("\n🔄 Test 6: Round-trip structured data processing")
	fmt.Println("-----------------------------------------------")
	
	// Read back the JSON Lines data
	readBackStream := stream.JSONToStream(&jsonlBuffer)
	
	// Process: Calculate order metrics and flatten structure
	processedStream := stream.Map(func(record stream.Record) stream.Record {
		orderID := stream.GetOr(record, "order_id", int64(0))
		total := stream.GetOr(record, "total", 0.0)
		
		// Extract customer info
		var customerName, customerEmail string
		if customer, ok := stream.Get[stream.Record](record, "customer"); ok {
			customerName = stream.GetOr(customer, "name", "")
			customerEmail = stream.GetOr(customer, "email", "")
		}
		
		// Count items
		var itemCount int
		if itemsStream, ok := stream.Get[stream.Stream[any]](record, "items"); ok {
			items, _ := stream.Collect(itemsStream)
			itemCount = len(items)
		}
		
		// Calculate metrics
		avgItemValue := 0.0
		if itemCount > 0 {
			avgItemValue = total / float64(itemCount)
		}
		
		return stream.R(
			"order_id", orderID,
			"customer_name", customerName,
			"customer_email", customerEmail,
			"item_count", itemCount,
			"total_amount", total,
			"avg_item_value", avgItemValue,
			"processed_at", time.Now().Format("2006-01-02 15:04:05"),
		)
	})(readBackStream)
	
	// Write processed results as pretty JSON
	var processedBuffer bytes.Buffer
	processedSink := stream.NewJSONSink(&processedBuffer).
		WithFormat(stream.JSONArray).
		WithPrettyPrint()
	
	err = processedSink.WriteStream(processedStream)
	if err != nil {
		fmt.Printf("Error writing processed JSON: %v\n", err)
		return
	}
	
	fmt.Println("Processed order metrics:")
	fmt.Println(processedBuffer.String())

	// Test 7: File-based JSON operations
	fmt.Println("\n📁 Test 7: File-based JSON operations")
	fmt.Println("-------------------------------------")
	
	// Write to file
	analyticsFile := "/tmp/order_analytics.json"
	err = stream.StreamToJSONFile(stream.FromSlice(salesData), analyticsFile)
	if err != nil {
		fmt.Printf("Error writing JSON file: %v\n", err)
		return
	}
	
	// Read back from file
	fileStream, err := stream.JSONToStreamFromFile(analyticsFile)
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		return
	}
	
	fileRecords, _ := stream.Collect(fileStream)
	fmt.Printf("Successfully read %d orders from JSON file\n", len(fileRecords))
	
	// Cleanup
	os.Remove(analyticsFile)

	// Test 8: Integration with existing StreamV2 operations
	fmt.Println("\n⚙️ Test 8: Integration with StreamV2 operations")
	fmt.Println("-----------------------------------------------")
	
	// Create larger dataset for aggregation
	orderData := []stream.Record{
		stream.R("region", "US", "category", "electronics", "amount", 1299.99),
		stream.R("region", "EU", "category", "electronics", "amount", 899.99),
		stream.R("region", "US", "category", "furniture", "amount", 299.99),
		stream.R("region", "ASIA", "category", "electronics", "amount", 1199.99),
		stream.R("region", "EU", "category", "furniture", "amount", 399.99),
		stream.R("region", "US", "category", "electronics", "amount", 799.99),
	}
	
	// Convert to JSON, then back to stream, then aggregate
	var tempBuffer bytes.Buffer
	stream.StreamToJSON(stream.FromSlice(orderData), &tempBuffer)
	
	reloadedStream := stream.JSONToStream(&tempBuffer)
	
	// Group by region with aggregations
	results, _ := stream.Collect(
		stream.GroupBy([]string{"region"}, 
			stream.FieldSumSpec[float64]("total_sales", "amount"),
			stream.FieldAvgSpec[float64]("avg_order", "amount"),
		)(reloadedStream),
	)
	
	// Output results as pretty JSON
	var resultsBuffer bytes.Buffer
	resultsSink := stream.NewJSONSink(&resultsBuffer).
		WithFormat(stream.JSONArray).
		WithPrettyPrint()
	
	resultsSink.WriteRecords(results)
	
	fmt.Println("Regional sales aggregation (JSON → Stream → GroupBy → JSON):")
	fmt.Println(resultsBuffer.String())

	fmt.Println("🎉 JSON Structured Data I/O Complete!")
	fmt.Println("✨ Full support for nested objects, arrays, and Stream[T] fields!")
}