# Technical Test - Case Study 2
## Digitalisasi Manajemen Produksi & Inventori

Project ini adalah Proof of Concept (PoC) untuk membantu proses produksi dan pencatatan stok bahan baku pada proses produksi kemeja.

Sistem menggunakan **BOM (Bill of Materials)** untuk menentukan kebutuhan bahan baku, melakukan **reservasi material** saat Work Order dibuat, dan memperbarui stok ketika Work Order selesai.

### Teknologi
- **Backend:** Golang
- **Frontend:** React + Vite
- **Database:** MySQL
- **Database Access:** `database/sql`
- **ORM:** Tidak menggunakan ORM

---

## Quick Start

### Prerequisites
Pastikan sudah terinstall:
- Go 1.22+
- Node.js 18+ dan npm
- MySQL 8+

---

### 1. Setup Database

Buat database MySQL:
```sql
CREATE DATABASE case_study_2
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;
```

Kemudian jalankan file `backend/db/init.sql`. File tersebut berisi:
- Struktur tabel
- Relasi antar tabel
- Seed material
- Seed product
- Seed BOM

Contoh menggunakan MySQL CLI:
```bash
mysql -u root -p case_study_2 < backend/db/init.sql
```
*Jika menggunakan phpMyAdmin atau HeidiSQL, buka file `backend/db/init.sql`, pilih database `case_study_2`, lalu jalankan seluruh isinya.*

---

### 2. Konfigurasi Database

Backend menggunakan environment variable berikut:

| Variable      | Default                 | Deskripsi                  |
| ------------- | ----------------------- | -------------------------- |
| `SERVER_PORT` | `8080`                  | Port HTTP backend          |
| `DB_DRIVER`   | `mysql`                 | Driver database            |
| `DB_DSN`      | Connection string MySQL | Connection string database |

Contoh untuk Windows CMD jika MySQL lokal menggunakan user root tanpa password:
```cmd
set "DB_DSN=root:@tcp(127.0.0.1:3306)/case_study_2?parseTime=true&multiStatements=true"
```
*Jika MySQL menggunakan password, sesuaikan `DB_DSN`. Jangan menyimpan password database pribadi di repository.*

---

### 3. Jalankan Backend

Buka terminal:
```bash
cd backend
go run ./cmd/server
```
Backend berjalan di: `http://localhost:8080` (biarkan terminal backend tetap berjalan).

---

### 4. Jalankan Frontend

Buka terminal baru:
```bash
cd frontend
npm install
npm run dev
```
Frontend berjalan di: `http://localhost:5173` (buka alamat tersebut di browser).

---

## Data Seed / Dummy Data

Project sudah menyediakan seed data agar reviewer bisa langsung mencoba tanpa memasukkan data dari awal.

### Material
| SKU      | Nama Material | Satuan | On Hand |
| -------- | ------------- | ------ | ------- |
| `RM-001` | Kain          | meter  | 1000    |
| `RM-002` | Benang        | gram   | 5000    |
| `RM-003` | Kancing       | pcs    | 1000    |

### Product
| SKU      | Nama Produk | Satuan | On Hand |
| -------- | ----------- | ------ | ------- |
| `FG-001` | Kemeja      | pcs    | 10      |

### BOM FG-001 Version 1
Kebutuhan bahan untuk membuat 1 Kemeja:
| Material | Kebutuhan |
| -------- | --------- |
| Kain     | 1.5 meter |
| Benang   | 50 gram   |
| Kancing  | 5 pcs     |

---

## Demo Flow

Reviewer dapat mencoba alur utama berikut menggunakan seed data:

1. **Buka Inventory**
   - Lihat material dan product.
   - Informasi stok yang ditampilkan: `On Hand`, `Reserved`, `Available` (`Available = On Hand - Reserved`).

2. **Buka BOM**
   - Pilih: `FG-001 - Kemeja`.
   - Kemudian lihat BOM Version 1 dan kebutuhan material.

3. **Buat Work Order**
   - Buat Work Order dengan Product: `FG-001`, Quantity: `2`.
   - Sistem akan menghitung kebutuhan material:
     - Kain: $1.5 \times 2 = 3\text{ meter}$
     - Benang: $50 \times 2 = 100\text{ gram}$
     - Kancing: $5 \times 2 = 10\text{ pcs}$

