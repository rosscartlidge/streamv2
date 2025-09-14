package stream

import (
	"fmt"
	"testing"
)

// TestInnerJoin tests inner join functionality
func TestInnerJoin(t *testing.T) {
	t.Run("BasicInnerJoin", func(t *testing.T) {
		// Left stream: users
		users := []Record{
			NewRecord().Int("id", 1).String("name", "Alice").Build(),
			NewRecord().Int("id", 2).String("name", "Bob").Build(),
			NewRecord().Int("id", 3).String("name", "Charlie").Build(),
		}
		
		// Right stream: profiles
		profiles := []Record{
			NewRecord().Int("userId", 1).String("department", "Engineering").Build(),
			NewRecord().Int("userId", 2).String("department", "Sales").Build(),
			// No profile for Charlie (id=3)
		}
		
		userStream := FromRecordsUnsafe(users)
		profileStream := FromRecordsUnsafe(profiles)
		
		// Inner join on id = userId
		joined := InnerJoin(profileStream, "id", "userId")(userStream)
		results, err := Collect(joined)
		if err != nil {
			t.Fatalf("Failed to collect join results: %v", err)
		}
		
		// Should have 2 results (Alice and Bob, no Charlie)
		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
		
		// Verify Alice's record
		alice := results[0]
		if GetOr(alice, "id", int64(0)) != 1 {
			t.Errorf("Expected id=1, got %v", alice["id"])
		}
		if GetOr(alice, "name", "") != "Alice" {
			t.Errorf("Expected name=Alice, got %v", alice["name"])
		}
		if GetOr(alice, "department", "") != "Engineering" {
			t.Errorf("Expected department=Engineering, got %v", alice["department"])
		}
		
		// Verify Bob's record
		bob := results[1]
		if GetOr(bob, "id", int64(0)) != 2 {
			t.Errorf("Expected id=2, got %v", bob["id"])
		}
		if GetOr(bob, "name", "") != "Bob" {
			t.Errorf("Expected name=Bob, got %v", bob["name"])
		}
		if GetOr(bob, "department", "") != "Sales" {
			t.Errorf("Expected department=Sales, got %v", bob["department"])
		}
	})

	t.Run("FieldConflicts", func(t *testing.T) {
		// Both streams have "name" field
		left := []Record{
			NewRecord().Int("id", 1).String("name", "Alice User").Build(),
		}
		right := []Record{
			NewRecord().Int("userId", 1).String("name", "Alice Profile").Build(),
		}
		
		leftStream := FromRecordsUnsafe(left)
		rightStream := FromRecordsUnsafe(right)
		
		// Inner join with default prefixes
		joined := InnerJoin(rightStream, "id", "userId")(leftStream)
		results, err := Collect(joined)
		if err != nil {
			t.Fatalf("Failed to collect join results: %v", err)
		}
		
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		
		result := results[0]
		// Should have prefixed names due to conflict
		if GetOr(result, "left.name", "") != "Alice User" {
			t.Errorf("Expected left.name=Alice User, got %v", result["left.name"])
		}
		if GetOr(result, "right.name", "") != "Alice Profile" {
			t.Errorf("Expected right.name=Alice Profile, got %v", result["right.name"])
		}
	})

	t.Run("CustomPrefixes", func(t *testing.T) {
		left := []Record{
			NewRecord().Int("id", 1).String("name", "Alice User").Build(),
		}
		right := []Record{
			NewRecord().Int("userId", 1).String("name", "Alice Profile").Build(),
		}
		
		leftStream := FromRecordsUnsafe(left)
		rightStream := FromRecordsUnsafe(right)
		
		// Inner join with custom prefixes
		joined := InnerJoin(rightStream, "id", "userId", WithPrefixes("user.", "profile."))(leftStream)
		results, err := Collect(joined)
		if err != nil {
			t.Fatalf("Failed to collect join results: %v", err)
		}
		
		result := results[0]
		if GetOr(result, "user.name", "") != "Alice User" {
			t.Errorf("Expected user.name=Alice User, got %v", result["user.name"])
		}
		if GetOr(result, "profile.name", "") != "Alice Profile" {
			t.Errorf("Expected profile.name=Alice Profile, got %v", result["profile.name"])
		}
	})

	t.Run("NoPrefixes", func(t *testing.T) {
		left := []Record{
			NewRecord().Int("id", 1).String("name", "Alice User").Build(),
		}
		right := []Record{
			NewRecord().Int("userId", 1).String("name", "Alice Profile").Build(),
		}
		
		leftStream := FromRecordsUnsafe(left)
		rightStream := FromRecordsUnsafe(right)
		
		// Inner join with no prefixes (right wins)
		joined := InnerJoin(rightStream, "id", "userId", WithPrefixes("", ""))(leftStream)
		results, err := Collect(joined)
		if err != nil {
			t.Fatalf("Failed to collect join results: %v", err)
		}
		
		result := results[0]
		if GetOr(result, "name", "") != "Alice Profile" {
			t.Errorf("Expected name=Alice Profile (right wins), got %v", result["name"])
		}
	})
}

