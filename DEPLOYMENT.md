# Deployment Guide (Sharing Vision Backend)

## Arsitektur produksi

- `article-service`: internal business service di `:8001` (HTTP) dan `:9001` (gRPC)
- `gateway`: public API di `:8000` (proxy semua request REST ke article-service via gRPC)
- `health-service`: endpoint kesiapan di `:8002`

## Reverse proxy domain publik

Gunakan template:
- `deploy/nginx-be-sharing-vision.conf`

Proxy `server_name be-sharing-vision.meetsin.id` ke `127.0.0.1:8000`.

Pastikan sertifikat TLS valid untuk `be-sharing-vision.meetsin.id` dan auto-renew aktif.

## Start (systemd)

Sediakan env:
- `DB_DSN`
- `ALLOWED_ORIGINS`
- `ARTICLE_SERVICE_GRPC_TARGET=127.0.0.1:9001`
- `ARTICLE_HTTP_ADDRESS=:8001`
- `ARTICLE_GRPC_ADDRESS=:9001`
- `GATEWAY_HTTP_ADDRESS=:8000`
- `HEALTH_HTTP_ADDRESS=:8002`

Gunakan unit service:
- `deploy/article.service` (article service)
- `deploy/gateway.service` (public gateway)
- `deploy/health.service` (health endpoint)

Contoh urutan:
1. `article-service` aktif dulu (HTTP + gRPC)
2. `gateway` menyambung ke `ARTICLE_SERVICE_GRPC_TARGET`
3. `health-service` jalan mandiri

## Checklist live check

```bash
curl -I https://be-sharing-vision.meetsin.id/health
curl -s https://be-sharing-vision.meetsin.id/article/10/0
curl -s https://be-sharing-vision.meetsin.id/article/1
curl -s https://sharing-vision-frontend-two.vercel.app/api/article/10/0
```

## Catatan keamanan

- Semua endpoint CRUD tetap melakukan validasi ketat.
- CORS sesuai allowlist.
- Timeout/restriksi request body aktif.
- Header keamanan dijalankan di masing-masing service.
