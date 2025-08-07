package migrations

import (
	"zatrano/configs/logconfig"
	"zatrano/models"

	"gorm.io/gorm"
)

func MigrateInvitationParticipantsTable(db *gorm.DB) error {
	logconfig.SLog.Info("InvitationParticipant tablosu migrate ediliyor...")
	if err := db.AutoMigrate(&models.InvitationParticipant{}); err != nil {
		return err
	}
	logconfig.SLog.Info("InvitationParticipant tablosu migrate işlemi tamamlandı.")
	return nil
}
