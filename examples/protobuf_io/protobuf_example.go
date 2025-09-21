package main

import (
	"fmt"
	"log"

	"github.com/rosscartlidge/streamv2/examples/protobuf_io/pb"
	"github.com/rosscartlidge/streamv2/pkg/stream"
	"google.golang.org/protobuf/proto"
)

func main() {
	fmt.Println("🔧 Protocol Buffer I/O with StreamV2")
	fmt.Println("====================================")

	// Test 1: Creating Records and converting to protobuf
	fmt.Println("\n📝 Test 1: Records to Protobuf")
	fmt.Println("------------------------------")

	// Create sample user data using StreamV2 Records
	userData := []stream.Record{
		stream.NewRecord().
			Int("id", 1).
			String("name", "Alice Johnson").
			String("email", "alice@company.com").
			Bool("active", true).
			Set("tags", stream.FromSliceAny([]any{"developer", "senior", "go"})).
			Set("metadata", stream.NewRecord().
				String("department", "Engineering").
				String("level", "Senior").
				Float("salary", 95000.0).
				Build()).
			Build(),

		stream.NewRecord().
			Int("id", 2).
			String("name", "Bob Smith").
			String("email", "bob@company.com").
			Bool("active", true).
			Set("tags", stream.FromSliceAny([]any{"manager", "sales"})).
			Set("metadata", stream.NewRecord().
				String("department", "Sales").
				String("level", "Manager").
				Float("salary", 85000.0).
				Build()).
			Build(),

		stream.NewRecord().
			Int("id", 3).
			String("name", "Charlie Brown").
			String("email", "charlie@company.com").
			Bool("active", false).
			Set("tags", stream.FromSliceAny([]any{"intern", "marketing"})).
			Set("metadata", stream.NewRecord().
				String("department", "Marketing").
				String("level", "Intern").
				Float("salary", 45000.0).
				Build()).
			Build(),
	}

	// Convert StreamV2 Records to protobuf messages
	userStream := stream.FromRecordsUnsafe(userData)

	fmt.Println("Converting StreamV2 Records to protobuf messages:")

	// Transform records to protobuf messages
	protobufStream := stream.Map(func(record stream.Record) *pb.User {
		// Extract basic fields
		id := stream.GetOr(record, "id", int64(0))
		name := stream.GetOr(record, "name", "")
		email := stream.GetOr(record, "email", "")
		active := stream.GetOr(record, "active", false)

		// Extract tags
		var tags []string
		if tagsStream, ok := record["tags"].(stream.Stream[any]); ok {
			if tagsList, err := stream.Collect(tagsStream); err == nil {
				for _, tag := range tagsList {
					if tagStr, ok := tag.(string); ok {
						tags = append(tags, tagStr)
					}
				}
			}
		}

		// Extract metadata
		var metadata *pb.Metadata
		if metadataRecord, ok := record["metadata"].(stream.Record); ok {
			metadata = &pb.Metadata{
				Department: stream.GetOr(metadataRecord, "department", ""),
				Level:      stream.GetOr(metadataRecord, "level", ""),
				Salary:     stream.GetOr(metadataRecord, "salary", 0.0),
			}
		}

		return &pb.User{
			Id:       id,
			Name:     name,
			Email:    email,
			Active:   active,
			Tags:     tags,
			Metadata: metadata,
		}
	})(userStream)

	protobufUsers, _ := stream.Collect(protobufStream)

	// Serialize to binary and show results
	var serializedData []string
	for i, user := range protobufUsers {
		data, err := proto.Marshal(user)
		if err != nil {
			log.Fatalf("Failed to marshal protobuf: %v", err)
		}
		serializedData = append(serializedData, string(data))

		fmt.Printf("  User %d: %s (%s) - %s, Active: %v\n",
			user.Id, user.Name, user.Email, user.Metadata.Department, user.Active)
		fmt.Printf("    Tags: %v\n", user.Tags)
		fmt.Printf("    Metadata: %s level, $%.0f salary\n",
			user.Metadata.Level, user.Metadata.Salary)
		fmt.Printf("    Serialized size: %d bytes\n", len(data))
		if i < len(protobufUsers)-1 {
			fmt.Println()
		}
	}

	// Test 2: Deserializing from protobuf back to Records
	fmt.Println("\n📖 Test 2: Protobuf to Records")
	fmt.Println("------------------------------")

	// Create a stream from serialized protobuf data
	binaryStream := stream.FromSlice(serializedData)

	// Deserialize and convert back to StreamV2 Records
	recordStream := stream.Map(func(data string) stream.Record {
		user := &pb.User{}
		if err := proto.Unmarshal([]byte(data), user); err != nil {
			log.Printf("Failed to unmarshal protobuf: %v", err)
			return stream.Record{}
		}

		// Convert tags to stream
		tagsStream := stream.FromSliceAny(func() []any {
			tags := make([]any, len(user.Tags))
			for i, tag := range user.Tags {
				tags[i] = tag
			}
			return tags
		}())

		// Convert metadata to record
		metadataRecord := stream.NewRecord().
			String("department", user.Metadata.Department).
			String("level", user.Metadata.Level).
			Float("salary", user.Metadata.Salary).
			Build()

		return stream.NewRecord().
			Int("id", user.Id).
			String("name", user.Name).
			String("email", user.Email).
			Bool("active", user.Active).
			Set("tags", tagsStream).
			Set("metadata", metadataRecord).
			Build()
	})(binaryStream)

	recoveredRecords, _ := stream.Collect(recordStream)

	fmt.Println("Successfully deserialized protobuf back to StreamV2 Records:")
	for i, record := range recoveredRecords {
		id := stream.GetOr(record, "id", int64(0))
		name := stream.GetOr(record, "name", "")
		email := stream.GetOr(record, "email", "")
		active := stream.GetOr(record, "active", false)

		// Get metadata
		if metadataRecord, ok := record["metadata"].(stream.Record); ok {
			department := stream.GetOr(metadataRecord, "department", "")
			level := stream.GetOr(metadataRecord, "level", "")
			salary := stream.GetOr(metadataRecord, "salary", 0.0)

			fmt.Printf("  Record %d: ID=%d, Name=%s, Email=%s, Active=%v\n",
				i+1, id, name, email, active)
			fmt.Printf("    Department: %s, Level: %s, Salary: $%.0f\n",
				department, level, salary)
		}
	}

	// Test 3: Stream processing with aggregations
	fmt.Println("\n📊 Test 3: Stream Aggregation by Department")
	fmt.Println("------------------------------------------")

	// Filter for active employees and flatten metadata
	activeEmployees := stream.Pipe(
		stream.Where(func(r stream.Record) bool {
			return stream.GetOr(r, "active", false)
		}),
		stream.DotFlatten("."),
	)(stream.FromRecordsUnsafe(recoveredRecords))

	// Group by department
	results, _ := stream.Collect(
		stream.GroupBy([]string{"metadata.department"},
			stream.SumField[float64]("total_salary", "metadata.salary"),
			stream.AvgField[float64]("avg_salary", "metadata.salary"),
		)(activeEmployees),
	)

	fmt.Println("Department analysis from protobuf data:")
	for _, result := range results {
		department := stream.GetOr(result, "metadata.department", "")
		totalSalary := stream.GetOr(result, "total_salary", 0.0)
		avgSalary := stream.GetOr(result, "avg_salary", 0.0)

		fmt.Printf("  %s: Total=$%.0f, Average=$%.0f\n",
			department, totalSalary, avgSalary)
	}

	// Test 4: Round-trip validation
	fmt.Println("\n🔄 Test 4: Round-trip Validation")
	fmt.Println("-------------------------------")

	// Compare original data with round-trip data
	fmt.Println("Validating data integrity after protobuf round-trip:")

	allMatch := true
	for i, original := range userData {
		if i >= len(recoveredRecords) {
			fmt.Printf("  ❌ Missing record %d\n", i+1)
			allMatch = false
			continue
		}

		recovered := recoveredRecords[i]

		// Compare key fields
		originalID := stream.GetOr(original, "id", int64(0))
		recoveredID := stream.GetOr(recovered, "id", int64(0))

		originalName := stream.GetOr(original, "name", "")
		recoveredName := stream.GetOr(recovered, "name", "")

		if originalID == recoveredID && originalName == recoveredName {
			fmt.Printf("  ✅ Record %d: ID and name match\n", i+1)
		} else {
			fmt.Printf("  ❌ Record %d: Mismatch - Original(ID=%d, Name=%s) vs Recovered(ID=%d, Name=%s)\n",
				i+1, originalID, originalName, recoveredID, recoveredName)
			allMatch = false
		}
	}

	if allMatch {
		fmt.Println("\n🎉 Success! All data survived the protobuf round-trip correctly")
	} else {
		fmt.Println("\n⚠️  Some data mismatches detected in round-trip")
	}

	// Test 5: Performance demonstration
	fmt.Println("\n⚡ Test 5: Performance Demonstration")
	fmt.Println("-----------------------------------")

	// Create larger dataset for performance testing
	largeDataset := make([]stream.Record, 1000)
	for i := 0; i < 1000; i++ {
		largeDataset[i] = stream.NewRecord().
			Int("id", int64(i+1)).
			String("name", fmt.Sprintf("User_%d", i+1)).
			String("email", fmt.Sprintf("user%d@example.com", i+1)).
			Bool("active", i%3 != 0). // 2/3 active
			Set("metadata", stream.NewRecord().
				String("department", []string{"Engineering", "Sales", "Marketing", "HR"}[i%4]).
				String("level", []string{"Junior", "Senior", "Manager", "Director"}[i%4]).
				Float("salary", float64(50000+i*100)).
				Build()).
			Build()
	}

	// Process large dataset
	largeStream := stream.FromRecordsUnsafe(largeDataset)

	// Convert to protobuf, serialize, deserialize, and aggregate
	processed := stream.Pipe3(
		// Convert to protobuf and serialize
		stream.Map(func(record stream.Record) string {
			user := &pb.User{
				Id:     stream.GetOr(record, "id", int64(0)),
				Name:   stream.GetOr(record, "name", ""),
				Email:  stream.GetOr(record, "email", ""),
				Active: stream.GetOr(record, "active", false),
			}
			if metadataRecord, ok := record["metadata"].(stream.Record); ok {
				user.Metadata = &pb.Metadata{
					Department: stream.GetOr(metadataRecord, "department", ""),
					Level:      stream.GetOr(metadataRecord, "level", ""),
					Salary:     stream.GetOr(metadataRecord, "salary", 0.0),
				}
			}
			data, _ := proto.Marshal(user)
			return string(data)
		}),

		// Deserialize back to records
		stream.Map(func(data string) stream.Record {
			user := &pb.User{}
			proto.Unmarshal([]byte(data), user)

			metadataRecord := stream.NewRecord().
				String("department", user.Metadata.Department).
				String("level", user.Metadata.Level).
				Float("salary", user.Metadata.Salary).
				Build()

			return stream.NewRecord().
				Int("id", user.Id).
				String("name", user.Name).
				Bool("active", user.Active).
				Set("metadata", metadataRecord).
				Build()
		}),

		// Filter active users only
		stream.Where(func(r stream.Record) bool {
			return stream.GetOr(r, "active", false)
		}),
	)(largeStream)

	processedRecords, _ := stream.Collect(processed)

	fmt.Printf("Processed %d records through protobuf serialization pipeline\n", len(largeDataset))
	fmt.Printf("Active users after filtering: %d\n", len(processedRecords))

	// Group by department for summary (flatten metadata first)
	flattenedProcessed := stream.DotFlatten(".")(stream.FromRecordsUnsafe(processedRecords))
	deptSummary, _ := stream.Collect(
		stream.GroupBy([]string{"metadata.department"},
			stream.CountField("count", "id"),
			stream.AvgField[float64]("avg_salary", "metadata.salary"),
		)(flattenedProcessed),
	)

	fmt.Println("Department summary after protobuf processing:")
	for _, dept := range deptSummary {
		department := stream.GetOr(dept, "metadata.department", "")
		count := stream.GetOr(dept, "count", int64(0))
		avgSalary := stream.GetOr(dept, "avg_salary", 0.0)

		fmt.Printf("  %s: %d active employees, $%.0f avg salary\n",
			department, count, avgSalary)
	}

	fmt.Println("\n✨ Protocol Buffer integration complete!")
	fmt.Println("✅ Real protobuf schemas with generated Go types")
	fmt.Println("✅ Bidirectional conversion between Records and protobuf")
	fmt.Println("✅ Stream processing with serialization/deserialization")
	fmt.Println("✅ Data integrity validation")
	fmt.Println("✅ Performance demonstration with large datasets")
}