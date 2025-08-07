package models

// CardBank represents the many-to-many relationship between Card and Bank with additional IBAN field
type CardBank struct {
	BaseModel
	CardID uint   `gorm:"index;not null"`
	BankID uint   `gorm:"index;not null"`
	IBAN   string `gorm:"size:50;not null"`
	Card   Card   `gorm:"foreignKey:CardID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Bank   Bank   `gorm:"foreignKey:BankID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// TableName returns the table name for the CardBank model
func (CardBank) TableName() string {
	return "card_banks"
}
