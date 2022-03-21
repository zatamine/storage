package storage

type model interface {
	ID() string
	Status() int64
}

type Storage[T model] interface {
	FindOne(id string) (*T, error)
	FindAll() ([]T, error)
	Create(*T) error
	Update(T) error
	Delete(id string) error
}
