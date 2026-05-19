// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package ossx

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestR2UploadAndDelete(t *testing.T) {
	Init(Config{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		Bucket:          os.Getenv("R2_BUCKET"),
		PublicDomain:    os.Getenv("R2_PUBLIC_DOMAIN"),
	})

	if !Enabled() {
		t.Fatal("R2 not configured — check env vars")
	}

	key := MakeKey("test", "hello.txt")
	body := bytes.NewReader([]byte("R2 upload test " + time.Now().String()))

	// Upload
	url, err := Upload(context.Background(), key, body, "text/plain")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	t.Logf("Uploaded to: %s", url)

	// Verify the URL looks correct
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("Bad URL: %s", url)
	}

	// Cleanup
	if err := Delete(context.Background(), key); err != nil {
		t.Logf("Cleanup warning: %v", err)
	}
	t.Logf("Upload + Delete OK")
}

func TestPresignedURL(t *testing.T) {
	Init(Config{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		Bucket:          os.Getenv("R2_BUCKET"),
		PublicDomain:    os.Getenv("R2_PUBLIC_DOMAIN"),
	})

	if !Enabled() {
		t.Skip("R2 not configured")
	}

	url, err := PresignedPutURL(context.Background(), "test/presigned.txt", 10*time.Minute)
	if err != nil {
		t.Fatalf("PresignedPutURL failed: %v", err)
	}
	if !strings.Contains(url, "X-Amz-Signature") {
		t.Errorf("Expected signed URL, got: %s", url)
	}
	t.Logf("Presigned URL OK")
}
