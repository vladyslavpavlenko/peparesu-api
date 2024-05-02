package models

import "time"

// User is the user model.
type User struct {
	ID          uint
	FirstName   string   `gorm:"size:255;"`
	LastName    string   `gorm:"size:255;"`
	Email       string   `gorm:"size:255;unique;not null;" json:"-"`
	Password    string   `gorm:"size:255" json:"-"`
	UserTypeID  uint     `gorm:"not null"`
	UserType    UserType `json:"-"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Restaurants []Restaurant `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;" json:"-"`
}
