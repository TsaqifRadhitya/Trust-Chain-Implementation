# API Gateway Specification (Nginx)

Dokumen ini berisi spesifikasi teknis untuk konfigurasi API Gateway berbasis Nginx. API Gateway ini bertindak sebagai pintu masuk utama (Reverse Proxy) untuk mengakses layanan-layanan di backend dan menangani otorisasi secara terpusat.

---

## 1. Informasi Umum

- **Service**: API Gateway (Nginx)
- **Port Exposed**: `8080` (Docker Host)
- **Base URL**: `http://localhost:8080/`
- **Konfigurasi Utama**: `nginx.conf`

---

## 2. Arsitektur dan Alur Kerja

Nginx beroperasi sebagai *Reverse Proxy* dan memanfaatkan modul `auth_request` untuk memvalidasi token JWT secara terpusat sebelum meneruskan request ke layanan yang membutuhkan otentikasi (Protected Routes).

1. **Public Routes (Tanpa Otentikasi)**: Request diteruskan langsung ke *upstream* tanpa melalui proses validasi token (contoh: Login, Refresh Token).
2. **Protected Routes (Butuh Otentikasi)**: 
   - Nginx akan mencegat request dan melakukan *sub-request* secara internal ke endpoint validator (`/auth/validate`).
   - Jika validator mengembalikan status `200 OK`, request utama diizinkan lewat dan diteruskan ke *downstream service*.
   - Nginx akan menangkap header `X-User-Id` dari respons validator dan menyuntikkannya ke dalam header request ke *downstream service*.
   - Jika validator mengembalikan status `401 Unauthorized` atau HTTP status error lainnya, request langsung ditolak oleh Nginx tanpa menyentuh *downstream service*.

---

## 3. Daftar Routing (Endpoints)

Nginx dikonfigurasi untuk menangani beberapa lokasi *routing* berikut:

### A. Auth Service Routes (Public)
Digunakan untuk keperluan autentikasi seperti login dan rotasi token.

- **Path**: `/api/v1/auth/`
- **Upstream**: `http://backend/api/v1/auth/`
- **Otentikasi**: Tidak ada (Public)
- **Tingkah Laku**: Proxy meneruskan semua *header* asli klien (`Host`, `X-Real-IP`, `X-Forwarded-For`) ke upstream.

### B. Internal Auth Validator (Internal)
Endpoint khusus yang hanya bisa diakses oleh Nginx secara internal (tidak dapat dipanggil langsung dari luar) untuk keperluan `auth_request`.

- **Path**: `/auth/validate` (Internal Route)
- **Upstream**: `http://backend/api/v1/auth/validate`
- **Tingkah Laku**: 
  - Nginx mematikan pengiriman *body* pada request (`proxy_pass_request_body off;`).
  - Meneruskan token JWT (via header `Authorization`) ke *backend* untuk diverifikasi.

### C. Setting Service Routes (Protected)
Digunakan untuk mengakses konfigurasi spesifik user. Endpoint ini dilindungi penuh oleh Nginx.

- **Path**: `/api/v1/settings`
- **Upstream**: `http://backend/api/v1/settings`
- **Otentikasi**: `auth_request /auth/validate`
- **Tingkah Laku**: 
  - Harus lolos pengecekan `auth_request` sebelum dikirimkan.
  - Jika token valid, Nginx akan mengekstrak variabel internal (`$upstream_http_x_user_id`) dari respons validator dan menginjeksinya ke header request `X-User-Id` yang diteruskan menuju *backend*.

---

## 4. Upstream Configuration

Konfigurasi saat ini mendefinisikan blok `upstream` di dalam Nginx untuk menyalurkan *traffic* ke instance dalam internal network Docker:
- **`backend`**: Proxy menyasar nama service docker `backend` pada port `8080`. (Service ini saat ini mewadahi logic dari Auth Service sekaligus Setting Service).

---

## 5. Cara Menjalankan

API Gateway Nginx dan konfigurasi upstream-nya telah diintegrasikan langsung dalam file root `docker-compose.yaml`. Untuk menjalankan API gateway secara utuh beserta service terkait, jalankan perintah:

```bash
docker-compose up -d --build
```
Setelah kontainer aktif, layanan Nginx akan siap menerima panggilan HTTP di port `8080`.
