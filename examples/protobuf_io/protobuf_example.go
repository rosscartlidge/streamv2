package main

import (
	"bytes"
	"fmt"
	"github.com/rosscartlidge/streamv2/pkg/stream"
	
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/proto"
)

func main() {
	fmt.Println("🔧 Protocol Buffer I/O with StreamV2")
	fmt.Println("====================================")

	// For this example, we'll create a simple protobuf schema dynamically
	// In real applications, you'd typically have generated .pb.go files from .proto files
	
	fmt.Println("\n🏗️ Creating dynamic protobuf schema")
	fmt.Println("----------------------------------")
	
	// Create a simple User message schema dynamically
	userSchema := createUserMessageDescriptor()
	fmt.Printf("Created schema for message: %s\n", userSchema.Name())
	fmt.Printf("Fields: %d\n", userSchema.Fields().Len())
	
	// Test 1: Creating Records and converting to protobuf
	fmt.Println("\n📝 Test 1: Records to Protobuf")
	fmt.Println("------------------------------")
	
	// Create sample user data using StreamV2 Records
	userData := []stream.Record{
		stream.NewRecord().
			Int("id", 1001).
			String("name", "Alice Johnson").
			String("email", "alice@example.com").
			Bool("active", true).
			Set("tags", stream.FromSliceAny([]any{"developer", "golang", "streaming"})).
			Set("metadata", stream.NewRecord().String("department", "engineering").String("level", "senior").Build()).
			Build(),
		stream.NewRecord().
			Int("id", 1002).
			String("name", "Bob Smith").
			String("email", "bob@example.com").
			Bool("active", false).
			Set("tags", stream.FromSliceAny([]any{"analyst", "data"})).
			Set("metadata", stream.NewRecord().String("department", "analytics").String("level", "junior").Build()).
			Build(),
	}
	
	fmt.Printf("Created %d user records\n", len(userData))
	for i, record := range userData {
		name := stream.GetOr(record, "name", "")
		email := stream.GetOr(record, "email", "")
		fmt.Printf("  %d: %s (%s)\n", i+1, name, email)
	}
	
	// Test 2: Serialize to protobuf binary format
	fmt.Println("\n💾 Test 2: Binary protobuf serialization")
	fmt.Println("---------------------------------------")
	
	var protobufBuffer bytes.Buffer
	err := stream.StreamToProtobuf(stream.FromSlice(userData), &protobufBuffer, userSchema)
	if err != nil {
		fmt.Printf("Error serializing to protobuf: %v\n", err)
		return
	}
	
	fmt.Printf("Serialized to binary protobuf: %d bytes\n", protobufBuffer.Len())
	
	// Test 3: Deserialize from protobuf binary
	fmt.Println("\n📖 Test 3: Binary protobuf deserialization")
	fmt.Println("-----------------------------------------")
	
	protobufStream := stream.ProtobufToStream(&protobufBuffer, userSchema)
	deserializedRecords, _ := stream.Collect(protobufStream)
	
	fmt.Printf("Deserialized %d records from protobuf\n", len(deserializedRecords))
	for i, record := range deserializedRecords {
		name := stream.GetOr(record, "name", "")
		id := stream.GetOr(record, "id", int64(0))
		active := stream.GetOr(record, "active", false)
		fmt.Printf("  %d: ID=%d, Name=%s, Active=%v\n", i+1, id, name, active)
		
		// Show tags (converted from repeated field to Stream[any])
		if tagsStream, ok := stream.Get[stream.Stream[any]](record, "tags"); ok {
			tags, _ := stream.Collect(tagsStream)
			fmt.Printf("      Tags: %v\n", tags)
		}
		
		// Show metadata (converted from map field to Record)
		if metadata, ok := stream.Get[stream.Record](record, "metadata"); ok {
			dept := stream.GetOr(metadata, "department", "")
			level := stream.GetOr(metadata, "level", "")
			fmt.Printf("      Metadata: dept=%s, level=%s\n", dept, level)
		}
	}
	
	// Test 4: Protobuf JSON format
	fmt.Println("\n🎨 Test 4: Protobuf JSON format")
	fmt.Println("-------------------------------")
	
	var jsonBuffer bytes.Buffer
	jsonSink := stream.NewProtobufSink(&jsonBuffer, userSchema).WithFormat(stream.ProtobufJSON)
	err = jsonSink.WriteRecords(userData)
	if err != nil {
		fmt.Printf("Error serializing to protobuf JSON: %v\n", err)
		return
	}
	
	fmt.Println("Protobuf JSON output:")
	fmt.Println(jsonBuffer.String())
	
	// Test 5: Read protobuf JSON back
	fmt.Println("\n🔄 Test 5: Protobuf JSON round-trip")
	fmt.Println("----------------------------------")
	
	jsonSource := stream.NewProtobufSource(&jsonBuffer, userSchema).WithFormat(stream.ProtobufJSON)
	jsonRecords, _ := stream.Collect(jsonSource.ToStream())
	
	fmt.Printf("Read back %d records from protobuf JSON\n", len(jsonRecords))
	
	// Test 6: Integration with StreamV2 operations
	fmt.Println("\n⚙️ Test 6: StreamV2 integration")
	fmt.Println("------------------------------")
	
	// Create a larger dataset
	largeDataset := make([]stream.Record, 10)
	departments := []string{"engineering", "analytics", "marketing", "sales"}
	
	for i := 0; i < 10; i++ {
		largeDataset[i] = stream.NewRecord().
			Int("id", int64(2000+i)).
			String("name", fmt.Sprintf("User%d", i)).
			String("email", fmt.Sprintf("user%d@example.com", i)).
			Bool("active", i%2 == 0).
			Set("metadata", stream.NewRecord().
				String("department", departments[i%len(departments)]).
				Float("salary", float64(50000 + i*5000)).
				Build()).
			Build()
	}
	
	// Serialize to protobuf, then read back and process
	var largeBuffer bytes.Buffer
	stream.StreamToProtobuf(stream.FromSlice(largeDataset), &largeBuffer, userSchema)
	
	// Read back and group by department
	reloadedStream := stream.ProtobufToStream(&largeBuffer, userSchema)
	
	// Extract department from nested metadata and group
	departmentStream := stream.Map(func(record stream.Record) stream.Record {
		id := stream.GetOr(record, "id", int64(0))
		name := stream.GetOr(record, "name", "")
		active := stream.GetOr(record, "active", false)
		
		var department string
		var salary float64
		if metadata, ok := stream.Get[stream.Record](record, "metadata"); ok {
			department = stream.GetOr(metadata, "department", "")
			salary = stream.GetOr(metadata, "salary", 0.0)
		}
		
		return stream.NewRecord().
			Int("id", id).
			String("name", name).
			String("department", department).
			Float("salary", salary).
			Bool("active", active).
			Build()
	})(reloadedStream)
	
	// Group by department
	results, _ := stream.Collect(
		stream.GroupBy([]string{"department"},
			stream.SumField[float64]("total_salary", "salary"),
			stream.AvgField[float64]("avg_salary", "salary"),
		)(departmentStream),
	)
	
	fmt.Println("Department analysis from protobuf data:")
	for _, result := range results {
		dept := stream.GetOr(result, "department", "")
		count := stream.GetOr(result, "group_count", int64(0))
		avgSalary := stream.GetOr(result, "avg_salary", 0.0)
		fmt.Printf("  %s: %d employees, avg salary $%.0f\n", dept, count, avgSalary)
	}
	
	// Test 7: Performance comparison
	fmt.Println("\n⚡ Test 7: Performance comparison")
	fmt.Println("--------------------------------")
	
	// Compare sizes: JSON vs Protobuf
	var jsonCompareBuffer bytes.Buffer
	stream.StreamToJSON(stream.FromSlice(largeDataset), &jsonCompareBuffer)
	
	var protobufCompareBuffer bytes.Buffer  
	stream.StreamToProtobuf(stream.FromSlice(largeDataset), &protobufCompareBuffer, userSchema)
	
	fmt.Printf("JSON size: %d bytes\n", jsonCompareBuffer.Len())
	fmt.Printf("Protobuf size: %d bytes\n", protobufCompareBuffer.Len())
	fmt.Printf("Compression ratio: %.1fx smaller\n", float64(jsonCompareBuffer.Len())/float64(protobufCompareBuffer.Len()))
	
	fmt.Println("\n🎉 Protocol Buffer I/O Complete!")
	fmt.Println("✨ High-performance binary serialization with StreamV2!")
}

