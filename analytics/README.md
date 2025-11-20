# Analytics Package

A comprehensive analytics system for gamification platforms that provides real-time metrics, aggregation, export capabilities, and live dashboard support.

## Features

- **Real-time Metrics**: Track all gamification events with low-latency processing
- **Aggregation**: Daily, weekly, and monthly aggregation pipelines
- **Export**: Integration with external analytics platforms (Segment, HTTP webhooks)
- **Streaming**: Real-time event streaming for live dashboards
- **Comprehensive Tracking**: Points, badges, levels, achievements, and user engagement

## Quick Start

```go
package main

import (
    "context"
    "gamifykit/analytics"
    "gamifykit/engine"
)

func main() {
    // Create analytics service
    analyticsSvc := analytics.NewAnalyticsService()

    // Create gamification service
    svc := engine.NewGamifyService(storage, bus, rules)

    // Register analytics hook
    svc.RegisterHook(analyticsSvc.GetHook())

    // Start analytics in background
    ctx := context.Background()
    analyticsSvc.Start(ctx)

    // Your gamification logic here...
    svc.AddPoints(ctx, userID, "xp", 100)

    // Get real-time stats
    stats := analyticsSvc.GetRealtimeStats()
    fmt.Printf("Points awarded in last 24h: %d\n", stats["points_awarded_24h"])
}
```

## Components

### ComprehensiveMetrics

Tracks all gamification events with detailed metrics:

```go
metrics := analytics.NewComprehensiveMetrics()

// Get daily active users
dau := metrics.GetDailyActiveUsers("2024-01-01")

// Get points awarded by metric
xpPoints := metrics.GetPointsAwardedByMetric("xp")

// Get real-time stats (last 24h)
points, badges, levels := metrics.GetRealtimeStats()
```

### AggregationEngine

Handles periodic aggregation of metrics:

```go
aggregator := analytics.NewAggregationEngine(metrics, 1*time.Hour)

// Force immediate aggregation
aggregator.AggregateNow()

// Get daily aggregated data
dailyData, exists := aggregator.GetAggregatedData(analytics.PeriodDaily, "2024-01-01")

// Export to JSON
jsonData, _ := aggregator.ExportData(analytics.PeriodDaily)
```

### StreamPublisher

Provides real-time event streaming:

```go
publisher := analytics.NewStreamPublisher(metrics)

// Subscribe to real-time events
subscriber := analytics.NewInMemorySubscriber("dashboard")
publisher.Subscribe("dashboard", subscriber)

// Events are automatically published when gamification events occur
```

### Export System

Export aggregated data to external platforms:

```go
// HTTP webhook export
httpExporter := analytics.NewHTTPExporter("https://api.example.com/analytics", "api-key", 10)

// Segment analytics export
segmentExporter := analytics.NewSegmentExporter("write-key")

// Combine multiple exporters
multiExporter := analytics.NewMultiExporter(httpExporter, segmentExporter)

// Create export manager
exportManager := analytics.NewExportManager(multiExporter)

// Export data
exportManager.ExportData(ctx, aggregatedData)
```

## Configuration

Create analytics with custom configuration:

```go
config := &analytics.AnalyticsConfig{
    AggregationInterval: 30 * time.Minute,
    MaxRecentEvents:     1000,
    ExportInterval:      6 * time.Hour,
    EnableStreaming:     true,
    Exporters: []analytics.ExporterConfig{
        {
            Type:      "http",
            Endpoint:  "https://analytics.example.com/webhook",
            APIKey:    "secret-key",
            BatchSize: 50,
        },
        {
            Type:   "segment",
            APIKey: "segment-write-key",
        },
    },
}

analyticsSvc := analytics.NewAnalyticsServiceWithConfig(config)
```

## Dashboard Integration

Create live dashboards with real-time data:

```go
// Create dashboard manager
dashboard := analytics.NewDashboardManager(publisher, metrics, 100)

// Get dashboard data
data := dashboard.GetDashboardData()

// Data includes:
// - Real-time stats (last 24h)
// - Top metrics by points
// - Recent events stream
```

## WebSocket Streaming

Stream real-time events to web clients:

```go
// Create WebSocket subscriber
wsSubscriber := analytics.NewWebSocketSubscriber("client-123", 100)

// Subscribe to real-time events
analyticsSvc.SubscribeToRealtime("client-123", wsSubscriber)

// Read events in WebSocket handler
for {
    event, err := wsSubscriber.ReadEvent(ctx)
    if err != nil {
        break
    }

    // Send to WebSocket client
    ws.SendJSON(event)
}
```

## Metrics Tracked

### User Engagement
- Daily/Weekly/Monthly Active Users (DAU/WAU/MAU)
- User retention rates
- Session tracking

### Points System
- Points awarded/spent by day, week, month
- Points by metric type
- Top users by points

### Badge System
- Badges awarded by day, type
- Unique badge holders
- Badge completion rates

### Level System
- Levels reached by metric
- Level distribution analysis
- Level progression tracking

### Achievement System
- Achievements unlocked by type
- Completion rates
- Achievement popularity

## Export Formats

### Aggregated Data JSON
```json
{
  "period": "daily",
  "key": "2024-01-01",
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-02T00:00:00Z",
  "active_users": 150,
  "points_awarded": 12500,
  "points_spent": 3200,
  "points_by_metric": {
    "xp": 8500,
    "coins": 4000
  },
  "badges_awarded": 45,
  "badges_by_type": {
    "first_steps": 12,
    "early_adopter": 8
  },
  "levels_reached": 23,
  "levels_by_metric": {
    "xp": 18,
    "skill": 5
  },
  "achievements_unlocked": 8,
  "achievements_by_type": {
    "speed_demon": 3,
    "completionist": 2
  }
}
```

### Real-time Stream Events
```json
{
  "type": "points_awarded",
  "user_id": "user123",
  "metric": "xp",
  "points": 100,
  "timestamp": "2024-01-01T10:30:00Z",
  "metadata": {
    "source": "quest_completion",
    "multiplier": 1.5
  }
}
```

## Performance

- **Low Latency**: Real-time processing with minimal overhead
- **Scalable Aggregation**: Efficient background aggregation
- **Batch Exports**: Configurable batch sizes for external APIs
- **Memory Efficient**: Configurable limits for recent events storage

## Integration Examples

See `integration.go` for complete examples of:
- Basic analytics setup
- Advanced configuration
- External export integration
- Dashboard data consumption
- WebSocket streaming

## Testing

Run the analytics test suite:

```bash
go test ./analytics/...
```

Includes benchmarks for performance validation:

```bash
go test -bench=. ./analytics/
```
