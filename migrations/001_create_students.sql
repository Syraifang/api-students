CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade VARCHAR(2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks tambahan untuk mempercepat fitur pencarian nama (ILIKE)
CREATE INDEX IF NOT EXISTS students_name_lower_idx ON students (LOWER(name));