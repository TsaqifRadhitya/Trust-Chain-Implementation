# API Integration Specification - Auth & Settings Service

Dokumen ini berisi spesifikasi API Contract untuk **Auth Service** dan **Setting Service** sebagai panduan integrasi bagi tim Frontend (FE).

---

## 1. Informasi Umum (General Info)
- **Base URL**: `http://localhost:8080/api/v1` (atau sesuaikan dengan host deployment).
- **Format Response**: Semua response menggunakan envelope JSON standar dengan format:
  - **Success Response (2xx)**:
    ```json
    {
      "status": 200,
      "message": "Deskripsi singkat keberhasilan",
      "data": {} // Isi data bervariasi sesuai endpoint (dapat berupa object atau array)
    }
    ```
  - **Error Response (4xx / 5xx)**:
    ```json
    {
      "status": 400,
      "message": "Deskripsi singkat kesalahan",
      "error": "Detail pesan error / validation error"
    }
    ```

---

## 2. Alur Autentikasi (Authentication Flow)
1. Frontend mengirim email & password ke endpoint `/auth/login`.
2. Backend merespon dengan membawa `token` (Access Token, exp: 1 jam) dan `refresh_token` (Refresh Token, exp: 7 hari).
3. Simpan `token` di state aplikasi (memory/cookie) dan `refresh_token` di storage aman (secure HTTP-only cookie atau localStorage dengan proteksi).
4. Untuk setiap request yang membutuhkan autentikasi (Protected Routes), kirim Access Token melalui header HTTP:
   ```http
   Authorization: Bearer <access_token>
   ```
5. Jika request menghasilkan error **401 Unauthorized** (Access Token expired), Frontend harus memanggil endpoint `/auth/refresh` secara otomatis (silakan gunakan interceptor Axios) menggunakan `refresh_token` untuk mendapatkan token baru, kemudian mengulangi request awal.

---

## 3. Daftar Endpoint (Endpoints List)

### A. Auth Service

#### 1. Login User
Digunakan untuk menukarkan email & password dengan token autentikasi.

- **Endpoint**: `/auth/login`
- **Method**: `POST`
- **Headers**: `Content-Type: application/json`
- **Request Body**:
  ```json
  {
    "email": "admin@trustchain.com",
    "password": "password123"
  }
  ```
- **Constraint / Validasi**:
  - `email`: Wajib diisi (required), berformat email valid.
  - `password`: Wajib diisi (required).

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Login successful",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIsInR...", // Gunakan untuk Bearer token
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR...", // Simpan untuk memperbarui token
      "user": {
        "email": "admin@trustchain.com",
        "name": "Admin TrustChain"
      }
    }
  }
  ```

- **Response Error (400 Bad Request - Validasi Gagal)**:
  ```json
  {
    "status": 400,
    "message": "Invalid request payload",
    "error": "Key: 'LoginRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
  }
  ```

- **Response Error (401 Unauthorized - Kredensial Salah)**:
  ```json
  {
    "status": 401,
    "message": "Login failed",
    "error": "email atau password salah"
  }
  ```

---

#### 2. Refresh Token
Digunakan saat Access Token telah expired untuk mendapatkan Access Token & Refresh Token baru tanpa memaksa user login kembali (Token Rotation).

- **Endpoint**: `/auth/refresh`
- **Method**: `POST`
- **Headers**: `Content-Type: application/json`
- **Request Body**:
  ```json
  {
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR..."
  }
  ```
- **Constraint / Validasi**:
  - `refresh_token`: Wajib diisi (required).

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Token refreshed successfully",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIsInR_new...", // Access Token baru
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR_new_refresh..." // Refresh Token baru (rotasi)
    }
  }
  ```

- **Response Error (401 Unauthorized - Token Invalid / Expired)**:
  ```json
  {
    "status": 401,
    "message": "Token refresh failed",
    "error": "refresh token tidak valid"
  }
  ```

---

#### 3. Auth Validator (Nginx auth_request)
Endpoint internal yang dirancang khusus untuk digunakan oleh modul `auth_request` di Nginx guna memvalidasi token JWT secara terpusat. Endpoint ini bersifat opsional (tidak memblokir request) dan melempar status otentikasi melalui header.

- **Endpoint**: `/auth/validate`
- **Method**: `GET`
- **Headers**:
  - `Authorization: Bearer <access_token>` *(Opsional)*

- **Perilaku Khusus (Behavior)**:
  - **Selalu mengembalikan HTTP Status `200 OK`**. Tujuannya agar Nginx selalu mengizinkan request lewat dan meneruskannya ke *backend/downstream service* utama.
  - Jika token disertakan dan **valid**, endpoint ini akan mengekstrak ID pengguna dan menyisipkannya pada HTTP Header `X-User-Id` ke dalam *response*. Nginx dapat menangkap *header* ini dan meneruskannya ke layanan downstream.
  - Jika token **tidak ada** atau **invalid/expired**, header `X-User-Id` tidak akan disertakan (kosong). Layanan *downstream* dapat mendeteksi ketidakhadiran *header* ini untuk mengenali status pengguna sebagai tamu (*Guest*).

