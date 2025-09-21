package main

import (
	"fmt"
	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Println("=== Join Filters Examples ===")
	fmt.Println()

	// Example 1: Basic InnerJoin - Only matching records
	fmt.Println("1. InnerJoin - Basic Usage (Only Matching Records)")
	fmt.Println("─────────────────────────────────────────────────")

	// Users table
	users := []stream.Record{
		stream.NewRecord().Int("id", 1).String("username", "alice").String("email", "alice@example.com").Build(),
		stream.NewRecord().Int("id", 2).String("username", "bob").String("email", "bob@example.com").Build(),
		stream.NewRecord().Int("id", 3).String("username", "charlie").String("email", "charlie@example.com").Build(),
	}

	// Profiles table (Charlie has no profile)
	profiles := []stream.Record{
		stream.NewRecord().Int("userId", 1).String("firstName", "Alice").String("lastName", "Johnson").String("department", "Engineering").Build(),
		stream.NewRecord().Int("userId", 2).String("firstName", "Bob").String("lastName", "Smith").String("department", "Sales").Build(),
	}

	userStream := stream.FromRecordsUnsafe(users)
	profileStream := stream.FromRecordsUnsafe(profiles)

	// Inner join - only users with profiles
	innerJoined := stream.InnerJoin(profileStream, "id", "userId")(userStream)
	innerResults, _ := stream.Collect(innerJoined)

	fmt.Printf("Inner Join: %d results (only users with profiles)\n", len(innerResults))
	for i, result := range innerResults {
		username := stream.GetOr(result, "username", "")
		firstName := stream.GetOr(result, "firstName", "")
		department := stream.GetOr(result, "department", "")
		fmt.Printf("  %d: %s (%s) - %s\n", i+1, username, firstName, department)
	}

	fmt.Println()

	// Example 2: LeftJoin - All left records + matching right
	fmt.Println("2. LeftJoin - All Left Records + Matching Right")
	fmt.Println("──────────────────────────────────────────────")

	// Reset streams
	userStream2 := stream.FromRecordsUnsafe(users)
	profileStream2 := stream.FromRecordsUnsafe(profiles)

	// Left join - all users, with profiles when available
	leftJoined := stream.LeftJoin(profileStream2, "id", "userId")(userStream2)
	leftResults, _ := stream.Collect(leftJoined)

	fmt.Printf("Left Join: %d results (all users)\n", len(leftResults))
	for i, result := range leftResults {
		username := stream.GetOr(result, "username", "")
		firstName := stream.GetOr(result, "firstName", "NO PROFILE")
		department := stream.GetOr(result, "department", "NO PROFILE")
		fmt.Printf("  %d: %s (%s) - %s\n", i+1, username, firstName, department)
	}

	fmt.Println()

	// Example 3: RightJoin - All right records + matching left
	fmt.Println("3. RightJoin - All Right Records + Matching Left")
	fmt.Println("───────────────────────────────────────────────")

	// Extended profiles with one that has no matching user
	extendedProfiles := []stream.Record{
		stream.NewRecord().Int("userId", 1).String("firstName", "Alice").String("lastName", "Johnson").String("department", "Engineering").Build(),
		stream.NewRecord().Int("userId", 2).String("firstName", "Bob").String("lastName", "Smith").String("department", "Sales").Build(),
		stream.NewRecord().Int("userId", 99).String("firstName", "Diana").String("lastName", "Wilson").String("department", "HR").Build(), // No matching user
	}

	userStream3 := stream.FromRecordsUnsafe(users)
	extendedProfileStream := stream.FromRecordsUnsafe(extendedProfiles)

	// Right join - all profiles, with users when available
	rightJoined := stream.RightJoin(extendedProfileStream, "id", "userId")(userStream3)
	rightResults, _ := stream.Collect(rightJoined)

	fmt.Printf("Right Join: %d results (all profiles)\n", len(rightResults))
	for i, result := range rightResults {
		username := stream.GetOr(result, "username", "NO USER")
		firstName := stream.GetOr(result, "firstName", "")
		department := stream.GetOr(result, "department", "")
		fmt.Printf("  %d: %s (%s) - %s\n", i+1, username, firstName, department)
	}

	fmt.Println()

	// Example 4: FullJoin - All records from both sides
	fmt.Println("4. FullJoin - All Records from Both Sides")
	fmt.Println("────────────────────────────────────────")

	userStream4 := stream.FromRecordsUnsafe(users)
	extendedProfileStream2 := stream.FromRecordsUnsafe(extendedProfiles)

	// Full join - all users and all profiles
	fullJoined := stream.FullJoin(extendedProfileStream2, "id", "userId")(userStream4)
	fullResults, _ := stream.Collect(fullJoined)

	fmt.Printf("Full Join: %d results (all users and all profiles)\n", len(fullResults))
	for i, result := range fullResults {
		username := stream.GetOr(result, "username", "NO USER")
		firstName := stream.GetOr(result, "firstName", "NO PROFILE")
		department := stream.GetOr(result, "department", "NO PROFILE")
		fmt.Printf("  %d: %s (%s) - %s\n", i+1, username, firstName, department)
	}

	fmt.Println()

	// Example 5: Field Conflict Resolution
	fmt.Println("5. Field Conflict Resolution")
	fmt.Println("───────────────────────────")

	// Both tables have 'name' field
	customers := []stream.Record{
		stream.NewRecord().Int("id", 1).String("name", "Customer Alice").String("type", "premium").Build(),
		stream.NewRecord().Int("id", 2).String("name", "Customer Bob").String("type", "standard").Build(),
	}

	orders := []stream.Record{
		stream.NewRecord().Int("customerId", 1).String("name", "Order #001").Float("amount", 99.99).Build(),
		stream.NewRecord().Int("customerId", 2).String("name", "Order #002").Float("amount", 49.99).Build(),
	}

	customerStream := stream.FromRecordsUnsafe(customers)
	orderStream := stream.FromRecordsUnsafe(orders)

	fmt.Println("a) Default prefixes (left./right.):")
	defaultJoined := stream.InnerJoin(orderStream, "id", "customerId")(customerStream)
	defaultResults, _ := stream.Collect(defaultJoined)

	for i, result := range defaultResults {
		customerName := stream.GetOr(result, "left.name", "")
		orderName := stream.GetOr(result, "right.name", "")
		amount := stream.GetOr(result, "amount", 0.0)
		fmt.Printf("  %d: Customer='%s', Order='%s', Amount=$%.2f\n", i+1, customerName, orderName, amount)
	}

	fmt.Println("b) Custom prefixes (customer./order.):")
	customerStream2 := stream.FromRecordsUnsafe(customers)
	orderStream2 := stream.FromRecordsUnsafe(orders)

	customJoined := stream.InnerJoin(orderStream2, "id", "customerId", 
		stream.WithPrefixes("customer.", "order."))(customerStream2)
	customResults, _ := stream.Collect(customJoined)

	for i, result := range customResults {
		customerName := stream.GetOr(result, "customer.name", "")
		orderName := stream.GetOr(result, "order.name", "")
		amount := stream.GetOr(result, "amount", 0.0)
		fmt.Printf("  %d: Customer='%s', Order='%s', Amount=$%.2f\n", i+1, customerName, orderName, amount)
	}

	fmt.Println("c) No prefixes (right side wins):")
	customerStream3 := stream.FromRecordsUnsafe(customers)
	orderStream3 := stream.FromRecordsUnsafe(orders)

	noPrefixJoined := stream.InnerJoin(orderStream3, "id", "customerId",
		stream.WithPrefixes("", ""))(customerStream3)
	noPrefixResults, _ := stream.Collect(noPrefixJoined)

	for i, result := range noPrefixResults {
		name := stream.GetOr(result, "name", "")
		amount := stream.GetOr(result, "amount", 0.0)
		fmt.Printf("  %d: Name='%s' (order won), Amount=$%.2f\n", i+1, name, amount)
	}

	fmt.Println()

	// Example 6: Multiple Joins in Pipeline
	fmt.Println("6. Multiple Joins in Pipeline")
	fmt.Println("────────────────────────────")

	// Sample data for complex joins
	employees := []stream.Record{
		stream.NewRecord().Int("empId", 1).String("name", "Alice").Int("deptId", 10).Int("managerId", 3).Build(),
		stream.NewRecord().Int("empId", 2).String("name", "Bob").Int("deptId", 20).Int("managerId", 3).Build(),
		stream.NewRecord().Int("empId", 3).String("name", "Charlie").Int("deptId", 10).Int("managerId", 0).Build(), // Manager
	}

	departments := []stream.Record{
		stream.NewRecord().Int("deptId", 10).String("deptName", "Engineering").String("location", "NYC").Build(),
		stream.NewRecord().Int("deptId", 20).String("deptName", "Sales").String("location", "SF").Build(),
	}

	managers := []stream.Record{
		stream.NewRecord().Int("managerId", 3).String("managerName", "Charlie Boss").String("title", "Director").Build(),
	}

	empStream := stream.FromRecordsUnsafe(employees)
	deptStream := stream.FromRecordsUnsafe(departments)
	managerStream := stream.FromRecordsUnsafe(managers)

	// Chain multiple joins
	enrichedEmployees := stream.Pipe(
		// First join with departments
		stream.InnerJoin(deptStream, "deptId", "deptId", stream.WithPrefixes("emp.", "dept.")),
		// Then join with managers
		stream.LeftJoin(managerStream, "emp.managerId", "managerId", stream.WithPrefixes("", "mgr.")),
	)(empStream)

	enrichedResults, _ := stream.Collect(enrichedEmployees)

	fmt.Printf("Enriched employees: %d results\n", len(enrichedResults))
	for i, result := range enrichedResults {
		name := stream.GetOr(result, "emp.name", "")
		dept := stream.GetOr(result, "dept.deptName", "")
		location := stream.GetOr(result, "dept.location", "")
		manager := stream.GetOr(result, "mgr.managerName", "NO MANAGER")
		fmt.Printf("  %d: %s - %s (%s) - Reports to: %s\n", i+1, name, dept, location, manager)
	}

	fmt.Println()

	// Example 7: Join with Aggregation
	fmt.Println("7. Join with Aggregation - Order Analytics")
	fmt.Println("────────────────────────────────────────")

	orderRecords := []stream.Record{
		stream.NewRecord().String("orderId", "ORD_001").Int("customerId", 1).Float("amount", 99.99).String("status", "completed").Build(),
		stream.NewRecord().String("orderId", "ORD_002").Int("customerId", 1).Float("amount", 149.99).String("status", "completed").Build(),
		stream.NewRecord().String("orderId", "ORD_003").Int("customerId", 2).Float("amount", 49.99).String("status", "pending").Build(),
		stream.NewRecord().String("orderId", "ORD_004").Int("customerId", 1).Float("amount", 199.99).String("status", "completed").Build(),
	}

	customerData := []stream.Record{
		stream.NewRecord().Int("customerId", 1).String("customerName", "Alice Corp").String("tier", "premium").Build(),
		stream.NewRecord().Int("customerId", 2).String("customerName", "Bob Inc").String("tier", "standard").Build(),
	}

	orderAnalyticsStream := stream.FromRecordsUnsafe(orderRecords)
	customerDataStream := stream.FromRecordsUnsafe(customerData)

	// Join orders with customer data, then filter and analyze
	orderAnalytics := stream.Pipe3(
		// Join with customer data
		stream.InnerJoin(customerDataStream, "customerId", "customerId", stream.WithPrefixes("order.", "cust.")),
		// Filter for completed orders only
		stream.Where(func(r stream.Record) bool {
			status := stream.GetOr(r, "order.status", "")
			return status == "completed"
		}),
		// Select relevant fields for analysis
		stream.Select("order.orderId", "cust.customerName", "cust.tier", "order.amount"),
	)(orderAnalyticsStream)

	analyticsResults, _ := stream.Collect(orderAnalytics)

	fmt.Printf("Completed order analytics: %d results\n", len(analyticsResults))
	totalRevenue := 0.0
	for i, result := range analyticsResults {
		orderId := stream.GetOr(result, "order.orderId", "")
		customer := stream.GetOr(result, "cust.customerName", "")
		tier := stream.GetOr(result, "cust.tier", "")
		amount := stream.GetOr(result, "order.amount", 0.0)
		totalRevenue += amount
		fmt.Printf("  %d: %s - %s (%s tier) - $%.2f\n", i+1, orderId, customer, tier, amount)
	}
	fmt.Printf("Total Revenue: $%.2f\n", totalRevenue)

	fmt.Println()

	// Example 8: Performance Considerations
	fmt.Println("8. Performance Considerations - Hash Join Algorithm")
	fmt.Println("─────────────────────────────────────────────────")

	// Create larger datasets to demonstrate performance characteristics
	fmt.Println("Creating large datasets for performance demo...")

	// Large left stream (10,000 transactions)
	largeTransactions := make([]stream.Record, 10000)
	for i := 0; i < 10000; i++ {
		customerId := (i % 100) + 1 // 100 unique customers
		largeTransactions[i] = stream.NewRecord().
			String("txnId", fmt.Sprintf("TXN_%05d", i)).
			Int("customerId", int64(customerId)).
			Float("amount", float64((i%1000)+1)*1.99).
			Build()
	}

	// Smaller right stream (100 customers) - this gets loaded into hash table
	largeCustomers := make([]stream.Record, 100)
	for i := 0; i < 100; i++ {
		largeCustomers[i] = stream.NewRecord().
			Int("customerId", int64(i+1)).
			String("customerName", fmt.Sprintf("Customer_%03d", i+1)).
			String("segment", []string{"premium", "standard", "basic"}[i%3]).
			Build()
	}

	largeTxnStream := stream.FromRecordsUnsafe(largeTransactions)
	largeCustomerStream := stream.FromRecordsUnsafe(largeCustomers)

	fmt.Printf("Processing: %d transactions × %d customers\n", len(largeTransactions), len(largeCustomers))
	fmt.Println("Hash Join: Right stream (customers) loaded into memory, left stream (transactions) processed lazily")

	// Perform the join
	largeJoinResult := stream.InnerJoin(largeCustomerStream, "customerId", "customerId")(largeTxnStream)

	// Count results without collecting all into memory
	count := 0
	for {
		_, err := largeJoinResult()
		if err != nil {
			break
		}
		count++
	}

	fmt.Printf("Successfully joined %d records\n", count)
	fmt.Println("✅ Efficient for large left streams with smaller right lookup tables")
	fmt.Println("⚠️  Right stream must be finite and fit in memory")

	fmt.Println()

	fmt.Println("=== Summary ===")
	fmt.Println("✅ InnerJoin: Only matching records from both streams")
	fmt.Println("✅ LeftJoin: All left records + matching right records")
	fmt.Println("✅ RightJoin: All right records + matching left records")
	fmt.Println("✅ FullJoin: All records from both streams")
	fmt.Println("✅ Field Conflicts: Configurable prefixes (default: left./right.)")
	fmt.Println("✅ Hash Join: O(R + L) time, O(R) space complexity")
	fmt.Println("✅ Pipeline Integration: Composable with other stream operations")
	fmt.Println("✅ Production Ready: Handles large datasets efficiently")
}