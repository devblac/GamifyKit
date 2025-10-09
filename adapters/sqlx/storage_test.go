package sqlx

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gamifykit/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Skip SQL tests if requested
	if os.Getenv("SKIP_SQL_TESTS") == "true" {
		os.Exit(0)
	}

	os.Exit(m.Run())
}

// testDBConfig returns database configuration for testing
func testDBConfig(driver Driver) Config {
	config := DefaultConfig(driver)

	// Override with test-specific values
	switch driver {
	case DriverPostgres:
		config.DSN = os.Getenv("TEST_POSTGRES_DSN")
		if config.DSN == "" {
			config.DSN = "postgres://gamifykit_test:gamifykit_test@localhost/gamifykit_test?sslmode=disable"
		}
	case DriverMySQL:
		config.DSN = os.Getenv("TEST_MYSQL_DSN")
		if config.DSN == "" {
			config.DSN = "gamifykit_test:gamifykit_test@tcp(localhost:3306)/gamifykit_test?parseTime=true"
		}
	}

	config.MaxOpenConns = 5
	config.MaxIdleConns = 2

	return config
}

// skipIfNoDB skips the test if the specified database is not available
func skipIfNoDB(t *testing.T, driver Driver) *Store {
	config := testDBConfig(driver)

	store, err := New(config)
	if err != nil {
		t.Skipf("Database %s not available: %v", driver, err)
		return nil
	}

	// Clean up test data after test
	t.Cleanup(func() {
		cleanupTestData(t, store, driver)
		store.Close()
	})

	return store
}

// cleanupTestData removes test data from the database
func cleanupTestData(t *testing.T, store *Store, driver Driver) {
	ctx := context.Background()

	// Delete test data (users starting with "test-")
	testTables := []string{"user_points", "user_badges", "user_levels"}

	for _, table := range testTables {
		query := `DELETE FROM ` + table + ` WHERE user_id LIKE 'test-%'`
		if driver == DriverPostgres {
			query = `DELETE FROM ` + table + ` WHERE user_id LIKE 'test-%'`
		}
		_, err := store.db.ExecContext(ctx, query)
		if err != nil {
			t.Logf("Warning: failed to cleanup test data from %s: %v", table, err)
		}
	}
}

func TestStore_Postgres_AddPoints(t *testing.T) {
	store := skipIfNoDB(t, DriverPostgres)
	if store == nil {
		return
	}

	testAddPoints(t, store)
}

func TestStore_MySQL_AddPoints(t *testing.T) {
	store := skipIfNoDB(t, DriverMySQL)
	if store == nil {
		return
	}

	testAddPoints(t, store)
}

func testAddPoints(t *testing.T, store *Store) {
	ctx := context.Background()

	userID := core.UserID("test-user")
	metric := core.MetricXP

	// Clean up any existing data
	cleanupUserData(t, store, userID)

	// Test adding points
	total, err := store.AddPoints(ctx, userID, metric, 50)
	require.NoError(t, err)
	assert.Equal(t, int64(50), total)

	// Test adding more points
	total, err = store.AddPoints(ctx, userID, metric, 25)
	require.NoError(t, err)
	assert.Equal(t, int64(75), total)

	// Test subtracting points
	total, err = store.AddPoints(ctx, userID, metric, -30)
	require.NoError(t, err)
	assert.Equal(t, int64(45), total)
}

func TestStore_Postgres_AddPoints_ZeroDelta(t *testing.T) {
	testAddPointsZeroDelta(t, DriverPostgres)
}

func TestStore_MySQL_AddPoints_ZeroDelta(t *testing.T) {
	testAddPointsZeroDelta(t, DriverMySQL)
}

