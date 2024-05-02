package models

// UserType is the user type model.
type UserType struct {
	ID    uint   `gorm:"primaryKey"`
	Title string `gorm:"size:255;not null"`
}
