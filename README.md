# TrustChain Implementation

TrustChain adalah platform sistem informasi terintegrasi berbasis blockchain yang dirancang untuk menarik (*pull*), memvalidasi, mencatat transaksi/log dari sistem *Enterprise Resource Planning* (ERP) milik klien ke dalam buku besar (*ledger*) yang transparan dan dapat diverifikasi.

## 🏗 Arsitektur Sistem

Sistem ini dibangun menggunakan arsitektur *Microservices* yang diorkestrasi melalui Docker Compose, dengan komponen-komponen utama sebagai berikut:

### 1. API Gateway (Nginx)
Berperan sebagai *Reverse Proxy* dan pintu masuk utama (*entry point*) untuk seluruh *traffic* eksternal. Nginx juga memvalidasi setiap token akses secara terpusat dengan memanfaatkan modul `auth_request` ke Auth Service.

### 2. Frontend (`/fe`)
Aplikasi web *Single Page Application* (SPA) interaktif.
- **Tech Stack**: React 19, TypeScript, Vite, TailwindCSS v4, React Router, dan React Query.
- **Fitur Utama**: Dashboard pengguna, manajemen *settings* integrasi ERP, dan tampilan antarmuka Blockchain Explorer untuk melacak blok dan transaksi secara *real-time*.

### 3. Backend Services (`/backend`)
Layanan API internal yang memproses logika bisnis dan database. Dibangun menggunakan **Go (Golang)** dengan *framework* **Gin**.
- **Auth Service**: Menangani proses otentikasi (JWT), manajemen akun, penyimpanan konfigurasi ERP pengguna, dan verifikasi internal token untuk API Gateway.
- **Explorer Service** (Spesifikasi): Mengelola sinkronisasi data (*Background Worker*) yang mengambil data dari ERP klien, mencatatnya ke dalam blockchain, dan menyediakan data tersebut untuk *Frontend Explorer*.
- **Database**: Menggunakan PostgreSQL 15.

---

## 🚀 Cara Menjalankan (*Quick Start*)

Pastikan Anda telah menginstal [Docker](https://www.docker.com/) dan [Docker Compose](https://docs.docker.com/compose/) di mesin Anda.

1. **Jalankan Seluruh Layanan Backend & Nginx**
   Jalankan perintah berikut di *root directory* proyek:
   ```bash
   docker-compose up --build -d
   ```
   *Perintah ini akan menjalankan container untuk Nginx (port 8080), Auth Backend, dan PostgreSQL secara bersamaan.*

2. **Menjalankan Frontend (Lokal)**
   Buka terminal baru, masuk ke folder `fe`, instal dependensi, lalu jalankan *development server*:
   ```bash
   cd fe
   npm install
   npm run dev
   ```

3. **Akses Aplikasi**
   - **Frontend**: Biasanya berjalan di `http://localhost:5173`
   - **API Gateway (Backend/Nginx)**: Tersedia di `http://localhost:8080/api/v1`

---

## 📚 Struktur Direktori & Dokumentasi

Untuk detail lebih lanjut mengenai masing-masing komponen, Anda dapat membaca spesifikasi *API Contract* dan dokumentasi implementasinya di folder masing-masing:

- `nginx/spec.md`: Spesifikasi dan cara kerja API Gateway.
- `backend/auth_service/spec.md`: Spesifikasi API Auth & Settings, beserta Endpoint Internal (Service-to-Service).
- `explorer_service/spec.md`: Spesifikasi API Blockchain Explorer & Background Data Sync Worker.
- `fe/implementation.md`: Rancangan arsitektur dan UI untuk implementasi fitur Blockchain Explorer di Frontend.

---

## 🛠 Konfigurasi Lingkungan (*Environment*)

### Backend
Konfigurasi environment dapat disesuaikan pada file `docker-compose.yaml` (untuk Nginx dan integrasi container) serta `.env` di dalam masing-masing *service* (misal `backend/auth_service/.env`).

### Frontend
Buat file `fe/.env` berdasarkan `fe/.env.example` untuk mengatur URL API:
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```
