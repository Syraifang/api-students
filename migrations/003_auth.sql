-- 1. Tambahkan kolom password dan role ke tabel students yang sudah ada
ALTER TABLE students ADD COLUMN IF NOT EXISTS password VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE students ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'student';

-- 2. Buat tabel refresh_tokens yang menginduk ke tabel students
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL UNIQUE REFERENCES students (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Indeks untuk mempercepat pencarian token mahasiswa
CREATE INDEX IF NOT EXISTS refresh_tokens_student_id_idx ON refresh_tokens (student_id);