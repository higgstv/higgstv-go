package service

import (
	"context"
	"errors"

	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/models"
	"github.com/higgstv/higgstv-go/pkg/uuidutil"
)

// ChannelService 頻道服務
type ChannelService struct {
	channelRepo database.ChannelRepository
	userRepo    database.UserRepository
}

// NewChannelService 建立頻道服務
func NewChannelService(channelRepo database.ChannelRepository, userRepo database.UserRepository) *ChannelService {
	return &ChannelService{
		channelRepo: channelRepo,
		userRepo:    userRepo,
	}
}

// CreateDefaultChannel 建立預設頻道
func (s *ChannelService) CreateDefaultChannel(ctx context.Context, userID, username string) (*models.Channel, error) {
	channel := &models.Channel{
		ID:            uuidutil.NewBase64UUID(),
		Type:          models.ChannelTypeDefault,
		Name:          username + "'s Channel",
		Desc:          "",
		Tags:          []int{},
		ContentsSeq:   "",
		Contents:      []models.Program{},
		ContentsOrder: []int{},
		Owners:        []string{userID},
		Permission:    []models.ChannelPermission{},
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, err
	}

	// 新增頻道到使用者的 own_channels
	if err := s.userRepo.AddChannel(ctx, username, channel.ID); err != nil {
		return nil, err
	}

	return channel, nil
}

// CreateUnclassifiedChannel 建立未分類頻道
func (s *ChannelService) CreateUnclassifiedChannel(ctx context.Context, userID, username string) (*models.Channel, error) {
	channel := &models.Channel{
		ID:            uuidutil.NewBase64UUID(),
		Type:          models.ChannelTypeUnclassified,
		Name:          "Unclassified",
		Desc:          "",
		Tags:          []int{},
		ContentsSeq:   "",
		Contents:      []models.Program{},
		ContentsOrder: []int{},
		Owners:        []string{userID},
		Permission:    []models.ChannelPermission{},
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, err
	}

	// 新增頻道到使用者的 own_channels
	if err := s.userRepo.AddChannel(ctx, username, channel.ID); err != nil {
		return nil, err
	}

	// 設定使用者的 unclassified_channel
	if err := s.userRepo.SetUnclassifiedChannel(ctx, username, channel.ID); err != nil {
		return nil, err
	}

	return channel, nil
}

// AddChannel 新增頻道
func (s *ChannelService) AddChannel(ctx context.Context, userID, username, name string, tags []int) (*models.Channel, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	channel := &models.Channel{
		ID:            uuidutil.NewBase64UUID(),
		Type:          models.ChannelTypeDefault,
		Name:          name,
		Desc:          "",
		Tags:          tags,
		ContentsSeq:   "",
		Contents:      []models.Program{},
		ContentsOrder: []int{},
		Owners:        []string{userID},
		Permission:    []models.ChannelPermission{},
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, err
	}

	// 新增頻道到使用者的 own_channels
	if err := s.userRepo.AddChannel(ctx, username, channel.ID); err != nil {
		return nil, err
	}

	return channel, nil
}

// GetChannel 取得頻道
func (s *ChannelService) GetChannel(ctx context.Context, channelID string) (*models.Channel, error) {
	return s.channelRepo.FindByID(ctx, channelID)
}

// UpdateChannel 更新頻道
func (s *ChannelService) UpdateChannel(ctx context.Context, channelID string, name string, desc string, tags []int) error {
	update := make(map[string]interface{})
	if name != "" {
		update["name"] = name
	}
	if desc != "" {
		update["desc"] = desc
	}
	if tags != nil {
		update["tags"] = tags
	}

	if len(update) == 0 {
		return errors.New("no fields to update")
	}

	return s.channelRepo.Update(ctx, channelID, update)
}

// ListChannels 列出頻道
func (s *ChannelService) ListChannels(ctx context.Context, filter database.Filter, sort database.Sort, limit, skip int64) ([]models.Channel, error) {
	return s.channelRepo.ListChannels(ctx, filter, sort, limit, skip)
}

// IsAdmin 檢查是否為頻道管理員
func (s *ChannelService) IsAdmin(ctx context.Context, channelID, userID string) (bool, error) {
	return s.channelRepo.IsAdmin(ctx, channelID, userID)
}

// AddOwners 新增擁有者
func (s *ChannelService) AddOwners(ctx context.Context, channelID string, userIDs []string) error {
	return s.channelRepo.AddOwners(ctx, channelID, userIDs)
}

