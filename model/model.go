package model

type User struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement"`
	Email        string `gorm:"unique;not null;size:255"`
	Name         string `gorm:"not null;size:255"`
	PasswordHash string `gorm:"not null;size:255"`
}

type RegularExpense struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement"`
	UserID      uint64 `gorm:"index"`
	Name        string `gorm:"not null;size:50"`
	Description string
	NextDate    *string `gorm:"type:date;index"`
	Frequency   string  `gorm:"type:interval"`
	Amount      uint    `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}

type Expense struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement"`
	UserID           uint64 `gorm:"index"`
	RegularExpenseID uint64 `gorm:"index"`
	Date             string `gorm:"type:date;not null"`

	User           User           `gorm:"foreignKey:UserID"`
	RegularExpense RegularExpense `gorm:"foreignKey:RegularExpenseID"`
}
