package book

import (
	"fmt"

	"github.com/FAdemoglu/homeworkfourtwo/internal/domain/author"
	"gorm.io/gorm"
)

type Book struct {
	gorm.Model
	BookName   string
	AuthorID   int
	StockCode  int
	ISBNNumber int
	PageNumber int
	Price      int
	StockCount int
	Author     author.Author `gorm:"foreignKey:AuthorID;references:AuthorId"`
	IsDeleted  bool
}

func (Book) TableName() string {
	return "Book"
}

func (b *Book) ToString() string {
	return fmt.Sprintf("ID : %d, Name : %s, Code : %s, CountryCode : %s,CreatedAt : %s", b.ISBNNumber, b.BookName, b.StockCode, b.ISBNNumber, b.PageNumber)
}
