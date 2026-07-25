package book

type Repository interface {
	Create(b Book) (Book, error)
	List() ([]Book, error)
	Get(id int) (Book, error)
	Update(id int, b Book) (Book, error)
	Delete(id int) error
}
