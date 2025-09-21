package main

import (
	"fmt"
	"time"
	"github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
	fmt.Println("🕒 StreamV2 Event-Time Windowing with Watermarks")
	fmt.Println("================================================")

	demonstrateOutOfOrderEvents()
	demonstrateSlidingWindows()
	demonstrateSessionWindows()
	demonstrateWatermarkStrategies()
	demonstrateLateDataHandling()
}

// ============================================================================
// OUT-OF-ORDER EVENT PROCESSING
// ============================================================================

func demonstrateOutOfOrderEvents() {
	fmt.Println("\n📅 Test 1: Out-of-Order Event Processing")
	fmt.Println("----------------------------------------")

	baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)
	
	// Simulate financial trades arriving out of order
	trades := []stream.Record{
		stream.NewRecord().String("symbol", "AAPL").Float("price", 150.25).String("timestamp", baseTime.Add(30*time.Second).Format(time.RFC3339)).Int("volume", 100).Build(),
		stream.NewRecord().String("symbol", "AAPL").Float("price", 149.50).String("timestamp", baseTime.Format(time.RFC3339)).Int("volume", 200).Build(), // Late arrival!
		stream.NewRecord().String("symbol", "AAPL").Float("price", 151.00).String("timestamp", baseTime.Add(90*time.Second).Format(time.RFC3339)).Int("volume", 150).Build(),
		stream.NewRecord().String("symbol", "AAPL").Float("price", 150.75).String("timestamp", baseTime.Add(45*time.Second).Format(time.RFC3339)).Int("volume", 75).Build(), // Late arrival!
		stream.NewRecord().String("symbol", "AAPL").Float("price", 152.00).String("timestamp", baseTime.Add(150*time.Second).Format(time.RFC3339)).Int("volume", 300).Build(),
	}

	fmt.Printf("Processing %d trades (some arriving out of order):\n", len(trades))
	for i, trade := range trades {
		timestamp := stream.GetOr(trade, "timestamp", "")
		price := stream.GetOr(trade, "price", 0.0)
		volume := stream.GetOr(trade, "volume", int64(0))
		fmt.Printf("  Trade %d: $%.2f (%d shares) at %s\n", i+1, price, volume, timestamp)
	}

	// Create 1-minute event-time windows with 30-second watermark lateness
	windowedTrades := stream.EventTimeTumblingWindow(
		1*time.Minute,
		stream.WithTimestampExtractor(stream.NewRecordTimestampExtractor("timestamp")),
		stream.WithAllowedLateness(30*time.Second),
		stream.WithLateDataPolicy(stream.DropLateData),
	)(stream.FromSlice(trades))

	fmt.Println("\n📊 1-Minute Trading Windows (by event time):")
	windowNum := 1
	for {
		window, err := windowedTrades()
		if err == stream.EOS {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}

		windowTrades, _ := stream.Collect(window)
		if len(windowTrades) == 0 {
			continue
		}

		// Calculate window statistics
		var totalVolume int64
		var weightedPrice float64
		var totalValue float64
		var minPrice, maxPrice float64
		var windowStart, windowEnd time.Time

		for i, trade := range windowTrades {
			price := stream.GetOr(trade, "price", 0.0)
			volume := stream.GetOr(trade, "volume", int64(0))
			timestamp := stream.GetOr(trade, "timestamp", "")
			
			if tradeTime, err := time.Parse(time.RFC3339, timestamp); err == nil {
				if i == 0 {
					windowStart = tradeTime.Truncate(1 * time.Minute)
					windowEnd = windowStart.Add(1 * time.Minute)
					minPrice = price
					maxPrice = price
				} else {
					if price < minPrice {
						minPrice = price
					}
					if price > maxPrice {
						maxPrice = price
					}
				}
			}

			totalVolume += volume
			totalValue += price * float64(volume)
		}

		if totalVolume > 0 {
			weightedPrice = totalValue / float64(totalVolume)
		}

		fmt.Printf("  Window %d [%s - %s):\n", windowNum, 
			windowStart.Format("15:04:05"), windowEnd.Format("15:04:05"))
		fmt.Printf("    Trades: %d\n", len(windowTrades))
		fmt.Printf("    Volume: %d shares\n", totalVolume)
		fmt.Printf("    Price Range: $%.2f - $%.2f\n", minPrice, maxPrice)
		fmt.Printf("    Volume-Weighted Price: $%.2f\n", weightedPrice)
		fmt.Printf("    Total Value: $%.2f\n", totalValue)

		windowNum++
	}
}

// ============================================================================
// SLIDING WINDOWS
// ============================================================================

