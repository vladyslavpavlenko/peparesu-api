package models

import "time"

// User is the user model.
type User struct {
	ID         uint
	FirstName  string `gorm:"size:255;" json:"-"`
	LastName   string `gorm:"size:255;" json:"-"`
	Email      string `gorm:"size:255;unique;not null;" json:"-"`
	Password   string `gorm:"size:255" json:"-"`
	UserTypeID uint   `gorm:"not null"`
	UserType   UserType
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
