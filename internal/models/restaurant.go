package models

// Restaurant is the restaurant model.
type Restaurant struct {
	ID          uint   `gorm:"primaryKey"`
	OwnerID     uint   `gorm:"not null;index"`
	Owner       User   `gorm:"foreignKey:OwnerID" json:"-"`
	Title       string `gorm:"size:255;not null"`
	Type        string `gorm:"size:255;not null"`
	Description string `gorm:"size:1000;"`
	Address     string `gorm:"size:255;"`
	Phone       string `gorm:"size:255;"`
	Menus       []Menu `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}