4. **Reservation**
   - Jika stok cukup, material akan masuk ke `reserved` (Kain: `reserved +3`, Benang: `reserved +100`, Kancing: `reserved +10`).
   - `on_hand` belum berkurang pada tahap ini. Status Work Order menjadi `RESERVED`.

5. **Complete Work Order**
   - Buka detail Work Order dan klik **Complete Work Order**.
   - Setelah berhasil:
     - Kain: 1000 → 997 meter
     - Benang: 5000 → 4900 gram
     - Kancing: 1000 → 990 pcs
     - Kemeja: 10 → 12 pcs
     - Status Work Order berubah menjadi `COMPLETED`.

---

## Business Flow

Alur utama sistem:

```text
Product
   ↓
BOM
   ↓
Create Work Order
   ↓
BOM Explosion
   ↓
Check Available Stock
   ↓
Reserve Material
   ↓
Production
   ↓
Complete Work Order
   ↓
Raw Material Stock ↓ & Finished Goods Stock ↑
```

### BOM Explosion
Kebutuhan material dihitung dengan rumus:
$$\text{Required Quantity} = \text{BOM Quantity} \times \text{Work Order Quantity}$$

*Contoh:* Kain $= 1.5\text{ meter} \times 2 = 3\text{ meter}$.

---

## Concurrency & Transaction

Stok material perlu dijaga agar tidak dialokasikan melebihi jumlah yang tersedia ketika beberapa Work Order diproses pada waktu yang hampir bersamaan.

1. **Database Transaction**
   Proses reservation dijalankan dalam satu transaksi:
   ```text
   BEGIN → Cek stock → Reserve material → Create Work Order → COMMIT
   ```
   Jika salah satu langkah gagal: `ROLLBACK`. Dengan cara ini tidak ada material yang ter-reserve sebagian.

2. **Row Locking**
   Material yang sedang diproses dikunci menggunakan:
   ```sql
   SELECT ... FROM materials WHERE id = ? FOR UPDATE;
   ```
   Dengan cara ini transaksi lain harus menunggu sampai transaksi yang sedang menggunakan data tersebut selesai.

3. **Available Stock**
   Sistem menggunakan rumus `Available = On Hand - Reserved`. Reservation hanya menambah `reserved`. `on_hand` baru berkurang ketika Work Order selesai.

4. **Urutan Penguncian**
   Material dikunci berdasarkan `material_id` secara ascending. Tujuannya untuk mengurangi risiko deadlock ketika beberapa transaksi membutuhkan material yang sama.

5. **Atomic Transaction**
   Pada saat Complete Work Order, perubahan berikut dilakukan dalam satu transaksi:
   - Mengurangi `on_hand` material
   - Mengurangi `reserved` material
   - Mengisi `issued_quantity`
   - Menambah stok finished goods
   - Mengubah status Work Order menjadi `COMPLETED`
   Jika salah satu proses gagal, semua perubahan dalam transaksi di-rollback.

---

## API Endpoints

### Materials
| Method | Endpoint            | Kegunaan                  |
| ------ | ------------------- | ------------------------- |
| GET    | `/api/materials`     | Mengambil semua material  |
| POST   | `/api/materials`     | Menambah material         |
| GET    | `/api/materials/:id` | Mengambil detail material |
| PUT    | `/api/materials/:id` | Mengubah material         |
| DELETE | `/api/materials/:id` | Menghapus material        |

### Products
| Method | Endpoint           | Kegunaan                 |
| ------ | ------------------ | ------------------------ |
| GET    | `/api/products`     | Mengambil semua product  |
| POST   | `/api/products`     | Menambah product         |
| GET    | `/api/products/:id` | Mengambil detail product |
| PUT    | `/api/products/:id` | Mengubah product         |
| DELETE | `/api/products/:id` | Menghapus product        |

### BOM
| Method | Endpoint               | Kegunaan                    |
| ------ | ---------------------- | --------------------------- |
| POST   | `/api/boms`            | Membuat BOM baru            |
| GET    | `/api/products/:id/bom` | Mengambil BOM aktif product |

*BOM menggunakan versioning. Jika recipe berubah, sistem dapat membuat versi BOM baru tanpa mengubah versi sebelumnya.*

