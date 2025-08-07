package models

type Card struct {
	BaseModel
	// Required fields
	IsActive bool   `gorm:"not null;index"`
	IsFree   bool   `gorm:"not null;index"`
	UserID   uint   `gorm:"uniqueIndex;not null"`
	Slug     string `gorm:"size:255;not null;uniqueIndex"`

	// Optional fields
	Name       string `gorm:"size:100"`
	Title      string `gorm:"size:255"`
	Photo      string `gorm:"size:255"`
	Telephone  string `gorm:"size:20"`
	Email      string `gorm:"size:100"`
	Location   string `gorm:"size:255"`
	WebsiteUrl string `gorm:"size:255"`
	StoreUrl   string `gorm:"size:255"`
	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	// Has many relationships with junction tables
	CardBanks       []CardBank        `gorm:"foreignKey:CardID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CardSocialMedia []CardSocialMedia `gorm:"foreignKey:CardID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// TableName returns the table name for the Card model
func (Card) TableName() string {
	return "cards"
}
