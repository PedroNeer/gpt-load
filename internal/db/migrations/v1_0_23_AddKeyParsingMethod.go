package db

import (
	"gorm.io/gorm"
)

type V1_0_23_AddKeyParsingMethod struct{}

func (m *V1_0_23_AddKeyParsingMethod) Up(tx *gorm.DB) error {
	// 添加字段，设置默认值为 'none'
	return tx.Exec("ALTER TABLE groups ADD COLUMN key_parsing_method VARCHAR(50) DEFAULT 'none' NOT NULL").Error
}

func (m *V1_0_23_AddKeyParsingMethod) Down(tx *gorm.DB) error {
	return tx.Exec("ALTER TABLE groups DROP COLUMN key_parsing_method").Error
}