- **Response (200 OK)**:
  ```http
  HTTP/1.1 200 OK
  X-User-Id: 1   (Opsional, hanya jika token valid)
  ```

---

### B. Setting Service

#### 3. Get User Settings (Protected Route)
Mengambil konfigurasi ERP & sensitivitas fraud detection milik user yang sedang login.

- **Endpoint**: `/settings`
- **Method**: `GET`
- **Headers**: 
  - `Authorization: Bearer <access_token>`
  
- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Settings retrieved successfully",
    "data": {
      "id": 1,
      "user_id": 1,
      "erp_type": "SAP S/4HANA",
      "endpoint": "https://erp.internal.company.com/api/v2/transactions",
      "api_key": "sk-trustchain-default-placeholder",
      "volume_sensitivity": 85,
      "geo_threshold": 50,
      "velocity_limit": 70,
      "updated_at": "2026-05-26T14:04:08.320933+07:00"
    }
  }
  ```

- **Response Error (401 Unauthorized - Token Expired / Hilang)**:
  ```json
  {
    "status": 401,
    "message": "Unauthorized access",
    "error": "Invalid or expired token"
  }
  ```

---

#### 4. Update/Upsert User Settings (Protected Route)
Memperbarui konfigurasi ERP & sensitivitas fraud detection milik user.

- **Endpoint**: `/settings`
- **Method**: `PUT`
- **Headers**: 
  - `Authorization: Bearer <access_token>`
  - `Content-Type: application/json`
- **Request Body**:
  ```json
  {
    "erp_type": "Oracle Fusion",
    "endpoint": "https://oracle.internal.company.com/api/v1/data",
    "api_key": "sk-oracle-updated-key-999",
    "volume_sensitivity": 90,
    "geo_threshold": 40,
    "velocity_limit": 65
  }
  ```
- **Constraint / Validasi**:
  - `erp_type`: Wajib diisi.
  - `endpoint`: Wajib diisi, harus berformat URL valid.
  - `api_key`: Wajib diisi.
  - `volume_sensitivity`: Wajib diisi, angka minimal 0, maksimal 100.
  - `geo_threshold`: Wajib diisi, angka minimal 0, maksimal 100.
  - `velocity_limit`: Wajib diisi, angka minimal 0, maksimal 100.

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Settings updated successfully",
    "data": {
      "id": 1,
      "user_id": 1,
      "erp_type": "Oracle Fusion",
      "endpoint": "https://oracle.internal.company.com/api/v1/data",
      "api_key": "sk-oracle-updated-key-999",
      "volume_sensitivity": 90,
      "geo_threshold": 40,
      "velocity_limit": 65,
      "updated_at": "2026-05-26T14:04:30.8946674+07:00"
    }
  }
  ```

- **Response Error (400 Bad Request - Validasi Input Gagal)**:
  ```json
  {
    "status": 400,
    "message": "Invalid request payload",
    "error": "Key: 'UpdateSettingRequest.VolumeSensitivity' Error:Field validation for 'VolumeSensitivity' failed on the 'max' tag"
  }
  ```

---

### C. Internal Service API (Service-to-Service Communication)

Endpoint di bawah ini didesain **hanya untuk digunakan oleh service internal** (misalnya `explorer_service` atau `case_service`), dan tidak boleh diekspos ke publik oleh Nginx. Komunikasi ini biasanya diamankan melalui *Internal Network* (VPC/Docker Network) atau menggunakan otorisasi antar-service.

#### 5. Get User Data by ID (Internal Route)
Digunakan oleh service lain untuk mengambil data pengguna spesifik tanpa memerlukan Bearer token.

- **Endpoint**: `/internal/users/:id`
- **Method**: `GET`
- **Headers**: Tidak ada (Hanya bisa diakses via internal network)

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "User found",
    "data": {
      "id": 1,
      "email": "admin@trustchain.com",
      "name": "Admin TrustChain"
    }
  }
  ```

- **Response Error (404 Not Found)**:
  ```json
  {
    "status": 404,
    "message": "User not found",
    "error": "Detail pesan error"
  }
  ```

---

#### 6. Get All Users (Internal Route)
Digunakan oleh service lain untuk mengambil seluruh data pengguna tanpa memerlukan Bearer token.

- **Endpoint**: `/internal/users`
- **Method**: `GET`
- **Headers**: Tidak ada (Hanya bisa diakses via internal network)

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Users found",
    "data": [
      {
        "id": 1,
        "email": "admin@trustchain.com",
        "name": "Admin TrustChain"
      },
      {
        "id": 2,
        "email": "user2@trustchain.com",
        "name": "User 2"
      }
    ]
  }
  ```
