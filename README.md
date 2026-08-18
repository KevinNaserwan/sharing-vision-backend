# sharing-vision-backend

Backend Golang (Gin + GORM) untuk use case Post Article.

## Endpoints

- `POST /article/` : create article
- `GET /article/<limit>/<offset>` : list with pagination
- `GET /article/<id>` : get article detail
- `POST /article/<id>` : update article
- `PUT /article/<id>` : update article
- `PATCH /article/<id>` : update article
- `DELETE /article/<id>` : delete article

## Validasi Payload

- `title` (required, min 20)
- `content` (required, min 200)
- `category` (required, min 3)
- `status` (required, must be `publish`, `draft`, `thrash`)

## Setup Lokal

### 1) Database

Buat database MySQL:

```bash
CREATE DATABASE article;
```

### 2) Environment

Buat `.env`/shell:

```bash
export DB_DSN='root:root@tcp(127.0.0.1:3306)/article?charset=utf8mb4&parseTime=True&loc=Local'
```

### 3) Migrasi

```bash
cd sharing-vision-backend
go run ./migrations
```

### 4) Run API

```bash
go run ./app
```

Server berjalan di `http://localhost:8000`.

## Deploy Backend (opsional, gratis)

### Render (gratis)

1. Push repo ke GitHub.
2. Buat Web Service baru.
3. Build Command: `go build -o app ./app`
4. Start Command: `./app`
5. Tambah environment variable `DB_DSN` mengarah ke MySQL kamu.

### Railway (gratis tier)

1. Import repository.
2. Environment variable: `DB_DSN`.
3. Start command: `go run ./app`.

### Docker

```bash
docker build -t sharing-vision-backend .
docker run -p 8000:8000 -e DB_DSN='user:pass@tcp(host:3306)/article?charset=utf8mb4&parseTime=True&loc=Local' sharing-vision-backend
```

## Postman Collection

File: `postman-collection.json`
