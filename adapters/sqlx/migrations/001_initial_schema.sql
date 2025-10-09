-- Initial schema for GamifyKit SQL storage
-- This migration creates the necessary tables for storing user gamification data

-- Table for storing user points by metric
CREATE TABLE user_points (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    metric VARCHAR(255) NOT NULL,
    points BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, metric)
);

-- Table for storing user badges
CREATE TABLE user_badges (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    badge VARCHAR(255) NOT NULL,
    awarded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, badge)
);

-- Table for storing user levels by metric
CREATE TABLE user_levels (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    metric VARCHAR(255) NOT NULL,
    level BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, metric)
);

-- Indexes for performance
CREATE INDEX idx_user_points_user_id ON user_points(user_id);
CREATE INDEX idx_user_points_metric ON user_points(metric);
CREATE INDEX idx_user_badges_user_id ON user_badges(user_id);
CREATE INDEX idx_user_levels_user_id ON user_levels(user_id);
CREATE INDEX idx_user_levels_metric ON user_levels(metric);

-- Comments for documentation
COMMENT ON TABLE user_points IS 'Stores gamification points for users by metric (e.g., XP, coins)';
COMMENT ON TABLE user_badges IS 'Stores badges awarded to users';
COMMENT ON TABLE user_levels IS 'Stores user levels for different metrics';
