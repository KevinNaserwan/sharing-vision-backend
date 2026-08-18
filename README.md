# sharing-vision-backend

Backend microservice **Post Article** (Go + Gin + GORM + MySQL).

## Struktur Proyek
- `cmd/api`: HTTP API
- `cmd/migrate`: runner migrasi SQL
- `internal`: config, middleware, service, repository, model, storage
- `migrations`: skema tabel `posts`
- `postman-collection.json`: koleksi endpoint untuk pengujian
- `deploy/nginx-be-sharing-vision.conf`: template reverse proxy domain publik

## Tabel `posts`

```sql
CREATE TABLE IF NOT EXISTS posts (
  id INT AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(200) NOT NULL,
  content TEXT NOT NULL,
  category VARCHAR(100) NOT NULL,
  status VARCHAR(100) NOT NULL CHECK (status IN ('publish', 'draft', 'thrash')),
  created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

## Prasyarat
- Go 1.22+
- MySQL/MariaDB
- Database `article`

## Konfigurasi Environment

```bash
cp .env.example .env
```

Isi `.env`:

```bash
APP_ENV=production
SERVER_ADDRESS=:8000
DB_DSN=user:password@tcp(db_host:3306)/article?charset=utf8mb4&parseTime=True&loc=Local
ALLOWED_ORIGINS=https://be-sharing-vision.meetsin.id,https://*.vercel.app,http://localhost:5173
REQUEST_TIMEOUT_SECONDS=15
READ_HEADER_TIMEOUT_SECONDS=5
MAX_REQUEST_BODY_BYTES=1048576
ENABLE_REQUEST_LOGGING=true
```

## Menjalankan

### 1) Migrasi

```bash
go run ./cmd/migrate
```

### 2) Menjalankan API

```bash
source .env
go run ./cmd/api
```

Service berjalan di `:8000`.

### 3) Docker

```bash
docker build -t sharing-vision-backend:live .
docker run -d \
  --name sharing-vision-backend \
  --network svnet \
  -p 8000:8000 \
  --env-file .env \
  sharing-vision-backend:live
```

## Endpoint

Prefix: `/article`

1. `POST /article/` → Create article baru
2. `GET /article/{limit}/{offset}` → List artikel dengan pagination
3. `GET /article/{id}` → Detail artikel by id
4. `PUT /article/{id}` → Update artikel by id
5. `PATCH /article/{id}` → Update artikel by id
6. `POST /article/{id}` → Update by id (alias), atau hapus jika `?action=delete`
7. `DELETE /article/{id}` → Hapus artikel by id

Health:
- `GET /health`
- `GET /ready`

## Validasi Request

Semua input harus memenuhi:
- `title`: required, minimal 20 karakter
- `content`: required, minimal 200 karakter
- `category`: required, minimal 3 karakter
- `status`: required, harus salah satu dari `publish`, `draft`, `thrash`

Jika validasi gagal: response `400` dengan detail error per field.

## Postman

Import `postman-collection.json`.

```bash
cat postman-collection.json
```

## Deployment (VPS)

1. Build image backend
2. Jalankan container
3. Pasang reverse proxy host untuk domain publik ke `127.0.0.1:8000`
4. Aktifkan TLS (Let’s Encrypt) pada domain `be-sharing-vision.meetsin.id`

Template reverse proxy: `deploy/nginx-be-sharing-vision.conf`

## Keamanan
- Validasi dan sanitasi payload di service layer
- CORS allowlist sesuai `ALLOWED_ORIGINS`
- Security headers (HSTS, CSP, X-Content-Type-Options, X-Frame-Options, dll)
- Recover JSON handler untuk panic
- Batas maksimal body request
- `.env` dikecualikan via `.gitignore`

## Referensi
- [cmd/api/main.go](/cmd/api/main.go)
- [cmd/migrate/main.go](/cmd/migrate/main.go)
- [internal/config/config.go](/internal/config/config.go)
- [internal/service/post_service.go](/internal/service/post_service.go)
- [internal/handler/article_handler.go](/internal/handler/article_handler.go)
- [internal/repository/post_repository.go](/internal/repository/post_repository.go)
- [postman-collection.json](/postman-collection.json)
