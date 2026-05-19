// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

func IsImageMimeType(mime string) bool {
	imageMimes := map[string]bool{
		"image/jpeg":    true,
		"image/png":     true,
		"image/gif":     true,
		"image/webp":    true,
		"image/svg+xml": true,
		"image/bmp":     true,
	}
	return imageMimes[mime]
}

func IsImageURL(url string) bool {
	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"}
	for _, ext := range imageExts {
		if len(url) >= len(ext) &&
			url[len(url)-len(ext):] == ext {
			return true
		}
	}
	return false
}
