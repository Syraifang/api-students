-- 1. Tabel roles
CREATE TABLE IF NOT EXISTS roles (
    name VARCHAR(20) PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
('admin', 'Akses penuh terhadap seluruh data dan pengaturan'),
('staff', 'Boleh melihat data seluruh user, tetapi tidak boleh mengubah'),
('user', 'Hanya boleh mengelola datanya sendiri')
ON CONFLICT (name) DO NOTHING;

-- 2. Tabel permissions (dasar)
CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

INSERT INTO permissions (name, description) VALUES
('user:list', 'Melihat daftar seluruh user'),
('user:read:any', 'Melihat data user mana pun'),
('user:update:any', 'Mengubah data user mana pun'),
('user:delete', 'Menghapus user'),
('role:assign', 'Mengubah role milik user lain')
ON CONFLICT (name) DO NOTHING;

-- 3. Tabel role_permissions (dasar)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_name VARCHAR(20) NOT NULL REFERENCES roles (name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions (name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO role_permissions (role_name, permission_name) VALUES
('admin', 'user:list'), ('admin', 'user:read:any'), ('admin', 'user:update:any'), ('admin', 'user:delete'), ('admin', 'role:assign'),
('staff', 'user:list'), ('staff', 'user:read:any')
ON CONFLICT DO NOTHING;

-- 4. Penyesuaian role lama di tabel students
UPDATE students SET role = 'user' WHERE role = 'student' OR role NOT IN (SELECT name FROM roles);
ALTER TABLE students DROP CONSTRAINT IF EXISTS students_role_fkey;
ALTER TABLE students ADD CONSTRAINT students_role_fkey FOREIGN KEY (role) REFERENCES roles (name) ON UPDATE CASCADE;
CREATE INDEX IF NOT EXISTS students_role_idx ON students (role);