package migrations

import (
	"zatrano/configs/logconfig"
	"zatrano/models"

	"gorm.io/gorm"
)

func MigrateBanksTable(db *gorm.DB) error {
	logconfig.SLog.Info("Bank tablosu migrate ediliyor...")
	if err := db.AutoMigrate(&models.Bank{}); err != nil {

		return err
	}
	logconfig.SLog.Info("Bank tablosu migrate işlemi tamamlandı.")
	return nil
}