func demonstrateSlidingWindows() {
	fmt.Println("\n🔄 Test 2: Sliding Windows for Moving Averages")
	fmt.Println("----------------------------------------------")

	baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)
	
	// IoT sensor readings every 30 seconds
	sensorData := []stream.Record{
		stream.NewRecord().String("sensor", "temp-01").Float("value", 23.5).String("timestamp", baseTime.Format(time.RFC3339)).Build(),
		stream.NewRecord().String("sensor", "temp-01").Float("value", 24.1).String("timestamp", baseTime.Add(30*time.Second).Format(time.RFC3339)).Build(),
		stream.NewRecord().String("sensor", "temp-01").Float("value", 23.8).String("timestamp", baseTime.Add(60*time.Second).Format(time.RFC3339)).Build(),
		stream.NewRecord().String("sensor", "temp-01").Float("value", 25.2).String("timestamp", baseTime.Add(90*time.Second).Format(time.RFC3339)).Build(),
		stream.NewRecord().String("sensor", "temp-01").Float("value", 24.7).String("timestamp", baseTime.Add(120*time.Second).Format(time.RFC3339)).Build(),
		stream.NewRecord().String("sensor", "temp-01").Float("value", 23.9).String("timestamp", baseTime.Add(150*time.Second).Format(time.RFC3339)).Build(),
	}

	fmt.Printf("Sensor readings every 30 seconds:\n")
	for _, reading := range sensorData {
		timestamp := stream.GetOr(reading, "timestamp", "")
		value := stream.GetOr(reading, "value", 0.0)
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			fmt.Printf("  %s: %.1f°C\n", t.Format("15:04:05"), value)
		}
	}

	// Create 2-minute sliding windows that slide every 30 seconds
	slidingWindows := stream.EventTimeSlidingWindow(
		2*time.Minute,  // Window size
		30*time.Second, // Slide interval
		stream.WithTimestampExtractor(stream.NewRecordTimestampExtractor("timestamp")),
		stream.WithAllowedLateness(10*time.Second),
	)(stream.FromSlice(sensorData))

	fmt.Println("\n📈 2-Minute Sliding Windows (30s slide):")
	windowNum := 1
	for {
		window, err := slidingWindows()
		if err == stream.EOS {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}

		readings, _ := stream.Collect(window)
		if len(readings) == 0 {
			continue
		}

		// Calculate moving average
		var sum float64
		var count int
		var windowStart time.Time

		for i, reading := range readings {
			value := stream.GetOr(reading, "value", 0.0)
			timestamp := stream.GetOr(reading, "timestamp", "")
			
			if i == 0 {
				if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
					windowStart = t.Truncate(30 * time.Second)
				}
			}
			
			sum += value
			count++
		}

		average := sum / float64(count)
		windowEnd := windowStart.Add(2 * time.Minute)

		fmt.Printf("  Window %d [%s - %s): %.1f°C avg (%d readings)\n", 
			windowNum, windowStart.Format("15:04:05"), windowEnd.Format("15:04:05"), 
			average, count)

		windowNum++
	}
}

// ============================================================================
// SESSION WINDOWS
// ============================================================================

func demonstrateSessionWindows() {
	fmt.Println("\n🔗 Test 3: Session Windows for User Activity")
	fmt.Println("--------------------------------------------")

	baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)
	
	// User activity events with gaps
	userEvents := []stream.Record{
		stream.NewRecord().String("user", "alice").String("action", "login").String("timestamp", baseTime.Format(time.RFC3339)).Build(),
		stream.NewRecord().String("user", "alice").String("action", "view_page").String("timestamp", baseTime.Add(30*time.Second).Format(time.RFC3339)).Build(),
		stream.NewRecord().String("user", "alice").String("action", "click_link").String("timestamp", baseTime.Add(45*time.Second).Format(time.RFC3339)).Build(),
		stream.NewRecord().String("user", "alice").String("action", "view_page").String("timestamp", baseTime.Add(60*time.Second).Format(time.RFC3339)).Build(),
		// 5-minute gap (session timeout)
		stream.NewRecord().String("user", "alice").String("action", "login").String("timestamp", baseTime.Add(6*time.Minute).Format(time.RFC3339)).Build(),
		stream.NewRecord().String("user", "alice").String("action", "search").String("timestamp", baseTime.Add(6*time.Minute+30*time.Second).Format(time.RFC3339)).Build(),
		stream.NewRecord().String("user", "alice").String("action", "logout").String("timestamp", baseTime.Add(7*time.Minute).Format(time.RFC3339)).Build(),
	}

	fmt.Printf("User activity events:\n")
	for _, event := range userEvents {
		timestamp := stream.GetOr(event, "timestamp", "")
		action := stream.GetOr(event, "action", "")
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			fmt.Printf("  %s: %s\n", t.Format("15:04:05"), action)
		}
	}

	// Create session windows with 2-minute timeout
	sessionWindows := stream.EventTimeSessionWindow(
		2*time.Minute, // Session timeout
		stream.WithTimestampExtractor(stream.NewRecordTimestampExtractor("timestamp")),
		stream.WithAllowedLateness(30*time.Second),
	)(stream.FromSlice(userEvents))

	fmt.Println("\n👤 User Sessions (2-minute timeout):")
	sessionNum := 1
	for {
		session, err := sessionWindows()
		if err == stream.EOS {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}

		sessionEvents, _ := stream.Collect(session)
		if len(sessionEvents) == 0 {
			continue
		}

		// Analyze session
		var sessionStart, sessionEnd time.Time
		actions := make([]string, 0, len(sessionEvents))

		for i, event := range sessionEvents {
			timestamp := stream.GetOr(event, "timestamp", "")
			action := stream.GetOr(event, "action", "")
			actions = append(actions, action)

			if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
				if i == 0 {
					sessionStart = t
				}
				sessionEnd = t
			}
		}

		duration := sessionEnd.Sub(sessionStart)
		
		fmt.Printf("  Session %d: %s - %s (%.0f sec)\n", 
			sessionNum, sessionStart.Format("15:04:05"), sessionEnd.Format("15:04:05"), duration.Seconds())
		fmt.Printf("    Events: %d\n", len(sessionEvents))
		fmt.Printf("    Actions: %v\n", actions)

		sessionNum++
	}
}

