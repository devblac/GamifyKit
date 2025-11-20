package leaderboard

import (
	"context"
	"testing"

	"gamifykit/core"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func TestRedisBoard_Update(t *testing.T) {
	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
	}

	board := NewRedisBoard(client, "test:leaderboard")
	defer client.Del(ctx, "test:leaderboard") // cleanup

	// Test updating scores
	board.Update("alice", 100)
	board.Update("bob", 150)
	board.Update("charlie", 75)

	// Verify scores
	aliceEntry, exists := board.Get("alice")
	require.True(t, exists)
	assert.Equal(t, int64(100), aliceEntry.Score)

	bobEntry, exists := board.Get("bob")
	require.True(t, exists)
	assert.Equal(t, int64(150), bobEntry.Score)

	charlieEntry, exists := board.Get("charlie")
	require.True(t, exists)
	assert.Equal(t, int64(75), charlieEntry.Score)
}

func TestRedisBoard_TopN(t *testing.T) {
	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
	}

	board := NewRedisBoard(client, "test:leaderboard:topn")
	defer client.Del(ctx, "test:leaderboard:topn") // cleanup

	// Add test data
	board.Update("alice", 100)
	board.Update("bob", 150)
	board.Update("charlie", 75)
	board.Update("dave", 200)

	// Get top 3
	top3 := board.TopN(3)
	require.Len(t, top3, 3)

	// Should be ordered by score descending
	assert.Equal(t, core.UserID("dave"), top3[0].User)
	assert.Equal(t, int64(200), top3[0].Score)

	assert.Equal(t, core.UserID("bob"), top3[1].User)
	assert.Equal(t, int64(150), top3[1].Score)

	assert.Equal(t, core.UserID("alice"), top3[2].User)
	assert.Equal(t, int64(100), top3[2].Score)
}

func TestRedisBoard_Remove(t *testing.T) {
	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
	}

	board := NewRedisBoard(client, "test:leaderboard:remove")
	defer client.Del(ctx, "test:leaderboard:remove") // cleanup

	// Add and then remove a user
	board.Update("alice", 100)
	_, exists := board.Get("alice")
	require.True(t, exists)

	board.Remove("alice")
	_, exists = board.Get("alice")
	assert.False(t, exists)
}

func TestRedisBoard_Get(t *testing.T) {
	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
	}

	board := NewRedisBoard(client, "test:leaderboard:get")
	defer client.Del(ctx, "test:leaderboard:get") // cleanup

	// Test getting non-existent user
	_, exists := board.Get("nonexistent")
	assert.False(t, exists)

	// Add user and get their entry
	board.Update("alice", 100)
	entry, exists := board.Get("alice")
	require.True(t, exists)
	assert.Equal(t, core.UserID("alice"), entry.User)
	assert.Equal(t, int64(100), entry.Score)
}