package stream

import (
	"testing"
)

// TestDotFlatten tests the DotFlatten function
func TestDotFlatten(t *testing.T) {
	t.Run("SimpleNestedRecord", func(t *testing.T) {
		// Create a record with nested structure
		nestedRecord := NewRecord().
			String("name", "Alice").
			Set("address", NewRecord().
				String("street", "123 Main St").
				String("city", "New York").
				Int("zip", 10001).
				Build()).
			Int("age", 30).
			Build()

		records := []Record{nestedRecord}
		stream, _ := FromRecords(records)

		// Apply dot flatten
		flattened := DotFlatten(".")(stream)
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened results: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		result := results[0]

		// Check that nested fields are flattened with dot notation
		if GetOr(result, "name", "") != "Alice" {
			t.Errorf("Expected name=Alice, got %v", result["name"])
		}
		if GetOr(result, "address.street", "") != "123 Main St" {
			t.Errorf("Expected address.street=123 Main St, got %v", result["address.street"])
		}
		if GetOr(result, "address.city", "") != "New York" {
			t.Errorf("Expected address.city=New York, got %v", result["address.city"])
		}
		if GetOr(result, "address.zip", int64(0)) != 10001 {
			t.Errorf("Expected address.zip=10001, got %v", result["address.zip"])
		}
		if GetOr(result, "age", int64(0)) != 30 {
			t.Errorf("Expected age=30, got %v", result["age"])
		}

		// Check that original nested field is gone
		if _, exists := result["address"]; exists {
			t.Errorf("Original nested 'address' field should be removed")
		}
	})

	t.Run("DeeplyNestedRecord", func(t *testing.T) {
		// Create deeply nested structure
		deepRecord := NewRecord().
			String("id", "user123").
			Set("profile", NewRecord().
				String("name", "Bob").
				Set("contact", NewRecord().
					String("email", "bob@example.com").
					Set("phone", NewRecord().
						String("home", "555-1234").
						String("mobile", "555-5678").
						Build()).
					Build()).
				Build()).
			Build()

		records := []Record{deepRecord}
		stream, _ := FromRecords(records)

		// Apply dot flatten
		flattened := DotFlatten(".")(stream)
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened results: %v", err)
		}

		result := results[0]

		// Check deeply nested fields
		if GetOr(result, "id", "") != "user123" {
			t.Errorf("Expected id=user123, got %v", result["id"])
		}
		if GetOr(result, "profile.name", "") != "Bob" {
			t.Errorf("Expected profile.name=Bob, got %v", result["profile.name"])
		}
		if GetOr(result, "profile.contact.email", "") != "bob@example.com" {
			t.Errorf("Expected profile.contact.email=bob@example.com, got %v", result["profile.contact.email"])
		}
		if GetOr(result, "profile.contact.phone.home", "") != "555-1234" {
			t.Errorf("Expected profile.contact.phone.home=555-1234, got %v", result["profile.contact.phone.home"])
		}
		if GetOr(result, "profile.contact.phone.mobile", "") != "555-5678" {
			t.Errorf("Expected profile.contact.phone.mobile=555-5678, got %v", result["profile.contact.phone.mobile"])
		}
	})

	t.Run("CustomSeparator", func(t *testing.T) {
		nestedRecord := NewRecord().
			String("name", "Charlie").
			Set("meta", NewRecord().
				String("type", "user").
				String("role", "admin").
				Build()).
			Build()

		records := []Record{nestedRecord}
		stream, _ := FromRecords(records)

		// Use custom separator
		flattened := DotFlatten("_")(stream)
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened results: %v", err)
		}

		result := results[0]

		if GetOr(result, "meta_type", "") != "user" {
			t.Errorf("Expected meta_type=user, got %v", result["meta_type"])
		}
		if GetOr(result, "meta_role", "") != "admin" {
			t.Errorf("Expected meta_role=admin, got %v", result["meta_role"])
		}
	})

	t.Run("SpecificFields", func(t *testing.T) {
		// Create record with multiple nested fields, but only flatten some
		userRecord := NewRecord().
			String("name", "Alice").
			String("email", "alice@example.com").
			Build()
		
		addressRecord := NewRecord().
			String("street", "123 Main St").
			String("city", "New York").
			Int("zip", 10001).
			Build()
		
		prefsRecord := NewRecord().
			String("theme", "dark").
			Bool("notifications", true).
			Build()

		record := NewRecord().
			String("id", "user_123").
			Int("age", 30).
			Set("user", userRecord).
			Set("address", addressRecord).
			Set("preferences", prefsRecord).
			Build()

		records := []Record{record}
		stream := FromRecordsUnsafe(records)

		// Apply dot flatten only to user and address (not preferences)
		flattened := DotFlatten(".", "user", "address")(stream)
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened results: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		result := results[0]

		// Check that specified fields were flattened
		if GetOr(result, "user.name", "") != "Alice" {
			t.Errorf("Expected user.name=Alice, got %v", result["user.name"])
		}
		if GetOr(result, "user.email", "") != "alice@example.com" {
			t.Errorf("Expected user.email=alice@example.com, got %v", result["user.email"])
		}
		if GetOr(result, "address.street", "") != "123 Main St" {
			t.Errorf("Expected address.street=123 Main St, got %v", result["address.street"])
		}
		if GetOr(result, "address.city", "") != "New York" {
			t.Errorf("Expected address.city=New York, got %v", result["address.city"])
		}

		// Check that non-flattened fields remain as nested records
		if _, exists := result["user"]; exists {
			t.Errorf("Original nested 'user' field should be removed after flattening")
		}
		if _, exists := result["address"]; exists {
			t.Errorf("Original nested 'address' field should be removed after flattening")
		}
		
		// Check that preferences remain as nested record (not flattened)
		if prefs, exists := result["preferences"]; !exists {
			t.Errorf("Expected 'preferences' field to remain as nested record")
		} else if prefsRecord, ok := prefs.(Record); !ok {
			t.Errorf("Expected 'preferences' to still be a Record, got %T", prefs)
		} else {
			if GetOr(prefsRecord, "theme", "") != "dark" {
				t.Errorf("Expected preferences.theme=dark, got %v", prefsRecord["theme"])
			}
		}

		// Check that other fields remain unchanged
		if GetOr(result, "id", "") != "user_123" {
			t.Errorf("Expected id=user_123, got %v", result["id"])
		}
		if GetOr(result, "age", int64(0)) != 30 {
			t.Errorf("Expected age=30, got %v", result["age"])
		}
	})
}