// ============================================================================
// WATERMARK STRATEGIES
// ============================================================================

func demonstrateWatermarkStrategies() {
	fmt.Println("\n💧 Test 4: Different Watermark Strategies")
	fmt.Println("-----------------------------------------")

	baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)

	// Test different watermark strategies
	fmt.Println("Watermark Strategy Comparison:")

	// Conservative watermark (allows more lateness)
	conservativeWM := stream.BoundedOutOfOrdernessWatermark(60 * time.Second)
	conservativeResult := conservativeWM(baseTime.Add(2 * time.Minute))
	fmt.Printf("  Conservative (60s lateness): Event at %s → Watermark at %s\n", 
		baseTime.Add(2*time.Minute).Format("15:04:05"), conservativeResult.Format("15:04:05"))

	// Aggressive watermark (allows less lateness)
	aggressiveWM := stream.BoundedOutOfOrdernessWatermark(10 * time.Second)
	aggressiveResult := aggressiveWM(baseTime.Add(2 * time.Minute))
	fmt.Printf("  Aggressive (10s lateness):   Event at %s → Watermark at %s\n", 
		baseTime.Add(2*time.Minute).Format("15:04:05"), aggressiveResult.Format("15:04:05"))

	// Periodic watermark (updates every 30 seconds)
	periodicWM := stream.PeriodicWatermarkGenerator(30*time.Second, 
		stream.BoundedOutOfOrdernessWatermark(30*time.Second))
	periodicResult := periodicWM(baseTime.Add(2 * time.Minute))
	
	fmt.Printf("  Periodic (30s intervals):   Event at %s → Watermark at %s\n",
		baseTime.Add(2*time.Minute).Format("15:04:05"), periodicResult.Format("15:04:05"))

	fmt.Println("\n💡 Watermark Trade-offs:")
	fmt.Println("  • Conservative: Fewer dropped events, higher latency")
	fmt.Println("  • Aggressive: Lower latency, more dropped late events")
	fmt.Println("  • Periodic: Reduces computation overhead for high-volume streams")
}

// ============================================================================
// LATE DATA HANDLING
// ============================================================================

func demonstrateLateDataHandling() {
	fmt.Println("\n⏰ Test 5: Late Data Handling Policies")
	fmt.Println("-------------------------------------")

	baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)
	
	// Events that will trigger window firing, then late data
	eventsWithLateData := []stream.Record{
		stream.NewRecord().String("id", "1").String("timestamp", baseTime.Add(70*time.Second).Format(time.RFC3339)).Int("value", 100).Build(), // Triggers first window
		stream.NewRecord().String("id", "2").String("timestamp", baseTime.Add(30*time.Second).Format(time.RFC3339)).Int("value", 200).Build(), // Late for first window
		stream.NewRecord().String("id", "3").String("timestamp", baseTime.Add(130*time.Second).Format(time.RFC3339)).Int("value", 300).Build(), // Triggers second window
	}

	fmt.Println("Testing DROP late data policy:")
	
	windowedStream := stream.EventTimeTumblingWindow(
		1*time.Minute,
		stream.WithTimestampExtractor(stream.NewRecordTimestampExtractor("timestamp")),
		stream.WithAllowedLateness(15*time.Second), // Short lateness window
		stream.WithLateDataPolicy(stream.DropLateData),
	)(stream.FromSlice(eventsWithLateData))

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

		windowEvents, _ := stream.Collect(window)
		if len(windowEvents) == 0 {
			continue
		}

		fmt.Printf("  Window %d: %d events (late events dropped)\n", windowNum, len(windowEvents))
		for _, event := range windowEvents {
			id := stream.GetOr(event, "id", "")
			value := stream.GetOr(event, "value", int64(0))
			fmt.Printf("    Event %s: value=%d\n", id, value)
		}

		windowNum++
	}

	fmt.Println("\n📋 Summary:")
	fmt.Println("  ✅ Event-time windowing handles out-of-order data correctly")
	fmt.Println("  ✅ Sliding windows provide overlapping analysis periods")
	fmt.Println("  ✅ Session windows group related activities automatically")
	fmt.Println("  ✅ Configurable watermark strategies balance latency vs completeness")
	fmt.Println("  ✅ Late data policies provide flexible handling of delayed events")

	fmt.Println("\n🎉 Event-Time Windowing Complete!")
	fmt.Println("🚀 StreamV2 now supports enterprise-grade event-time processing!")
}