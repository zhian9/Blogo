// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

// Package ossx provides Cloudflare R2 object storage integration.
// R2 is S3-compatible — uses AWS SDK v2 with a custom endpoint.
package ossx

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var client *s3.Client
var bucket string
var publicDomain string
var uploadDir string

// Config holds R2 connection parameters.
type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicDomain    string
	Endpoint        string // optional, auto-generated from AccountID if empty
	UploadDir       string
}

// Enabled returns true if R2 is configured.
func Enabled() bool { return client != nil }

// Init creates the S3 client. Call once during bootstrap.
func Init(cfg Config) {
	if cfg.AccountID == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.Bucket == "" {
		log.Println("[ossx] R2 not configured — uploads will use local disk")
		return
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)
	}

	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				SigningRegion:     "auto",
				HostnameImmutable: true,
			}, nil
		},
	)

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Printf("[ossx] failed to create S3 client: %v", err)
		return
	}

	client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	bucket = cfg.Bucket
	publicDomain = cfg.PublicDomain
	uploadDir = cfg.UploadDir
	log.Println("[ossx] R2 client initialized, bucket:", bucket)
}

// PresignedPutURL generates a temporary URL the frontend can use to
// directly upload a file to R2 without going through the server.
func PresignedPutURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if client == nil {
		return "", fmt.Errorf("ossx: R2 not configured")
	}
	presignClient := s3.NewPresignClient(client)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) {
		o.Expires = ttl
	})
	if err != nil {
		return "", fmt.Errorf("ossx: presign failed: %w", err)
	}
	return req.URL, nil
}

// PublicURL returns the CDN / public URL for an uploaded file.
// Uses publicDomain if set, otherwise falls back to R2 dev domain.
func PublicURL(key string) string {
	if key == "" {
		return ""
	}
	if publicDomain != "" {
		return fmt.Sprintf("https://%s/%s", publicDomain, key)
	}
	return fmt.Sprintf("https://pub-%s.r2.dev/%s", bucket, key)
}

// UploadDir returns the configured directory prefix for uploads.
func UploadDir() string {
	return uploadDir
}

// MakeKey builds an S3 object key: <uploadDir>/<category>/<year>/<month>/<filename>
func MakeKey(category, filename string) string {
	now := time.Now()
	return fmt.Sprintf("%s/%s/%s/%s/%s",
		uploadDir, category,
		now.Format("2006"), now.Format("01"),
		filename,
	)
}

// Upload reads data from r and writes it to R2 at the given key.
// Returns the public URL on success.
func Upload(ctx context.Context, key string, r io.Reader, contentType string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("ossx: R2 not configured")
	}
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("ossx: upload failed: %w", err)
	}
	return PublicURL(key), nil
}

// Delete removes a file from R2.
func Delete(ctx context.Context, key string) error {
	if client == nil {
		return nil
	}
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// ParseKey extracts the S3 key from a full URL.
// e.g. "https://cdn.example.com/uploads/avatar/2026/05/abc.png" → "uploads/avatar/2026/05/abc.png"
func ParseKey(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// strip leading slash
	if len(u.Path) > 0 && u.Path[0] == '/' {
		return u.Path[1:], nil
	}
	return u.Path, nil
}