// TestCrossFlatten tests the CrossFlatten function
func TestCrossFlatten(t *testing.T) {
	t.Run("SingleStreamField", func(t *testing.T) {
		// Create record with stream field
		tagsStream := FromSliceAny([]any{"golang", "streaming", "data"})
		record := NewRecord().
			Int("id", 1).
			String("name", "Alice").
			Set("tags", tagsStream).
			Build()

		records := []Record{record}
		stream := FromRecordsUnsafe(records)

		// Apply cross flatten with default separator (expand all stream fields)
		flattened := CrossFlatten(".")(stream)
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened results: %v", err)
		}

		// Should create 3 records (one for each tag)
		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		// Check each result
		expectedTags := []string{"golang", "streaming", "data"}
		for i, result := range results {
			if GetOr(result, "id", int64(0)) != 1 {
				t.Errorf("Result %d: Expected id=1, got %v", i, result["id"])
			}
			if GetOr(result, "name", "") != "Alice" {
				t.Errorf("Result %d: Expected name=Alice, got %v", i, result["name"])
			}
			if GetOr(result, "tags", "") != expectedTags[i] {
				t.Errorf("Result %d: Expected tags=%s, got %v", i, expectedTags[i], result["tags"])
			}
		}
	})

	t.Run("MultipleStreamFields", func(t *testing.T) {
		// Create record with multiple stream fields (cartesian product)
		colorsStream := FromSliceAny([]any{"red", "blue"})
		sizesStream := FromSliceAny([]any{"S", "M"})
		
		record := NewRecord().
			String("product", "shirt").
			Float("price", 29.99).
			Set("colors", colorsStream).
			Set("sizes", sizesStream).
			Build()

		records := []Record{record}
		stream := FromRecordsUnsafe(records)

		// Apply cross flatten
		flattened := CrossFlatten(".")(stream)
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened results: %v", err)
		}

		// Should create 4 records (2 colors × 2 sizes)
		if len(results) != 4 {
			t.Fatalf("Expected 4 results (2×2), got %d", len(results))
		}

		// Check that all records have the non-stream fields
		for i, result := range results {
			product := GetOr(result, "product", "")
			price := GetOr(result, "price", 0.0)

			if product != "shirt" {
				t.Errorf("Result %d: Expected product=shirt, got %v", i, product)
			}
			if price != 29.99 {
				t.Errorf("Result %d: Expected price=29.99, got %v", i, price)
			}
		}
	})

	t.Run("NoStreamFields", func(t *testing.T) {
		// Record with no stream fields - should pass through unchanged
		record := NewRecord().
			String("name", "Diana").
			Int("age", 25).
			Build()

		records := []Record{record}
		stream := FromRecordsUnsafe(records)

		// Apply cross flatten
		flattened := CrossFlatten(".")(stream)
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened results: %v", err)
		}

		// Should return original record unchanged
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		result := results[0]
		if GetOr(result, "name", "") != "Diana" {
			t.Errorf("Expected name=Diana, got %v", result["name"])
		}
		if GetOr(result, "age", int64(0)) != 25 {
			t.Errorf("Expected age=25, got %v", result["age"])
		}
	})

	t.Run("SpecificFields", func(t *testing.T) {
		// Create record with multiple stream fields, but only expand some
		colorsStream := FromSliceAny([]any{"red", "blue"})
		sizesStream := FromSliceAny([]any{"S", "M"})
		categoriesStream := FromSliceAny([]any{"shirts", "pants"})
		
		record := NewRecord().
			String("product", "clothing").
			Float("price", 19.99).
			Set("colors", colorsStream).
			Set("sizes", sizesStream).
			Set("categories", categoriesStream).
			Build()

		records := []Record{record}
		stream := FromRecordsUnsafe(records)

		// Apply cross flatten only to colors and sizes (not categories)
		flattened := CrossFlatten(".", "colors", "sizes")(stream)
		results, err := Collect(flattened)
		if err != nil {
			t.Fatalf("Failed to collect flattened results: %v", err)
		}

		// Should create 4 records (2 colors × 2 sizes), categories should remain as stream
		if len(results) != 4 {
			t.Fatalf("Expected 4 results (2×2), got %d", len(results))
		}

		// Check that all records have the expected fields
		for i, result := range results {
			product := GetOr(result, "product", "")
			price := GetOr(result, "price", 0.0)
			color := GetOr(result, "colors", "")
			size := GetOr(result, "sizes", "")

			if product != "clothing" {
				t.Errorf("Result %d: Expected product=clothing, got %v", i, product)
			}
			if price != 19.99 {
				t.Errorf("Result %d: Expected price=19.99, got %v", i, price)
			}
			if color == "" {
				t.Errorf("Result %d: Expected color to be expanded, got empty", i)
			}
			if size == "" {
				t.Errorf("Result %d: Expected size to be expanded, got empty", i)
			}
			
			// Categories should still be a stream (not expanded)
			if _, exists := result["categories"]; !exists {
				t.Errorf("Result %d: Expected categories field to remain", i)
			}
		}
	})
}

