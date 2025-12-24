package storage

import (
	"context"
	"io"
	"time"
)

type PutInput struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
}

type FileInfo struct {
	Key      string    `json:"key"`
	Size     int64     `json:"size"`
	Updated  time.Time `json:"updated"`
	URL      string    `json:"url,omitempty"`
}

type Storage interface {
	Put(ctx context.Context, in PutInput) (FileInfo, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string, limit int) ([]FileInfo, error)
	GetURL(ctx context.Context, key string, expire time.Duration) (string, error)
}
