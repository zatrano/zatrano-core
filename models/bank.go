package models

type Bank struct {
	BaseModel
	IsActive bool   `gorm:"not null;index"`
	Name     string `gorm:"size:255;not null;index"`
}

func (Bank) TableName() string {
	return "banks"
}
