package main

import (
	"fmt"
	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Println("=== Flatten Filters Examples ===")
	fmt.Println()

	// Example 1: Basic DotFlatten - Flatten all nested records
	fmt.Println("1. DotFlatten - Basic Usage (Flatten All Nested Records)")
	fmt.Println("─────────────────────────────────────────────────────────")
	
	userRecord := stream.NewRecord().
		String("id", "user_123").
		String("email", "alice@example.com").
		Set("profile", stream.NewRecord().
			String("firstName", "Alice").
			String("lastName", "Johnson").
			Int("age", 30).
			Build()).
		Set("address", stream.NewRecord().
			String("street", "123 Main St").
			String("city", "New York").
			String("state", "NY").
			Int("zipCode", 10001).
			Build()).
		Build()

	records := []stream.Record{userRecord}
	inputStream := stream.FromRecordsUnsafe(records)

	// Flatten all nested records using dot notation
	flattened := stream.DotFlatten(".")(inputStream)
	results, _ := stream.Collect(flattened)

	fmt.Println("Original nested structure flattened with dots:")
	for k, v := range results[0] {
		fmt.Printf("  %s: %v\n", k, v)
	}

	fmt.Println()

	// Example 2: DotFlatten with specific fields
	fmt.Println("2. DotFlatten - Selective Field Flattening")
	fmt.Println("────────────────────────────────────────────")

	// Same record as above
	inputStream2 := stream.FromRecordsUnsafe(records)

	// Flatten only the profile field, leave address as nested record
	profileOnly := stream.DotFlatten(".", "profile")(inputStream2)
	results2, _ := stream.Collect(profileOnly)

	fmt.Println("Only 'profile' field flattened, 'address' remains nested:")
	for k, v := range results2[0] {
		if nestedRecord, ok := v.(stream.Record); ok && k == "address" {
			fmt.Printf("  %s: nested record with fields:\n", k)
			for nk, nv := range nestedRecord {
				fmt.Printf("    %s.%s: %v\n", k, nk, nv)
			}
		} else {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}

	fmt.Println()

	// Example 3: Basic CrossFlatten - Expand all stream fields
	fmt.Println("3. CrossFlatten - Basic Usage (Expand All Stream Fields)")
	fmt.Println("──────────────────────────────────────────────────────────")

	productRecord := stream.NewRecord().
		String("productId", "shirt_001").
		String("name", "Cotton T-Shirt").
		Float("basePrice", 19.99).
		Set("colors", stream.FromSliceAny([]any{"red", "blue", "green"})).
		Set("sizes", stream.FromSliceAny([]any{"S", "M", "L", "XL"})).
		Build()

	productRecords := []stream.Record{productRecord}
	productStream := stream.FromRecordsUnsafe(productRecords)

	// Expand all stream fields (creates cartesian product)
	expanded := stream.CrossFlatten(".")(productStream)
	expandedResults, _ := stream.Collect(expanded)

	fmt.Printf("Original 1 record expanded to %d records (3 colors × 4 sizes):\n", len(expandedResults))
	for i, result := range expandedResults {
		if i >= 6 { // Show first 6 for brevity
			fmt.Printf("  ... and %d more combinations\n", len(expandedResults)-6)
			break
		}
		color := stream.GetOr(result, "colors", "")
		size := stream.GetOr(result, "sizes", "")
		price := stream.GetOr(result, "basePrice", 0.0)
		fmt.Printf("  %d: %s %s shirt - $%.2f\n", i+1, color, size, price)
	}

	fmt.Println()

	// Example 4: CrossFlatten with specific fields
	fmt.Println("4. CrossFlatten - Selective Stream Expansion")
	fmt.Println("──────────────────────────────────────────────")

	inventoryRecord := stream.NewRecord().
		String("warehouseId", "WH_001").
		String("productType", "clothing").
		Set("colors", stream.FromSliceAny([]any{"red", "blue"})).
		Set("sizes", stream.FromSliceAny([]any{"S", "M", "L"})).
		Set("locations", stream.FromSliceAny([]any{"A1", "A2", "B1", "B2"})).
		Build()

	inventoryRecords := []stream.Record{inventoryRecord}
	inventoryStream := stream.FromRecordsUnsafe(inventoryRecords)

	// Expand only colors and sizes, keep locations as stream for later processing
	partialExpansion := stream.CrossFlatten(".", "colors", "sizes")(inventoryStream)
	partialResults, _ := stream.Collect(partialExpansion)

	fmt.Printf("Expanded colors×sizes (%d records), locations remain as stream:\n", len(partialResults))
	for i, result := range partialResults {
		if i >= 4 { // Show first 4 for brevity
			fmt.Printf("  ... and %d more combinations\n", len(partialResults)-4)
			break
		}
		color := stream.GetOr(result, "colors", "")
		size := stream.GetOr(result, "sizes", "")
		fmt.Printf("  %d: %s/%s (locations stream preserved)\n", i+1, color, size)
	}

	fmt.Println()

	// Example 5: Combining DotFlatten and CrossFlatten
	fmt.Println("5. Combined Flatten Operations")
	fmt.Println("──────────────────────────────")

	complexRecord := stream.NewRecord().
		String("orderId", "ORD_001").
		Set("customer", stream.NewRecord().
			String("name", "Alice Johnson").
			String("email", "alice@example.com").
			Set("address", stream.NewRecord().
				String("street", "123 Main St").
				String("city", "New York").
				Build()).
			Build()).
		Set("items", stream.FromSliceAny([]any{"laptop", "mouse"})).
		Set("shippingOptions", stream.FromSliceAny([]any{"standard", "express"})).
		Build()

	complexRecords := []stream.Record{complexRecord}
	complexStream := stream.FromRecordsUnsafe(complexRecords)

	// First flatten nested records, then expand specific streams
	pipeline := stream.Pipe(
		stream.DotFlatten("."),                           // Flatten all nested records
		stream.CrossFlatten(".", "items"),               // Expand only items stream
	)(complexStream)

	pipelineResults, _ := stream.Collect(pipeline)

	fmt.Printf("After DotFlatten + CrossFlatten: %d records\n", len(pipelineResults))
	for i, result := range pipelineResults {
		item := stream.GetOr(result, "items", "")
		customerName := stream.GetOr(result, "customer.name", "")
		address := stream.GetOr(result, "customer.address.city", "")
		fmt.Printf("  %d: Order for %s in %s - Item: %s\n", i+1, customerName, address, item)
	}

	fmt.Println()

	// Example 6: Using Flatten with FlatMap for stream-level processing
	fmt.Println("6. Flatten with FlatMap - Processing Multiple Records")
	fmt.Println("───────────────────────────────────────────────────────")

	// Multiple order records
	order1 := stream.NewRecord().
		String("orderId", "ORD_001").
		String("customer", "Alice").
		Set("items", stream.FromSliceAny([]any{"laptop", "mouse"})).
		Build()

	order2 := stream.NewRecord().
		String("orderId", "ORD_002").
		String("customer", "Bob").
		Set("items", stream.FromSliceAny([]any{"keyboard", "monitor", "cables"})).
		Build()

	orders := []stream.Record{order1, order2}
	ordersStream := stream.FromRecordsUnsafe(orders)

	// Use FlatMap to expand items for each order
	itemizedOrders := stream.FlatMap(func(order stream.Record) stream.Stream[stream.Record] {
		// Convert single order to stream and expand its items
		singleOrderStream := stream.FromRecordsUnsafe([]stream.Record{order})
		return stream.CrossFlatten(".", "items")(singleOrderStream)
	})(ordersStream)

	itemizedResults, _ := stream.Collect(itemizedOrders)

	fmt.Printf("Expanded %d orders into %d itemized records:\n", len(orders), len(itemizedResults))
	for i, result := range itemizedResults {
		orderId := stream.GetOr(result, "orderId", "")
		customer := stream.GetOr(result, "customer", "")
		item := stream.GetOr(result, "items", "")
		fmt.Printf("  %d: %s - %s ordered %s\n", i+1, orderId, customer, item)
	}

	fmt.Println()

	// Example 7: Real-world data processing pipeline
	fmt.Println("7. Real-World Example - E-commerce Analytics")
	fmt.Println("────────────────────────────────────────────────")

	// Simulate e-commerce data with nested customer info and multiple products
	salesRecord := stream.NewRecord().
		String("transactionId", "TXN_12345").
		String("timestamp", "2024-01-15T14:30:00Z").
		Set("customer", stream.NewRecord().
			String("customerId", "CUST_001").
			String("name", "Alice Johnson").
			String("tier", "premium").
			Set("location", stream.NewRecord().
				String("country", "US").
				String("state", "NY").
				String("city", "New York").
				Build()).
			Build()).
		Set("products", stream.FromSliceAny([]any{"laptop", "mouse", "keyboard"})).
		Set("channels", stream.FromSliceAny([]any{"online", "mobile"})).
		Float("totalAmount", 1299.99).
		Build()

	salesRecords := []stream.Record{salesRecord}
	salesStream := stream.FromRecordsUnsafe(salesRecords)

	// Complete analytics pipeline
	analyticsResult := stream.Pipe3(
		// 1. Flatten all nested customer data
		stream.DotFlatten("."),
		// 2. Expand products to analyze each product sale separately  
		stream.CrossFlatten(".", "products"),
		// 3. Select relevant fields for analysis
		stream.Select("transactionId", "customer.name", "customer.tier", 
			"customer.location.country", "products", "totalAmount"),
	)(salesStream)

	analyticsResults, _ := stream.Collect(analyticsResult)

	fmt.Printf("Analytics pipeline result: %d product-level records:\n", len(analyticsResults))
	for i, result := range analyticsResults {
		txnId := stream.GetOr(result, "transactionId", "")
		customer := stream.GetOr(result, "customer.name", "")
		tier := stream.GetOr(result, "customer.tier", "")
		country := stream.GetOr(result, "customer.location.country", "")
		product := stream.GetOr(result, "products", "")
		amount := stream.GetOr(result, "totalAmount", 0.0)
		fmt.Printf("  %d: %s - %s (%s, %s) bought %s - $%.2f\n", 
			i+1, txnId, customer, tier, country, product, amount)
	}

	fmt.Println()

	fmt.Println("=== Summary ===")
	fmt.Println("✅ DotFlatten: Converts nested records to flat structure with dot notation")
	fmt.Println("✅ CrossFlatten: Expands stream fields into multiple records (cartesian product)")
	fmt.Println("✅ Selective Processing: Both support field-specific operations")
	fmt.Println("✅ Composable: Chain with other filters for complex data transformations")
	fmt.Println("✅ FlatMap Integration: Process entire streams of complex records")
	fmt.Println("✅ Real-world Ready: Handle complex nested data with streams efficiently")
}