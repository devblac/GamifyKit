package sqlx

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"gamifykit/core"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Driver represents the database driver type
type Driver string

const (
	DriverPostgres Driver = "postgres"
	DriverMySQL    Driver = "mysql"
)

// Config holds SQL database configuration
type Config struct {
	Driver          Driver
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns sensible defaults for SQL configuration
func DefaultConfig(driver Driver) Config {
	config := Config{
		Driver:          driver,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	switch driver {
	case DriverPostgres:
		config.DSN = "postgres://gamifykit:gamifykit@localhost/gamifykit?sslmode=disable"
	case DriverMySQL:
		config.DSN = "gamifykit:gamifykit@tcp(localhost:3306)/gamifykit?parseTime=true"
	}

	return config
}

// Store implements the engine.Storage interface using SQL database as the backend.
// Uses optimistic locking and transactions for data consistency.
type Store struct {
	db     *sqlx.DB
	driver Driver
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// New creates a new SQL-backed storage with the provided configuration
func New(config Config) (*Store, error) {
	db, err := sqlx.Open(string(config.Driver), config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			// Log close error but prioritize the ping error
			// In error cleanup, we don't fail the operation for close errors
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &Store{db: db, driver: config.Driver}

	// Run migrations
	if err := store.runMigrations(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			// Log close error but prioritize the migration error
			// In error cleanup, we don't fail the operation for close errors
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// NewWithDB creates a Store using an existing sqlx.DB (useful for testing)
func NewWithDB(db *sqlx.DB, driver Driver) *Store {
	return &Store{db: db, driver: driver}
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// runMigrations executes database migrations
func (s *Store) runMigrations(ctx context.Context) error {
	// Read migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		// Execute migration
		if _, err := s.db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// AddPoints atomically adds points to a user's metric with transaction safety
func (s *Store) AddPoints(ctx context.Context, userID core.UserID, metric core.Metric, delta int64) (int64, error) {
	if delta == 0 {
		return 0, errors.New("delta cannot be zero")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current points (or 0 if not exists)
	var currentPoints sql.NullInt64
	query := `
		SELECT points FROM user_points
		WHERE user_id = $1 AND metric = $2
	`
	if s.driver == DriverMySQL {
		query = `
			SELECT points FROM user_points
			WHERE user_id = ? AND metric = ?
		`
	}

	err = tx.QueryRowContext(ctx, query, userID, metric).Scan(&currentPoints)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to get current points: %w", err)
	}

	newPoints := currentPoints.Int64 + delta

	// Check for overflow (basic check)
	if (delta > 0 && newPoints < currentPoints.Int64) || (delta < 0 && newPoints > currentPoints.Int64) {
		return 0, errors.New("integer overflow in AddPoints")
	}

	// Insert or update points
	if currentPoints.Valid {
		// Update existing
		updateQuery := `
			UPDATE user_points
			SET points = $1, updated_at = $2
			WHERE user_id = $3 AND metric = $4
		`
		if s.driver == DriverMySQL {
			updateQuery = `
				UPDATE user_points
				SET points = ?, updated_at = ?
				WHERE user_id = ? AND metric = ?
			`
		}
		_, err = tx.ExecContext(ctx, updateQuery, newPoints, time.Now().UTC(), userID, metric)
	} else {
		// Insert new
		insertQuery := `
			INSERT INTO user_points (user_id, metric, points, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`
		if s.driver == DriverMySQL {
			insertQuery = `
				INSERT INTO user_points (user_id, metric, points, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?)
			`
		}
		_, err = tx.ExecContext(ctx, insertQuery, userID, metric, newPoints, time.Now().UTC(), time.Now().UTC())
	}

	if err != nil {
		return 0, fmt.Errorf("failed to update points: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newPoints, nil
}

// AwardBadge adds a badge to the user's badge collection
func (s *Store) AwardBadge(ctx context.Context, userID core.UserID, badge core.Badge) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if badge already exists
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM user_badges
			WHERE user_id = $1 AND badge = $2
		)
	`
	if s.driver == DriverMySQL {
		checkQuery = `
			SELECT EXISTS(
				SELECT 1 FROM user_badges
				WHERE user_id = ? AND badge = ?
			)
		`
	}

	err = tx.QueryRowContext(ctx, checkQuery, userID, badge).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check badge existence: %w", err)
	}

	if exists {
		// Badge already awarded, commit and return
		return tx.Commit()
	}

	// Insert new badge
	insertQuery := `
		INSERT INTO user_badges (user_id, badge, awarded_at)
		VALUES ($1, $2, $3)
	`
	if s.driver == DriverMySQL {
		insertQuery = `
			INSERT INTO user_badges (user_id, badge, awarded_at)
			VALUES (?, ?, ?)
		`
	}

	_, err = tx.ExecContext(ctx, insertQuery, userID, badge, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to award badge: %w", err)
	}

	return tx.Commit()
}

// GetState retrieves the complete user state from the database
func (s *Store) GetState(ctx context.Context, userID core.UserID) (core.UserState, error) {
	state := core.UserState{
		UserID:  userID,
		Points:  make(map[core.Metric]int64),
		Badges:  make(map[core.Badge]struct{}),
		Levels:  make(map[core.Metric]int64),
		Updated: time.Now().UTC(),
	}

	// Get points
	pointsQuery := `
		SELECT metric, points FROM user_points
		WHERE user_id = $1
	`
	if s.driver == DriverMySQL {
		pointsQuery = `
			SELECT metric, points FROM user_points
			WHERE user_id = ?
		`
	}

	pointsRows, err := s.db.QueryContext(ctx, pointsQuery, userID)
	if err != nil {
		return core.UserState{}, fmt.Errorf("failed to get points: %w", err)
	}
	defer pointsRows.Close()

	for pointsRows.Next() {
		var metric core.Metric
		var points int64
		if err := pointsRows.Scan(&metric, &points); err != nil {
			return core.UserState{}, fmt.Errorf("failed to scan points: %w", err)
		}
		state.Points[metric] = points
	}

	// Get badges
	badgesQuery := `
		SELECT badge FROM user_badges
		WHERE user_id = $1
	`
	if s.driver == DriverMySQL {
		badgesQuery = `
			SELECT badge FROM user_badges
			WHERE user_id = ?
		`
	}

	badgesRows, err := s.db.QueryContext(ctx, badgesQuery, userID)
	if err != nil {
		return core.UserState{}, fmt.Errorf("failed to get badges: %w", err)
	}
	defer badgesRows.Close()

	for badgesRows.Next() {
		var badge core.Badge
		if err := badgesRows.Scan(&badge); err != nil {
			return core.UserState{}, fmt.Errorf("failed to scan badge: %w", err)
		}
		state.Badges[badge] = struct{}{}
	}

	// Get levels
	levelsQuery := `
		SELECT metric, level FROM user_levels
		WHERE user_id = $1
	`
	if s.driver == DriverMySQL {
		levelsQuery = `
			SELECT metric, level FROM user_levels
			WHERE user_id = ?
		`
	}

	levelsRows, err := s.db.QueryContext(ctx, levelsQuery, userID)
	if err != nil {
		return core.UserState{}, fmt.Errorf("failed to get levels: %w", err)
	}
	defer levelsRows.Close()

	for levelsRows.Next() {
		var metric core.Metric
		var level int64
		if err := levelsRows.Scan(&metric, &level); err != nil {
			return core.UserState{}, fmt.Errorf("failed to scan level: %w", err)
		}
		state.Levels[metric] = level
	}

	return state, nil
}

// SetLevel sets the user's level for a specific metric
func (s *Store) SetLevel(ctx context.Context, userID core.UserID, metric core.Metric, level int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if level already exists
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM user_levels
			WHERE user_id = $1 AND metric = $2
		)
	`
	if s.driver == DriverMySQL {
		checkQuery = `
			SELECT EXISTS(
				SELECT 1 FROM user_levels
				WHERE user_id = ? AND metric = ?
			)
		`
	}

	err = tx.QueryRowContext(ctx, checkQuery, userID, metric).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check level existence: %w", err)
	}

	if exists {
		// Update existing
		updateQuery := `
			UPDATE user_levels
			SET level = $1, updated_at = $2
			WHERE user_id = $3 AND metric = $4
		`
		if s.driver == DriverMySQL {
			updateQuery = `
				UPDATE user_levels
				SET level = ?, updated_at = ?
				WHERE user_id = ? AND metric = ?
			`
		}
		_, err = tx.ExecContext(ctx, updateQuery, level, time.Now().UTC(), userID, metric)
	} else {
		// Insert new
		insertQuery := `
			INSERT INTO user_levels (user_id, metric, level, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`
		if s.driver == DriverMySQL {
			insertQuery = `
				INSERT INTO user_levels (user_id, metric, level, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?)
			`
		}
		_, err = tx.ExecContext(ctx, insertQuery, userID, metric, level, time.Now().UTC(), time.Now().UTC())
	}

	if err != nil {
		return fmt.Errorf("failed to set level: %w", err)
	}

	return tx.Commit()
}
