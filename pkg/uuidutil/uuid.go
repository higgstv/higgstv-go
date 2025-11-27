package uuidutil

import (
	"encoding/base64"
	"github.com/google/uuid"
)

// NewBase64UUID 產生新的 Base64 編碼 UUID
func NewBase64UUID() string {
	id := uuid.New()
	return base64.URLEncoding.EncodeToString(id[:])
}

// ParseBase64UUID 解析 Base64 編碼的 UUID
func ParseBase64UUID(s string) (uuid.UUID, error) {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.FromBytes(data)
}

