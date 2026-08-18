# sharing-vision-backend

Backend microservice untuk use case **Post Article** (Golang + Gin + GORM) dengan MySQL.

## Arsitektur

- `cmd/api`   : entrypoint HTTP API
- `cmd/migrate`: runner migrasi SQL
- `internal`  : config, middleware, service, repository, model, storage, docs
- `migrations`: file SQL DDL
- `postman-collection.json`: request collection untuk semua endpoint

## Struktur Tabel `posts`

```sql
CREATE TABLE posts (
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

## Menyiapkan Database

```sql
CREATE DATABASE article;
```

## Menjalankan Migrasi

```bash
cp .env.example .env
# edit .env sesuai environment
touch .env
source .env

go run ./cmd/migrate
```

## Menjalankan API Lokal

```bash
cp .env.example .env
# edit DB_DSN jika perlu
source .env

go run ./cmd/api
```

Service berjalan di `:8000`.

## Endpoint

Semua endpoint prefiks `/article`.

1. `POST /article/` → create article
2. `GET /article/{limit}/{offset}` → list pagination
3. `GET /article/{id}` → get by id
4. `PUT /article/{id}` → update by id
5. `PATCH /article/{id}` → update by id
6. `POST /article/{id}` → update by id (alias), atau delete jika `action=delete`
7. `DELETE /article/{id}` → delete by id

Health check
- `GET /health`
- `GET /ready`
- `GET /swagger.json` (opsional untuk dokumentasi OpenAPI bila dibutuhkan)

## Validasi Payload

Sebelum create/update, wajib dipenuhi:

- `title`: required, minimal **20 karakter**
- `content`: required, minimal **200 karakter**
- `category`: required, minimal **3 karakter**
- `status`: required, salah satu dari `publish`, `draft`, `thrash`

Jika gagal validasi: response `400` dengan format:

```json
{
  "message": "validation failed (n issue)",
  "errors": {
    "title": "minimum 20 characters"
  }
}
```

## Deployment ke VPS (Domain `https://be-sharing-vision.meetsin.id`)

Backend ini dijalankan di VPS langsung (Docker), bukan Vercel.

### 1) Build image

```bash
docker build --network host --build-arg GOPROXY_URL=https://goproxy.cn,direct -t sharing-vision-backend:live .
```

### 2) Run container (internal)

Pastikan MySQL (`article` database + tabel `posts`) sudah tersedia.

```bash
docker run -d \
  --name sharing-vision-backend \
  --network svnet \
  -p 8000:8000 \
  -e APP_ENV=production \
  -e SERVER_ADDRESS=:8000 \
  -e DB_DSN='root:svroot123@tcp(sv-mysql:3306)/article?charset=utf8mb4&parseTime=True&loc=Local' \
  -e ALLOWED_ORIGINS='https://be-sharing-vision.meetsin.id,https://*.vercel.app,http://localhost:5173,http://localhost:3000' \
  sharing-vision-backend:live
```

### 3) Health check

```bash
curl https://be-sharing-vision.meetsin.id/health
```

### 4) HTTPS (SSL)

Aktifkan SSL via webserver reverse proxy host (contoh Nginx/Apache) dengan ACME certificate untuk domain `be-sharing-vision.meetsin.id` lalu arahkan proxy ke:

- `http://127.0.0.1:8000` (mode HTTP), atau
- `http://127.0.0.1:8000` (mode HTTPS terminasi TLS di reverse proxy).

Endpoint yang digunakan FE:
- `https://be-sharing-vision.meetsin.id`
- `https://be-sharing-vision.meetsin.id/docs`

## Postman

Import `postman-collection.json`.

```bash
cat postman-collection.json
```

Catatan: Postman Collection di repository ini sudah menjadi sumber dokumentasi request utama. Swagger UI tidak dijadikan alat verifikasi penilaian.

## Keamanan yang diterapkan

- Validasi ketat payload
- CORS origin allowlist
- Header keamanan dasar (X-Content-Type-Options, CSP, X-Frame-Options, dll)
- Recovery handler (JSON)
- Maksimal ukuran body request
- `.env` diproteksi via `.gitignore`

## File penting

- [cmd/api/main.go](/cmd/api/main.go)
- [cmd/migrate/main.go](/cmd/migrate/main.go)
- [internal/config/config.go](/internal/config/config.go)
- [internal/service/post_service.go](/internal/service/post_service.go)
- [internal/handler/article_handler.go](/internal/handler/article_handler.go)
- [internal/repository/post_repository.go](/internal/repository/post_repository.go)
- [internal/docs/openapi.json](/internal/docs/openapi.json)
- [postman-collection.json](/postman-collection.json)
