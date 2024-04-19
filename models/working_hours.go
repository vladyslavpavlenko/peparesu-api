package models

// WorkingHours is the working hours type model.
type WorkingHours struct {
	ID           uint       `gorm:"primaryKey"`
	RestaurantID uint       `gorm:"not null;index"`
	Restaurant   Restaurant `gorm:"foreignKey:RestaurantID"`
	DayOfWeek    uint       `gorm:"not null"`
	IsClosed     bool       `gorm:"default:false"`
	OpenTime     string     `gorm:"type:time"`
	CloseTime    string     `gorm:"type:time"`
}
