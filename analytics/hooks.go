package analytics

import (
    "sync"
    "time"

    "gamifykit/core"
)

// Hook receives domain events for KPI aggregation.
type Hook interface { OnEvent(e core.Event) }

// DAU tracks daily active users.
type DAU struct {
    mu   sync.Mutex
    days map[string]map[core.UserID]struct{}
}

func NewDAU() *DAU { return &DAU{days: map[string]map[core.UserID]struct{}{}} }

func (d *DAU) OnEvent(e core.Event) {
    day := time.Unix(e.Time.Unix(), 0).UTC().Format("2006-01-02")
    d.mu.Lock(); defer d.mu.Unlock()
    m := d.days[day]
    if m == nil { m = map[core.UserID]struct{}{}; d.days[day] = m }
    m[e.UserID] = struct{}{}
}

func (d *DAU) Count(day string) int {
    d.mu.Lock(); defer d.mu.Unlock()
    return len(d.days[day])
}


