package migrations

import (
	"zatrano/configs/logconfig"
	"zatrano/models"

	"gorm.io/gorm"
)

func MigrateCardSocialMediaTable(db *gorm.DB) error {
	logconfig.SLog.Info("CardSocialMedia tablosu migrate ediliyor...")
	if err := db.AutoMigrate(&models.CardSocialMedia{}); err != nil {
		return err
	}
	logconfig.SLog.Info("CardSocialMedia tablosu migrate işlemi tamamlandı.")
	return nil
}
