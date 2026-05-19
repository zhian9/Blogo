// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// MD5 hash
func MD5(b []byte) string {
	h := md5.New()
	_, _ = h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// MD5String 返回md5加密后的字符串
func MD5String(s string) string {
	return MD5([]byte(s))
}

// SHA1 hash
func SHA1(b []byte) string {
	h := sha1.New()
	_, _ = h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// SHA1String 返回sha1加密后的字符串
func SHA1String(s string) string {
	return SHA1([]byte(s))
}

// GeneratePassword 封装bcrypt.GenerateFromPassword
func GeneratePassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CompareHashAndPassword 封装 bcrypt.CompareHashAndPassword
func CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
