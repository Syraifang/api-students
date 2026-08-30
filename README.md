# Praktikum Backend - Students API (Pertemuan 3)

API ini dibuat menggunakan Go Fiber dan PostgreSQL dengan mengimplementasikan arsitektur Repository Pattern.

## Persiapan database

Untuk menjalankan aplikasi ini secara lokal, Anda harus membuat database kosong terlebih dahulu di PostgreSQL:

1. Buka terminal/psql dan jalankan:
   `CREATE DATABASE api_students;`
2. Jalankan file migrasi untuk membuat tabel dan indeks:
   `psql -U postgres -d api_students -f migrations/001_create_students.sql`

## Skema Tabel (students)

| Kolom | Tipe Data | Keterangan |
| :--- | :--- | :--- |
| `id` | SERIAL | Primary Key |
| `nim` | VARCHAR(20) | UNIQUE, NOT NULL |
| `name` | VARCHAR(100) | NOT NULL |
| `grade` | VARCHAR(2) | NOT NULL |
| `is_active` | BOOLEAN | NOT NULL, DEFAULT TRUE |
| `created_at` | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

*Terdapat juga indeks tambahan pada kolom `name` (LOWER) untuk mempercepat pencarian (ILIKE).*

## Konfigurasi Environment (.env)

Buat file `.env` di *root* direktori (sejajar dengan `main.go`) dan isi dengan variabel berikut (lihat `.env.example` sebagai referensi):

*   `APP_PORT`: Port untuk menjalankan server (contoh: 3000)
*   `DB_HOST`: Host database (contoh: localhost)
*   `DB_PORT`: Port PostgreSQL (contoh: 5432)
*   `DB_USER`: Username PostgreSQL (contoh: postgres)
*   `DB_PASSWORD`: Kata sandi user
*   `DB_NAME`: Nama database (api_students)
*   `DB_SSLMODE`: Mode SSL (disable)
*   `DB_MAX_CONNS`: Batas maksimal koneksi *pool* (contoh: 10)