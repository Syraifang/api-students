-- 1. Tambah permissions khusus student
INSERT INTO permissions (name, description) VALUES
('student:list', 'Melihat daftar seluruh student'),
('student:read:any', 'Melihat data student mana pun'),
('student:create', 'Menambah data student baru'),
('student:update:any', 'Mengubah data student mana pun'),
('student:delete', 'Menghapus data student')
ON CONFLICT (name) DO NOTHING;

-- 2. Pasangkan ke role admin dan staff
INSERT INTO role_permissions (role_name, permission_name) VALUES
('admin', 'student:list'), ('admin', 'student:read:any'), ('admin', 'student:create'), ('admin', 'student:update:any'), ('admin', 'student:delete'),
('staff', 'student:list'), ('staff', 'student:read:any'), ('staff', 'student:create')
ON CONFLICT DO NOTHING;

-- 3. Tambah owner_id untuk cek kepemilikan (ownership)
ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id BIGINT;

-- Isi data lama agar tidak error saat dikunci
UPDATE students SET owner_id = id WHERE owner_id IS NULL;

-- Kunci owner_id
ALTER TABLE students DROP CONSTRAINT IF EXISTS students_owner_fkey;
ALTER TABLE students ADD CONSTRAINT students_owner_fkey FOREIGN KEY (owner_id) REFERENCES students(id) ON DELETE CASCADE;