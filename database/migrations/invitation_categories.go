package migrations

import (
	"zatrano/configs/logconfig"
	"zatrano/models"

	"gorm.io/gorm"
)

func MigrateInvitationCategoriesTable(db *gorm.DB) error {
	logconfig.SLog.Info("InvitationCategory tablosu migrate ediliyor...")
	if err := db.AutoMigrate(&models.InvitationCategory{}); err != nil {
		return err
	}
	logconfig.SLog.Info("InvitationCategory tablosu migrate işlemi tamamlandı.")
	return nil
}
