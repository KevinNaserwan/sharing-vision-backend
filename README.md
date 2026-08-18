# sharing-vision-backend

Backend Golang (Gin + GORM) untuk use case Post Article.

## Fitur yang sudah ada

1. Endpoint:
   - `POST /article/`
   - `GET /article/<limit>/<offset>`
   - `GET /article/<id>`
   - `POST /article/<id>`
   - `PUT /article/<id>`
   - `PATCH /article/<id>`
   - `DELETE /article/<id>`

2. Validasi payload:
   - `title`: wajib, minimal 20 karakter
   - `content`: wajib, minimal 200 karakter
   - `category`: wajib, minimal 3 karakter
   - `status`: hanya `publish`, `draft`, `thrash`

3. Migrasi tabel:
   - SQL migration: `migrations/0001_create_posts.sql`
   - Runner: `go run ./migrations`

4. Postman collection:
   - `postman-collection.json` berisi request untuk semua endpoint.

## Setup

1. Siapkan MySQL dan buat DB `article`.

2. Set DSN:

```bash
export DB_DSN='root:root@tcp(127.0.0.1:3306)/article?charset=utf8mb4&parseTime=True&loc=Local'
```

3. Jalankan migrasi:

```bash
cd sharing-vision-backend
go run ./migrations
```

4. Jalankan service:

```bash
go run ./app
```

Server berjalan di `http://localhost:8000`.
