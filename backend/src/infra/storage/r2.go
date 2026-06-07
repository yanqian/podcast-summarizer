package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	appconfig "podcast-summarizer/src/config"
)

type R2Uploader struct {
	client *s3.Client
	bucket string
	base   string
}

func NewR2Uploader(ctx context.Context, cfg appconfig.Config) (*R2Uploader, error) {
	if cfg.R2Bucket == "" || cfg.R2Endpoint == "" || cfg.R2AccessKey == "" || cfg.R2SecretKey == "" {
		return nil, fmt.Errorf("missing R2 configuration")
	}
	endpoint := sanitizeEndpoint(cfg.R2Endpoint)
	creds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(cfg.R2AccessKey, cfg.R2SecretKey, ""),
	)
	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("auto"),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(endpoint)
	})
	return &R2Uploader{client: client, bucket: cfg.R2Bucket, base: cfg.R2PublicBaseURL}, nil
}

func (u *R2Uploader) Upload(ctx context.Context, key string, localPath string) (string, error) {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", err
	}
	_, err = u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return "", err
	}
	if u.base != "" {
		return fmt.Sprintf("%s/%s", u.base, key), nil
	}
	return key, nil
}

// sanitizeEndpoint drops any path suffix from the endpoint to avoid duplicating bucket/key segments.
func sanitizeEndpoint(raw string) string {
	if raw == "" {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return strings.TrimRight(raw, "/")
	}
	parsed.Path = ""
	parsed.RawPath = ""
	return strings.TrimRight(parsed.String(), "/")
}