func testAddPointsZeroDelta(t *testing.T, driver Driver) {
	store := skipIfNoDB(t, driver)
	if store == nil {
		return
	}

	ctx := context.Background()

	userID := core.UserID("test-user-zero")
	metric := core.MetricXP

	_, err := store.AddPoints(ctx, userID, metric, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delta cannot be zero")
}

func TestStore_Postgres_AwardBadge(t *testing.T) {
	store := skipIfNoDB(t, DriverPostgres)
	if store == nil {
		return
	}

	testAwardBadge(t, store)
}

func TestStore_MySQL_AwardBadge(t *testing.T) {
	store := skipIfNoDB(t, DriverMySQL)
	if store == nil {
		return
	}

	testAwardBadge(t, store)
}

func testAwardBadge(t *testing.T, store *Store) {
	ctx := context.Background()

	userID := core.UserID("test-user")
	badge := core.Badge("first-win")

	// Clean up any existing data
	cleanupUserData(t, store, userID)

	// Test awarding badge
	err := store.AwardBadge(ctx, userID, badge)
	require.NoError(t, err)

	// Test awarding same badge again (should be idempotent)
	err = store.AwardBadge(ctx, userID, badge)
	require.NoError(t, err)

	// Verify badge exists
	state, err := store.GetState(ctx, userID)
	require.NoError(t, err)
	assert.Contains(t, state.Badges, badge)
}

func TestStore_Postgres_GetState(t *testing.T) {
	store := skipIfNoDB(t, DriverPostgres)
	if store == nil {
		return
	}

	testGetState(t, store)
}

func TestStore_MySQL_GetState(t *testing.T) {
	store := skipIfNoDB(t, DriverMySQL)
	if store == nil {
		return
	}

	testGetState(t, store)
}

func testGetState(t *testing.T, store *Store) {
	ctx := context.Background()

	userID := core.UserID("test-user-state")

	// Clean up any existing data
	cleanupUserData(t, store, userID)

	// Set up some state
	_, err := store.AddPoints(ctx, userID, core.MetricXP, 100)
	require.NoError(t, err)
	_, err = store.AddPoints(ctx, userID, core.MetricPoints, 50)
	require.NoError(t, err)

	err = store.AwardBadge(ctx, userID, core.Badge("winner"))
	require.NoError(t, err)

	err = store.SetLevel(ctx, userID, core.MetricXP, 5)
	require.NoError(t, err)

	// Get state
	state, err := store.GetState(ctx, userID)
	require.NoError(t, err)

	assert.Equal(t, userID, state.UserID)
	assert.Equal(t, int64(100), state.Points[core.MetricXP])
	assert.Equal(t, int64(50), state.Points[core.MetricPoints])
	assert.Contains(t, state.Badges, core.Badge("winner"))
	assert.Equal(t, int64(5), state.Levels[core.MetricXP])
	assert.True(t, time.Since(state.Updated) < time.Second)
}

func TestStore_Postgres_SetLevel(t *testing.T) {
	store := skipIfNoDB(t, DriverPostgres)
	if store == nil {
		return
	}

	testSetLevel(t, store)
}

func TestStore_MySQL_SetLevel(t *testing.T) {
	store := skipIfNoDB(t, DriverMySQL)
	if store == nil {
		return
	}

	testSetLevel(t, store)
}

func testSetLevel(t *testing.T, store *Store) {
	ctx := context.Background()

	userID := core.UserID("test-user-level")
	metric := core.MetricXP

	// Clean up any existing data
	cleanupUserData(t, store, userID)

	// Test setting level
	err := store.SetLevel(ctx, userID, metric, 10)
	require.NoError(t, err)

	// Verify level was set
	state, err := store.GetState(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(10), state.Levels[metric])

	// Test updating level
	err = store.SetLevel(ctx, userID, metric, 15)
	require.NoError(t, err)

	state, err = store.GetState(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(15), state.Levels[metric])
}

func TestStore_Postgres_EmptyUser(t *testing.T) {
	store := skipIfNoDB(t, DriverPostgres)
	if store == nil {
		return
	}

	testEmptyUser(t, store)
}

func TestStore_MySQL_EmptyUser(t *testing.T) {
	store := skipIfNoDB(t, DriverMySQL)
	if store == nil {
		return
	}

	testEmptyUser(t, store)
}

func testEmptyUser(t *testing.T, store *Store) {
	ctx := context.Background()

	userID := core.UserID("nonexistent-user")

	// Clean up any existing data
	cleanupUserData(t, store, userID)

	// Get state for user that doesn't exist
	state, err := store.GetState(ctx, userID)
	require.NoError(t, err)

	assert.Equal(t, userID, state.UserID)
	assert.Empty(t, state.Points)
	assert.Empty(t, state.Badges)
	assert.Empty(t, state.Levels)
	assert.True(t, time.Since(state.Updated) < time.Second)
}

func TestStore_Postgres_ConcurrentAccess(t *testing.T) {
	store := skipIfNoDB(t, DriverPostgres)
	if store == nil {
		return
	}

	testConcurrentAccess(t, store)
}

func TestStore_MySQL_ConcurrentAccess(t *testing.T) {
	store := skipIfNoDB(t, DriverMySQL)
	if store == nil {
		return
	}

	testConcurrentAccess(t, store)
}

func testConcurrentAccess(t *testing.T, store *Store) {
	ctx := context.Background()

	userID := core.UserID("test-user-concurrent")
	metric := core.MetricXP

	// Clean up any existing data
	cleanupUserData(t, store, userID)

	// Run concurrent operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(delta int64) {
			_, err := store.AddPoints(ctx, userID, metric, delta)
			assert.NoError(t, err)
			done <- true
		}(int64(i + 1))
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state (sum of 1+2+...+10 = 55)
	state, err := store.GetState(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(55), state.Points[metric])
}