### Work Order
| Method | Endpoint                        | Kegunaan                                      |
| ------ | ------------------------------- | --------------------------------------------- |
| POST   | `/api/work-orders`              | Membuat Work Order dan melakukan reservation  |
| GET    | `/api/work-orders/:id`          | Mengambil detail Work Order                   |
| POST   | `/api/work-orders/:id/complete` | Menyelesaikan Work Order dan memperbarui stok |

---

## Testing

### Backend
Jalankan:
```bash
cd backend
gofmt -w .
go build ./...
go test ./...
```
Test mencakup beberapa bagian seperti: validasi Material, validasi Product, validasi BOM, BOM explosion, Work Order, material reservation, insufficient stock, Complete Work Order, transaction rollback, dan MySQL integration test.

### Frontend
Jalankan:
```bash
cd frontend
npm install
npm run build
```
Build frontend berhasil menggunakan Vite.

---

## Assumptions

Beberapa asumsi yang digunakan pada PoC:
- Satu Work Order hanya menggunakan satu Product.
- Satu Work Order menggunakan satu versi BOM.
- Setiap Product hanya memiliki satu BOM aktif pada satu waktu.
- Reservation menggunakan konsep `Available = On Hand - Reserved`.
- Reservation bersifat all-or-nothing.
- Work Order diselesaikan secara penuh dan belum mendukung partial completion.
- Inventory menggunakan satu lokasi.
- BOM yang digunakan pada PoC masih single-level.
- Authentication dan login tidak dibuat karena tidak menjadi fokus PoC.

---

## Scope

### In Scope
- Material CRUD API
- Product CRUD API
- Inventory viewing
- BOM creation & versioning & detail
- BOM explosion
- Work Order & Material reservation (`SELECT ... FOR UPDATE`, atomic transaction)
- Complete Work Order & Stock update
- React frontend PoC
- Unit testing & Integration testing

### Out of Scope
- Authentication / login (JWT / RBAC)
- Multi-warehouse / Multi-location inventory
- Sales Order & Purchase Order
- Shipping / fulfillment
- Omnichannel integration
- Partial Work Order completion / Scrap material tracking
- Production deployment / CI/CD pipeline

---

## Status Pengerjaan

| Komponen                                    | Status       |
| ------------------------------------------- | ------------ |
| Material & Product API                      | Selesai      |
| BOM & Versioning                            | Selesai      |
| Work Order & BOM Explosion                  | Selesai      |
| Material Reservation & Concurrency Handling | Selesai      |
| Complete Work Order                         | Selesai      |
| React Frontend PoC                          | Selesai      |
| Authentication                              | Belum dibuat |
| Multi-warehouse                             | Belum dibuat |
| Omnichannel Integration                     | Belum dibuat |

---

## Known Limitations

1. **Belum Ada Authentication**: Project belum memiliki login atau authentication. Semua endpoint dapat diakses langsung karena fokus PoC adalah business flow produksi dan konsistensi stok.
2. **Satu Lokasi Inventory**: Stok saat ini belum dibedakan berdasarkan gudang atau lokasi.
3. **Belum Mendukung Partial Completion**: Work Order harus diselesaikan penuh. Belum ada proses penyelesaian bertahap.
4. **Belum Ada Sales Order / Purchase Order**: Project hanya berfokus pada proses produksi dan inventory.
5. **Belum Ada Integrasi Omnichannel**: Belum ada integrasi dengan marketplace, website, atau channel penjualan eksternal.
6. **Belum Production Deployment**: Project ini dibuat sebagai PoC untuk technical test.

---

## Struktur Project

```text
case-study-2/
├── backend/
│   ├── cmd/
│   │   ├── server/
│   │   └── testcomplete/
│   ├── config/
│   ├── db/
│   │   └── init.sql
│   └── internal/
│       ├── handler/
│       ├── model/
│       ├── repository/
│       └── service/
│
├── frontend/
│   └── src/
│       ├── api/
│       ├── components/
│       └── pages/
│
└── README.md
```

---

## Architecture

Backend menggunakan pola layered architecture:

```text
React Frontend
      ↓
   REST API
      ↓
   Handler     (Menangani request dan response HTTP)
      ↓
   Service     (Menangani aturan bisnis)
      ↓
 Repository   (Menangani query database)
      ↓
 MySQL / InnoDB (Sumber utama data inventory dan transaction)
```