package model

type User struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string `gorm:"unique;not null;size:255" json:"email"`
	Name         string `gorm:"not null;size:255" json:"name"`
	PasswordHash string `gorm:"not null;size:255" json:"password_hash"`
}

type RegularExpense struct {
	ID          uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64  `gorm:"index" json:"user_id"`
	Name        string  `gorm:"not null;size:50" json:"name"`
	Description string  `json:"description,omitempty"`
	NextDate    *string `gorm:"type:date;index" json:"next_date"`
	Frequency   string  `gorm:"type:interval" json:"frequency"`
	Amount      uint    `gorm:"not null" json:"amount"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

type Expense struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint64 `gorm:"index" json:"user_id"`
	RegularExpenseID uint64 `gorm:"index" json:"regular_expense_id"`
	Date             string `gorm:"type:date;not null" json:"date"`

	User           User           `gorm:"foreignKey:UserID" json:"-"`
	RegularExpense RegularExpense `gorm:"foreignKey:RegularExpenseID" json:"-"`
}
