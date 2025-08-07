package models

type InvitationCategory struct {
	BaseModel

	// Zorunlu Alanlar
	IsActive bool   `gorm:"not null;index"`
	Template string `gorm:"type:varchar(255);not null"`
	Name     string `gorm:"type:varchar(255);not null;index"`
	Icon     string `gorm:"type:varchar(50);not null"`

	// İlişki Tanımı
	Invitations []Invitation `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (InvitationCategory) TableName() string {
	return "invitation_categories"
}
