package youtube

import "fmt"

// GetThumbnailURL 取得 YouTube 影片縮圖 URL
func GetThumbnailURL(videoID string) string {
	return fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
}

