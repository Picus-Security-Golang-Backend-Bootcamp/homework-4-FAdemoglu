package author

import "gorm.io/gorm"

type AuthorRepository struct {
	db *gorm.DB
}

func NewAuthorRepository(db *gorm.DB) *AuthorRepository {
	return &AuthorRepository{
		db: db,
	}
}

func (r *AuthorRepository) Migration() error {
	return r.db.AutoMigrate(&Author{})
}

func (r *AuthorRepository) Insert(author Author) error {
	return r.db.Create(&author).Error
}

func (r *AuthorRepository) InsertCsvDatas(authors []Author) {
	for _, c := range authors {
		r.db.Where(Author{AuthorId: c.AuthorId}).Attrs(Author{AuthorId: c.AuthorId, Name: c.Name}).FirstOrCreate(&c)
	}
}
func (r *AuthorRepository) InsertSampleData() {
	cities := []Author{
		{AuthorId: 1, Name: "Machiavelli"},
	}

	for _, c := range cities {
		r.db.Create(&c)
	}
}