// createUserMessageDescriptor creates a simple User message descriptor dynamically
// In real applications, you'd use generated descriptors from .proto files
func createUserMessageDescriptor() protoreflect.MessageDescriptor {
	// This is a simplified example - in practice you'd load from .proto files
	// or use generated descriptors
	
	// Create a basic file descriptor
	_ = &descriptorpb.FileDescriptorProto{
		Name:    proto.String("user.proto"),
		Package: proto.String("example"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("User"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("id"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
					},
					{
						Name:   proto.String("name"), 
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:   proto.String("email"),
						Number: proto.Int32(3),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:   proto.String("active"),
						Number: proto.Int32(4),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum(),
					},
					{
						Name:   proto.String("tags"),
						Number: proto.Int32(5),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
					},
					{
						Name:     proto.String("metadata"),
						Number:   proto.Int32(6),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".example.Metadata"),
					},
				},
			},
			{
				Name: proto.String("Metadata"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("department"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:   proto.String("level"),
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:   proto.String("salary"),
						Number: proto.Int32(3),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum(),
					},
				},
			},
		},
	}
	
	// This would normally be done by protoc compiler or runtime descriptor building
	// For this example, we'll return a mock descriptor that has the basic structure
	// In practice, you'd use protodesc.NewFile() with proper dependency resolution
	
	// Return a simple mock - this is just for demonstration
	// Real implementation would build proper descriptors
	return createMockUserDescriptor()
}

// Mock descriptor for demonstration - in practice use real protobuf descriptors
func createMockUserDescriptor() protoreflect.MessageDescriptor {
	// This is a placeholder implementation
	// In real applications, you would:
	// 1. Generate .pb.go files from .proto files using protoc
	// 2. Use the generated descriptors: someMessage{}.ProtoReflect().Descriptor()
	// 3. Or use protodesc.NewFile() to build descriptors dynamically
	
	fmt.Println("Note: Using simplified mock descriptor for demonstration")
	fmt.Println("In production, use generated .pb.go files or protodesc.NewFile()")
	
	// For now, return nil - the example will show the API structure
	// without full protobuf functionality
	return nil
}