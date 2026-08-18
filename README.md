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

```sql
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

## Deploy Backend (gratis, disarankan)

### Render

1. Buat New Web Service.
2. Set build command: `go build -o app ./app`
3. Set start command: `./app`
4. Tambah env `DB_DSN` ke MySQL host.

### Railway

1. Import repo.
2. Env: `DB_DSN`
3. Start command: `go run ./app`

### Docker

```bash
docker build -t sharing-vision-backend .
docker run -p 8000:8000 -e DB_DSN='user:pass@tcp(host:3306)/article?charset=utf8mb4&parseTime=True&loc=Local' sharing-vision-backend
```

## Postman Collection

File: `postman-collection.json`

## Repo

- Backend: https://github.com/KevinNaserwan/sharing-vision-backend
