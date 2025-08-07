package models

import "time"

type Invitation struct {
	BaseModel

	InvitationKey string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Image         string    `gorm:"type:varchar(255);not null"`
	UserID        uint      `gorm:"index;not null"`
	CategoryID    uint      `gorm:"index;not null"`
	IsConfirmed   bool      `gorm:"not null;default:false;index"`
	IsParticipant bool      `gorm:"not null;default:false;index"`
	IsFree        bool      `gorm:"not null;index"`
	Description   string    `gorm:"type:text"`
	Venue         string    `gorm:"type:varchar(255)"`
	Address       string    `gorm:"type:varchar(255)"`
	Location      string    `gorm:"type:varchar(255)"`
	Link          string    `gorm:"type:varchar(255)"`
	Telephone     string    `gorm:"type:varchar(20)"`
	Note          string    `gorm:"type:text"`
	Date          time.Time `gorm:"index"`
	Time          string    `gorm:"type:varchar(10)"`

	User             *User                   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Category         *InvitationCategory     `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	InvitationDetail *InvitationDetail       `gorm:"foreignKey:InvitationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Participants     []InvitationParticipant `gorm:"foreignKey:InvitationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Invitation) TableName() string {
	return "invitations"
}
