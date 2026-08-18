# sharing-vision-backend

Backend microservice untuk use case **Post Article** (Golang + Gin + GORM) dengan MySQL.

## Arsitektur

- `cmd/api`   : entrypoint HTTP API
- `cmd/migrate`: runner migrasi SQL
- `internal`  : config, middleware, service, repository, model, storage, docs
- `migrations`: file SQL DDL
- `postman-collection.json`: contoh request untuk seluruh endpoint

## Struktur Tabel `posts`

```sql
CREATE TABLE posts (
  id INT AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(200) NOT NULL,
  content TEXT NOT NULL,
  category VARCHAR(100) NOT NULL,
  status VARCHAR(100) NOT NULL,
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
- `GET /docs` (Swagger UI), `GET /swagger.json`

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

## Deployment HTTPS ke `https://be-sharing-vision.meetsin.id`

Direkomendasikan deploy di VPS gratis berbiaya rendah/ gratis trial dengan Docker:

### 1) Build image

```bash
docker build -t sharing-vision-backend:latest /home/meetsin/sharing-vision-backend
```

### 2) Run service

```bash
docker run -d \
  --name sharing-vision-backend \
  -p 127.0.0.1:8000:8000 \
  -e APP_ENV=production \
  -e DB_DSN='user:password@tcp(mysql_host:3306)/article?charset=utf8mb4&parseTime=True&loc=Local' \
  sharing-vision-backend:latest
```

### 3) Reverse proxy + SSL (Nginx + Certbot)

1. Salin `deploy/nginx-be-sharing-vision.conf` ke `/etc/nginx/sites-available/` lalu symlink ke `sites-enabled`.
2. Pasang DNS `A record` `be-sharing-vision.meetsin.id` ke server.
3. Enable site dan reload nginx.
4. Jalankan Certbot:

```bash
certbot --nginx -d be-sharing-vision.meetsin.id
```

### 4) Opsi systemd

Copy binary ke `/opt/sharing-vision-backend/article-service`, ubah `YOUR_DSN` di `deploy/article.service`, lalu aktifkan:

```bash
sudo cp deploy/article.service /etc/systemd/system/article.service
sudo systemctl daemon-reload
sudo systemctl enable --now article.service
```

## Postman

Import file `postman-collection.json`.

```bash
cat postman-collection.json
```

## Keamanan yang diterapkan

- Validasi ketat payload
- CORS origin allowlist
- Header keamanan dasar (X-Content-Type-Options, CSP, X-Frame-Options, dll)
- Recovery handler (JSON)
- Maksimal ukuran body request
- Hardening rekomendasi untuk service (systemd hardening + reverse proxy)

## Swagger

Buka: `https://be-sharing-vision.meetsin.id/docs`
