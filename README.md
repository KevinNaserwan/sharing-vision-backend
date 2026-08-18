# sharing-vision-backend

Backend microservice (Go + Gin + MySQL) untuk use case **Post Article**.

## Arsitektur
- `cmd/api` (legacy): service tunggal tetap tersedia untuk kebutuhan kompatibilitas lokal.
- `cmd/article-service`: **article-service internal**  
  - HTTP CRUD untuk artikel (`/article/...`)  
  - gRPC CRUD & event stream (`ArticleService`)
  - Publikasi event: `created`, `updated`, `deleted`
- `cmd/gateway`: **gateway publik** untuk memenuhi endpoint HTTP yang dipakai frontend
- `cmd/health-service`: endpoint `/health` dan `/ready` terpisah

`gateway` berjalan sebagai entrypoint publik. `article-service` dan `health-service` dipisah untuk memisahkan concern.

## Struktur Direktori
- `internal/config` → konfigurasi app
- `internal/handler` → handler HTTP
- `internal/service` → business logic + validasi
- `internal/repository` → akses database
- `internal/model` → model `posts`
- `internal/middleware` → security, cors, timeout middleware
- `internal/articlepb` → definisi interface gRPC + codec JSON
- `internal/pubsub` → in-memory event bus (pub/sub)
- `cmd/migrate` → migrasi database
- `migrations` → skema tabel `posts`
- `postman-collection.json` → koleksi Postman sesuai requirement

## Database `posts`

```sql
CREATE TABLE IF NOT EXISTS posts (
  id INT AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(200) NOT NULL,
  content TEXT NOT NULL,
  category VARCHAR(100) NOT NULL,
  status VARCHAR(100) NOT NULL CHECK (status IN ('publish', 'draft', 'thrash')),
  created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## Environment

```bash
cp .env.example .env
```

Isi `.env`:

```bash
APP_ENV=production
SERVER_ADDRESS=:8000
DB_DSN=user:password@tcp(db_host:3306)/article?charset=utf8mb4&parseTime=True&loc=Local
ALLOWED_ORIGINS=https://be-sharing-vision.meetsin.id,https://sharing-vision-frontend-two.vercel.app,https://*.vercel.app,http://localhost:5173
REQUEST_TIMEOUT_SECONDS=15
READ_HEADER_TIMEOUT_SECONDS=5
MAX_REQUEST_BODY_BYTES=1048576
ENABLE_REQUEST_LOGGING=true

# Microservice routing
ARTICLE_HTTP_ADDRESS=:8001
ARTICLE_GRPC_ADDRESS=:9001
GATEWAY_HTTP_ADDRESS=:8000
HEALTH_HTTP_ADDRESS=:8002
ARTICLE_SERVICE_GRPC_TARGET=127.0.0.1:9001
ENABLE_EVENT_CONSUMER=true
```

## Menjalankan

### 1) Migrasi

```bash
go run ./cmd/migrate
```

### 2) Menjalankan microservices

```bash
source .env

# internal article service + gRPC
go run ./cmd/article-service

# API gateway publik (jalankan di terminal lain)
go run ./cmd/gateway

# health service terpisah (opsional)
go run ./cmd/health-service
```

### 3) Menjalankan versi legacy (opsional)

```bash
go run ./cmd/api
```

### 4) Docker

```bash
docker build -t sharing-vision-backend:live .
docker run -d \
  --name sharing-vision-article \
  --network svnet \
  -p 8001:8001 -p 9001:9001 \
  --env-file .env \
  --entrypoint /app/article-service \
  sharing-vision-backend:live \
  /app/article-service

docker run -d \
  --name sharing-vision-gateway \
  --network svnet \
  -p 8000:8000 \
  --env-file .env \
  sharing-vision-backend:live \
  /app/gateway
```

Jika Dockerfile dipakai default (tanpa override command), container menjalankan `gateway`.

## Endpoint (HTTP via gateway/public API)

Prefix: `/article`

1. `POST /article/` → membuat artikel baru
2. `GET /article/{limit}/{offset}` → list dengan pagination
3. `GET /article/{id}` → detail artikel
4. `PUT /article/{id}` → update artikel
5. `PATCH /article/{id}` → update artikel
6. `POST /article/{id}` → update artikel (alias), optional `?action=delete` untuk hapus
7. `DELETE /article/{id}` → hapus artikel
8. `GET /health`, `GET /ready`

## Validasi Input

- `title`: required, minimal 20 karakter
- `content`: required, minimal 200 karakter
- `category`: required, minimal 3 karakter
- `status`: required, salah satu dari `publish`, `draft`, `thrash`

Jika validasi gagal: HTTP `400` dengan payload error validasi.

## gRPC (internal)

`cmd/article-service` mengekspos gRPC `ArticleService` pada `ARTICLE_GRPC_ADDRESS`.

Service:
- `CreateArticle`
- `ListArticles`
- `GetArticle`
- `UpdateArticle`
- `DeleteArticle`
- `SubscribeEvents` (stream event untuk kebutuhan pub/sub/logging)

Encoding gRPC: JSON codec internal (tanpa file `.proto` generation dalam repository).

## Postman

File: `postman-collection.json`

```
cat postman-collection.json
```

## Smoke test cepat

```bash
BASE_URL="${BASE_URL:-https://be-sharing-vision.meetsin.id}"
FE_URL="${FE_URL:-https://sharing-vision-frontend-two.vercel.app}"

curl -sS "$BASE_URL/health"
curl -sS "$BASE_URL/article/10/0"
curl -sS "$FE_URL/api/article/10/0"
```

## Deployment

Lihat [`DEPLOYMENT.md`](./DEPLOYMENT.md) untuk:
- mapping reverse proxy nginx
- checklist live check
- verifikasi HTTPS + status endpoint

## Keamanan

- Validasi ketat pada service layer sebelum operasi database
- CORS allowlist (`ALLOWED_ORIGINS`)
- Security headers (HSTS, CSP, X-Content-Type-Options, X-Frame-Options, dll)
- Recover handler untuk panic
- Batas maksimal body request
- `.env` tidak disimpan di repo (`.gitignore`)

## CI/CD

- Backend CI tersedia di GitHub Actions: `.github/workflows/ci.yml`
- Pipeline memeriksa:
  - install dependency Go
  - pengecekan format `gofmt`
  - unit test
  - `go vet`
- CI otomatis jalan pada setiap `push` dan `pull_request`.

## Daftar Checklist untuk Deployment

- Backend test/CI (lulus): open tab Actions dan cek workflow `backend-ci` berstatus ✅
- Frontend test/CI (lulus): refer repo [sharing-vision-frontend](https://github.com/KevinNaserwan/sharing-vision-frontend), workflow `frontend-ci` berstatus ✅
