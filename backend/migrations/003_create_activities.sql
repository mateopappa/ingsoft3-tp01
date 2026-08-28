-- 003_create_activities.sql
CREATE TABLE IF NOT EXISTS activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    description VARCHAR(255) NOT NULL,
    duration_seconds INTEGER NOT NULL CHECK (duration_seconds > 0),
    activity_date DATE NOT NULL DEFAULT CURRENT_DATE,
    note TEXT,
    favorite BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activities_user_date ON activities (user_id, activity_date DESC);
CREATE INDEX IF NOT EXISTS idx_activities_user_category ON activities (user_id, category_id);
CREATE INDEX IF NOT EXISTS idx_activities_user_favorite ON activities (user_id, favorite) WHERE favorite = TRUE;
