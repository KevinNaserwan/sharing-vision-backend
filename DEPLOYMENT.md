# sharing-vision-backend Deployment Guide

## Arsitektur Microservices
- Backend service: `sharing-vision-backend` (Go/Gin)
- Database: MySQL (`article`)
- Frontend service: `sharing-vision-frontend` (Vercel static)

## Konfigurasi reverse proxy
Pastikan domain API diarahkan ke backend service, contoh:
- `be-sharing-vision.meetsin.id` -> `127.0.0.1:8000`
- sertifikat TLS harus valid untuk `be-sharing-vision.meetsin.id`

Template contoh nginx berada di:
- `deploy/nginx-be-sharing-vision.conf`

## Checklist live check (wajib)
1. `curl -I https://be-sharing-vision.meetsin.id/health`
2. `curl -s https://be-sharing-vision.meetsin.id/article/10/0`
3. `curl -s https://sharing-vision-frontend-two.vercel.app/api/article/10/0`
4. Akses UI:
   - All Posts
   - Add New
   - Edit
   - Trashed
   - Preview

## Command verifikasi cepat
```bash
BASE_URL=https://be-sharing-vision.meetsin.id FE_URL=https://sharing-vision-frontend-two.vercel.app \
./scripts/verify-live.sh
```
