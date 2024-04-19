package models

// MenuItem is the menu item model.
type MenuItem struct {
	ID          uint   `gorm:"primaryKey"`
	MenuID      uint   `gorm:"not null;index"`
	Menu        Menu   `gorm:"foreignKey:MenuID"`
	Title       string `gorm:"size:255;not null"`
	Description string `gorm:"size:255"`
	LikesCount  uint
	PriceUAH    uint
}
