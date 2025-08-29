package keypool

import (
	"context"
	"fmt"
	"gpt-load/internal/channel"
	"gpt-load/internal/config"
	"gpt-load/internal/models"
	"gpt-load/internal/types"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

// KeyTestResult holds the validation result for a single key.
type KeyTestResult struct {
	KeyValue string `json:"key_value"`
	IsValid  bool   `json:"is_valid"`
	Error    string `json:"error,omitempty"`
}

// KeyValidator provides methods to validate API keys.
type KeyValidator struct {
	DB              *gorm.DB
	channelFactory  *channel.Factory
	SettingsManager *config.SystemSettingsManager
	keypoolProvider *KeyProvider
}

type KeyValidatorParams struct {
	dig.In
	DB              *gorm.DB
	ChannelFactory  *channel.Factory
	SettingsManager *config.SystemSettingsManager
	KeypoolProvider *KeyProvider
}

// NewKeyValidator creates a new KeyValidator.
func NewKeyValidator(params KeyValidatorParams) *KeyValidator {
	return &KeyValidator{
		DB:              params.DB,
		channelFactory:  params.ChannelFactory,
		SettingsManager: params.SettingsManager,
		keypoolProvider: params.KeypoolProvider,
	}
}

// ValidateSingleKey performs a validation check on a single API key.
func (s *KeyValidator) ValidateSingleKey(key *models.APIKey, group *models.Group) (bool, error) {
	if group.EffectiveConfig.AppUrl == "" {
		group.EffectiveConfig = s.SettingsManager.GetEffectiveConfig(group.Config)
	}

	// 解析密钥（如果需要）
	if err := s.keypoolProvider.parseKey(key, group.ID); err != nil {
		logrus.WithError(err).Warnf("Failed to parse key for validation, using raw key")
		// 解析失败时使用原始密钥
		key.ParsedKey = &types.ParsedKey{
			RawKey:    key.KeyValue,
			ActualKey: key.KeyValue,
			Params:    make(map[string]string),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(group.EffectiveConfig.KeyValidationTimeoutSeconds)*time.Second)
	defer cancel()

	ch, err := s.channelFactory.GetChannel(group)
	if err != nil {
		return false, fmt.Errorf("failed to get channel for group %s: %w", group.Name, err)
	}

	isValid, validationErr := ch.ValidateKey(ctx, key, group)

	var errorMsg string
	if !isValid && validationErr != nil {
		errorMsg = validationErr.Error()
	}
	s.keypoolProvider.UpdateStatus(key, group, isValid, errorMsg)

	if !isValid {
		logrus.WithFields(logrus.Fields{
			"error":    validationErr,
			"key_id":   key.ID,
			"group_id": group.ID,
		}).Debug("Key validation failed")
		return false, validationErr
	}

	logrus.WithFields(logrus.Fields{
		"key_id":   key.ID,
		"is_valid": isValid,
	}).Debug("Key validation successful")

	return true, nil
}

// TestMultipleKeys performs a synchronous validation for a list of key values within a specific group.
func (s *KeyValidator) TestMultipleKeys(group *models.Group, keyValues []string) ([]KeyTestResult, error) {
	results := make([]KeyTestResult, len(keyValues))

	// Find which of the provided keys actually exist in the database for this group
	var existingKeys []models.APIKey
	if err := s.DB.Where("group_id = ? AND key_value IN ?", group.ID, keyValues).Find(&existingKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to query keys from DB: %w", err)
	}
	existingKeyMap := make(map[string]models.APIKey)
	for _, k := range existingKeys {
		existingKeyMap[k.KeyValue] = k
	}

	for i, kv := range keyValues {
		apiKey, exists := existingKeyMap[kv]
		if !exists {
			results[i] = KeyTestResult{
				KeyValue: kv,
				IsValid:  false,
				Error:    "Key does not exist in this group or has been removed.",
			}
			continue
		}

		isValid, validationErr := s.ValidateSingleKey(&apiKey, group)

		results[i] = KeyTestResult{
			KeyValue: kv,
			IsValid:  isValid,
			Error:    "",
		}
		if validationErr != nil {
			results[i].Error = validationErr.Error()
		}
	}

	return results, nil
}
