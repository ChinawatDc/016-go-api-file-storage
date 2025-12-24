package storage

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Storage struct {
	client    *s3.Client
	presigner *s3.PresignClient

	bucket        string
	prefix        string
	publicBaseURL string
}

type S3Options struct {
	Region         string
	Bucket         string
	Prefix         string
	PublicBaseURL  string
	Endpoint       string
	ForcePathStyle bool

	AccessKey string
	SecretKey string
}

func NewS3(ctx context.Context, opt S3Options) (*S3Storage, error) {
	if opt.Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}

	loadOpts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(opt.Region),
	}
	if opt.AccessKey != "" && opt.SecretKey != "" {
		loadOpts = append(loadOpts, awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opt.AccessKey, opt.SecretKey, ""),
		))
	}

	cfg, err := awscfg.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, err
	}

	s3opt := func(o *s3.Options) {
		o.Region = opt.Region

		if opt.Endpoint != "" {
			o.EndpointResolver = s3.EndpointResolverFromURL(opt.Endpoint)
		}

		o.UsePathStyle = opt.ForcePathStyle
	}

	client := s3.NewFromConfig(cfg, s3opt)

	return &S3Storage{
		client:        client,
		presigner:     s3.NewPresignClient(client),
		bucket:        opt.Bucket,
		prefix:        cleanPrefix(opt.Prefix),
		publicBaseURL: strings.TrimRight(opt.PublicBaseURL, "/"),
	}, nil
}

func cleanPrefix(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

func (s *S3Storage) Put(ctx context.Context, in PutInput) (FileInfo, error) {
	key := s.prefix + strings.TrimLeft(in.Key, "/")

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        in.Body,
		ContentType: aws.String(in.ContentType),
	})
	if err != nil {
		return FileInfo{}, err
	}

fi := FileInfo{Key: key, Size: in.Size, Updated: time.Now()}

	if s.publicBaseURL != "" {
		fi.URL = s.publicBaseURL + "/" + url.PathEscape(key)
	}

	return fi, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	key = strings.TrimLeft(key, "/")
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Storage) List(ctx context.Context, prefix string, limit int) ([]FileInfo, error) {
	pfx := strings.TrimLeft(prefix, "/")
	out, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucket),
		Prefix:  aws.String(pfx),
		MaxKeys: int32Ptr(int32(limit)),
	})
	if err != nil {
		return nil, err
	}

	files := make([]FileInfo, 0, len(out.Contents))
	for _, o := range out.Contents {
		fi := FileInfo{
			Key:     aws.ToString(o.Key),
			Size:    aws.ToInt64(o.Size),
			Updated: aws.ToTime(o.LastModified),
		}
		if s.publicBaseURL != "" {
			fi.URL = s.publicBaseURL + "/" + url.PathEscape(fi.Key)
		}
		files = append(files, fi)
	}
	return files, nil
}

func (s *S3Storage) GetURL(ctx context.Context, key string, expire time.Duration) (string, error) {
	key = strings.TrimLeft(key, "/")

	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + url.PathEscape(key), nil
	}
	ps, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expire))
	if err != nil {
		return "", err
	}
	return ps.URL, nil
}

func int32Ptr(v int32) *int32 { return &v }

func join(parts ...string) string { return path.Join(parts...) }
var _ = types.ObjectCannedACLPrivate
