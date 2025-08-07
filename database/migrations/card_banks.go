package migrations

import (
	"zatrano/configs/logconfig"
	"zatrano/models"

	"gorm.io/gorm"
)

func MigrateCardBanksTable(db *gorm.DB) error {
	logconfig.SLog.Info("CardBank tablosu migrate ediliyor...")
	if err := db.AutoMigrate(&models.CardBank{}); err != nil {
		return err
	}
	logconfig.SLog.Info("CardBank tablosu migrate işlemi tamamlandı.")
	return nil
}
