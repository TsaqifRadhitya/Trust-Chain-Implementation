# API Integration Specification - Blockchain Explorer Service

Dokumen ini berisi spesifikasi API Contract untuk **Blockchain Explorer Service** sebagai panduan integrasi bagi tim Frontend (FE).

---

## 1. Informasi Umum (General Info)
- **Base URL**: `http://localhost:8080/api/v1/explorer` (Akan diarahkan Nginx ke service explorer/backend).
- **Format Response**: Menggunakan envelope JSON standar.
  - **Success Response (2xx)**:
    ```json
    {
      "status": 200,
      "message": "Deskripsi sukses",
      "data": {} // Berisi object atau array
    }
    ```

---

## 2. Daftar Endpoint (Endpoints List)

### A. Blocks

#### 1. Get Recent Blocks
Mengambil daftar blok terbaru dengan dukungan *pagination* atau *limit*.

- **Endpoint**: `/blocks`
- **Method**: `GET`
- **Query Params**:
  - `limit` (opsional, default 10): Jumlah blok yang diambil.
  - `page` (opsional, default 1): Halaman data.

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Recent blocks retrieved successfully",
    "data": [
      {
        "height": 12040,
        "hash": "0xabc123...",
        "timestamp": "2026-05-26T14:30:00Z",
        "tx_count": 45,
        "miner": "0xMinerAddress..."
      }
    ]
  }
  ```

#### 2. Get Block Detail
Mengambil informasi lengkap dari satu blok beserta daftar transaksinya.

- **Endpoint**: `/blocks/:hash_or_height`
- **Method**: `GET`

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Block details retrieved",
    "data": {
      "height": 12040,
      "hash": "0xabc123...",
      "parent_hash": "0xabc122...",
      "timestamp": "2026-05-26T14:30:00Z",
      "size": 14500,
      "miner": "0xMinerAddress...",
      "transactions": [
        "0xtx1...", "0xtx2..."
      ]
    }
  }
  ```

---

### B. Transactions

#### 3. Get Recent Transactions
Mengambil daftar transaksi terbaru di seluruh jaringan.

- **Endpoint**: `/transactions`
- **Method**: `GET`
- **Query Params**:
  - `limit` (opsional, default 10).

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Recent transactions retrieved",
    "data": [
      {
        "hash": "0xdef456...",
        "block_height": 12040,
        "from": "0xSender...",
        "to": "0xReceiver...",
        "value": "1.5",
        "fee": "0.001",
        "timestamp": "2026-05-26T14:30:05Z"
      }
    ]
  }
  ```

#### 4. Get Transaction Detail
Mengambil detail informasi satu transaksi.

- **Endpoint**: `/transactions/:hash`
- **Method**: `GET`

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Transaction details retrieved",
    "data": {
      "hash": "0xdef456...",
      "status": "success",
      "block_height": 12040,
      "timestamp": "2026-05-26T14:30:05Z",
      "from": "0xSender...",
      "to": "0xReceiver...",
      "value": "1.5",
      "fee": "0.001",
      "gas_used": 21000,
      "data": "0x..."
    }
  }
  ```

---

### C. Addresses / Accounts

#### 5. Get Address Detail
Mengambil saldo (balance) dan daftar transaksi yang terkait dengan suatu address.

- **Endpoint**: `/address/:address`
- **Method**: `GET`

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Address details retrieved",
    "data": {
      "address": "0xSender...",
      "balance": "150.75",
      "tx_count": 12,
      "transactions": [
        {
          "hash": "0xdef456...",
          "type": "out", // 'in' atau 'out'
          "value": "1.5",
          "timestamp": "2026-05-26T14:30:05Z"
        }
      ]
    }
  }
  ```

---

### D. Global Search (Opsional/Utility)

#### 6. Search
Endpoint cerdas yang mendeteksi input user dan langsung mengembalikan tipe kecocokan (apakah itu blok, transaksi, atau akun).

- **Endpoint**: `/search`
- **Method**: `GET`
- **Query Params**:
  - `q` (wajib): String hash atau address.

- **Response Success (200 OK)**:
  ```json
  {
    "status": 200,
    "message": "Search result found",
    "data": {
      "type": "transaction", // Bisa "block", "transaction", atau "address"
      "hash_or_id": "0xdef456..."
    }
  }
  ```
  *(Frontend dapat memanfaatkannya untuk langsung me-routing user ke `/explorer/tx/0xdef456...`)*


---

## 3. Background Worker (Data Sync)

Selain melayani API publik untuk Explorer, `explorer_service` juga memiliki sebuah **Background Worker** (Daemon/Cron Job) yang bertugas secara spesifik menarik (*pull*) data secara periodik dari sistem klien (misal ERP) untuk kemudian dimasukkan ke dalam jaringan Blockchain.

### Alur Kerja Worker:
1. **Mengambil Konfigurasi Klien (Inter-Service Communication)**: 
   Worker akan memanggil endpoint internal yang disediakan oleh `auth_service` (yakni `GET /api/v1/settings` yang terhubung ke _Setting Service_) guna mengumpulkan daftar konfigurasi aktif seluruh klien.
2. **Koneksi ke Endpoint Klien**: 
   Berdasarkan dari hasil *fetch* data *settings* tersebut, Worker akan membaca field `endpoint`, `api_key`, serta parameter-parameter lain seperti `erp_type`.
3. **Data Fetching (Pull)**: 
   Worker melakukan HTTP request ke alamat `endpoint` klien tersebut, menggunakan `api_key` sebagai bagian dari *header* / *auth* untuk mengamankan komunikasi dan mengambil log atau transaksi klien terbaru.
4. **Proses & Pencatatan ke Ledger**: 
   Data yang berhasil didapatkan akan diproses lebih lanjut, divalidasi, dan dicatat (*minting*/*mining*) ke dalam *block* baru di dalam *chain*. Transaksi-transaksi ini kemudian menjadi sumber data yang akan disajikan secara transparan ke *Frontend* melalui endpoint-endpoint API Explorer (`/blocks` dan `/transactions`) di atas.
