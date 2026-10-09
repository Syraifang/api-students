# SIAKAD Mini - RESTful API Back-End (UTS Praktikum Pengembangan Back-End)

API backend layanan akademik sederhana yang mengelola data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS), dibangun menggunakan **Go (Golang)**, **Fiber Framework**, dan **PostgreSQL**[cite: 7, 8].

## Fitur Utama & Aturan Bisnis
1. **Autentikasi & Otorisasi**: Menggunakan JWT (JSON Web Token) dengan peran (*role*) `admin` dan `mahasiswa`[cite: 7, 8].
2. **Manajemen Akademik (Admin)**: CRUD data mahasiswa beserta akun login-nya dan pengelolaan data mata kuliah[cite: 8].
3. **Sistem KRS (Mahasiswa)**:
   - Validasi batas SKS otomatis berdasarkan IPK Terakhir ($IPK \ge 3.00$ maks 24 SKS, $2.50-2.99$ maks 21 SKS, $< 2.50$ maks 18 SKS)[cite: 8].
   - Pencegahan duplikasi pengambilan mata kuliah di tahun akademik yang sama[cite: 8].
   - Pengamanan kuota kelas menggunakan mekanisme *row-locking* transaksi database (`FOR UPDATE`)[cite: 8].
   - Pembatasan akses data (*ownership*) sehingga mahasiswa hanya dapat mengakses KRS miliknya sendiri[cite: 8].

## Daftar 10 Endpoint API
1. `POST /api/v1/auth/login` - Login pengguna dan mendapatkan access token[cite: 9]
2. `GET /api/v1/auth/me` - Menampilkan profil pengguna yang sedang login[cite: 9]
3. `GET /api/v1/students` - Daftar mahasiswa dengan pagination, filter, dan search (Admin)[cite: 9]
4. `POST /api/v1/students` - Menambah mahasiswa sekaligus akun user-nya dalam satu transaksi (Admin)[cite: 9]
5. `GET /api/v1/students/{id}` - Detail mahasiswa beserta total SKS yang diambil[cite: 9]
6. `PUT /api/v1/students/{id}` - Memperbarui data mahasiswa (Admin)[cite: 9]
7. `DELETE /api/v1/students/{id}` - Soft delete data mahasiswa (Admin)[cite: 9]
8. `GET /api/v1/courses` - Daftar mata kuliah beserta sisa kuota dinamis[cite: 9]
9. `POST /api/v1/enrollments` - Mengambil mata kuliah / menambah KRS dengan validasi IPK & kuota (Mahasiswa)[cite: 9]
10. `DELETE /api/v1/enrollments/{id}` - Membatalkan mata kuliah dari KRS[cite: 9]

## Persiapan & Menjalankan Aplikasi
1. Buat database PostgreSQL baru:
   ```sql
   CREATE DATABASE api_students;