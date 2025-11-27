package models

import (
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChannelType 頻道類型
type ChannelType string

const (
	ChannelTypeDefault      ChannelType = "default"
	ChannelTypeUnclassified ChannelType = "unclassified"
)

// ChannelCover 頻道封面
type ChannelCover struct {
	Default string `bson:"default" json:"default"`
}

// ChannelPermission 頻道權限
type ChannelPermission struct {
	UserID string `bson:"user_id" json:"user_id"`
	Admin  bool   `bson:"admin" json:"admin"`
	Read   bool   `bson:"read" json:"read"`
	Write  bool   `bson:"write" json:"write"`
}

// Channel 頻道模型
type Channel struct {
	ID            string             `bson:"_id" json:"_id"` // 讀取時使用自訂解碼，寫入時直接使用字串
	Type          ChannelType        `bson:"type" json:"type"`
	Name          string             `bson:"name" json:"name"`
	Desc          string             `bson:"desc" json:"desc"`
	Tags          []int              `bson:"tags" json:"tags"`
	Cover         *ChannelCover      `bson:"cover,omitempty" json:"cover,omitempty"`
	ContentsSeq   string             `bson:"contents_seq" json:"contents_seq"`
	Contents      []Program          `bson:"contents" json:"contents"`
	ContentsOrder []int              `bson:"contents_order" json:"contents_order"`
	Owners        []string           `bson:"owners" json:"owners"`
	Permission    []ChannelPermission `bson:"permission" json:"permission"`
	Created       time.Time          `bson:"created" json:"created"`
	LastModified  time.Time          `bson:"last_modified" json:"last_modified"`
}

// UnmarshalBSON 自訂解碼 UUID（支援 UUID binary 和字串兩種格式）
func (c *Channel) UnmarshalBSON(data []byte) error {
	// 使用臨時結構來避免遞迴呼叫 UnmarshalBSON
	aux := &struct {
		IDRaw        bson.RawValue        `bson:"_id"`
		Type         ChannelType         `bson:"type"`
		Name         string               `bson:"name"`
		Desc         string               `bson:"desc"`
		Tags         []int                `bson:"tags"`
		Cover         *ChannelCover        `bson:"cover,omitempty"`
		ContentsSeq   string               `bson:"contents_seq"`
		Contents      []Program            `bson:"contents"`
		ContentsOrder []int                `bson:"contents_order"`
		Owners        []string             `bson:"owners"`
		Permission    []ChannelPermission  `bson:"permission"`
		Created       time.Time            `bson:"created"`
		LastModified  time.Time            `bson:"last_modified"`
	}{}
	
	// 解碼所有欄位（包括 _id 作為 RawValue）
	if err := bson.Unmarshal(data, aux); err != nil {
		return err
	}
	
	// 複製其他欄位
	c.Type = aux.Type
	c.Name = aux.Name
	c.Desc = aux.Desc
	c.Tags = aux.Tags
	c.Cover = aux.Cover
	c.ContentsSeq = aux.ContentsSeq
	c.Contents = aux.Contents
	c.ContentsOrder = aux.ContentsOrder
	c.Owners = aux.Owners
	c.Permission = aux.Permission
	c.Created = aux.Created
	c.LastModified = aux.LastModified
	
	// 處理 _id 欄位（可能是 UUID binary 或字串）
	switch aux.IDRaw.Type {
	case bson.TypeBinary:
		// UUID binary (subtype 4)
		var binary primitive.Binary
		if err := aux.IDRaw.Unmarshal(&binary); err == nil {
			if binary.Subtype == 4 { // UUID subtype
				c.ID = base64.URLEncoding.EncodeToString(binary.Data)
			} else {
				c.ID = base64.URLEncoding.EncodeToString(binary.Data)
			}
		}
	case bson.TypeString:
		// 字串格式
		var idStr string
		if err := aux.IDRaw.Unmarshal(&idStr); err == nil {
			c.ID = idStr
		}
	default:
		// 嘗試直接解碼為字串
		var idStr string
		if err := aux.IDRaw.Unmarshal(&idStr); err == nil {
			c.ID = idStr
		} else {
			// 最後嘗試 binary
			var binary primitive.Binary
			if err := aux.IDRaw.Unmarshal(&binary); err == nil {
				c.ID = base64.URLEncoding.EncodeToString(binary.Data)
			}
		}
	}
	
	return nil
}

// ChannelWithOwnersInfo 頻道資訊（含擁有者資訊，用於 getchannelinfo API）
type ChannelWithOwnersInfo struct {
	Channel
	OwnersInfo []UserBasicInfo `json:"owners_info"`
}

