package service

import (
	"context"
	"errors"

	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/models"
	"github.com/higgstv/higgstv-go/pkg/youtube"
)

// ProgramService 節目服務
type ProgramService struct {
	programRepo  database.ProgramRepository
	channelRepo database.ChannelRepository
}

// NewProgramService 建立節目服務
func NewProgramService(programRepo database.ProgramRepository, channelRepo database.ChannelRepository) *ProgramService {
	return &ProgramService{
		programRepo:  programRepo,
		channelRepo: channelRepo,
	}
}

// AddProgram 新增節目
func (s *ProgramService) AddProgram(ctx context.Context, channelID string, name, youtubeID, desc string, duration int, tags []int, updateCover bool) (*models.Program, error) {
	if name == "" || youtubeID == "" {
		return nil, errors.New("name and youtube_id are required")
	}

	program := &models.Program{
		Name:      name,
		Desc:      desc,
		Duration:  duration,
		Type:      models.ProgramTypeYouTube,
		YouTubeID: youtubeID,
		Tags:      tags,
	}

	if err := s.programRepo.AddProgram(ctx, channelID, program); err != nil {
		return nil, err
	}

	// 如果需要更新頻道封面
	if updateCover {
		thumbnailURL := youtube.GetThumbnailURL(youtubeID)
		update := map[string]interface{}{
			"cover.default": thumbnailURL,
		}
		if err := s.channelRepo.Update(ctx, channelID, update); err != nil {
			// 封面更新失敗不影響節目新增，只記錄錯誤
			// 可以考慮加入 logger 記錄
			_ = err // 忽略錯誤，繼續執行
		}
	}

	return program, nil
}

// UpdateProgram 更新節目
func (s *ProgramService) UpdateProgram(ctx context.Context, channelID string, programID int, name, youtubeID, desc string, duration *int, tags []int, updateCover bool) (*models.Program, error) {
	update := make(map[string]interface{})

	if name != "" {
		update["contents.$.name"] = name
	}
	if youtubeID != "" {
		update["contents.$.youtube_id"] = youtubeID
	}
	if desc != "" {
		update["contents.$.desc"] = desc
	}
	if duration != nil {
		update["contents.$.duration"] = *duration
	}
	if tags != nil {
		update["contents.$.tags"] = tags
	}

	if len(update) == 0 {
		return nil, errors.New("no fields to update")
	}

	// 更新節目
	if err := s.programRepo.UpdateProgram(ctx, channelID, programID, update); err != nil {
		return nil, err
	}

	// 如果需要更新頻道封面
	if updateCover {
		thumbnailURL := youtube.GetThumbnailURL(youtubeID)
		coverUpdate := map[string]interface{}{
			"cover.default": thumbnailURL,
		}
		if err := s.channelRepo.Update(ctx, channelID, coverUpdate); err != nil {
			// 封面更新失敗不影響節目更新，只記錄錯誤
			_ = err // 忽略錯誤，繼續執行
		}
	}

	// 重新查詢頻道以取得更新後的節目
	channel, err := s.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, errors.New("channel not found")
	}

	// 找出更新後的節目
	for _, program := range channel.Contents {
		if program.ID == programID {
			return &program, nil
		}
	}

	return nil, errors.New("program not found")
}

// DeletePrograms 刪除節目
func (s *ProgramService) DeletePrograms(ctx context.Context, channelID string, programIDs []int) error {
	if len(programIDs) == 0 {
		return errors.New("program IDs are required")
	}
	return s.programRepo.DeletePrograms(ctx, channelID, programIDs)
}

// MoveProgram 移動節目到另一個頻道
func (s *ProgramService) MoveProgram(ctx context.Context, sourceChannelID, targetChannelID string, programIDs []int) error {
	if len(programIDs) == 0 {
		return errors.New("program IDs are required")
	}

	// 取得來源頻道
	sourceChannel, err := s.channelRepo.FindByID(ctx, sourceChannelID)
	if err != nil {
		return err
	}
	if sourceChannel == nil {
		return errors.New("source channel not found")
	}

	// 取得目標頻道
	targetChannel, err := s.channelRepo.FindByID(ctx, targetChannelID)
	if err != nil {
		return err
	}
	if targetChannel == nil {
		return errors.New("target channel not found")
	}

	// 找出要移動的節目
	var programsToMove []models.Program
	programIDMap := make(map[int]bool)
	for _, id := range programIDs {
		programIDMap[id] = true
	}

	for _, program := range sourceChannel.Contents {
		if programIDMap[program.ID] {
			programsToMove = append(programsToMove, program)
		}
		// 不需要移動的節目會保留在原頻道中（DeletePrograms 只刪除指定的節目）
	}

	// 從來源頻道刪除
	if err := s.programRepo.DeletePrograms(ctx, sourceChannelID, programIDs); err != nil {
		return err
	}

	// 新增到目標頻道
	for _, program := range programsToMove {
		if err := s.programRepo.AddProgram(ctx, targetChannelID, &program); err != nil {
			return err
		}
	}

	return nil
}

// SetOrder 設定節目順序
func (s *ProgramService) SetOrder(ctx context.Context, channelID string, order []int) error {
	return s.programRepo.SetOrder(ctx, channelID, order)
}

