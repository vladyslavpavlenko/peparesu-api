package models

// MenuItem is the menu item model.
type MenuItem struct {
	ID          uint   `gorm:"primaryKey"`
	MenuID      uint   `gorm:"not null;index"`
	Menu        Menu   `gorm:"foreignKey:MenuID" json:"-"`
	Picture     string `gorm:"size:255;"`
	Title       string `gorm:"size:255;not null"`
	Description string `gorm:"size:1000"`
	LikesCount  uint
	PriceUAH    uint
}
