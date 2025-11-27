package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// YouTube URL 正則表達式
	youtubeURLRegex = regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com/watch\?v=|youtu\.be/|youtube\.com/embed/)([a-zA-Z0-9_-]{11})`)
	// YouTube ID 正則表達式
	youtubeIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)
)

// RegisterCustomValidators 註冊自訂驗證器
func RegisterCustomValidators(v *validator.Validate) error {
	// YouTube URL 驗證
	if err := v.RegisterValidation("youtube_url", validateYouTubeURL); err != nil {
		return err
	}

	// YouTube ID 驗證
	if err := v.RegisterValidation("youtube_id", validateYouTubeID); err != nil {
		return err
	}

	// 使用者名稱驗證（只允許字母、數字、底線）
	if err := v.RegisterValidation("username", validateUsername); err != nil {
		return err
	}

	// 密碼強度驗證（至少 6 個字元）
	if err := v.RegisterValidation("password", validatePassword); err != nil {
		return err
	}

	return nil
}

// validateYouTubeURL 驗證 YouTube URL
func validateYouTubeURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return true // 空值由 required 驗證處理
	}
	return youtubeURLRegex.MatchString(url)
}

// validateYouTubeID 驗證 YouTube ID
func validateYouTubeID(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	if id == "" {
		return true // 空值由 required 驗證處理
	}
	return youtubeIDRegex.MatchString(id)
}

// validateUsername 驗證使用者名稱
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if username == "" {
		return true // 空值由 required 驗證處理
	}
	
	// 長度檢查
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	// 只允許字母、數字、底線
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return usernameRegex.MatchString(username)
}

// validatePassword 驗證密碼強度
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if password == "" {
		return true // 空值由 required 驗證處理
	}
	
	// 至少 6 個字元
	return len(password) >= 6
}

// SanitizeInput 清理輸入字串
func SanitizeInput(s string) string {
	// 移除前後空白
	s = strings.TrimSpace(s)
	
	// 移除控制字元
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	
	return s
}

// ValidateAndSanitize 驗證並清理字串
func ValidateAndSanitize(s string, minLen, maxLen int) (string, error) {
	sanitized := SanitizeInput(s)
	
	if len(sanitized) < minLen {
		return "", fmt.Errorf("字串長度必須至少 %d 個字元", minLen)
	}
	
	if len(sanitized) > maxLen {
		return "", fmt.Errorf("字串長度不能超過 %d 個字元", maxLen)
	}
	
	return sanitized, nil
}

