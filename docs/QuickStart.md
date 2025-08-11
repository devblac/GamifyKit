Quick Start

Install
```
go get gamifykit
```

Example
```go
package main

import (
    "context"
    "fmt"

    "gamifykit/core"
    "gamifykit/engine"
    mem "gamifykit/adapters/memory"
    "gamifykit/realtime"
)

func main() {
    ctx := context.Background()
    bus := engine.NewEventBus(engine.DispatchAsync)
    store := mem.New()
    hub := realtime.NewHub()

    svc := engine.NewGamifyService(store, bus, engine.DefaultRuleEngine())
    defer svc.Close()

    user := core.UserID("alice")
    _, _ = svc.AddPoints(ctx, user, core.MetricXP, 50)
    state, _ := svc.GetState(ctx, user)
    fmt.Println("XP:", state.Points[core.MetricXP])
    _ = hub // subscribe to realtime if needed
}
```


