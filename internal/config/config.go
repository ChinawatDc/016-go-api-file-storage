package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	MaxUploadMB int64
	AllowedExt  map[string]bool

	Provider string

	S3Region          string
	S3Bucket          string
	S3Prefix          string
	S3PublicBaseURL   string
	S3PresignExpireMin int64
	S3Endpoint        string
	S3ForcePathStyle  bool

	GCSBucket         string
	GCSPrefix         string
	GCSPublicBaseURL  string
	GCSSignExpireMin  int64
	GCSCredFile       string
}

func Load() Config {
	_ = godotenv.Load(".env")

	cfg := Config{
		AppPort: getenv("APP_PORT", "8080"),

		MaxUploadMB: getenvInt64("MAX_UPLOAD_MB", 20),
		AllowedExt:  parseAllowedExt(getenv("ALLOWED_EXT", "jpg,jpeg,png,pdf,txt")),

		Provider: strings.ToLower(getenv("STORAGE_PROVIDER", "s3")),

		S3Region:           getenv("S3_REGION", "ap-southeast-1"),
		S3Bucket:           getenv("S3_BUCKET", ""),
		S3Prefix:           getenv("S3_PREFIX", "uploads/"),
		S3PublicBaseURL:    getenv("S3_PUBLIC_BASE_URL", ""),
		S3PresignExpireMin: getenvInt64("S3_PRESIGN_EXPIRE_MIN", 15),
		S3Endpoint:         getenv("S3_ENDPOINT", ""),
		S3ForcePathStyle:   getenvBool("S3_FORCE_PATH_STYLE", false),

		GCSBucket:        getenv("GCS_BUCKET", ""),
		GCSPrefix:        getenv("GCS_PREFIX", "uploads/"),
		GCSPublicBaseURL: getenv("GCS_PUBLIC_BASE_URL", ""),
		GCSSignExpireMin: getenvInt64("GCS_SIGN_EXPIRE_MIN", 15),
		GCSCredFile:      getenv("GCS_CREDENTIALS_FILE", ""),
	}

	log.Println("STORAGE_PROVIDER:", cfg.Provider)
	return cfg
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func getenvInt64(k string, def int64) int64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func getenvBool(k string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(k)))
	if v == "" {
		return def
	}
	return v == "true" || v == "1" || v == "yes"
}

func parseAllowedExt(csv string) map[string]bool {
	m := map[string]bool{}
	for _, p := range strings.Split(csv, ",") {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}
		p = strings.TrimPrefix(p, ".")
		m[p] = true
	}
	return m
}