// TestLeftJoin tests left join functionality
func TestLeftJoin(t *testing.T) {
	t.Run("BasicLeftJoin", func(t *testing.T) {
		// Left stream: users
		users := []Record{
			NewRecord().Int("id", 1).String("name", "Alice").Build(),
			NewRecord().Int("id", 2).String("name", "Bob").Build(),
			NewRecord().Int("id", 3).String("name", "Charlie").Build(), // No matching profile
		}
		
		// Right stream: profiles  
		profiles := []Record{
			NewRecord().Int("userId", 1).String("department", "Engineering").Build(),
			NewRecord().Int("userId", 2).String("department", "Sales").Build(),
		}
		
		userStream := FromRecordsUnsafe(users)
		profileStream := FromRecordsUnsafe(profiles)
		
		// Left join - all users should be included
		joined := LeftJoin(profileStream, "id", "userId")(userStream)
		results, err := Collect(joined)
		if err != nil {
			t.Fatalf("Failed to collect join results: %v", err)
		}
		
		// Should have 3 results (all users)
		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}
		
		// Charlie should be present without department
		charlie := results[2]
		if GetOr(charlie, "name", "") != "Charlie" {
			t.Errorf("Expected name=Charlie, got %v", charlie["name"])
		}
		if _, exists := charlie["department"]; exists {
			t.Errorf("Charlie should not have department field, got %v", charlie["department"])
		}
	})
}

// TestRightJoin tests right join functionality
func TestRightJoin(t *testing.T) {
	t.Run("BasicRightJoin", func(t *testing.T) {
		// Left stream: users (missing user for id=4)
		users := []Record{
			NewRecord().Int("id", 1).String("name", "Alice").Build(),
			NewRecord().Int("id", 2).String("name", "Bob").Build(),
		}
		
		// Right stream: profiles
		profiles := []Record{
			NewRecord().Int("userId", 1).String("department", "Engineering").Build(),
			NewRecord().Int("userId", 2).String("department", "Sales").Build(),
			NewRecord().Int("userId", 4).String("department", "HR").Build(), // No matching user
		}
		
		userStream := FromRecordsUnsafe(users)
		profileStream := FromRecordsUnsafe(profiles)
		
		// Right join - all profiles should be included
		joined := RightJoin(profileStream, "id", "userId")(userStream)
		results, err := Collect(joined)
		if err != nil {
			t.Fatalf("Failed to collect join results: %v", err)
		}
		
		// Should have 3 results (all profiles)
		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}
		
		// Find the HR department record (should have no name)
		var hrRecord Record
		for _, record := range results {
			if GetOr(record, "department", "") == "HR" {
				hrRecord = record
				break
			}
		}
		
		if hrRecord == nil {
			t.Fatalf("HR department record not found")
		}
		
		if _, exists := hrRecord["name"]; exists {
			t.Errorf("HR record should not have name field, got %v", hrRecord["name"])
		}
	})
}

// TestFullJoin tests full outer join functionality
func TestFullJoin(t *testing.T) {
	t.Run("BasicFullJoin", func(t *testing.T) {
		// Left stream: users
		users := []Record{
			NewRecord().Int("id", 1).String("name", "Alice").Build(),
			NewRecord().Int("id", 2).String("name", "Bob").Build(),
			NewRecord().Int("id", 3).String("name", "Charlie").Build(), // No matching profile
		}
		
		// Right stream: profiles
		profiles := []Record{
			NewRecord().Int("userId", 1).String("department", "Engineering").Build(),
			NewRecord().Int("userId", 2).String("department", "Sales").Build(),
			NewRecord().Int("userId", 4).String("department", "HR").Build(), // No matching user
		}
		
		userStream := FromRecordsUnsafe(users)
		profileStream := FromRecordsUnsafe(profiles)
		
		// Full join - should include all users and all profiles
		joined := FullJoin(profileStream, "id", "userId")(userStream)
		results, err := Collect(joined)
		if err != nil {
			t.Fatalf("Failed to collect join results: %v", err)
		}
		
		// Should have 4 results (Alice + Bob matched, Charlie unmatched, HR unmatched)
		if len(results) != 4 {
			t.Fatalf("Expected 4 results, got %d", len(results))
		}
		
		// Verify we have all expected records
		names := make(map[string]bool)
		departments := make(map[string]bool)
		
		for _, record := range results {
			if name, exists := record["name"]; exists {
				names[name.(string)] = true
			}
			if dept, exists := record["department"]; exists {
				departments[dept.(string)] = true
			}
		}
		
		expectedNames := []string{"Alice", "Bob", "Charlie"}
		for _, name := range expectedNames {
			if !names[name] {
				t.Errorf("Expected to find name %s in results", name)
			}
		}
		
		expectedDepts := []string{"Engineering", "Sales", "HR"}
		for _, dept := range expectedDepts {
			if !departments[dept] {
				t.Errorf("Expected to find department %s in results", dept)
			}
		}
	})
}

// TestJoinPerformance tests join with larger datasets
func TestJoinPerformance(t *testing.T) {
	t.Run("LargeDataset", func(t *testing.T) {
		// Create 1000 users
		users := make([]Record, 1000)
		for i := 0; i < 1000; i++ {
			users[i] = NewRecord().
				Int("id", int64(i)).
				String("name", fmt.Sprintf("User%d", i)).
				Build()
		}
		
		// Create 500 profiles (only half have profiles)
		profiles := make([]Record, 500)
		for i := 0; i < 500; i++ {
			profiles[i] = NewRecord().
				Int("userId", int64(i)).
				String("department", fmt.Sprintf("Dept%d", i%10)).
				Build()
		}
		
		userStream := FromRecordsUnsafe(users)
		profileStream := FromRecordsUnsafe(profiles)
		
		// Inner join
		joined := InnerJoin(profileStream, "id", "userId")(userStream)
		results, err := Collect(joined)
		if err != nil {
			t.Fatalf("Failed to collect join results: %v", err)
		}
		
		// Should have 500 results (users with profiles)
		if len(results) != 500 {
			t.Fatalf("Expected 500 results, got %d", len(results))
		}
	})
}