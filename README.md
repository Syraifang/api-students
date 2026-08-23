# Students API - Praktikum Backend Lanjut Modul 2

Repositori ini berisi RESTful API sederhana yang dibangun menggunakan bahasa Go dan menggunakan kerangka kerja (framework) **Fiber v2** serta penyimpanan berbasis memori sementara.

## Dokumen Kontrak API

Berikut adalah kontrak API lengkap untuk mengakses layanan Students:

| Metode | Endpoint | Parameter / Query | Contoh Body Permintaan (Request Body) | Status HTTP | Contoh Respons (Response Body) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/students` | `page`, `limit`, `search`, `sort`, `order`, `is_active` | *Tidak ada* | `200 OK` | `{"success":true, "message":"daftar mahasiswa berhasil diambil", "data":[...], "meta":{...}}` |
| **GET** | `/api/v1/students/:id` | `id` (pada URL) | *Tidak ada* | `200 OK`<br>`404 Not Found` | `{"success":true, "message":"mahasiswa ditemukan", "data":{...}}` |
| **POST** | `/api/v1/students` | *Tidak ada* | `{"nim":"123456789", "name":"John Doe", "grade":95.5, "is_active":true}` | `201 Created`<br>`400 Bad Request`<br>`409 Conflict`<br>`415 Unsupported Media Type`<br>`422 Unprocessable Entity` | `{"success":true, "message":"mahasiswa berhasil dibuat", "data":{...}}` |
| **PUT** | `/api/v1/students/:id` | `id` (pada URL) | `{"nim":"123456789", "name":"John Doe (PUT)", "grade":80.0, "is_active":false}` | `200 OK`<br>`400 Bad Request`<br>`404 Not Found`<br>`409 Conflict`<br>`415 Unsupported Media Type`<br>`422 Unprocessable Entity` | `{"success":true, "message":"data mahasiswa berhasil diganti seluruhnya", "data":{...}}` |
| **PATCH** | `/api/v1/students/:id` | `id` (pada URL) | `{"is_active":true}` | `200 OK`<br>`400 Bad Request`<br>`404 Not Found`<br>`409 Conflict`<br>`415 Unsupported Media Type`<br>`422 Unprocessable Entity` | `{"success":true, "message":"data mahasiswa berhasil diperbarui sebagian", "data":{...}}` |
| **DELETE** | `/api/v1/students/:id` | `id` (pada URL) | *Tidak ada* | `204 No Content`<br>`400 Bad Request`<br>`404 Not Found` | *(Kosong / Tanpa bodi respons)* |