// TestFlattenIntegration tests both flatten functions working together
func TestFlattenIntegration(t *testing.T) {
	t.Run("DotThenCross", func(t *testing.T) {
		// Create complex nested record with streams
		tagsStream := FromSliceAny([]any{"tech", "news"})
		record := NewRecord().
			String("id", "article_1").
			Set("meta", NewRecord().
				String("author", "Alice").
				String("category", "programming").
				Build()).
			Set("tags", tagsStream).
			Build()

		records := []Record{record}
		stream := FromRecordsUnsafe(records)

		// First apply dot flatten
		dotFlattened := DotFlatten(".")(stream)

		// Then apply cross flatten
		crossFlattened := CrossFlatten(".")(dotFlattened)

		results, err := Collect(crossFlattened)
		if err != nil {
			t.Fatalf("Failed to collect results: %v", err)
		}

		// Should have 2 records (one for each tag)
		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}

		// Check both records have flattened nested fields and expanded tags
		expectedTags := []string{"tech", "news"}
		for i, result := range results {
			if GetOr(result, "id", "") != "article_1" {
				t.Errorf("Result %d: Expected id=article_1, got %v", i, result["id"])
			}
			if GetOr(result, "meta.author", "") != "Alice" {
				t.Errorf("Result %d: Expected meta.author=Alice, got %v", i, result["meta.author"])
			}
			if GetOr(result, "meta.category", "") != "programming" {
				t.Errorf("Result %d: Expected meta.category=programming, got %v", i, result["meta.category"])
			}
			// Tags should be expanded (exact field name depends on implementation)
			if GetOr(result, "tags", "") != expectedTags[i] {
				t.Errorf("Result %d: Expected tags=%s, got %v", i, expectedTags[i], result["tags"])
			}
		}
	})
}