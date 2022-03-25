package book

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (r *BookRepository) FindAll() []Book {
	var books []Book
	r.db.Find(&books)
	return books
}

func (r *BookRepository) GetAllBooksWithAuthorInformation() ([]Book, error) {
	var books []Book
	result := r.db.Find(&books)

	if result.Error != nil {
		return nil, result.Error
	}

	return books, nil
}

func (r *BookRepository) DeleteById(id int) error {
	var exists bool
	result := r.db.Delete(&Book{}, id)

	if err := result.Scan(&exists); err != nil {
		fmt.Printf("Bu id ile bir kitap bulunamadı")
		return errors.New("Bu id ile bir kitap bulunamadı")
	} else if !exists {
		fmt.Printf("Bu id ile bir kitap bulunamadı")
		return errors.New("Bu id ile bir kitap bulunamadı")
	}
	if result.Error != nil {
		fmt.Printf("Bu id ile kayıtlı bir kitap bulunmamakta")
		return result.Error
	}
	fmt.Printf("Silme işlemi başarılı")
	return nil
}
func (r *BookRepository) Create(b Book) error {
	result := r.db.Create(b)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *BookRepository) Update(id int, count int) error {
	var book Book
	result := r.db.First(&book, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Printf("Bu id ile bir kitap bulunmuyor id : %d", id)
		return result.Error
	}
	if result.Error != nil {
		return result.Error
	}
	if count < 0 {
		return fmt.Errorf("Alınacak kitap sayısı 1'den küçük olamaz")
	}
	book.StockCount -= count
	resultSave := r.db.Save(book)
	if resultSave.Error != nil {
		return resultSave.Error
	}
	fmt.Printf("Kitap satma işlemi başarılı")
	return nil

}

func (r *BookRepository) SearchByAuthorAndBookName(searched string) []Book {
	var books []Book
	r.db.Preload("Author").Joins("JOIN author on author.AuthorId=book.AuthorID").Where("book.BookName LIKE ?", "%"+searched+"%").Or("author.Name LIKE ?", "%"+searched+"%").Find(&books)
	fmt.Println(books)
	return books
}

func (r *BookRepository) InsertCsvDatas(books []Book) {
	for _, c := range books {
		r.db.Where(Book{ISBNNumber: c.ISBNNumber}).Attrs(Book{BookName: c.BookName, AuthorID: c.AuthorID, StockCode: c.StockCode, ISBNNumber: c.ISBNNumber, PageNumber: c.PageNumber, Price: c.Price, StockCount: c.StockCount, IsDeleted: c.IsDeleted}).FirstOrCreate(&c)
	}
}

func (r *BookRepository) Migration() {
	r.db.AutoMigrate(&Book{})
}
func (r *BookRepository) InsertSampleData() {
	books := []Book{
		{BookName: "Savas Sanati", AuthorID: 1, StockCode: 123123123, ISBNNumber: 123123123, PageNumber: 237, Price: 20, StockCount: 15, IsDeleted: false},
	}

	for _, c := range books {
		r.db.Where(Book{ISBNNumber: c.ISBNNumber}).Attrs(Book{BookName: c.BookName, AuthorID: c.AuthorID, StockCode: c.StockCode, ISBNNumber: c.ISBNNumber, PageNumber: c.PageNumber, Price: c.Price, StockCount: c.StockCode, IsDeleted: c.IsDeleted}).FirstOrCreate(&c)
	}
	fmt.Printf("Başarılı bir şekilde veriler eklendi\n")
}
