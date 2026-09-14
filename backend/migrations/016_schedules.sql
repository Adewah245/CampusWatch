-- 016_schedules.sql
-- Purpose: define institution operating hours and break windows.

CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID NOT NULL,
    day_of_week SMALLINT NOT NULL,
    opening_time TIME NOT NULL,
    break_start TIME,
    break_end TIME,
    closing_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT schedules_institution_fk FOREIGN KEY (institution_id)
        REFERENCES institutions (id) ON DELETE CASCADE,
    CONSTRAINT schedules_day_check CHECK (day_of_week >= 0 AND day_of_week <= 6),
    CONSTRAINT schedules_break_pair_check CHECK ((break_start IS NULL) = (break_end IS NULL)),
    CONSTRAINT schedules_time_order_check CHECK (
        opening_time < closing_time AND (break_start IS NULL OR (opening_time < break_start AND break_start < break_end AND break_end < closing_time))
    ),
    CONSTRAINT schedules_unique_day UNIQUE (institution_id, day_of_week)
);