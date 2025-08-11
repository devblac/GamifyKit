package redis

import (
    "context"
    "errors"
    "gamifykit/core"
)

// Store is a placeholder for a Redis-backed Storage implementation.
// Not implemented in this initial version.
type Store struct{}

func New() *Store { return &Store{} }

func (s *Store) AddPoints(context.Context, core.UserID, core.Metric, int64) (int64, error) {
    return 0, errors.New("redis adapter not implemented yet")
}
func (s *Store) AwardBadge(context.Context, core.UserID, core.Badge) error {
    return errors.New("redis adapter not implemented yet")
}
func (s *Store) GetState(context.Context, core.UserID) (core.UserState, error) {
    return core.UserState{}, errors.New("redis adapter not implemented yet")
}
func (s *Store) SetLevel(context.Context, core.UserID, core.Metric, int64) error {
    return errors.New("redis adapter not implemented yet")
}


