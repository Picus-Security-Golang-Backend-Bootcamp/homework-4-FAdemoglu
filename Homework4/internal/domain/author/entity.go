package author

import (
	"fmt"

	"gorm.io/gorm"
)

type Author struct {
	AuthorId int `gorm:"column:AuthorId"`
	Name     string
}

func (Author) TableName() string {
	return "Author"
}

func (b *Author) ToString() string {
	return fmt.Sprintf("ID : %d, Name : %s", b.AuthorId, b.Name)
}

func (b *Author) BeforeDelete(tx *gorm.DB) (err error) {
	fmt.Printf("Author (%s) deleting... \n", b.Name)
	return nil
}
