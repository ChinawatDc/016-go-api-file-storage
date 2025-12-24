package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/ChinawatDc/016-go-api-file-storage/internal/config"
	"github.com/ChinawatDc/016-go-api-file-storage/internal/http/handlers"
	"github.com/ChinawatDc/016-go-api-file-storage/internal/storage"
)

func main() {
	cfg := config.Load()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.MaxMultipartMemory = cfg.MaxUploadMB * 1024 * 1024

	ctx := context.Background()

	var st storage.Storage
	switch cfg.Provider {
	case "s3":
		s, err := storage.NewS3(ctx, storage.S3Options{
			Region:         cfg.S3Region,
			Bucket:         cfg.S3Bucket,
			Prefix:         cfg.S3Prefix,
			PublicBaseURL:  cfg.S3PublicBaseURL,
			Endpoint:       cfg.S3Endpoint,
			ForcePathStyle: cfg.S3ForcePathStyle,
			AccessKey:      os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretKey:      os.Getenv("AWS_SECRET_ACCESS_KEY"),
		})
		if err != nil {
			log.Fatal(err)
		}
		st = s

	case "gcs":
		g, err := storage.NewGCS(ctx, storage.GCSOptions{
			Bucket:        cfg.GCSBucket,
			Prefix:        cfg.GCSPrefix,
			PublicBaseURL: cfg.GCSPublicBaseURL,
			CredFile:      cfg.GCSCredFile,
		})
		if err != nil {
			log.Fatal(err)
		}
		st = g

	default:
		log.Fatalf("unknown STORAGE_PROVIDER=%s (use s3 or gcs)", cfg.Provider)
	}

	h := handlers.NewFileHandler(st, cfg.AllowedExt, cfg.MaxUploadMB, "uploads/")

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	api := r.Group("/files")
	{
		api.POST("/upload", h.Upload)
		api.GET("/url", h.GetURL)
		api.GET("/list", h.List)
		api.DELETE("", h.Delete)
	}

	log.Println("listening on :" + cfg.AppPort)
	_ = r.Run(":" + cfg.AppPort)
}
