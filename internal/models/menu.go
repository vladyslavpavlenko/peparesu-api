package models

// Menu is the menu model.
type Menu struct {
	ID           uint       `gorm:"primaryKey"`
	RestaurantID uint       `gorm:"not null;index"`
	Restaurant   Restaurant `gorm:"foreignKey:RestaurantID" json:"-"`
	Title        string     `gorm:"size:255;not null"`
	MenuItems    []MenuItem `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}
