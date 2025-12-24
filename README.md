# 016-go-api-file-storage

แลปนี้เป็นส่วนหนึ่งของซีรีส์ **Go API Course**  
หัวข้อ: **การจัดเก็บไฟล์ด้วย Go API – S3 & Google Cloud Storage (GCS)**

---

## 🎯 เป้าหมายของแลป

- อัปโหลดไฟล์ผ่าน API
- จัดเก็บไฟล์บน **S3-compatible storage** (AWS S3 / MinIO)
- จัดเก็บไฟล์บน **Google Cloud Storage (GCS)**
- สร้าง Public URL หรือ Signed URL สำหรับดาวน์โหลดไฟล์
- ลบไฟล์ และ list ไฟล์ตาม prefix
- ออกแบบ storage layer ให้สลับ provider ได้

---

## 🧱 Tech Stack

- Go
- Gin Framework
- AWS SDK v2 (S3-compatible)
- Google Cloud Storage SDK
- Docker (MinIO สำหรับทดสอบ)
- Environment Config (`godotenv`)

---

## 📁 โครงสร้างโปรเจกต์

```
016-go-api-file-storage/
├─ cmd/api/main.go
├─ internal/
│  ├─ config/config.go
│  ├─ http/handlers/file_handler.go
│  ├─ storage/
│  │  ├─ storage.go
│  │  ├─ s3_storage.go
│  │  ├─ gcs_storage.go
│  │  └─ helpers.go
│  └─ utils/file_utils.go
├─ docker-compose.yml
├─ .env
└─ README.md
```

---

## ⚙️ Environment Variables

### MinIO / S3

```env
APP_PORT=8080
MAX_UPLOAD_MB=20
ALLOWED_EXT=jpg,jpeg,png,pdf,txt

STORAGE_PROVIDER=s3

S3_REGION=us-east-1
S3_BUCKET=my-bucket
S3_PREFIX=uploads/

AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin123

S3_ENDPOINT=http://localhost:9000
S3_FORCE_PATH_STYLE=true
S3_PUBLIC_BASE_URL=http://localhost:9000/my-bucket
S3_PRESIGN_EXPIRE_MIN=15
```

---

### Google Cloud Storage

```env
STORAGE_PROVIDER=gcs

GCS_BUCKET=my-gcs-bucket
GCS_PREFIX=uploads/

GCS_CREDENTIALS_FILE=./service-account.json
GCS_SIGN_EXPIRE_MIN=15
```

---

## 🐳 รัน MinIO

```bash
docker compose up -d
```

---

## ▶️ รัน API

```bash
go run cmd/api/main.go
```

---

## 🔐 API Endpoints

- `POST /files/upload`
- `GET /files/url`
- `GET /files/list`
- `DELETE /files`

---

## 🧠 Key Concepts

- S3-compatible API (รองรับ MinIO)
- Signed URL สำหรับ bucket private
- Storage interface abstraction
- Path-style endpoint

---

## 🚀 Next Steps

- Pre-signed PUT
- DB metadata
- Multipart upload
- JWT protection

---

MIT License
# 016-go-api-file-storage
