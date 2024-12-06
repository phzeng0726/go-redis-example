package domain

type User struct {
	ID    uint   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name  string `gorm:"column:name;size:100;not null" json:"name"`
	Email string `gorm:"column:email;size:100;unique;not null" json:"email"`
}

func (User) TableName() string {
	return "user"
}
