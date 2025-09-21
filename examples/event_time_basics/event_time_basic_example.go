package main

import (
	"fmt"
	"time"
	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Println("🕒 StreamV2 Event-Time Processing Foundation")
	fmt.Println("============================================")

	demonstrateTimestampExtraction()
	demonstrateWatermarkGeneration()
}

func demonstrateTimestampExtraction() {
	fmt.Println("\n📅 Test 1: Timestamp Extraction")
	fmt.Println("-------------------------------")

	baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)
	
	// Create test records with timestamps
	records := []stream.Record{
		stream.NewRecord().String("id", "1").String("timestamp", baseTime.Format(time.RFC3339)).Float("value", 10.5).Build(),
		stream.NewRecord().String("id", "2").Set("timestamp", baseTime.Add(30*time.Second)).Float("value", 20.7).Build(), // time.Time type
		stream.NewRecord().String("id", "3").String("timestamp", baseTime.Add(60*time.Second).Format(time.RFC3339)).Float("value", 15.2).Build(),
	}

	// Create timestamp extractor
	extractor := stream.NewRecordTimestampExtractor("timestamp")

	fmt.Println("Extracting timestamps from records:")
	for _, record := range records {
		id := stream.GetOr(record, "id", "")
		value := stream.GetOr(record, "value", 0.0)
		extractedTime := extractor(record)
		
		fmt.Printf("  Record %s: value=%.1f, event_time=%s\n", 
			id, value, extractedTime.Format("15:04:05"))
	}

	// Test with various timestamp formats
	fmt.Println("\nTesting different timestamp formats:")
	testFormats := []stream.Record{
		stream.NewRecord().String("format", "RFC3339").String("timestamp", baseTime.Format(time.RFC3339)).Build(),
		stream.NewRecord().String("format", "time.Time").Set("timestamp", baseTime).Build(),
		stream.NewRecord().String("format", "missing").String("other_field", "value").Build(), // No timestamp - should use current time
	}

	for _, record := range testFormats {
		format := stream.GetOr(record, "format", "")
		extractedTime := extractor(record)
		fmt.Printf("  %s format: %s\n", format, extractedTime.Format("15:04:05"))
	}
}

func demonstrateWatermarkGeneration() {
	fmt.Println("\n💧 Test 2: Watermark Generation")
	fmt.Println("-------------------------------")

	baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)

	// Test different watermark strategies
	fmt.Println("Testing watermark generation strategies:")

	// Conservative: Allow 60 seconds of lateness
	conservative := stream.BoundedOutOfOrdernessWatermark(60 * time.Second)
	wm1 := conservative(baseTime.Add(2 * time.Minute))
	fmt.Printf("  Conservative (60s lateness): Event at %s → Watermark at %s\n",
		baseTime.Add(2*time.Minute).Format("15:04:05"), wm1.Format("15:04:05"))

	// Aggressive: Allow only 10 seconds of lateness
	aggressive := stream.BoundedOutOfOrdernessWatermark(10 * time.Second)
	wm2 := aggressive(baseTime.Add(2 * time.Minute))
	fmt.Printf("  Aggressive (10s lateness):   Event at %s → Watermark at %s\n",
		baseTime.Add(2*time.Minute).Format("15:04:05"), wm2.Format("15:04:05"))

	// Real-time simulation
	fmt.Println("\nSimulating watermark progression:")
	moderateWM := stream.BoundedOutOfOrdernessWatermark(30 * time.Second)
	
	eventTimes := []time.Time{
		baseTime,
		baseTime.Add(30 * time.Second),
		baseTime.Add(45 * time.Second), // Slightly out of order
		baseTime.Add(90 * time.Second),
		baseTime.Add(2 * time.Minute),
	}

	for i, eventTime := range eventTimes {
		watermark := moderateWM(eventTime)
		lateness := eventTime.Sub(watermark)
		fmt.Printf("  Event %d at %s → Watermark %s (%.0fs lateness window)\n",
			i+1, eventTime.Format("15:04:05"), watermark.Format("15:04:05"), lateness.Seconds())
	}

	fmt.Println("\n💡 Key Concepts:")
	fmt.Println("  • Timestamp Extraction: Gets event time from data")
	fmt.Println("  • Watermarks: Signal progress in event time")
	fmt.Println("  • Lateness Allowance: Trade-off between completeness and latency")
	fmt.Println("  • Foundation: Ready for window trigger implementation")

	fmt.Println("\n🎉 Event-Time Processing Foundation Complete!")
	fmt.Println("🚀 StreamV2 now has the building blocks for advanced windowing!")
	
	demonstrateEventTimeWindowing()
}

