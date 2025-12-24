package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	cloudstorage "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GCSStorage struct {
	client *cloudstorage.Client
	bucket string
	prefix string

	publicBaseURL string

	googleAccessID string
	privateKey     []byte
}

type GCSOptions struct {
	Bucket        string
	Prefix        string
	PublicBaseURL string
	CredFile      string
}

func NewGCS(ctx context.Context, opt GCSOptions) (*GCSStorage, error) {
	if opt.Bucket == "" {
		return nil, fmt.Errorf("GCS_BUCKET is required")
	}

	var client *cloudstorage.Client
	var err error
	if opt.CredFile != "" {
		client, err = cloudstorage.NewClient(ctx, option.WithCredentialsFile(opt.CredFile))
	} else {
		client, err = cloudstorage.NewClient(ctx)
	}
	if err != nil {
		return nil, err
	}

	s := &GCSStorage{
		client:        client,
		bucket:        opt.Bucket,
		prefix:        cleanPrefix(opt.Prefix),
		publicBaseURL: strings.TrimRight(opt.PublicBaseURL, "/"),
	}

	if opt.CredFile != "" {
		accessID, pk, err := readServiceAccount(opt.CredFile)
		if err == nil {
			s.googleAccessID = accessID
			s.privateKey = pk
		}
	}

	return s, nil
}

func (g *GCSStorage) Put(ctx context.Context, in PutInput) (FileInfo, error) {
	key := g.prefix + strings.TrimLeft(in.Key, "/")

	w := g.client.Bucket(g.bucket).Object(key).NewWriter(ctx)
	w.ContentType = in.ContentType
	if _, err := ioCopy(w, in.Body); err != nil {
		_ = w.Close()
		return FileInfo{}, err
	}
	if err := w.Close(); err != nil {
		return FileInfo{}, err
	}

	fi := FileInfo{Key: key, Size: in.Size, Updated: time.Now()}
	if g.publicBaseURL != "" {
		fi.URL = g.publicBaseURL + "/" + url.PathEscape(key)
	}
	return fi, nil
}

func (g *GCSStorage) Delete(ctx context.Context, key string) error {
	key = strings.TrimLeft(key, "/")
	return g.client.Bucket(g.bucket).Object(key).Delete(ctx)
}

func (g *GCSStorage) List(ctx context.Context, prefix string, limit int) ([]FileInfo, error) {
	pfx := strings.TrimLeft(prefix, "/")
	it := g.client.Bucket(g.bucket).Objects(ctx, &cloudstorage.Query{Prefix: pfx})

	files := make([]FileInfo, 0, limit)
	for len(files) < limit {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		fi := FileInfo{
			Key:     attrs.Name,
			Size:    attrs.Size,
			Updated: attrs.Updated,
		}
		if g.publicBaseURL != "" {
			fi.URL = g.publicBaseURL + "/" + url.PathEscape(fi.Key)
		}
		files = append(files, fi)
	}
	return files, nil
}

func (g *GCSStorage) GetURL(ctx context.Context, key string, expire time.Duration) (string, error) {
	key = strings.TrimLeft(key, "/")

	if g.publicBaseURL != "" {
		return g.publicBaseURL + "/" + url.PathEscape(key), nil
	}

	if g.googleAccessID == "" || len(g.privateKey) == 0 {
		return "", fmt.Errorf("GCS signed url requires service account key (set GCS_CREDENTIALS_FILE)")
	}

	opts := &cloudstorage.SignedURLOptions{
		GoogleAccessID: g.googleAccessID,
		PrivateKey:     g.privateKey,
		Method:         "GET",
		Expires:        time.Now().Add(expire),
	}
	return cloudstorage.SignedURL(g.bucket, key, opts)
}
