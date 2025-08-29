package db

import (
	"gorm.io/gorm"
)

func MigrateDatabase(db *gorm.DB) error {
	// 执行所有迁移
	if err := V1_0_22_DropRetriesColumn(db); err != nil {
		return err
	}

	migration := V1_0_23_AddKeyParsingMethod{}
	if err := migration.Up(db); err != nil {
		return err
	}

	return nil
}