func demonstrateEventTimeWindowing() {
	fmt.Println("\n🪟 Test 3: Event-Time Tumbling Windows")
	fmt.Println("------------------------------------")

	baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)
	
	// Create test data with explicit timestamps (some out of order)
	testData := []stream.Record{
		stream.NewRecord().String("id", "1").String("event_time", baseTime.Format(time.RFC3339)).Int("value", 10).Build(),
		stream.NewRecord().String("id", "2").String("event_time", baseTime.Add(30*time.Second).Format(time.RFC3339)).Int("value", 20).Build(),
		stream.NewRecord().String("id", "3").String("event_time", baseTime.Add(90*time.Second).Format(time.RFC3339)).Int("value", 30).Build(),
		stream.NewRecord().String("id", "4").String("event_time", baseTime.Add(45*time.Second).Format(time.RFC3339)).Int("value", 25).Build(), // Out of order!
		stream.NewRecord().String("id", "5").String("event_time", baseTime.Add(150*time.Second).Format(time.RFC3339)).Int("value", 40).Build(),
	}

	fmt.Printf("Processing %d records with event-time windows:\n", len(testData))
	for _, record := range testData {
		id := stream.GetOr(record, "id", "")
		timestamp := stream.GetOr(record, "event_time", "")
		value := stream.GetOr(record, "value", int64(0))
		fmt.Printf("  Record %s: value=%d, event_time=%s\n", id, value, timestamp)
	}

	// Create 1-minute tumbling windows with 10-second lateness allowance
	windowedStream := stream.EventTimeTumblingWindow(
		1*time.Minute,
		stream.WithTimestampExtractor(stream.NewRecordTimestampExtractor("event_time")),
		stream.WithAllowedLateness(10*time.Second),
		stream.WithLateDataPolicy(stream.DropLateData),
	)(stream.FromSlice(testData))

	fmt.Println("\n📊 1-Minute Tumbling Windows (by event time):")
	windowNum := 1
	for {
		window, err := windowedStream()
		if err == stream.EOS {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}

		windowRecords, _ := stream.Collect(window)
		if len(windowRecords) == 0 {
			continue
		}

		// Calculate window statistics
		var totalValue int64
		var windowStart time.Time
		firstRecord := true

		for _, record := range windowRecords {
			value := stream.GetOr(record, "value", int64(0))
			timestamp := stream.GetOr(record, "event_time", "")
			
			if eventTime, err := time.Parse(time.RFC3339, timestamp); err == nil && firstRecord {
				windowStart = eventTime.Truncate(1 * time.Minute)
				firstRecord = false
			}

			totalValue += value
		}

		windowEnd := windowStart.Add(1 * time.Minute)
		fmt.Printf("  Window %d [%s - %s): %d records, total value=%d\n", 
			windowNum, windowStart.Format("15:04:05"), windowEnd.Format("15:04:05"), 
			len(windowRecords), totalValue)

		for _, record := range windowRecords {
			id := stream.GetOr(record, "id", "")
			value := stream.GetOr(record, "value", int64(0))
			fmt.Printf("    Record %s: value=%d\n", id, value)
		}

		windowNum++
	}

	fmt.Println("\n✨ Key Features Demonstrated:")
	fmt.Println("  • Out-of-order events handled correctly by event time")
	fmt.Println("  • Watermark-based window triggering")
	fmt.Println("  • Configurable lateness policies")
	fmt.Println("  • Production-ready windowing system")
}