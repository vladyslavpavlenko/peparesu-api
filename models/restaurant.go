package models

// Restaurant is the restaurant model.
type Restaurant struct {
	ID          uint   `gorm:"primaryKey"`
	OwnerID     uint   `gorm:"not null;index"`
	Owner       User   `gorm:"foreignKey:OwnerID"`
	Title       string `gorm:"size:255;not null"`
	Type        string `gorm:"size:255;not null"`
	Description string `gorm:"size:255;"`
	Address     string `gorm:"size:255;"`
	Phone       string `gorm:"size:255;"`
}
