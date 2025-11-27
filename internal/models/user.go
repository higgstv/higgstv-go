package models

import (
	"time"
)

// User 使用者模型
type User struct {
	ID                  string    `bson:"_id" json:"_id"`
	Username            string    `bson:"username" json:"username"`
	Email               string    `bson:"email" json:"email"`
	Password            string    `bson:"password" json:"-"` // 不序列化到 JSON
	OwnChannels         []string  `bson:"own_channels" json:"own_channels"`
	UnclassifiedChannel *string  `bson:"unclassified_channel,omitempty" json:"unclassified_channel,omitempty"`
	AccessKey           *string   `bson:"access_key,omitempty" json:"-"`
	Created             time.Time `bson:"created" json:"created"`
	LastModified        time.Time `bson:"last_modified" json:"last_modified"`
}

// UserBasicInfo 使用者基本資訊（用於 owners_info）
type UserBasicInfo struct {
	ID       string `json:"_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

