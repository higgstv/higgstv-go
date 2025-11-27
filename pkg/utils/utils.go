package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// GenerateRandomString 產生隨機字串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// SanitizeString 清理字串（移除危險字元）
func SanitizeString(s string) string {
	// 移除控制字元
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

// ValidateJSONPCallback 驗證 JSONP callback 參數
func ValidateJSONPCallback(callback string) bool {
	if callback == "" {
		return false
	}
	// 只允許字母、數字、底線和點
	for _, r := range callback {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.') {
			return false
		}
	}
	return true
}

