GamifyKit – High-performance gamification library for Go

Overview
GamifyKit is a modular, high-performance gamification engine for Go 1.22+. It provides ultra-low-latency, horizontally scalable APIs to add points, XP, levels, badges, achievements, challenges, leaderboards, realtime events, and analytics to any app with minimal code.

Quick Start
```go
import (
  "context"
  "gamifykit/core"
  "gamifykit/gamify"
  "gamifykit/realtime"
)

var ctx = context.Background()
svc := gamify.New(
    // defaults to in-memory storage, async dispatch, default rules
    gamify.WithRealtime(realtime.NewHub()),
)

userID := core.UserID("u1")
svc.AddPoints(ctx, userID, core.MetricXP, 50)

unsub := svc.Subscribe(core.EventAchievementUnlocked, func(ctx context.Context, e core.Event) {
    // handle achievement
})
defer unsub()
```

See `docs/QuickStart.md` and `cmd/demo-server`.

License: Apache-2.0


