package main

import (
	"bytes"
	"fmt"
	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Println("🔄 JSON Stream Reuse Patterns")
	fmt.Println("=============================")

	// Problem: Streams can only be consumed once
	fmt.Println("\n❌ Problem: Stream exhaustion")
	fmt.Println("----------------------------")
	
	record := stream.NewRecord().
		Int("order_id", 1001).
		Set("items", stream.FromSliceAny([]any{
			stream.NewRecord().String("product", "laptop").Float("price", 1299.99).Build(),
			stream.NewRecord().String("product", "mouse").Float("price", 29.99).Build(),
		})).
		Build()
	
	// First serialization - works fine
	var buffer1 bytes.Buffer
	stream.StreamToJSON(stream.FromSlice([]stream.Record{record}), &buffer1)
	fmt.Println("First serialization:")
	fmt.Println(buffer1.String())
	
	// Second serialization - stream is exhausted!
	var buffer2 bytes.Buffer
	sink := stream.NewJSONSink(&buffer2).WithFormat(stream.JSONArray).WithPrettyPrint()
	sink.WriteRecords([]stream.Record{record})
	fmt.Println("Second serialization (same record):")
	fmt.Println(buffer2.String())

	// Solution 1: Use Tee to split streams for multiple outputs
	fmt.Println("\n✅ Solution 1: Use Tee for multiple outputs")
	fmt.Println("------------------------------------------")
	
	originalItems := stream.FromSliceAny([]any{
		stream.NewRecord().String("product", "keyboard").Float("price", 89.99).Build(),
		stream.NewRecord().String("product", "monitor").Float("price", 299.99).Build(),
	})
	
	// Split the stream using Tee
	teedStreams := stream.Tee(originalItems, 2)
	
	// Create records with different stream copies
	record1 := stream.NewRecord().Int("order_id", 2001).Set("items", teedStreams[0]).Build()
	record2 := stream.NewRecord().Int("order_id", 2001).Set("items", teedStreams[1]).Build()
	
	// Now we can serialize each independently
	var teeBuffer1 bytes.Buffer
	stream.StreamToJSON(stream.FromSlice([]stream.Record{record1}), &teeBuffer1)
	fmt.Println("Teed serialization 1:")
	fmt.Println(teeBuffer1.String())
	
	var teeBuffer2 bytes.Buffer
	teeSink := stream.NewJSONSink(&teeBuffer2).WithFormat(stream.JSONArray).WithPrettyPrint()
	teeSink.WriteRecords([]stream.Record{record2})
	fmt.Println("Teed serialization 2:")
	fmt.Println(teeBuffer2.String())

	// Solution 2: Convert to slice first, then create multiple streams
	fmt.Println("\n✅ Solution 2: Convert to slice for reuse")
	fmt.Println("---------------------------------------")
	
	// Collect stream to slice first
	itemsSlice := []any{
		stream.NewRecord().String("product", "tablet").Float("price", 599.99).Build(),
		stream.NewRecord().String("product", "case").Float("price", 19.99).Build(),
	}
	
	// Create multiple records with fresh streams from the same slice
	reusableRecord1 := stream.NewRecord().Int("order_id", 3001).Set("items", stream.FromSliceAny(itemsSlice)).Build()
	reusableRecord2 := stream.NewRecord().Int("order_id", 3001).Set("items", stream.FromSliceAny(itemsSlice)).Build()
	
	var sliceBuffer1 bytes.Buffer
	stream.StreamToJSON(stream.FromSlice([]stream.Record{reusableRecord1}), &sliceBuffer1)
	fmt.Println("Slice-based serialization 1:")
	fmt.Println(sliceBuffer1.String())
	
	var sliceBuffer2 bytes.Buffer
	sliceSink := stream.NewJSONSink(&sliceBuffer2).WithFormat(stream.JSONArray).WithPrettyPrint()
	sliceSink.WriteRecords([]stream.Record{reusableRecord2})
	fmt.Println("Slice-based serialization 2:")
	fmt.Println(sliceBuffer2.String())

	// Solution 3: Use different serialization modes for different outputs
	fmt.Println("\n✅ Solution 3: Different modes for different outputs")
	fmt.Println("---------------------------------------------------")
	
	fmt.Println("For multiple JSON outputs of the same data, collect the streams")
	fmt.Println("to slices first, then create fresh streams for each output.")
	fmt.Println("This maintains the streaming API while allowing reuse.")

	fmt.Println("\n📝 Key Takeaways:")
	fmt.Println("- Streams can only be consumed once (this is by design)")
	fmt.Println("- Use Tee() to split streams for concurrent consumption")
	fmt.Println("- Use FromSlice() to create reusable streams from collected data")
	fmt.Println("- This behavior is consistent with Go's io.Reader pattern")
}