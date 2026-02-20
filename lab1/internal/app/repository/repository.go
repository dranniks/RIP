package repository

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}
