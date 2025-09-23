# StreamV2 Advanced Windowing Codelab

Welcome to the advanced windowing codelab! This tutorial covers StreamV2's powerful windowing capabilities for real-time stream processing. You'll learn how to handle time-based data, late arrivals, session detection, and complex triggering scenarios.

## What You'll Learn

- 🕐 **Event-time vs Processing-time** - Understanding temporal semantics
- 🪟 **Window Types** - Tumbling, sliding, and session windows
- 💧 **Watermarks** - Handling out-of-order data
- ⚡ **Advanced Triggers** - Count, time, and custom firing conditions
- 🔄 **Session Windows** - Activity-based grouping
- ⏰ **Late Data Handling** - Strategies for delayed events
- 🎯 **WindowBuilder** - Fluent API for complex configurations

## Prerequisites

- Complete the [main StreamV2 codelab](STREAMV2_CODELAB.md)
- Understanding of time concepts in streaming systems
- Go 1.21+ installed

## Chapter Navigation

**Foundations:** [Understanding Time](#understanding-time-in-streaming) • [Chapter 1: Basic Windowing](#chapter-1-basic-windowing) • [Chapter 2: Sliding Windows](#chapter-2-sliding-windows)

**Advanced:** [Chapter 3: Session Windows](#chapter-3-session-windows) • [Chapter 4: Advanced Triggers](#chapter-4-advanced-triggers) • [Chapter 5: Event-Time Processing](#chapter-5-event-time-processing)

**Expert:** [Chapter 6: Complex Scenarios](#chapter-6-complex-real-world-scenarios) • [Chapter 7: WindowBuilder](#chapter-7-windowbuilder-fluent-api) • [Advanced Sorting](#advanced-sorting-for-infinite-streams)

---

## Understanding Time in Streaming

Before diving into windowing, it's crucial to understand the different concepts of time:

- **Event Time**: When the event actually happened
- **Processing Time**: When the event is processed by your system
- **Ingestion Time**: When the event entered your streaming system

```go
// Example: Financial trade
{
  "symbol": "AAPL",
  "price": 150.25,
  "eventTime": "2025-01-15T14:30:00Z",     // When trade occurred
  "ingestionTime": "2025-01-15T14:30:05Z", // When received by system
  "processingTime": "2025-01-15T14:30:10Z" // When processed
}
```

---

## Chapter 1: Basic Windowing

### What Are Windows? A Simple Analogy

Imagine you're a **security guard watching a building entrance**. Instead of trying to remember every person who entered all day, you keep a logbook with **hourly pages**:

- **Page 1**: 9:00-10:00 AM - Who entered during this hour?
- **Page 2**: 10:00-11:00 AM - Who entered during this hour?
- **Page 3**: 11:00-12:00 PM - Who entered during this hour?

That's exactly what **windowing** does with streaming data! It groups events that happen during specific time periods.

### Understanding Tumbling Windows

Tumbling windows are like those **hourly logbook pages** - fixed-size time periods that don't overlap:

```
Time:     9:00   10:00   11:00   12:00   13:00
Windows:  [--1--][--2--][--3--][--4--][--5--]
```

**Key Properties:**
- ✅ **Fixed size**: Each window is exactly the same duration
- ✅ **No overlap**: Each event belongs to exactly one window
- ✅ **Complete coverage**: No gaps between windows
- ✅ **Simple**: Easy to understand and implement

### Why Use Event Time Instead of Processing Time?

This is **crucial** to understand! There are two different concepts of "time":

#### Processing Time (When we process it)
```
10:00 AM: Customer places order
10:05 AM: Our system processes the order (5 minutes later!)
```

#### Event Time (When it actually happened)
```
10:00 AM: Customer places order (this is what matters for business!)
10:05 AM: Our system processes it (this doesn't change when the order happened)
```

**Why Event Time Matters:**
- 📊 **Business Accuracy**: Reports show when things actually happened
- 🔄 **Reproducible Results**: Re-processing old data gives same results
- 🚀 **Scale Resilience**: Works even when your system is overwhelmed

### Let's Start Simple: Counting Website Visits

Before diving into complex examples, let's build something everyone can understand - counting website visits per minute:

**The Scenario**: You run a website and want to know "How many users visited each minute?"

```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Println("🌐 Website Analytics: Visits Per Minute")
    fmt.Println("======================================")

    // Let's simulate website visits happening throughout the day
    baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC) // 2:00 PM

    fmt.Println("📊 Simulating website visits:")
    visits := []stream.Record{
        stream.NewRecord().
            String("user_id", "alice123").
            String("page", "/home").
            String("visit_time", baseTime.Format(time.RFC3339)).
            Build(),
        stream.NewRecord().
            String("user_id", "bob456").
            String("page", "/products").
            String("visit_time", baseTime.Add(15*time.Second).Format(time.RFC3339)).
            Build(),
        stream.NewRecord().
            String("user_id", "charlie789").
            String("page", "/contact").
            String("visit_time", baseTime.Add(45*time.Second).Format(time.RFC3339)).
            Build(),
        // Next minute
        stream.NewRecord().
            String("user_id", "diana101").
            String("page", "/about").
            String("visit_time", baseTime.Add(75*time.Second).Format(time.RFC3339)).
            Build(),
        stream.NewRecord().
            String("user_id", "eve202").
            String("page", "/home").
            String("visit_time", baseTime.Add(105*time.Second).Format(time.RFC3339)).
            Build(),
    }

    // Show what data we're processing
    for i, visit := range visits {
        userID := stream.GetOr(visit, "user_id", "")
        page := stream.GetOr(visit, "page", "")
        visitTime := stream.GetOr(visit, "visit_time", "")
        if t, err := time.Parse(time.RFC3339, visitTime); err == nil {
            fmt.Printf("  Visit %d: %s visited %s at %s\n",
                       i+1, userID, page, t.Format("15:04:05"))
        }
    }

    fmt.Println("\n🪟 Creating 1-minute windows to count visits...")

    // HERE'S THE MAGIC: Create 1-minute tumbling windows
    windowedStream := stream.EventTimeTumblingWindow(
        1*time.Minute, // Group visits into 1-minute buckets
        stream.WithTimestampExtractor(
            // Tell StreamV2: "Look for timestamps in the 'visit_time' field"
            stream.NewRecordTimestampExtractor("visit_time"),
        ),
    )(stream.FromSlice(visits))

    // Process each completed window
    windowNumber := 1
    for {
        window, err := windowedStream()
        if err == stream.EOS {
            fmt.Println("✅ Finished processing all windows!")
            break
        }

        // Collect all visits in this time window
        windowVisits, _ := stream.Collect(window)

        if len(windowVisits) > 0 {
            // Calculate window time range
            firstVisit := windowVisits[0]
            firstTime := stream.GetOr(firstVisit, "visit_time", "")
            if t, err := time.Parse(time.RFC3339, firstTime); err == nil {
                windowStart := t.Truncate(1 * time.Minute)
                windowEnd := windowStart.Add(1 * time.Minute)

                fmt.Printf("📈 Window %d [%s - %s]: %d visits\n",
                           windowNumber,
                           windowStart.Format("15:04:05"),
                           windowEnd.Format("15:04:05"),
                           len(windowVisits))

                // Show details of each visit in this window
                for _, visit := range windowVisits {
                    userID := stream.GetOr(visit, "user_id", "")
                    page := stream.GetOr(visit, "page", "")
                    visitTime := stream.GetOr(visit, "visit_time", "")
                    if t, err := time.Parse(time.RFC3339, visitTime); err == nil {
                        fmt.Printf("    👤 %s → %s (%s)\n",
                                   userID, page, t.Format("15:04:05"))
                    }
                }
            }
        }

        windowNumber++
        fmt.Println()
    }
}
```

## Understanding What Just Happened

Let's break down this example step by step:

### 1. The Data Structure
Each website visit is a `Record` with:
- **user_id**: Who visited (alice123, bob456, etc.)
- **page**: What page they visited (/home, /products, etc.)
- **visit_time**: When they visited (the timestamp)

### 2. The Windowing Magic
```go
stream.EventTimeTumblingWindow(1*time.Minute, ...)
```

This tells StreamV2: "Group all visits into 1-minute time buckets"

### 3. The Timestamp Extractor
```go
stream.NewRecordTimestampExtractor("visit_time")
```

This tells StreamV2: "To determine when each visit happened, look at the 'visit_time' field"

### 4. The Results
StreamV2 automatically groups visits by time:
- **Window 1 (14:00-14:01)**: alice123, bob456, charlie789 (3 visits)
- **Window 2 (14:01-14:02)**: diana101, eve202 (2 visits)

## What You'll See When You Run This

```
🌐 Website Analytics: Visits Per Minute
======================================
📊 Simulating website visits:
  Visit 1: alice123 visited /home at 14:00:00
  Visit 2: bob456 visited /products at 14:00:15
  Visit 3: charlie789 visited /contact at 14:00:45
  Visit 4: diana101 visited /about at 14:01:15
  Visit 5: eve202 visited /home at 14:01:45

🪟 Creating 1-minute windows to count visits...
📈 Window 1 [14:00:00 - 14:01:00]: 3 visits
    👤 alice123 → /home (14:00:00)
    👤 bob456 → /products (14:00:15)
    👤 charlie789 → /contact (14:00:45)

📈 Window 2 [14:01:00 - 14:02:00]: 2 visits
    👤 diana101 → /about (14:01:15)
    👤 eve202 → /home (14:01:45)

✅ Finished processing all windows!
```

## Key Insights

1. **Automatic Grouping**: StreamV2 automatically puts visits into the right time buckets
2. **Business Metrics**: You can now answer "How many visitors per minute?"
3. **Scalable**: This works whether you have 5 visits or 5 million visits
4. **Real-Time Ready**: In production, this processes live data as it arrives

You've just built your first windowing application! Next, we'll explore more sophisticated window types.

---

## Chapter 2: Sliding Windows

### Understanding Sliding Windows

Unlike tumbling windows, sliding windows overlap with each other. This creates a smooth, continuous analysis:

```
Time:     0s    30s   60s   90s   120s
Window 1: [----------2min----------]
Window 2:       [----------2min----------]
Window 3:             [----------2min----------]
Window 4:                   [----------2min----------]
```

Each window shows a 2-minute snapshot, but they slide every 30 seconds. This gives you:
- **Continuous monitoring**: Updates every 30 seconds instead of every 2 minutes
- **Smooth trends**: No sudden jumps between discrete time periods
- **Early detection**: Spot changes as soon as they happen

### Real-World Example: IoT Temperature Monitoring

Imagine you're monitoring server temperatures. You want to:
- Track temperature trends over 2-minute periods
- Get updates every 30 seconds (not wait 2 minutes)
- Detect overheating conditions quickly

### Building a Temperature Monitor

```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    baseTime := time.Now()

    // Simulate IoT sensor readings every 30 seconds
    fmt.Println("🌡️  IoT Temperature Monitor - Sliding Windows Demo")

    sensorData := []stream.Record{
        stream.NewRecord().
            Float("temperature", 23.5).
            String("sensor", "server-01").
            String("timestamp", baseTime.Format(time.RFC3339)).
            Build(),
        stream.NewRecord().
            Float("temperature", 24.1).
            String("sensor", "server-01").
            String("timestamp", baseTime.Add(30*time.Second).Format(time.RFC3339)).
            Build(),
        stream.NewRecord().
            Float("temperature", 23.8).
            String("sensor", "server-01").
            String("timestamp", baseTime.Add(60*time.Second).Format(time.RFC3339)).
            Build(),
        stream.NewRecord().
            Float("temperature", 25.2).
            String("sensor", "server-01").
            String("timestamp", baseTime.Add(90*time.Second).Format(time.RFC3339)).
            Build(),
        stream.NewRecord().
            Float("temperature", 26.1).
            String("sensor", "server-01").
            String("timestamp", baseTime.Add(120*time.Second).Format(time.RFC3339)).
            Build(),
    }

    fmt.Println("📊 Creating 2-minute sliding windows (slides every 30s)...")

    // 2-minute sliding windows that slide every 30 seconds
    slidingWindows := stream.EventTimeSlidingWindow(
        2*time.Minute,  // Window size: analyze 2 minutes of data
        30*time.Second, // Slide interval: new window every 30 seconds
        stream.WithTimestampExtractor(
            stream.NewRecordTimestampExtractor("timestamp"),
        ),
    )(stream.FromSlice(sensorData))

    // Process each sliding window
    windowNumber := 1
    for {
        window, err := slidingWindows()
        if err == stream.EOS {
            fmt.Println("✅ Temperature monitoring complete!")
            break
        }

        readings, _ := stream.Collect(window)
        if len(readings) == 0 {
            continue
        }

        // Calculate moving average and detect trends
        var sum float64
        var minTemp, maxTemp float64
        var windowStart, windowEnd time.Time

        for i, reading := range readings {
            temp := stream.GetOr(reading, "temperature", 0.0)
            timestamp := stream.GetOr(reading, "timestamp", "")

            if i == 0 {
                minTemp = temp
                maxTemp = temp
                if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
                    windowStart = t
                    windowEnd = t.Add(2 * time.Minute)
                }
            } else {
                if temp < minTemp { minTemp = temp }
                if temp > maxTemp { maxTemp = temp }
            }

            sum += temp
        }

        average := sum / float64(len(readings))
        tempRange := maxTemp - minTemp

        fmt.Printf("🪟 Window %d [%s - %s]:\\n",
            windowNumber,
            windowStart.Format("15:04:05"),
            windowEnd.Format("15:04:05"))
        fmt.Printf("   📈 Average: %.1f°C (%d readings)\\n", average, len(readings))
        fmt.Printf("   📊 Range: %.1f°C (%.1f - %.1f)\\n", tempRange, minTemp, maxTemp)

        // Alert on high temperatures
        if average > 25.0 {
            fmt.Printf("   🚨 HIGH TEMPERATURE ALERT! Average %.1f°C exceeds 25°C\\n", average)
        }

        // Detect rapid temperature changes
        if tempRange > 2.0 {
            fmt.Printf("   ⚠️  High temperature variance: %.1f°C range\\n", tempRange)
        }

        fmt.Println()
        windowNumber++
    }
}
```

### Why Sliding Windows Matter

1. **Smooth Analysis**: No sudden jumps between reporting periods
   - Tumbling windows: Updates every 2 minutes with potential gaps
   - Sliding windows: Updates every 30 seconds with continuous overlap

2. **Early Detection**: Spot problems as soon as they develop
   - Don't wait for the full window period
   - React to trends immediately

3. **Moving Averages**: Perfect for trend analysis
   - Smooth out temporary spikes
   - Show overall direction of change
   - Filter noise while preserving signal

### Trade-offs to Consider

**Memory Usage**: Sliding windows use more memory than tumbling windows
- Each data point appears in multiple windows
- More window state to maintain

**Computation**: More frequent calculations
- Processing happens more often
- CPU usage increases with shorter slide intervals

**When to Use Sliding Windows**:
- ✅ Real-time dashboards and monitoring
- ✅ Trend detection and moving averages
- ✅ Anomaly detection systems
- ❌ Simple batch reporting (use tumbling instead)
- ❌ Memory-constrained environments

---

## Chapter 3: Session Windows

Session windows group related activities based on activity gaps, perfect for user behavior analysis.

```go
func demonstrateSessionWindows() {
    // User activity with natural gaps
    userEvents := []stream.Record{
        stream.NewRecord().String("user", "alice").String("action", "login").String("timestamp", baseTime.Format(time.RFC3339)).Build(),
        stream.NewRecord().String("user", "alice").String("action", "view_page").String("timestamp", baseTime.Add(30*time.Second).Format(time.RFC3339)).Build(),
        stream.NewRecord().String("user", "alice").String("action", "click").String("timestamp", baseTime.Add(45*time.Second).Format(time.RFC3339)).Build(),
        // 5-minute gap - triggers new session
        stream.NewRecord().String("user", "alice").String("action", "login").String("timestamp", baseTime.Add(6*time.Minute).Format(time.RFC3339)).Build(),
        stream.NewRecord().String("user", "alice").String("action", "logout").String("timestamp", baseTime.Add(7*time.Minute).Format(time.RFC3339)).Build(),
    }

    // Create session windows with 2-minute timeout
    sessionWindows := stream.EventTimeSessionWindow(
        2*time.Minute, // Session timeout
        stream.WithTimestampExtractor(
            stream.NewRecordTimestampExtractor("timestamp"),
        ),
    )(stream.FromSlice(userEvents))

    sessionNum := 1
    for {
        session, err := sessionWindows()
        if err == stream.EOS {
            break
        }

        sessionEvents, _ := stream.Collect(session)
        if len(sessionEvents) == 0 {
            continue
        }

        // Analyze session duration and actions
        actions := make([]string, 0, len(sessionEvents))
        for _, event := range sessionEvents {
            action := stream.GetOr(event, "action", "")
            actions = append(actions, action)
        }

        fmt.Printf("Session %d: %d events - %v\\n", sessionNum, len(sessionEvents), actions)
        sessionNum++
    }
}
```

### Session Windows Use Cases

- **User Behavior**: Group user actions into meaningful sessions
- **Fraud Detection**: Detect suspicious activity patterns
- **Application Monitoring**: Group related system events

---

## Chapter 4: Advanced Triggers

The WindowBuilder provides fine-grained control over when windows fire.

```go
func demonstrateAdvancedTriggers() {
    events := createSampleEvents() // Your event data

    // Window that fires on multiple conditions
    complexWindow := stream.Window[stream.Record]().
        TriggerOnCount(5).                    // Fire after 5 elements
        TriggerOnTime(30*time.Second).       // OR after 30 seconds
        TriggerOnProcessingTime(1*time.Minute). // OR every minute
        AllowLateness(10*time.Second).       // Allow 10s late data
        AccumulationMode().                  // Accumulate late data
        Apply()

    windowedStream := complexWindow(stream.FromSlice(events))

    // Process triggered windows
    for {
        window, err := windowedStream()
        if err == stream.EOS {
            break
        }

        windowEvents, _ := stream.Collect(window)
        fmt.Printf("Window fired with %d events\\n", len(windowEvents))
    }
}
```

### Trigger Types

- **Count Triggers**: Fire after N elements
- **Time Triggers**: Fire after duration since window start
- **Processing Time Triggers**: Fire at regular intervals
- **Custom Triggers**: Implement your own logic

---

## Chapter 5: Handling Out-of-Order Data

Real-world streams often have out-of-order events. Watermarks help handle this.

```go
func demonstrateOutOfOrderHandling() {
    // Events arriving out of order
    outOfOrderEvents := []stream.Record{
        stream.NewRecord().String("id", "3").String("timestamp", baseTime.Add(60*time.Second).Format(time.RFC3339)).Int("value", 300).Build(),
        stream.NewRecord().String("id", "1").String("timestamp", baseTime.Format(time.RFC3339)).Int("value", 100).Build(), // Late!
        stream.NewRecord().String("id", "2").String("timestamp", baseTime.Add(30*time.Second).Format(time.RFC3339)).Int("value", 200).Build(), // Late!
        stream.NewRecord().String("id", "4").String("timestamp", baseTime.Add(90*time.Second).Format(time.RFC3339)).Int("value", 400).Build(),
    }

    // Configure watermark strategy
    windowedStream := stream.EventTimeTumblingWindow(
        1*time.Minute,
        stream.WithTimestampExtractor(
            stream.NewRecordTimestampExtractor("timestamp"),
        ),
        stream.WithAllowedLateness(30*time.Second), // Wait up to 30s for late data
        stream.WithLateDataPolicy(stream.DropLateData), // Or AccumulateLateData
    )(stream.FromSlice(outOfOrderEvents))

    for {
        window, err := windowedStream()
        if err == stream.EOS {
            break
        }

        windowEvents, _ := stream.Collect(window)
        fmt.Printf("Window contains %d events (late events handled)\\n", len(windowEvents))
    }
}
```

### Watermark Strategies

- **BoundedOutOfOrderness**: Allow fixed lateness duration
- **Periodic**: Update watermarks at intervals
- **Custom**: Implement domain-specific logic

---

## Chapter 6: Real-World Example - Financial Analytics

Let's build a complete financial trading analytics system using advanced windowing.

```go
func financialAnalyticsExample() {
    // Simulate trading data
    trades := generateTradingData() // Your trade generation logic

    // Multiple analysis windows running in parallel

    // 1. Real-time price monitoring (5-second tumbling)
    priceWindows := stream.EventTimeTumblingWindow(
        5*time.Second,
        stream.WithTimestampExtractor(
            stream.NewRecordTimestampExtractor("timestamp"),
        ),
    )(stream.FromSlice(trades))

    // 2. Volume-weighted average price (1-minute sliding)
    vwapWindows := stream.EventTimeSlidingWindow(
        1*time.Minute,
        10*time.Second,
        stream.WithTimestampExtractor(
            stream.NewRecordTimestampExtractor("timestamp"),
        ),
    )(stream.FromSlice(trades))

    // 3. Trading session detection (5-minute timeout)
    sessionWindows := stream.EventTimeSessionWindow(
        5*time.Minute,
        stream.WithTimestampExtractor(
            stream.NewRecordTimestampExtractor("timestamp"),
        ),
    )(stream.FromSlice(trades))

    // Process each window type
    go processRealTimePrices(priceWindows)
    go processVWAP(vwapWindows)
    go processTradingSessions(sessionWindows)
}

func processRealTimePrices(windows Stream[Stream[stream.Record]]) {
    for {
        window, err := windows()
        if err == stream.EOS {
            break
        }

        trades, _ := stream.Collect(window)
        if len(trades) == 0 {
            continue
        }

        // Calculate real-time statistics
        var totalVolume int64
        var weightedPrice float64
        var totalValue float64

        for _, trade := range trades {
            price := stream.GetOr(trade, "price", 0.0)
            volume := stream.GetOr(trade, "volume", int64(0))

            totalVolume += volume
            totalValue += price * float64(volume)
        }

        if totalVolume > 0 {
            weightedPrice = totalValue / float64(totalVolume)
        }

        // Alert on significant price movements
        if len(trades) > 0 {
            firstPrice := stream.GetOr(trades[0], "price", 0.0)
            lastPrice := stream.GetOr(trades[len(trades)-1], "price", 0.0)
            priceChange := (lastPrice - firstPrice) / firstPrice * 100

            if abs(priceChange) > 2.0 {
                fmt.Printf("🚨 Price Alert: %.2f%% change (%.2f → %.2f)\\n",
                    priceChange, firstPrice, lastPrice)
            }
        }

        fmt.Printf("📊 5s Window: %d trades, VWAP: $%.2f\\n", len(trades), weightedPrice)
    }
}
```

---

## Chapter 7: Performance Considerations

### Choosing Window Types

- **Tumbling**: Lowest memory usage, discrete analysis periods
- **Sliding**: Higher memory usage, smoother analysis
- **Session**: Variable memory usage, activity-based

### Trigger Optimization

```go
// Efficient: Fewer trigger evaluations
window := stream.Window[stream.Record]().
    TriggerOnCount(1000).        // Batch processing
    TriggerOnTime(5*time.Minute). // Reasonable timeout
    Apply()

// Inefficient: Too many trigger evaluations
window := stream.Window[stream.Record]().
    TriggerOnCount(1).           // Per-element processing
    TriggerOnTime(1*time.Second). // Very frequent timeouts
    Apply()
```

### Memory Management

- **Bounded Windows**: Prefer time-based bounds over count-based
- **Late Data**: Balance completeness vs resource usage
- **Window State**: Monitor memory usage in long-running applications

---

## Chapter 8: Best Practices

### Making Smart Windowing Decisions

Windowing is powerful, but with great power comes great responsibility! Here's how to make smart decisions that will save you from pain later.

### 1. Choosing the Right Timestamp

**The Golden Rule**: Use the timestamp that reflects your business logic.

```go
// ✅ GOOD: E-commerce order processing
stream.WithTimestampExtractor(
    stream.NewRecordTimestampExtractor("orderTimestamp"), // When customer placed order
)

// ✅ GOOD: Financial trading analysis
stream.WithTimestampExtractor(
    stream.NewRecordTimestampExtractor("tradeTimestamp"), // When trade executed
)

// ❌ AVOID: Using processing time for business analytics
stream.WithTimestampExtractor(
    stream.ProcessingTimeExtractor(), // When your system processed it
)
```

**Why This Matters**:
- **Business Accuracy**: Your analytics reflect reality, not system delays
- **Reproducibility**: Re-processing historical data gives same results
- **Debugging**: Easy to trace issues back to business events

**When Processing Time is OK**:
- System monitoring and performance metrics
- SLA tracking (when did we process this?)
- Infrastructure alerting

### 2. Watermark Strategy: The Lateness Dilemma

Watermarks control how long you wait for late data. It's a classic trade-off:

```go
// 🐌 CONSERVATIVE: High accuracy, higher latency
stream.WithAllowedLateness(60*time.Second), // Wait up to 1 minute
stream.WithLateDataPolicy(stream.AccumulateLateData), // Include everything

// ⚡ AGGRESSIVE: Low latency, may miss some data
stream.WithAllowedLateness(5*time.Second), // Wait only 5 seconds
stream.WithLateDataPolicy(stream.DropLateData), // Drop late arrivals
```

**How to Choose**:

1. **Start Conservative** (60s+ lateness)
   - Run for a few days in production
   - Measure how much data actually arrives late
   - Gradually reduce lateness tolerance

2. **Monitor Late Data**
   ```go
   // Add metrics to track late arrivals
   lateDataCount := 0
   stream.WithLateDataPolicy(func(record stream.Record) {
       lateDataCount++
       log.Printf("Late data: %v (total: %d)", record, lateDataCount)
       // Decide: include or drop
   })
   ```

3. **Business Requirements Win**
   - Financial systems: Usually need 100% accuracy (conservative)
   - Real-time dashboards: Usually prefer low latency (aggressive)
   - Fraud detection: Depends on use case

### 3. Window Size: Goldilocks Principle

Window size affects both accuracy and performance:

```go
// 🎯 RIGHT-SIZED: Match business needs
analytics := stream.EventTimeTumblingWindow(
    15*time.Minute, // Matches business reporting cadence
    // Business reports every 15 minutes? Perfect.
)

alerting := stream.EventTimeTumblingWindow(
    30*time.Second, // Fast enough for real-time alerts
    // Need sub-minute alerting? Great choice.
)

// 🚫 TOO SMALL: Wasted resources
microWindows := stream.EventTimeTumblingWindow(
    1*time.Second, // Probably overkill for most use cases
    // High CPU, frequent calculations, no business value
)

// 🚫 TOO LARGE: Delayed insights
batchWindows := stream.EventTimeTumblingWindow(
    24*time.Hour, // Might as well be batch processing
    // No real-time value, defeats purpose of streaming
)
```

**Guidelines by Use Case**:
- **Real-time monitoring**: 30 seconds - 5 minutes
- **Business dashboards**: 5 minutes - 1 hour
- **Analytics reports**: 15 minutes - 1 day
- **Fraud detection**: 1 second - 1 minute

### 4. Memory Management

Windowing systems can consume significant memory. Plan ahead:

```go
// ✅ BOUNDED: Predictable memory usage
tumblingWindow := stream.EventTimeTumblingWindow(
    5*time.Minute, // Fixed window size
    stream.WithAllowedLateness(30*time.Second), // Bounded lateness
)

// ⚠️ UNBOUNDED: Can grow indefinitely
sessionWindow := stream.EventTimeSessionWindow(
    time.Hour, // Very long session timeout
    // Sessions could accumulate for hours!
)
```

**Memory Planning Tips**:
1. **Estimate Window Contents**: events/second × window_duration = max events per window
2. **Account for Late Data**: multiply by (1 + lateness_duration/window_duration)
3. **Monitor in Production**: Add memory metrics and alerts
4. **Set Reasonable Limits**: Fail fast rather than OOM

### 5. Testing Windowing Logic

Windowing introduces complexity. Test thoroughly:

```go
func TestWindowingLogic(t *testing.T) {
    // Create predictable test data
    baseTime := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)
    events := []stream.Record{
        // Events with known timestamps
        makeEvent("user1", baseTime),
        makeEvent("user2", baseTime.Add(30*time.Second)),
        makeEvent("user1", baseTime.Add(90*time.Second)), // Next window
    }

    // Test with deterministic timestamps
    windows := stream.EventTimeTumblingWindow(
        1*time.Minute,
        stream.WithTimestampExtractor(testTimestampExtractor),
    )(stream.FromSlice(events))

    // Verify window contents
    window1, _ := windows()
    events1, _ := stream.Collect(window1)
    assert.Equal(t, 2, len(events1)) // First two events

    window2, _ := windows()
    events2, _ := stream.Collect(window2)
    assert.Equal(t, 1, len(events2)) // Third event
}
```

**Testing Best Practices**:
- Use fixed, predictable timestamps in tests
- Test edge cases (window boundaries, late data)
- Verify window contents, not just counts
- Test timeout and cleanup logic

### 6. Observability: Know What's Happening

Production windowing systems need monitoring:

```go
// Add metrics to your windowing pipeline
var (
    windowsProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "windows_processed_total"},
        []string{"window_type"},
    )
    lateDataDropped = prometheus.NewCounter(
        prometheus.CounterOpts{Name: "late_data_dropped_total"},
    )
    windowProcessingDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{Name: "window_processing_duration_seconds"},
    )
)

func monitoredWindowProcessing(window Stream[stream.Record]) {
    start := time.Now()
    defer func() {
        windowProcessingDuration.Observe(time.Since(start).Seconds())
        windowsProcessed.WithLabelValues("tumbling").Inc()
    }()

    // Process window...
}
```

**Key Metrics to Track**:
- Window processing latency
- Late data drop rate
- Memory usage per window type
- Window completion rate
- Error rates

---

## Chapter 9: Troubleshooting

### Common Issues

**No Windows Firing**
- Check timestamp extractor configuration
- Verify event timestamps are properly formatted
- Ensure watermark allows window completion

**Missing Events**
- Increase allowed lateness
- Check late data policy configuration
- Verify timestamp ordering

**High Memory Usage**
- Reduce window sizes
- Tune trigger conditions
- Monitor session window timeouts

### Debugging Tips

```go
// Add debugging to understand window behavior
windowedStream := stream.EventTimeTumblingWindow(
    1*time.Minute,
    stream.WithTimestampExtractor(debugTimestampExtractor),
    stream.WithAllowedLateness(30*time.Second),
)(debugWrappedStream(inputStream))

func debugTimestampExtractor(record stream.Record) time.Time {
    timestamp := stream.GetOr(record, "timestamp", "")
    t, err := time.Parse(time.RFC3339, timestamp)
    if err != nil {
        fmt.Printf("⚠️ Invalid timestamp: %s\\n", timestamp)
        return time.Now()
    }
    fmt.Printf("🕐 Extracted timestamp: %s\\n", t.Format("15:04:05"))
    return t
}
```

---

## Advanced Sorting for Infinite Streams

Traditional sorting requires loading all data into memory, which is impossible with infinite streams. StreamV2 provides windowed sorting strategies to handle this challenge.

### Window-Based Sorting

Sort data within bounded windows - perfect for real-time analytics:

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
    "github.com/rosscartlidge/streamv2/pkg/stream"
)

func main() {
    fmt.Printf("=== Windowed Sorting for Infinite Streams ===\\n\\n")

    // Simulate real-time sensor readings (temperature values)
    temperatures := []float64{23.1, 24.5, 22.8, 25.2, 21.9, 26.1, 23.7, 24.8, 22.3, 25.9}

    // Sort within count-based windows (useful for batch processing)
    fmt.Printf("🔢 Count-based windowed sorting:\\n")
    sortedWindows, _ := stream.Collect(
        stream.SortCountWindow(4, func(a, b float64) int {
            if a < b { return -1 }
            if a > b { return 1 }
            return 0
        })(stream.FromSlice(temperatures)))

    fmt.Printf("Original: %v\\n", temperatures)
    fmt.Printf("Sorted in windows of 4: %v\\n\\n", sortedWindows)

    // Real-world example: Top-K tracking for leaderboards
    fmt.Printf("🏆 Top-K for infinite leaderboards:\\n")
    scores := []int{85, 92, 78, 96, 81, 88, 94, 76, 90, 87, 95, 83, 91, 79, 93}

    topPlayers, _ := stream.Collect(
        stream.TopK(5, func(a, b int) int {
            return a - b  // ascending comparison for max-heap behavior
        })(stream.FromSlice(scores)))

    fmt.Printf("All scores: %v\\n", scores)
    fmt.Printf("Top 5 scores: %v\\n\\n", topPlayers)
}
```

### Time-Based Windowed Sorting

For event-time processing, sort within time windows:

```go
// Simulate timestamped events
type SensorReading struct {
    Timestamp time.Time
    Value     float64
    SensorID  string
}

// In practice, you'd use SortTimeWindow for real-time data:
// stream.SortTimeWindow(5*time.Second, func(a, b SensorReading) int {
//     return int(a.Value*100 - b.Value*100)  // Sort by sensor value
// })
```

### When to Use Each Approach

**Count Windows (`SortCountWindow`):**
- ✅ Batch processing scenarios
- ✅ Fixed-size result sets
- ✅ Memory-bounded sorting

**Time Windows (`SortTimeWindow`):**
- ✅ Real-time analytics
- ✅ Event-time ordering
- ✅ Time-series data

**Top-K (`TopK`/`BottomK`):**
- ✅ Leaderboards and rankings
- ✅ Anomaly detection (extreme values)
- ✅ Memory-efficient approximate sorting
- ✅ Streaming recommendations

### Performance Considerations

- **Window size** affects latency vs. accuracy trade-offs
- **Top-K** uses heap-based algorithms for O(n log k) performance
- **Time windows** require buffering and can handle late-arriving data
- **Memory usage** scales with window size, not total stream length

---

## Summary

You've learned how to use StreamV2's advanced windowing capabilities:

- ✅ **Event-time processing** for business-correct analytics
- ✅ **Multiple window types** for different use cases
- ✅ **Advanced triggers** for fine-grained control
- ✅ **Out-of-order handling** with watermarks
- ✅ **Session detection** for activity-based grouping
- ✅ **Windowed sorting** for infinite stream ordering
- ✅ **Top-K algorithms** for streaming leaderboards
- ✅ **Performance optimization** and best practices

### Next Steps

- Explore the [complete examples](../examples/) directory
- Check out [event-time windowing examples](../examples/event_time_windowing/)
- Read the [API documentation](api.md) for detailed function references
- Implement custom triggers for your specific use cases

### Related Resources

- [StreamV2 Main Codelab](STREAMV2_CODELAB.md) - Core concepts and basic operations
- [API Reference](api.md) - Complete function documentation
- [Examples](../examples/) - Real-world usage patterns

---

**Advanced windowing unlocks the full power of stream processing - use it to build real-time analytics that scale!** 🚀