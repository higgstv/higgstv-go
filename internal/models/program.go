package models

import "time"

// ProgramType 節目類型
type ProgramType string

const (
	// ProgramTypeYouTube YouTube 節目類型
	ProgramTypeYouTube ProgramType = "youtube"
)

// Program 節目模型
type Program struct {
	ID          int        `bson:"_id" json:"_id"`
	Name        string     `bson:"name" json:"name"`
	Desc        string     `bson:"desc" json:"desc"`
	Duration    int        `bson:"duration" json:"duration"` // 秒
	Type        ProgramType `bson:"type" json:"type"`
	YouTubeID   string     `bson:"youtube_id" json:"youtube_id"`
	Tags        []int      `bson:"tags" json:"tags"`
	Created     time.Time  `bson:"created" json:"created"`
	LastModified time.Time `bson:"last_modified" json:"last_modified"`
}