// cleanupUserData removes all data for a specific user
func cleanupUserData(t *testing.T, store *Store, userID core.UserID) {
	ctx := context.Background()

	tables := []string{"user_points", "user_badges", "user_levels"}
	for _, table := range tables {
		query := `DELETE FROM ` + table + ` WHERE user_id = $1`
		if store.driver == DriverMySQL {
			query = `DELETE FROM ` + table + ` WHERE user_id = ?`
		}
		_, err := store.db.ExecContext(ctx, query, userID)
		if err != nil {
			t.Logf("Warning: failed to cleanup user data: %v", err)
		}
	}
}

func TestConfig_DefaultConfig_Postgres(t *testing.T) {
	config := DefaultConfig(DriverPostgres)

	assert.Equal(t, DriverPostgres, config.Driver)
	assert.Contains(t, config.DSN, "postgres://")
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
}

func TestConfig_DefaultConfig_MySQL(t *testing.T) {
	config := DefaultConfig(DriverMySQL)

	assert.Equal(t, DriverMySQL, config.Driver)
	assert.Contains(t, config.DSN, "tcp(localhost:3306)")
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
}

// Benchmark tests
func BenchmarkStore_AddPoints_Postgres(b *testing.B) {
	store := setupBenchmarkStore(b, DriverPostgres)
	if store == nil {
		b.Skip("PostgreSQL not available")
		return
	}

	benchmarkAddPoints(b, store)
}

func BenchmarkStore_AddPoints_MySQL(b *testing.B) {
	store := setupBenchmarkStore(b, DriverMySQL)
	if store == nil {
		b.Skip("MySQL not available")
		return
	}

	benchmarkAddPoints(b, store)
}

func setupBenchmarkStore(b *testing.B, driver Driver) *Store {
	config := testDBConfig(driver)
	store, err := New(config)
	if err != nil {
		return nil
	}

	b.Cleanup(func() {
		store.Close()
	})

	return store
}

func benchmarkAddPoints(b *testing.B, store *Store) {
	ctx := context.Background()
	userID := core.UserID("bench-user")
	metric := core.MetricXP

	// Clean up
	cleanupUserData(&testing.T{}, store, userID)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			testUserID := core.UserID(fmt.Sprintf("bench-user-%d", i%100))
			_, _ = store.AddPoints(ctx, testUserID, metric, 1)
			i++
		}
	})
}
